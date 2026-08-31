package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
)

type BackupConfig struct {
	DatabaseHost      string
	DatabasePort      string
	DatabaseName      string
	DatabaseUser      string
	DatabasePassword  string
	DatabaseContainer string
	UploadDir         string
	BackupDir         string
	Hours             []int
	Retention         int
}

type BackupFile struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type BackupStatus struct {
	Timezone        string       `json:"timezone"`
	Schedule        []string     `json:"schedule"`
	ScheduleMinutes []int32      `json:"schedule_minutes"`
	Retention       int          `json:"retention"`
	NextRun         time.Time    `json:"next_run"`
	Files           []BackupFile `json:"files"`
}

type FullBackup struct {
	Path string
	Name string
	Size int64
}

type RestoreResult struct {
	BackupCreatedAt time.Time `json:"backup_created_at"`
	RestoredAt      time.Time `json:"restored_at"`
	UploadFiles     int       `json:"upload_files"`
	UploadsRestored bool      `json:"uploads_restored"`
}

type backupManifest struct {
	FormatVersion   int       `json:"format_version"`
	CreatedAt       time.Time `json:"created_at"`
	Timezone        string    `json:"timezone"`
	Database        string    `json:"database"`
	DatabaseSHA256  string    `json:"database_sha256"`
	UploadFiles     int       `json:"upload_files"`
	IncludesUploads bool      `json:"includes_uploads"`
}

type BackupPolicyStore interface {
	GetBackupPolicy(context.Context) (domain.BackupPolicy, error)
	SetBackupPolicy(context.Context, domain.BackupPolicy) error
}

type BackupService struct {
	config          BackupConfig
	policies        BackupPolicyStore
	location        *time.Location
	mu              sync.Mutex
	scheduleChanged chan struct{}
}

func NewBackupService(config BackupConfig, policies BackupPolicyStore) (*BackupService, error) {
	if config.Retention < 1 || len(config.Hours) == 0 {
		return nil, fmt.Errorf("invalid backup schedule")
	}
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		location = time.FixedZone("Asia/Tehran", 3*60*60+30*60)
	}
	config.BackupDir, err = filepath.Abs(config.BackupDir)
	if err != nil {
		return nil, err
	}
	config.UploadDir, err = filepath.Abs(config.UploadDir)
	if err != nil {
		return nil, err
	}
	if filepath.Dir(config.UploadDir) == config.UploadDir {
		return nil, fmt.Errorf("upload directory cannot be a filesystem root")
	}
	if err = os.MkdirAll(config.BackupDir, 0o700); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}
	if err = os.Chmod(config.BackupDir, 0o700); err != nil {
		return nil, fmt.Errorf("secure backup directory: %w", err)
	}
	return &BackupService{config: config, policies: policies, location: location, scheduleChanged: make(chan struct{}, 1)}, nil
}

func (s *BackupService) Run(ctx context.Context, logf func(string, ...any)) {
	for {
		policy, err := s.policy(ctx)
		if err != nil {
			logf("load backup policy failed: %v", err)
			policy = s.defaultPolicy()
		}
		next := s.nextRun(time.Now(), policy.ScheduleMinutes)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.scheduleChanged:
			timer.Stop()
			continue
		case <-timer.C:
			backup, err := s.CreateDatabaseBackup(ctx)
			if err != nil {
				logf("automatic database backup failed: %v", err)
			} else {
				logf("automatic database backup created: %s (%d bytes)", backup.Name, backup.Size)
			}
		}
	}
}

