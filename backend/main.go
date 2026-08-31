package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/clubmax/clubmax/internal/config"
	"github.com/clubmax/clubmax/internal/handler"
	"github.com/clubmax/clubmax/internal/repository/postgres"
	"github.com/clubmax/clubmax/internal/service"
)

type logNotifier struct{}

func (logNotifier) Send(_ context.Context, phone, code string) error {
	log.Printf("development OTP phone=%s code=%s", phone, code)
	return nil
}
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}
	store, err := postgres.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal(err)
	}
	defer store.DB.Close()
	users := service.NewUserService(store, postgres.OTPStore{Store: store}, store, logNotifier{}, cfg.OTPPepper)
	gyms := service.NewGymService(postgres.GymStore{Store: store}, store)
	finance := service.NewFinanceService(store, store, postgres.GymStore{Store: store})
	messages := service.NewMessageService(postgres.MessageStore{Store: store}, store)
	admin := service.NewAdminService(store)
	backups, err := service.NewBackupService(service.BackupConfig{
		DatabaseHost: cfg.ServerIP, DatabasePort: cfg.DatabasePort, DatabaseName: cfg.DatabaseName,
		DatabaseUser: cfg.DatabaseUser, DatabasePassword: cfg.DatabasePassword,
		DatabaseContainer: cfg.DatabaseContainer, UploadDir: cfg.UploadDir, BackupDir: cfg.BackupDir,
		Hours: cfg.BackupHours, Retention: cfg.BackupRetention,
	}, store)
	if err != nil {
		log.Fatal(err)
	}
	bootstrapAdmin(ctx, users)
	go runWithdrawalSettlement(ctx, finance)
	go backups.Run(ctx, log.Printf)
	srv := &http.Server{Addr: cfg.BackendAddress(), Handler: handler.New(users, gyms, finance, messages, admin, backups, cfg.JWTSecret, cfg.FrontendOrigin(), cfg.UploadDir), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("ClubMax API listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func runWithdrawalSettlement(ctx context.Context, finance *service.FinanceService) {
	if count, err := finance.CompleteDueWithdrawals(ctx); err != nil {
		log.Printf("withdrawal settlement startup check: %v", err)
	} else if count > 0 {
		log.Printf("completed %d due withdrawals", count)
	}
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		location = time.FixedZone("Asia/Tehran", 3*60*60+30*60)
	}
	for {
		now := time.Now().In(location)
		next := time.Date(now.Year(), now.Month(), now.Day(), 23, 30, 0, 0, location)
		if !now.Before(next) {
			next = next.AddDate(0, 0, 1)
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if count, err := finance.CompleteDueWithdrawals(ctx); err != nil {
				log.Printf("withdrawal settlement: %v", err)
			} else if count > 0 {
				log.Printf("completed %d due withdrawals", count)
			}
		}
	}
}

func bootstrapAdmin(ctx context.Context, s *service.UserService) {
	name, user, pass := os.Getenv("CLUBMAX_ADMIN_NAME"), os.Getenv("CLUBMAX_ADMIN_USERNAME"), os.Getenv("CLUBMAX_ADMIN_PASSWORD")
	if name == "" || user == "" || pass == "" {
		log.Print("admin bootstrap skipped: administrator settings are missing from clubmax.env")
		return
	}
	if _, e := s.EnsureAdmin(ctx, name, user, pass); e != nil {
		log.Printf("admin bootstrap: %v", e)
	}
}
