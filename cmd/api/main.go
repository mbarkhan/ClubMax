package main

import (
	"context"
	"github.com/gymax/gymax/internal/handler"
	"github.com/gymax/gymax/internal/repository/postgres"
	"github.com/gymax/gymax/internal/service"
	"log"
	"net/http"
	"os"
	"time"
)

type logNotifier struct{}

func (logNotifier) Send(_ context.Context, phone, code string) error {
	log.Printf("development OTP phone=%s code=%s", phone, code)
	return nil
}
func main() {
	ctx := context.Background()
	url := mustEnv("DATABASE_URL")
	pepper := mustEnv("OTP_PEPPER")
	jwtSecret := mustEnv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}
	store, e := postgres.New(ctx, url)
	if e != nil {
		log.Fatal(e)
	}
	defer store.DB.Close()
	users := service.NewUserService(store, postgres.OTPStore{Store: store}, logNotifier{}, pepper)
	gyms := service.NewGymService(postgres.GymStore{Store: store}, store)
	bootstrapAdmin(ctx, users)
	srv := &http.Server{Addr: env("HTTP_ADDR", "127.0.0.1:8080"), Handler: handler.New(users, gyms, jwtSecret), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("GyMax API listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
func bootstrapAdmin(ctx context.Context, s *service.UserService) {
	name, user, pass := os.Getenv("ADMIN_NAME"), os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD")
	if name == "" || user == "" || pass == "" {
		log.Print("admin bootstrap skipped: ADMIN_NAME/ADMIN_USERNAME/ADMIN_PASSWORD are required")
		return
	}
	if _, e := s.CreateAdmin(ctx, name, user, pass); e != nil {
		log.Printf("admin bootstrap: %v", e)
	}
}
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
