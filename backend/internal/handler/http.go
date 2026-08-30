package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/reference"
	"github.com/clubmax/clubmax/internal/repository"
	"github.com/clubmax/clubmax/internal/service"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type API struct {
	users     *service.UserService
	gyms      *service.GymService
	finance   *service.FinanceService
	messages  *service.MessageService
	admin     *service.AdminService
	jwtKey    []byte
	uploadDir string
}

func New(u *service.UserService, c *service.GymService, f *service.FinanceService, m *service.MessageService, admin *service.AdminService, jwtSecret, frontendOrigin, uploadDir string) http.Handler {
	a := &API{users: u, gyms: c, finance: f, messages: m, admin: admin, jwtKey: []byte(jwtSecret), uploadDir: uploadDir}
	r := chi.NewRouter()
	r.Use(recoverer, securityHeaders, locale, cors(frontendOrigin))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	r.Handle("/media/*", http.StripPrefix("/media/", http.FileServer(http.Dir(uploadDir))))
	r.Route("/v1", func(r chi.Router) {
		r.Get("/meta/locales", func(w http.ResponseWriter, _ *http.Request) {
			write(w, 200, map[string]any{"default": "fa", "supported": []string{"fa", "en"}, "direction": map[string]string{"fa": "rtl", "en": "ltr"}})
		})
		r.Get("/meta/locations", func(w http.ResponseWriter, r *http.Request) {
			if province := r.URL.Query().Get("province"); province != "" {
				counties, ok := reference.Counties(province)
				if !ok {
					problem(w, domain.ErrNotFound)
					return
				}
				write(w, 200, map[string]any{"province": province, "counties": counties})
				return
			}
			write(w, 200, map[string]any{"provinces": reference.ProvinceNames()})
		})
		r.With(rateLimit(10, time.Minute)).Post("/auth/admin/login", a.adminLogin)
		r.With(rateLimit(10, time.Minute)).Post("/auth/member/login", a.memberLogin)
		r.With(rateLimit(10, time.Minute)).Post("/auth/login", a.login)
		r.With(rateLimit(5, time.Minute)).Post("/auth/register", a.register)
		r.Post("/auth/login/start", a.startLogin)
		r.Post("/auth/login/complete", a.completeLogin)
		r.Post("/auth/register/start", a.startRegistration)
		r.Post("/auth/register/complete", a.completeRegistration)
		r.Group(func(r chi.Router) {
			r.Use(a.identity)
			r.Post("/auth/phone-change/start", a.startPhoneChange)
			r.Post("/auth/phone-change/complete", a.completePhoneChange)
			r.With(rateLimit(5, time.Minute)).Post("/users/me/deletion/start", a.startAccountDeletion)
			r.With(rateLimit(10, time.Minute)).Post("/users/me/deletion/complete", a.completeAccountDeletion)
			r.Get("/users/me", a.me)
			r.Post("/users/me/documents/{kind}", a.uploadUserDocument)
			r.Get("/users/{id}", a.getUser)
			r.Put("/users/{id}", a.updateUser)
			r.With(requireRole(domain.RoleAdmin)).Get("/users", a.listUsers)
			r.With(requireRole(domain.RoleAdmin)).Post("/admins", a.createAdmin)
			r.With(requireRole(domain.RoleAdmin)).Put("/users/{id}/promote", a.promoteUser)
			r.With(requireRole(domain.RoleAdmin)).Delete("/users/{id}", a.deleteUser)
			r.With(requireRole(domain.RoleAdmin)).Get("/admin/overview", a.adminOverview)
			r.With(requireRole(domain.RoleAdmin)).Get("/admin/gyms", a.adminGyms)
			r.With(requireRole(domain.RoleAdmin)).Get("/admin/payments", a.adminPayments)
			r.With(requireRole(domain.RoleCoach)).Post("/gyms", a.createGym)
			r.With(requireRole(domain.RoleAthlete)).Get("/gyms/catalog", a.gymCatalog)
			r.With(requireRole(domain.RoleAthlete)).Get("/gyms/catalog/{id}", a.gymCatalogItem)
			r.With(requireRole(domain.RoleAthlete)).Post("/gyms/{id}/pay", a.payGymTuition)
			r.With(requireRole(domain.RoleAthlete)).Post("/gyms/{id}/monthly-dues/pay", a.payMonthlyTuition)
			r.With(requireRole(domain.RoleAthlete)).Get("/payments/me", a.athletePayments)
			r.With(requireRole(domain.RoleAthlete)).Get("/registration/fee", a.registrationFee)
			r.Get("/gyms", a.listGyms)
			r.Get("/gyms/{id}", a.getGym)
			r.Put("/gyms/{id}", a.updateGym)
			r.Delete("/gyms/{id}", a.deleteGym)
			r.Get("/gyms/{id}/images", a.listGymImages)
			r.Post("/gyms/{id}/images", a.uploadGymImage)
			r.Delete("/gyms/{id}/images/{imageID}", a.deleteGymImage)
			r.With(requireRole(domain.RoleCoach, domain.RoleAdmin)).Post("/gyms/{id}/athletes/{athleteID}", a.addAthlete)
			r.With(requireRole(domain.RoleCoach, domain.RoleAdmin)).Get("/gyms/{id}/athletes", a.gymAthletes)
			r.With(requireRole(domain.RoleCoach)).Get("/gyms/{id}/athletes/{athleteID}/payments", a.gymAthletePayments)
			r.Get("/messages", a.inbox)
			r.Post("/messages", a.sendMessage)
			r.Put("/messages/{id}/read", a.markMessageRead)
			r.Get("/finance/policy", a.getPolicy)
			r.With(requireRole(domain.RoleAdmin)).Put("/finance/policy", a.setPolicy)
			r.With(requireRole(domain.RoleCoach)).Get("/wallet", a.wallet)
			r.With(requireRole(domain.RoleCoach)).Get("/wallet/ledger", a.ledger)
			r.Get("/withdrawals", a.withdrawals)
			r.With(requireRole(domain.RoleCoach)).Post("/withdrawals", a.requestWithdrawal)
			r.With(requireRole(domain.RoleAdmin)).Put("/withdrawals/{id}/review", a.reviewWithdrawal)
		})
	})
	return r
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var v service.RegisterUser
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.Register(r.Context(), v)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, http.StatusCreated, authResponse(token, u))
}

