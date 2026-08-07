package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"github.com/gymax/gymax/internal/service"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	users  *service.UserService
	gyms   *service.GymService
	jwtKey []byte
}

func New(u *service.UserService, c *service.GymService, jwtSecret string) http.Handler {
	a := &API{u, c, []byte(jwtSecret)}
	r := chi.NewRouter()
	r.Use(recoverer, securityHeaders, locale)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	r.Route("/v1", func(r chi.Router) {
		r.Get("/meta/locales", func(w http.ResponseWriter, _ *http.Request) {
			write(w, 200, map[string]any{"default": "fa", "supported": []string{"fa", "en"}, "direction": map[string]string{"fa": "rtl", "en": "ltr"}})
		})
		r.Post("/auth/admin/login", a.adminLogin)
		r.Post("/auth/register/start", a.startRegistration)
		r.Post("/auth/register/complete", a.completeRegistration)
		r.Group(func(r chi.Router) {
			r.Use(a.identity)
			r.Post("/auth/phone-change/start", a.startPhoneChange)
			r.Post("/auth/phone-change/complete", a.completePhoneChange)
			r.Get("/users/{id}", a.getUser)
			r.Put("/users/{id}", a.updateUser)
			r.With(requireRole(domain.RoleAdmin)).Get("/users", a.listUsers)
			r.With(requireRole(domain.RoleAdmin)).Post("/admins", a.createAdmin)
			r.With(requireRole(domain.RoleAdmin)).Delete("/users/{id}", a.deleteUser)
			r.With(requireRole(domain.RoleCoach)).Post("/gyms", a.createGym)
			r.Delete("/gyms/{id}", a.deleteGym)
			r.Post("/gyms/{id}/athletes/{athleteID}", a.addAthlete)
		})
	})
	return r
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
	respond(w, e, 200, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 900, "user": u})
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
	respond(w, e, http.StatusCreated, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 900, "user": u})
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
func (a *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	e := a.users.Delete(r.Context(), chi.URLParam(r, "id"), current(r).ID)
	if e == nil {
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
func (a *API) deleteGym(w http.ResponseWriter, r *http.Request) {
	e := a.gyms.Delete(r.Context(), chi.URLParam(r, "id"), current(r))
	if e == nil {
		w.WriteHeader(204)
		return
	}
	problem(w, e)
}
func (a *API) addAthlete(w http.ResponseWriter, r *http.Request) {
	e := a.gyms.AddAthlete(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "athleteID"), current(r))
	respond(w, e, 201, map[string]string{"message": message(w, "athlete_joined_gym")})
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
	case errors.Is(e, domain.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(e, domain.ErrConflict):
		status, code = 409, "conflict"
	case errors.Is(e, domain.ErrGymLimit):
		status, code = 422, "gym_limit_reached"
	}
	write(w, status, map[string]string{"error": code, "message": message(w, code)})
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
		}, jwt.WithIssuer("gymax"), jwt.WithExpirationRequired())
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
	claims := jwt.MapClaims{"sub": u.ID, "role": u.Role, "iat": now.Unix(), "exp": now.Add(15 * time.Minute).Unix(), "iss": "gymax"}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.jwtKey)
}
