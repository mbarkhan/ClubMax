package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ServerIP         string
	BindIP           string
	PublicHost       string
	FrontendPort     string
	BackendPort      string
	DatabasePort     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	OTPPepper        string
	JWTSecret        string
	UploadDir        string
}

func Load() (Config, error) {
	if err := loadFile(); err != nil {
		return Config{}, err
	}
	c := Config{
		ServerIP: required("CLUBMAX_SERVER_IP"), FrontendPort: required("CLUBMAX_FRONTEND_PORT"),
		BindIP: required("CLUBMAX_BIND_IP"), PublicHost: required("CLUBMAX_PUBLIC_HOST"),
		BackendPort: required("CLUBMAX_BACKEND_PORT"), DatabasePort: required("CLUBMAX_DATABASE_PORT"),
		DatabaseName: required("CLUBMAX_DATABASE_NAME"), DatabaseUser: required("CLUBMAX_DATABASE_USER"),
		DatabasePassword: required("CLUBMAX_DATABASE_PASSWORD"), OTPPepper: required("CLUBMAX_OTP_PEPPER"),
		JWTSecret: required("CLUBMAX_JWT_SECRET"),
		UploadDir: required("CLUBMAX_UPLOAD_DIR"),
	}
	if c.ServerIP == "" || c.BindIP == "" || c.PublicHost == "" || c.FrontendPort == "" || c.BackendPort == "" || c.DatabasePort == "" ||
		c.DatabaseName == "" || c.DatabaseUser == "" || c.DatabasePassword == "" || c.OTPPepper == "" || c.JWTSecret == "" || c.UploadDir == "" {
		return Config{}, fmt.Errorf("clubmax.env contains an empty required value")
	}
	return c, nil
}

func (c Config) BackendAddress() string { return c.BindIP + ":" + c.BackendPort }
func (c Config) FrontendOrigin() string { return "http://" + c.PublicHost + ":" + c.FrontendPort }
func (c Config) DatabaseURL() string {
	connection := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DatabaseUser, c.DatabasePassword),
		Host:   c.ServerIP + ":" + c.DatabasePort,
		Path:   c.DatabaseName,
	}
	connection.RawQuery = "sslmode=disable"
	return connection.String()
}

func loadFile() error {
	paths := []string{"../clubmax.env", "clubmax.env"}
	if explicit := os.Getenv("CLUBMAX_CONFIG_FILE"); explicit != "" {
		paths = append([]string{explicit}, paths...)
	}
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		file, err := os.Open(absolute)
		if err != nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, found := strings.Cut(line, "=")
			if !found || strings.TrimSpace(key) == "" {
				return fmt.Errorf("invalid configuration line in %s: %q", absolute, line)
			}
			if err := os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"'`)); err != nil {
				return err
			}
		}
		return scanner.Err()
	}
	return fmt.Errorf("clubmax.env was not found")
}

func required(key string) string { return strings.TrimSpace(os.Getenv(key)) }
