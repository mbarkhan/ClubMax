package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	ServerIP          string
	BindIP            string
	PublicHost        string
	FrontendURL       string
	FrontendPort      string
	BackendPort       string
	DatabasePort      string
	DatabaseName      string
	DatabaseUser      string
	DatabasePassword  string
	OTPPepper         string
	JWTSecret         string
	UploadDir         string
	BackupDir         string
	DatabaseContainer string
	BackupHours       []int
	BackupRetention   int
	ZibalMerchant     string
	ZibalGatewayURL   string
}

func Load() (Config, error) {
	if err := loadFile(); err != nil {
		return Config{}, err
	}
	c := Config{
		ServerIP: required("CLUBMAX_SERVER_IP"), FrontendPort: required("CLUBMAX_FRONTEND_PORT"),
		BindIP: required("CLUBMAX_BIND_IP"), PublicHost: required("CLUBMAX_PUBLIC_HOST"),
		FrontendURL: required("CLUBMAX_FRONTEND_ORIGIN"),
		BackendPort: required("CLUBMAX_BACKEND_PORT"), DatabasePort: required("CLUBMAX_DATABASE_PORT"),
		DatabaseName: required("CLUBMAX_DATABASE_NAME"), DatabaseUser: required("CLUBMAX_DATABASE_USER"),
		DatabasePassword: required("CLUBMAX_DATABASE_PASSWORD"), OTPPepper: required("CLUBMAX_OTP_PEPPER"),
		JWTSecret:         required("CLUBMAX_JWT_SECRET"),
		UploadDir:         required("CLUBMAX_UPLOAD_DIR"),
		BackupDir:         required("CLUBMAX_BACKUP_DIR"),
		DatabaseContainer: required("CLUBMAX_DATABASE_CONTAINER"),
		ZibalMerchant:     required("CLUBMAX_ZIBAL_MERCHANT"),
		ZibalGatewayURL:   required("CLUBMAX_ZIBAL_GATEWAY_URL"),
	}
	if c.ServerIP == "" || c.BindIP == "" || c.PublicHost == "" || c.FrontendPort == "" || c.BackendPort == "" || c.DatabasePort == "" ||
		c.DatabaseName == "" || c.DatabaseUser == "" || c.DatabasePassword == "" || c.OTPPepper == "" || c.JWTSecret == "" || c.UploadDir == "" || c.BackupDir == "" || c.ZibalMerchant == "" || c.ZibalGatewayURL == "" {
		return Config{}, fmt.Errorf("clubmax.env contains an empty required value")
	}
	if len(c.ZibalMerchant) > 200 {
		return Config{}, fmt.Errorf("CLUBMAX_ZIBAL_MERCHANT is invalid")
	}
	if !validZibalGatewayURL(c.ZibalGatewayURL) {
		return Config{}, fmt.Errorf("CLUBMAX_ZIBAL_GATEWAY_URL must be an official Zibal HTTPS endpoint")
	}
	var err error
	if c.BackupHours, err = parseHours(required("CLUBMAX_BACKUP_HOURS")); err != nil {
		return Config{}, err
	}
	if c.BackupRetention, err = strconv.Atoi(required("CLUBMAX_BACKUP_RETENTION")); err != nil || c.BackupRetention < 1 {
		return Config{}, fmt.Errorf("CLUBMAX_BACKUP_RETENTION must be a positive integer")
	}
	return c, nil
}

func validZibalGatewayURL(raw string) bool {
	gatewayURL, err := url.Parse(raw)
	return err == nil && gatewayURL.Scheme == "https" && gatewayURL.Host == "gateway.zibal.ir" && gatewayURL.User == nil && gatewayURL.RawQuery == "" && gatewayURL.Fragment == "" && (gatewayURL.Path == "" || gatewayURL.Path == "/")
}

func parseHours(value string) ([]int, error) {
	seen := make(map[int]bool)
	var hours []int
	for _, item := range strings.Split(value, ",") {
		hour, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil || hour < 0 || hour > 23 {
			return nil, fmt.Errorf("CLUBMAX_BACKUP_HOURS must contain comma-separated hours between 0 and 23")
		}
		if !seen[hour] {
			seen[hour] = true
			hours = append(hours, hour)
		}
	}
	if len(hours) == 0 {
		return nil, fmt.Errorf("CLUBMAX_BACKUP_HOURS is required")
	}
	sort.Ints(hours)
	return hours, nil
}

func (c Config) BackendAddress() string { return c.BindIP + ":" + c.BackendPort }
func (c Config) FrontendOrigin() string {
	if c.FrontendURL != "" {
		return strings.TrimRight(c.FrontendURL, "/")
	}
	return "http://" + c.PublicHost + ":" + c.FrontendPort
}
func (c Config) PaymentCallbackURL() string {
	return c.FrontendOrigin() + "/v1/payments/zibal/callback"
}
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
	if os.Getenv("CLUBMAX_CONFIG_FILE") == "" && os.Getenv("CLUBMAX_SERVER_IP") != "" {
		return nil
	}
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
