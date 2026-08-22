package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/user"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// AuthHandler serves the /api/v1/auth routes.
type AuthHandler struct {
	service appauth.Service
}

// NewAuthHandler builds the auth handler.
func NewAuthHandler(service appauth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

type emailCredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type googleAuthRequest struct {
	IDToken string `json:"id_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type upgradeAccountRequest struct {
	Method   string `json:"method"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	IDToken  string `json:"id_token,omitempty"`
}

type authSessionResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID           string               `json:"id"`
	Kind         user.Kind            `json:"kind"`
	Email        *string              `json:"email"`
	DisplayName  *string              `json:"display_name"`
	AuthMethods  []string             `json:"auth_methods"`
	Subscription subscriptionResponse `json:"subscription"`
	CreatedAt    time.Time            `json:"created_at"`
}

type subscriptionResponse struct {
	Plan             subscription.Plan    `json:"plan"`
	Status           subscription.Status  `json:"status"`
	CurrentPeriodEnd *time.Time           `json:"current_period_end"`
	Entitlements     entitlementsResponse `json:"entitlements"`
}

type entitlementsResponse struct {
	MaxWallets         int      `json:"max_wallets"`
	AIOperationsPerDay int      `json:"ai_operations_per_day"`
	Features           []string `json:"features"`
}

// CreateGuest handles POST /auth/guest.
func (h *AuthHandler) CreateGuest(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.CreateGuest(r.Context(), clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.Created(w, r, mapSession(session))
}

// RegisterEmail handles POST /auth/email/register.
func (h *AuthHandler) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	var body emailCredentialsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	session, err := h.service.RegisterEmail(r.Context(), appauth.EmailCredentials{
		Email:    body.Email,
		Password: body.Password,
	}, clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.Created(w, r, mapSession(session))
}

// LoginEmail handles POST /auth/email/login.
func (h *AuthHandler) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var body emailCredentialsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	session, err := h.service.LoginEmail(r.Context(), appauth.EmailCredentials{
		Email:    body.Email,
		Password: body.Password,
	}, clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.OK(w, r, mapSession(session))
}

// LoginGoogle handles POST /auth/google.
func (h *AuthHandler) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	var body googleAuthRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	session, err := h.service.LoginGoogle(r.Context(), appauth.GoogleCredentials{
		IDToken: body.IDToken,
	}, clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.OK(w, r, mapSession(session))
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	session, err := h.service.Refresh(r.Context(), body.RefreshToken, clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.OK(w, r, mapSession(session))
}

// Upgrade handles POST /auth/upgrade.
func (h *AuthHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}

	var body upgradeAccountRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}

	req := appauth.UpgradeRequest{Method: appauth.UpgradeMethod(strings.ToUpper(body.Method))}
	switch req.Method {
	case appauth.UpgradeEmail:
		req.Email = &appauth.EmailCredentials{Email: body.Email, Password: body.Password}
	case appauth.UpgradeGoogle:
		req.Google = &appauth.GoogleCredentials{IDToken: body.IDToken}
	default:
		response.Error(w, r, apperr.New(apperr.CodeValidation).WithMessage("Unsupported upgrade method."))
		return
	}

	session, err := h.service.Upgrade(r.Context(), userID, req, clientContext(r))
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.OK(w, r, mapSession(session))
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	if err := h.service.Logout(r.Context(), body.RefreshToken); err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.NoContent(w)
}

// Session handles GET /auth/session.
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	profile, err := h.service.Session(r.Context(), userID)
	if err != nil {
		response.Error(w, r, apperr.From(err))
		return
	}
	response.OK(w, r, mapProfile(profile))
}

func mapSession(session appauth.Session) authSessionResponse {
	return authSessionResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt.UTC(),
		User:         mapProfile(session.User),
	}
}

func mapProfile(profile appauth.Profile) userResponse {
	return userResponse{
		ID:           profile.ID.String(),
		Kind:         profile.Kind,
		Email:        profile.Email,
		DisplayName:  profile.DisplayName,
		AuthMethods:  mapAuthMethods(profile.AuthMethods),
		Subscription: mapSubscription(profile.Subscription, profile.Entitlements),
		CreatedAt:    profile.CreatedAt.UTC(),
	}
}

func mapSubscription(sub subscription.Subscription, entitlements subscription.Entitlements) subscriptionResponse {
	features := make([]string, 0, len(entitlements.Features))
	for _, feature := range entitlements.Features {
		features = append(features, string(feature))
	}
	return subscriptionResponse{
		Plan:             sub.Plan,
		Status:           sub.Status,
		CurrentPeriodEnd: sub.CurrentPeriodEnd,
		Entitlements: entitlementsResponse{
			MaxWallets:         entitlements.MaxWallets,
			AIOperationsPerDay: entitlements.AIOperationsPerDay,
			Features:           features,
		},
	}
}

func mapAuthMethods(providers []user.AuthProvider) []string {
	methods := make([]string, 0, len(providers))
	for _, provider := range providers {
		switch provider {
		case user.ProviderGuest:
			methods = append(methods, "GUEST")
		case user.ProviderGoogle:
			methods = append(methods, "GOOGLE")
		case user.ProviderEmail:
			methods = append(methods, "EMAIL")
		default:
			methods = append(methods, strings.ToUpper(string(provider)))
		}
	}
	return methods
}

func clientContext(r *http.Request) appauth.ClientContext {
	ua := r.UserAgent()
	ip := clientIP(r)
	return appauth.ClientContext{
		UserAgent: stringPtr(ua),
		IPAddress: ip,
	}
}

func clientIP(r *http.Request) *string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return &ip
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		raw := strings.TrimSpace(r.RemoteAddr)
		if raw == "" {
			return nil
		}
		return &raw
	}
	return &host
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