func (a *API) memberLogin(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.AuthenticateMember(r.Context(), v.Phone, v.Password)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, http.StatusOK, authResponse(token, u))
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.Authenticate(r.Context(), v.Identity, v.Password)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, http.StatusOK, authResponse(token, u))
}

func authResponse(token string, user *domain.User) map[string]any {
	return map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 28800, "user": user}
}

func cors(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if originAllowed(origin, allowedOrigin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin, configured string) bool {
	if origin == configured {
		return true
	}
	actual, actualErr := url.Parse(origin)
	expected, expectedErr := url.Parse(configured)
	if actualErr != nil || expectedErr != nil || actual.Scheme != expected.Scheme || actual.Port() != expected.Port() {
		return false
	}
	isLoopback := func(host string) bool { return host == "localhost" || host == "127.0.0.1" || host == "::1" }
	return isLoopback(actual.Hostname())
}
func (a *API) adminLogin(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.AuthenticateAdmin(r.Context(), v.Username, v.Password)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, 200, authResponse(token, u))
}
func (a *API) startLogin(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Phone string `json:"phone"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	respond(w, a.users.StartLogin(r.Context(), v.Phone), 202, map[string]string{"message": message(w, "verification_code_sent")})
}
func (a *API) completeLogin(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.CompleteLogin(r.Context(), v.Phone, v.Code)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, 200, authResponse(token, u))
}
func (a *API) startRegistration(w http.ResponseWriter, r *http.Request) {
	var v service.StartRegistration
	if decode(w, r, &v) != nil {
		return
	}
	respond(w, a.users.StartRegistration(r.Context(), v), http.StatusAccepted, map[string]string{"message": message(w, "verification_code_sent")})
}
func (a *API) completeRegistration(w http.ResponseWriter, r *http.Request) {
	var v service.CompleteRegistration
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.CompleteRegistration(r.Context(), v)
	if e != nil {
		problem(w, e)
		return
	}
	token, e := a.signToken(u)
	respond(w, e, http.StatusCreated, authResponse(token, u))
}
func (a *API) startPhoneChange(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Phone string `json:"phone"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	respond(w, a.users.StartPhoneChange(r.Context(), current(r).ID, v.Phone), 202, map[string]string{"message": message(w, "verification_code_sent")})
}
func (a *API) completePhoneChange(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	respond(w, a.users.CompletePhoneChange(r.Context(), current(r).ID, v.Phone, v.Code), 200, map[string]string{"message": message(w, "phone_updated")})
}
func (a *API) getUser(w http.ResponseWriter, r *http.Request) {
	actor := current(r)
	id := chi.URLParam(r, "id")
	if actor.Role != domain.RoleAdmin && actor.ID != id {
		problem(w, domain.ErrForbidden)
		return
	}
	u, e := a.users.Get(r.Context(), id)
	respond(w, e, 200, u)
}
func (a *API) me(w http.ResponseWriter, r *http.Request) {
	u, e := a.users.Get(r.Context(), current(r).ID)
	respond(w, e, 200, u)
}
func (a *API) uploadUserDocument(w http.ResponseWriter, r *http.Request) {
	kind := domain.DocumentKind(chi.URLParam(r, "kind"))
	r.Body = http.MaxBytesReader(w, r.Body, 500*1024+64*1024)
	file, _, err := r.FormFile("file")
	if err != nil {
		problem(w, domain.ErrInvalid)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 500*1024+1))
	if err != nil || len(data) == 0 || len(data) > 500*1024 {
		problem(w, domain.ErrInvalid)
		return
	}
	mime := http.DetectContentType(data)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "application/pdf": ".pdf"}
	ext, ok := extensions[mime]
	if !ok {
		problem(w, domain.ErrInvalid)
		return
	}
	relative := filepath.Join("users", current(r).ID, string(kind), uuid.NewString()+ext)
	absolute := filepath.Join(a.uploadDir, relative)
	if err = os.MkdirAll(filepath.Dir(absolute), 0750); err != nil {
		problem(w, err)
		return
	}
	if err = os.WriteFile(absolute, data, 0640); err != nil {
		problem(w, err)
		return
	}
	doc, e := a.users.AddDocument(r.Context(), current(r).ID, kind, filepath.ToSlash(relative), mime, int64(len(data)))
	if e != nil {
		_ = os.Remove(absolute)
		problem(w, e)
		return
	}
	respond(w, nil, 201, map[string]any{"document": doc, "url": fmt.Sprintf("/media/%s", filepath.ToSlash(relative))})
}
func (a *API) updateUser(w http.ResponseWriter, r *http.Request) {
	actor, id := current(r), chi.URLParam(r, "id")
	if actor.Role != domain.RoleAdmin && actor.ID != id {
		problem(w, domain.ErrForbidden)
		return
	}
	var v domain.User
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.UpdateProfile(r.Context(), id, &v)
	respond(w, e, 200, u)
}
func (a *API) startAccountDeletion(w http.ResponseWriter, r *http.Request) {
	e := a.users.StartAccountDeletion(r.Context(), current(r).ID)
	respond(w, e, http.StatusAccepted, map[string]string{"message": message(w, "verification_code_sent")})
}
func (a *API) completeAccountDeletion(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Code string `json:"code"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	userID := current(r).ID
	if e := a.users.CompleteAccountDeletion(r.Context(), userID, v.Code); e != nil {
		problem(w, e)
		return
	}
	if userID != "" {
		_ = os.RemoveAll(filepath.Join(a.uploadDir, "users", userID))
	}
	w.WriteHeader(http.StatusNoContent)
}
func (a *API) createAdmin(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	u, e := a.users.CreateAdmin(r.Context(), v.Name, v.Username, v.Password)
	respond(w, e, 201, u)
}
func (a *API) promoteUser(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	e := a.users.PromoteToAdmin(r.Context(), chi.URLParam(r, "id"), v.Username, v.Password)
	respond(w, e, 200, map[string]string{"message": "کاربر به مدیر سایت تبدیل شد."})
}
func (a *API) listUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var role *domain.Role
	if v := r.URL.Query().Get("role"); v != "" {
		x := domain.Role(v)
		role = &x
	}
	v, e := a.users.List(r.Context(), repository.UserFilter{Role: role, Search: r.URL.Query().Get("q"), Limit: limit, Offset: offset})
	respond(w, e, 200, v)
}
func (a *API) adminOverview(w http.ResponseWriter, r *http.Request) {
	v, e := a.admin.Overview(r.Context())
	respond(w, e, 200, v)
}
func (a *API) adminGyms(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	v, e := a.admin.Gyms(r.Context(), r.URL.Query().Get("q"), limit, offset)
	respond(w, e, 200, v)
}
func (a *API) adminPayments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	v, e := a.admin.Payments(r.Context(), r.URL.Query().Get("q"), limit, offset)
	respond(w, e, 200, v)
}
func (a *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	e := a.users.Delete(r.Context(), id, current(r).ID)
	if e == nil {
		_ = os.RemoveAll(filepath.Join(a.uploadDir, "users", id))
		w.WriteHeader(204)
		return
	}
	problem(w, e)
}
func (a *API) createGym(w http.ResponseWriter, r *http.Request) {
	var v domain.Gym
	if decode(w, r, &v) != nil {
		return
	}
	c, e := a.gyms.Create(r.Context(), &v, current(r).ID)
	respond(w, e, 201, c)
}
func (a *API) listGyms(w http.ResponseWriter, r *http.Request) {
	gyms, e := a.gyms.List(r.Context(), current(r))
	respond(w, e, 200, gyms)
}
func (a *API) gymCatalog(w http.ResponseWriter, r *http.Request) {
	v, e := a.gyms.Catalog(r.Context(), current(r))
	respond(w, e, 200, v)
}
func (a *API) gymCatalogItem(w http.ResponseWriter, r *http.Request) {
	v, e := a.gyms.CatalogItem(r.Context(), chi.URLParam(r, "id"), current(r))
	respond(w, e, 200, v)
}
func (a *API) registrationFee(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.RegistrationFee(r.Context())
	respond(w, e, 200, v)
}
func (a *API) payGymTuition(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.PayTuition(r.Context(), current(r).ID, chi.URLParam(r, "id"))
	respond(w, e, 201, v)
}
func (a *API) getGym(w http.ResponseWriter, r *http.Request) {
	gym, e := a.gyms.Get(r.Context(), chi.URLParam(r, "id"), current(r))
	respond(w, e, 200, gym)
}
func (a *API) updateGym(w http.ResponseWriter, r *http.Request) {
	var v domain.Gym
	if decode(w, r, &v) != nil {
		return
	}
	gym, e := a.gyms.Update(r.Context(), chi.URLParam(r, "id"), &v, current(r))
	respond(w, e, 200, gym)
}
func (a *API) listGymImages(w http.ResponseWriter, r *http.Request) {
	images, e := a.gyms.Images(r.Context(), chi.URLParam(r, "id"), current(r))
	respond(w, e, 200, images)
}
func (a *API) uploadGymImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 500*1024+64*1024)
	file, _, err := r.FormFile("image")
	if err != nil {
		problem(w, domain.ErrInvalid)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 500*1024+1))
	if err != nil || len(data) == 0 || len(data) > 500*1024 {
		problem(w, domain.ErrInvalid)
		return
	}
	mime := http.DetectContentType(data)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	ext, ok := extensions[mime]
	if !ok {
		problem(w, domain.ErrInvalid)
		return
	}
	gymID := chi.URLParam(r, "id")
	relative := filepath.Join("gyms", gymID, uuid.NewString()+ext)
	absolute := filepath.Join(a.uploadDir, relative)
	if err = os.MkdirAll(filepath.Dir(absolute), 0750); err != nil {
		problem(w, err)
		return
	}
	if err = os.WriteFile(absolute, data, 0640); err != nil {
		problem(w, err)
		return
	}
	image, e := a.gyms.AddImage(r.Context(), gymID, current(r), filepath.ToSlash(relative), mime, int64(len(data)))
	if e != nil {
		_ = os.Remove(absolute)
		problem(w, e)
		return
	}
	respond(w, nil, 201, map[string]any{"image": image, "url": fmt.Sprintf("/media/%s", filepath.ToSlash(relative))})
}
func (a *API) deleteGymImage(w http.ResponseWriter, r *http.Request) {
	image, e := a.gyms.DeleteImage(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "imageID"), current(r))
	if e != nil {
		problem(w, e)
		return
	}
	_ = os.Remove(filepath.Join(a.uploadDir, filepath.FromSlash(image.StorageKey)))
	w.WriteHeader(http.StatusNoContent)
}
func (a *API) deleteGym(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	images, _ := a.gyms.Images(r.Context(), id, current(r))
	e := a.gyms.Delete(r.Context(), id, current(r))
	if e == nil {
		for _, image := range images {
			_ = os.Remove(filepath.Join(a.uploadDir, filepath.FromSlash(image.StorageKey)))
		}
		w.WriteHeader(204)
		return
	}
	problem(w, e)
}
func (a *API) addAthlete(w http.ResponseWriter, r *http.Request) {
	e := a.gyms.AddAthlete(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "athleteID"), current(r))
	respond(w, e, 201, map[string]string{"message": message(w, "athlete_joined_gym")})
}
func (a *API) gymAthletes(w http.ResponseWriter, r *http.Request) {
	v, e := a.gyms.Athletes(r.Context(), chi.URLParam(r, "id"), current(r))
	respond(w, e, 200, v)
}
func (a *API) gymAthletePayments(w http.ResponseWriter, r *http.Request) {
	payments, e := a.finance.GymAthletePayments(r.Context(), current(r).ID, chi.URLParam(r, "id"), chi.URLParam(r, "athleteID"))
	respond(w, e, http.StatusOK, payments)
}
func (a *API) payMonthlyTuition(w http.ResponseWriter, r *http.Request) {
	payment, e := a.finance.PayMonthlyTuition(r.Context(), current(r).ID, chi.URLParam(r, "id"))
	respond(w, e, http.StatusCreated, payment)
}
func (a *API) athletePayments(w http.ResponseWriter, r *http.Request) {
	payments, e := a.finance.AthletePayments(r.Context(), current(r).ID)
	respond(w, e, http.StatusOK, payments)
}
func (a *API) inbox(w http.ResponseWriter, r *http.Request) {
	v, e := a.messages.Inbox(r.Context(), current(r).ID, 20, 0)
	respond(w, e, 200, v)
}
func (a *API) sendMessage(w http.ResponseWriter, r *http.Request) {
	var v domain.Message
	if decode(w, r, &v) != nil {
		return
	}
	m, e := a.messages.Send(r.Context(), current(r).ID, &v)
	respond(w, e, 201, m)
}
func (a *API) markMessageRead(w http.ResponseWriter, r *http.Request) {
	e := a.messages.MarkRead(r.Context(), current(r).ID, chi.URLParam(r, "id"))
	respond(w, e, 200, map[string]string{"message": "ok"})
}
func (a *API) getPolicy(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.Policy(r.Context())
	respond(w, e, 200, v)
}
func (a *API) setPolicy(w http.ResponseWriter, r *http.Request) {
	var v domain.WithdrawalPolicy
	if decode(w, r, &v) != nil {
		return
	}
	e := a.finance.SetPolicy(r.Context(), current(r).ID, v)
	respond(w, e, 200, map[string]string{"message": "ok"})
}
func (a *API) wallet(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.Balance(r.Context(), current(r).ID)
	respond(w, e, 200, map[string]int64{"balance": v})
}
func (a *API) ledger(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.Ledger(r.Context(), current(r).ID, 50, 0)
	respond(w, e, 200, v)
}
func (a *API) withdrawals(w http.ResponseWriter, r *http.Request) {
	v, e := a.finance.Withdrawals(r.Context(), current(r), 50, 0)
	respond(w, e, 200, v)
}
func (a *API) requestWithdrawal(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Amount int64 `json:"amount"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	x, e := a.finance.RequestWithdrawal(r.Context(), current(r).ID, v.Amount)
	respond(w, e, 201, x)
}
func (a *API) reviewWithdrawal(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Status           domain.WithdrawalStatus `json:"status"`
		BankTrackingCode *string                 `json:"bank_tracking_code"`
		RejectionReason  *string                 `json:"rejection_reason"`
	}
	if decode(w, r, &v) != nil {
		return
	}
	e := a.finance.ReviewWithdrawal(r.Context(), current(r).ID, chi.URLParam(r, "id"), v.Status, v.BankTrackingCode, v.RejectionReason)
	respond(w, e, 200, map[string]string{"message": "ok"})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		problem(w, domain.ErrInvalid)
		return e
	}
	return nil
}
func respond(w http.ResponseWriter, e error, status int, v any) {
	if e != nil {
		problem(w, e)
		return
	}
	write(w, status, v)
}
func problem(w http.ResponseWriter, e error) {
	status := 500
	code := "internal_error"
	switch {
	case errors.Is(e, domain.ErrInvalid):
		status, code = 400, "invalid_input"
	case errors.Is(e, domain.ErrInvalidOTP):
		status, code = 400, "invalid_otp"
	case errors.Is(e, domain.ErrForbidden):
		status, code = 403, "forbidden"
	case errors.Is(e, domain.ErrInvalidCredentials):
		status, code = 401, "invalid_credentials"
	case errors.Is(e, domain.ErrProtectedAdmin):
		status, code = 409, "protected_admin"
	case errors.Is(e, domain.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(e, domain.ErrConflict):
		status, code = 409, "conflict"
	case errors.Is(e, domain.ErrGymLimit):
		status, code = 422, "gym_limit_reached"
	case errors.Is(e, domain.ErrGymImageLimit):
		status, code = 422, "gym_image_limit_reached"
	case errors.Is(e, domain.ErrGymFull):
		status, code = 422, "gym_full"
	case errors.Is(e, domain.ErrInsufficientFunds):
		status, code = 422, "insufficient_funds"
	case errors.Is(e, domain.ErrWithdrawalLimit):
		status, code = 422, "withdrawal_limit"
	case errors.Is(e, domain.ErrInvalidState):
		status, code = 409, "invalid_state"
	case errors.Is(e, domain.ErrMessageRecipient):
		status, code = 403, "message_recipient"
	}
	payload := map[string]string{"error": code, "message": message(w, code)}
	var validation *domain.ValidationError
	if errors.As(e, &validation) {
		labels := map[string]string{"phone": "شماره موبایل", "role": "نقش کاربری", "password": "رمز عبور", "first_name": "نام", "last_name": "نام خانوادگی", "national_id": "کد ملی", "birth_date": "تاریخ تولد", "education": "تحصیلات", "gender": "جنسیت", "marital_status": "وضعیت تأهل", "province": "استان", "city": "استان یا شهرستان", "bank_iban": "شماره شبا", "has_medical_history": "سابقه بیماری", "blood_type": "گروه خونی", "emergency_phone": "شماره تماس اضطراری", "emergency_relation": "نسبت تماس اضطراری"}
		payload["field"] = validation.Field
		if label := labels[validation.Field]; label != "" {
			payload["message"] = label + " واردشده معتبر نیست."
		}
	}
	write(w, status, payload)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				problem(w, errors.New("panic"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *API) identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		if len(raw) < 8 || raw[:7] != "Bearer " {
			write(w, 401, map[string]string{"error": "unauthorized", "message": message(w, "unauthorized")})
			return
		}
		token, e := jwt.Parse(raw[7:], func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return a.jwtKey, nil
		}, jwt.WithIssuer("clubmax"), jwt.WithAudience("clubmax-web"), jwt.WithExpirationRequired())
		if e != nil || !token.Valid {
			write(w, 401, map[string]string{"error": "unauthorized", "message": message(w, "unauthorized")})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			write(w, 401, map[string]string{"error": "unauthorized", "message": message(w, "unauthorized")})
			return
		}
		id, _ := claims["sub"].(string)
		roleValue, _ := claims["role"].(string)
		role := domain.Role(roleValue)
		if id == "" || (role != domain.RoleAdmin && role != domain.RoleCoach && role != domain.RoleAthlete) {
			write(w, 401, map[string]string{"error": "unauthorized", "message": message(w, "unauthorized")})
			return
		}
		r = r.WithContext(withUser(r.Context(), &domain.User{ID: id, Role: role}))
		next.ServeHTTP(w, r)
	})
}
func requireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, v := range roles {
				if current(r).Role == v {
					next.ServeHTTP(w, r)
					return
				}
			}
			problem(w, domain.ErrForbidden)
		})
	}
}
func (a *API) signToken(u *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"sub": u.ID, "role": u.Role, "iat": now.Unix(), "exp": now.Add(8 * time.Hour).Unix(), "iss": "clubmax", "aud": "clubmax-web"}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.jwtKey)
}