func (s *BackupService) CreateDatabaseBackup(ctx context.Context) (BackupFile, error) {
	if !s.mu.TryLock() {
		return BackupFile{}, domain.ErrBackupInProgress
	}
	defer s.mu.Unlock()

	createdAt := time.Now().In(s.location)
	name := "clubmax-db-auto-" + createdAt.Format("20060102-150405") + ".dump"
	path := filepath.Join(s.config.BackupDir, name)
	if err := s.dumpDatabase(ctx, path); err != nil {
		return BackupFile{}, err
	}
	policy, err := s.policy(ctx)
	if err != nil {
		return BackupFile{}, err
	}
	if err := s.pruneAutomatic(policy.Retention); err != nil {
		return BackupFile{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return BackupFile{}, err
	}
	return BackupFile{Name: name, Size: info.Size(), CreatedAt: info.ModTime()}, nil
}

func (s *BackupService) CreateFullBackup(ctx context.Context, includeUploads bool) (FullBackup, error) {
	if !s.mu.TryLock() {
		return FullBackup{}, domain.ErrBackupInProgress
	}
	defer s.mu.Unlock()

	createdAt := time.Now().In(s.location)
	stamp := createdAt.Format("20060102-150405")
	databasePath := filepath.Join(s.config.BackupDir, ".clubmax-manual-"+stamp+".dump")
	defer os.Remove(databasePath)
	if err := s.dumpDatabase(ctx, databasePath); err != nil {
		return FullBackup{}, err
	}

	prefix := "clubmax-database-"
	if includeUploads {
		prefix = "clubmax-full-"
	}
	name := prefix + stamp + ".tar.gz"
	path := filepath.Join(s.config.BackupDir, name)
	partial := path + ".partial"
	defer os.Remove(partial)
	file, err := os.OpenFile(partial, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return FullBackup{}, err
	}
	gzipWriter, err := gzip.NewWriterLevel(file, gzip.BestSpeed)
	if err != nil {
		file.Close()
		return FullBackup{}, err
	}
	tarWriter := tar.NewWriter(gzipWriter)
	databaseHash, err := addFileToArchive(tarWriter, databasePath, "database/clubmax.dump", createdAt)
	uploadCount := 0
	if err == nil && includeUploads {
		uploadCount, err = s.addUploads(tarWriter)
	}
	if err == nil {
		manifest := backupManifest{
			FormatVersion: 2, CreatedAt: createdAt, Timezone: "Asia/Tehran",
			Database: s.config.DatabaseName, DatabaseSHA256: databaseHash, UploadFiles: uploadCount,
			IncludesUploads: includeUploads,
		}
		var data []byte
		data, err = json.MarshalIndent(manifest, "", "  ")
		if err == nil {
			err = addBytesToArchive(tarWriter, data, "manifest.json", createdAt)
		}
	}
	if closeErr := tarWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := gzipWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Sync(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return FullBackup{}, err
	}
	if err = os.Rename(partial, path); err != nil {
		return FullBackup{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return FullBackup{}, err
	}
	return FullBackup{Path: path, Name: name, Size: info.Size()}, nil
}

func (s *BackupService) RestoreFullBackup(ctx context.Context, source io.Reader) (RestoreResult, error) {
	if !s.mu.TryLock() {
		return RestoreResult{}, domain.ErrBackupInProgress
	}
	defer s.mu.Unlock()

	parent := filepath.Dir(s.config.UploadDir)
	if err := os.MkdirAll(parent, 0o750); err != nil {
		return RestoreResult{}, err
	}
	stage, err := os.MkdirTemp(parent, ".clubmax-restore-")
	if err != nil {
		return RestoreResult{}, err
	}
	defer os.RemoveAll(stage)
	if err = os.Chmod(stage, 0o700); err != nil {
		return RestoreResult{}, err
	}

	databasePath := filepath.Join(stage, "clubmax.dump")
	uploadsPath := filepath.Join(stage, "uploads")
	if err = os.Mkdir(uploadsPath, 0o750); err != nil {
		return RestoreResult{}, err
	}
	manifest, databaseHash, uploadCount, err := extractFullBackup(source, databasePath, uploadsPath)
	compatibleVersion := manifest.FormatVersion == 1 || manifest.FormatVersion == 2
	includesUploads := manifest.FormatVersion == 1 || manifest.IncludesUploads
	if err != nil || !compatibleVersion || manifest.Database != s.config.DatabaseName ||
		manifest.DatabaseSHA256 != databaseHash || manifest.UploadFiles != uploadCount ||
		(!includesUploads && uploadCount != 0) || manifest.CreatedAt.IsZero() {
		return RestoreResult{}, domain.ErrInvalidBackup
	}
	if err = validateDump(databasePath); err != nil {
		return RestoreResult{}, domain.ErrInvalidBackup
	}
	if err = s.restoreDatabase(ctx, databasePath); err != nil {
		return RestoreResult{}, fmt.Errorf("restore database: %w", err)
	}
	if includesUploads {
		if err = s.replaceUploads(uploadsPath); err != nil {
			return RestoreResult{}, fmt.Errorf("restore uploads: %w", err)
		}
	}
	s.notifyScheduleChanged()
	return RestoreResult{BackupCreatedAt: manifest.CreatedAt, RestoredAt: time.Now().In(s.location), UploadFiles: uploadCount, UploadsRestored: includesUploads}, nil
}

func (s *BackupService) Policy(ctx context.Context) (domain.BackupPolicy, error) {
	return s.policy(ctx)
}

func (s *BackupService) UpdatePolicy(ctx context.Context, policy domain.BackupPolicy) (domain.BackupPolicy, error) {
	if !validSchedule(policy.ScheduleMinutes) {
		return domain.BackupPolicy{}, domain.ErrInvalid
	}
	policy.ScheduleMinutes = normalizeSchedule(policy.ScheduleMinutes)
	if len(policy.ScheduleMinutes) == 0 || len(policy.ScheduleMinutes) > 24 || policy.Retention < 1 || policy.Retention > 1000 {
		return domain.BackupPolicy{}, domain.ErrInvalid
	}
	if err := s.policies.SetBackupPolicy(ctx, policy); err != nil {
		return domain.BackupPolicy{}, err
	}
	s.notifyScheduleChanged()
	s.mu.Lock()
	err := s.pruneAutomatic(policy.Retention)
	s.mu.Unlock()
	if err != nil {
		return domain.BackupPolicy{}, err
	}
	return s.policy(ctx)
}

func (s *BackupService) policy(ctx context.Context) (domain.BackupPolicy, error) {
	policy, err := s.policies.GetBackupPolicy(ctx)
	if err != nil {
		return domain.BackupPolicy{}, err
	}
	if !validSchedule(policy.ScheduleMinutes) {
		return domain.BackupPolicy{}, domain.ErrInvalid
	}
	policy.ScheduleMinutes = normalizeSchedule(policy.ScheduleMinutes)
	if len(policy.ScheduleMinutes) == 0 || policy.Retention < 1 || policy.Retention > 1000 {
		return domain.BackupPolicy{}, domain.ErrInvalid
	}
	return policy, nil
}

func validSchedule(schedule []int32) bool {
	if len(schedule) == 0 || len(schedule) > 24 {
		return false
	}
	for _, minute := range schedule {
		if minute < 0 || minute >= 24*60 {
			return false
		}
	}
	return true
}

func (s *BackupService) defaultPolicy() domain.BackupPolicy {
	minutes := make([]int32, len(s.config.Hours))
	for i, hour := range s.config.Hours {
		minutes[i] = int32(hour * 60)
	}
	return domain.BackupPolicy{ScheduleMinutes: minutes, Retention: s.config.Retention}
}

func normalizeSchedule(schedule []int32) []int32 {
	seen := make(map[int32]bool, len(schedule))
	result := make([]int32, 0, len(schedule))
	for _, minute := range schedule {
		if minute < 0 || minute >= 24*60 || seen[minute] {
			continue
		}
		seen[minute] = true
		result = append(result, minute)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (s *BackupService) notifyScheduleChanged() {
	select {
	case s.scheduleChanged <- struct{}{}:
	default:
	}
}

func extractFullBackup(source io.Reader, databasePath, uploadsPath string) (backupManifest, string, int, error) {
	const maxExpandedSize int64 = 20 << 30
	const maxUploadFiles = 100_000
	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		return backupManifest{}, "", 0, err
	}
	defer gzipReader.Close()
	reader := tar.NewReader(gzipReader)
	var manifest backupManifest
	var databaseHash string
	var databaseFound, manifestFound bool
	var uploadCount int
	var expanded int64
	for {
		header, nextErr := reader.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil || header.Size < 0 || header.Typeflag != tar.TypeReg {
			return backupManifest{}, "", 0, domain.ErrInvalidBackup
		}
		expanded += header.Size
		if expanded > maxExpandedSize {
			return backupManifest{}, "", 0, domain.ErrInvalidBackup
		}
		name := pathpkg.Clean(strings.TrimPrefix(header.Name, "./"))
		switch {
		case name == "database/clubmax.dump" && !databaseFound:
			databaseHash, err = extractArchiveFile(reader, databasePath, header.Size, true)
			databaseFound = err == nil
		case name == "manifest.json" && !manifestFound && header.Size <= 1<<20:
			data := make([]byte, header.Size)
			_, err = io.ReadFull(reader, data)
			if err == nil {
				err = json.Unmarshal(data, &manifest)
			}
			manifestFound = err == nil
		case strings.HasPrefix(name, "uploads/"):
			if uploadCount >= maxUploadFiles {
				return backupManifest{}, "", 0, domain.ErrInvalidBackup
			}
			relative := strings.TrimPrefix(name, "uploads/")
			if relative == "" || relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
				return backupManifest{}, "", 0, domain.ErrInvalidBackup
			}
			destination := filepath.Join(uploadsPath, filepath.FromSlash(relative))
			if !pathWithin(uploadsPath, destination) {
				return backupManifest{}, "", 0, domain.ErrInvalidBackup
			}
			if err = os.MkdirAll(filepath.Dir(destination), 0o750); err == nil {
				_, err = extractArchiveFile(reader, destination, header.Size, false)
			}
			if err == nil {
				uploadCount++
			}
		default:
			return backupManifest{}, "", 0, domain.ErrInvalidBackup
		}
		if err != nil {
			return backupManifest{}, "", 0, err
		}
	}
	if !databaseFound || !manifestFound {
		return backupManifest{}, "", 0, domain.ErrInvalidBackup
	}
	return manifest, databaseHash, uploadCount, nil
}

func extractArchiveFile(reader io.Reader, destination string, size int64, hashFile bool) (string, error) {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	target := io.Writer(file)
	if hashFile {
		target = io.MultiWriter(file, hash)
	}
	written, copyErr := io.CopyN(target, reader, size)
	syncErr := file.Sync()
	closeErr := file.Close()
	if copyErr != nil || written != size {
		return "", domain.ErrInvalidBackup
	}
	if syncErr != nil {
		return "", syncErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func pathWithin(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func (s *BackupService) Status(ctx context.Context) (BackupStatus, error) {
	files, err := filepath.Glob(filepath.Join(s.config.BackupDir, "clubmax-db-auto-*.dump"))
	if err != nil {
		return BackupStatus{}, err
	}
	backups := make([]BackupFile, 0, len(files))
	for _, path := range files {
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}
		backups = append(backups, BackupFile{Name: info.Name(), Size: info.Size(), CreatedAt: info.ModTime()})
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].CreatedAt.After(backups[j].CreatedAt) })
	if len(backups) > 10 {
		backups = backups[:10]
	}
	policy, err := s.policy(ctx)
	if err != nil {
		return BackupStatus{}, err
	}
	schedule := make([]string, len(policy.ScheduleMinutes))
	for i, minute := range policy.ScheduleMinutes {
		schedule[i] = fmt.Sprintf("%02d:%02d", minute/60, minute%60)
	}
	return BackupStatus{
		Timezone: "Asia/Tehran", Schedule: schedule, ScheduleMinutes: policy.ScheduleMinutes,
		Retention: policy.Retention, NextRun: s.nextRun(time.Now(), policy.ScheduleMinutes), Files: backups,
	}, nil
}

func (s *BackupService) nextRun(now time.Time, schedule []int32) time.Time {
	localNow := now.In(s.location)
	for _, minute := range schedule {
		candidate := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), int(minute)/60, int(minute)%60, 0, 0, s.location)
		if candidate.After(localNow) {
			return candidate
		}
	}
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, int(schedule[0])/60, int(schedule[0])%60, 0, 0, s.location)
}

func (s *BackupService) dumpDatabase(ctx context.Context, destination string) error {
	hostArgs := []string{
		"--host", s.config.DatabaseHost,
		"--port", s.config.DatabasePort,
		"--username", s.config.DatabaseUser,
		"--format=custom", "--no-owner", "--no-privileges", s.config.DatabaseName,
	}
	hostCommand := exec.CommandContext(ctx, "pg_dump", hostArgs...)
	hostCommand.Env = append(os.Environ(), "PGPASSWORD="+s.config.DatabasePassword)
	if err := writeCommandOutput(destination, hostCommand); err == nil {
		return nil
	} else if s.config.DatabaseContainer == "" {
		return err
	}

	dockerArgs := []string{
		"exec", s.config.DatabaseContainer, "pg_dump",
		"--username", s.config.DatabaseUser,
		"--format=custom", "--no-owner", "--no-privileges", s.config.DatabaseName,
	}
	if err := writeCommandOutput(destination, exec.CommandContext(ctx, "docker", dockerArgs...)); err != nil {
		return fmt.Errorf("database dump failed: %w", err)
	}
	return nil
}

func (s *BackupService) restoreDatabase(ctx context.Context, source string) error {
	common := []string{"--clean", "--if-exists", "--no-owner", "--no-privileges", "--single-transaction", "--exit-on-error"}
	hostArgs := append([]string{
		"--host", s.config.DatabaseHost,
		"--port", s.config.DatabasePort,
		"--username", s.config.DatabaseUser,
		"--dbname", s.config.DatabaseName,
	}, common...)
	hostArgs = append(hostArgs, source)
	hostCommand := exec.CommandContext(ctx, "pg_restore", hostArgs...)
	hostCommand.Env = append(os.Environ(), "PGPASSWORD="+s.config.DatabasePassword)
	var hostStderr bytes.Buffer
	hostCommand.Stderr = &hostStderr
	if err := hostCommand.Run(); err == nil {
		return nil
	} else if s.config.DatabaseContainer == "" {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(hostStderr.String()))
	}

	dockerArgs := append([]string{
		"exec", "-i", s.config.DatabaseContainer, "pg_restore",
		"--username", s.config.DatabaseUser,
		"--dbname", s.config.DatabaseName,
	}, common...)
	if err := runCommandInput(source, exec.CommandContext(ctx, "docker", dockerArgs...)); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}
	return nil
}

func runCommandInput(source string, command *exec.Cmd) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdin, err := command.StdinPipe()
	if err != nil {
		return err
	}
	if err = command.Start(); err != nil {
		return err
	}
	_, copyErr := io.Copy(stdin, file)
	closeErr := stdin.Close()
	runErr := command.Wait()
	if runErr != nil {
		return fmt.Errorf("%w: %s", runErr, strings.TrimSpace(stderr.String()))
	}
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (s *BackupService) replaceUploads(staged string) error {
	old := s.config.UploadDir + ".restore-old-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	hadCurrent := false
	if _, err := os.Stat(s.config.UploadDir); err == nil {
		hadCurrent = true
		if err = os.Rename(s.config.UploadDir, old); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(staged, s.config.UploadDir); err != nil {
		if hadCurrent {
			_ = os.Rename(old, s.config.UploadDir)
		}
		return err
	}
	if hadCurrent {
		_ = os.RemoveAll(old)
	}
	return nil
}

func writeCommandOutput(destination string, command *exec.Cmd) error {
	partial := destination + ".partial"
	_ = os.Remove(partial)
	file, err := os.OpenFile(partial, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		file.Close()
		_ = os.Remove(partial)
		return err
	}
	if err = command.Start(); err != nil {
		file.Close()
		_ = os.Remove(partial)
		return err
	}
	_, copyErr := io.Copy(file, stdout)
	runErr := command.Wait()
	syncErr := file.Sync()
	closeErr := file.Close()
	if runErr != nil {
		_ = os.Remove(partial)
		return fmt.Errorf("%w: %s", runErr, strings.TrimSpace(stderr.String()))
	}
	if copyErr != nil {
		_ = os.Remove(partial)
		return copyErr
	}
	if syncErr != nil {
		_ = os.Remove(partial)
		return syncErr
	}
	if closeErr != nil {
		_ = os.Remove(partial)
		return closeErr
	}
	if err = validateDump(partial); err != nil {
		_ = os.Remove(partial)
		return err
	}
	if err = os.Rename(partial, destination); err != nil {
		_ = os.Remove(partial)
		return err
	}
	return nil
}

func validateDump(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	magic := make([]byte, 5)
	if _, err = io.ReadFull(file, magic); err != nil {
		return fmt.Errorf("invalid database dump: %w", err)
	}
	if string(magic) != "PGDMP" {
		return errors.New("invalid database dump header")
	}
	return nil
}

func (s *BackupService) pruneAutomatic(retention int) error {
	files, err := filepath.Glob(filepath.Join(s.config.BackupDir, "clubmax-db-auto-*.dump"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	if len(files) <= retention {
		return nil
	}
	for _, path := range files[:len(files)-retention] {
		if err = os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func (s *BackupService) addUploads(writer *tar.Writer) (int, error) {
	count := 0
	err := filepath.WalkDir(s.config.UploadDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(s.config.UploadDir, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("invalid upload path")
		}
		if _, err = addFileToArchive(writer, path, filepath.ToSlash(filepath.Join("uploads", relative)), time.Time{}); err != nil {
			return err
		}
		count++
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return count, err
}

func addFileToArchive(writer *tar.Writer, source, archiveName string, fallbackTime time.Time) (string, error) {
	file, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	modified := info.ModTime()
	if modified.IsZero() {
		modified = fallbackTime
	}
	header := &tar.Header{Name: archiveName, Mode: 0o600, Size: info.Size(), ModTime: modified}
	if err = writer.WriteHeader(header); err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err = io.Copy(io.MultiWriter(writer, hash), file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func addBytesToArchive(writer *tar.Writer, data []byte, archiveName string, modified time.Time) error {
	header := &tar.Header{Name: archiveName, Mode: 0o600, Size: int64(len(data)), ModTime: modified}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	_, err := writer.Write(data)
	return err
}
