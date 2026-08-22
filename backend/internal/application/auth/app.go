package auth

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/user"
	"github.com/maxaicrypto/backend/internal/infrastructure/auth/refresh"
)

// App implements Service.
type App struct {
	users         user.Repository
	sessions      user.SessionRepository
	subscriptions subscription.Repository
	tokens        TokenIssuer
	passwords     PasswordHasher
	google        GoogleVerifier
	refreshTTL    time.Duration
	aiDailyLimit  int
}

// NewApp wires the authentication service.
func NewApp(
	users user.Repository,
	sessions user.SessionRepository,
	subscriptions subscription.Repository,
	tokens TokenIssuer,
	passwords PasswordHasher,
	google GoogleVerifier,
	cfg *config.Config,
) *App {
	return &App{
		users:         users,
		sessions:      sessions,
		subscriptions: subscriptions,
		tokens:        tokens,
		passwords:     passwords,
		google:        google,
		refreshTTL:    cfg.Auth.RefreshTokenTTL,
		aiDailyLimit:  cfg.AI.DailyLimit,
	}
}

// CreateGuest implements Service.
func (a *App) CreateGuest(ctx context.Context, client ClientContext) (Session, error) {
	subject := uuid.NewString()
	u, err := a.users.CreateGuest(ctx, subject)
	if err != nil {
		return Session{}, err
	}
	return a.issueSession(ctx, u, client)
}

// RegisterEmail implements Service.
func (a *App) RegisterEmail(ctx context.Context, creds EmailCredentials, client ClientContext) (Session, error) {
	email, err := normalizeEmail(creds.Email)
	if err != nil {
		return Session{}, err
	}
	if err := validatePassword(creds.Password); err != nil {
		return Session{}, err
	}
	if _, found, err := a.users.FindIdentity(ctx, user.ProviderEmail, email); err != nil {
		return Session{}, err
	} else if found {
		return Session{}, apperr.New(apperr.CodeEmailAlreadyRegistered)
	}

	hash, err := a.passwords.Hash(creds.Password)
	if err != nil {
		return Session{}, apperr.Wrap(apperr.CodeInternal, err)
	}

	identity := user.Identity{
		Provider:      user.ProviderEmail,
		Subject:       email,
		Email:         &email,
		PasswordHash:  &hash,
		EmailVerified: true,
	}
	u, err := a.users.CreateRegistered(ctx, user.User{
		Kind:  user.KindRegistered,
		Email: &email,
	}, identity)
	if err != nil {
		return Session{}, err
	}
	return a.issueSession(ctx, u, client)
}

// LoginEmail implements Service.
func (a *App) LoginEmail(ctx context.Context, creds EmailCredentials, client ClientContext) (Session, error) {
	email, err := normalizeEmail(creds.Email)
	if err != nil {
		return Session{}, err
	}
	identity, found, err := a.users.FindIdentity(ctx, user.ProviderEmail, email)
	if err != nil {
		return Session{}, err
	}
	if !found || identity.PasswordHash == nil {
		return Session{}, apperr.New(apperr.CodeInvalidCredentials)
	}
	if err := a.passwords.Verify(*identity.PasswordHash, creds.Password); err != nil {
		return Session{}, apperr.New(apperr.CodeInvalidCredentials)
	}
	u, err := a.users.GetByID(ctx, identity.UserID)
	if err != nil {
		return Session{}, err
	}
	return a.issueSession(ctx, u, client)
}

// LoginGoogle implements Service.
func (a *App) LoginGoogle(ctx context.Context, creds GoogleCredentials, client ClientContext) (Session, error) {
	verified, err := a.google.Verify(ctx, creds.IDToken)
	if err != nil {
		return Session{}, err
	}
	if !verified.EmailVerified {
		return Session{}, apperr.New(apperr.CodeAuthentication).
			WithMessage("The Google account email is not verified.")
	}

	if u, found, err := a.users.FindByIdentity(ctx, user.ProviderGoogle, verified.Subject); err != nil {
		return Session{}, err
	} else if found {
		return a.issueSession(ctx, u, client)
	}

	email := strings.ToLower(strings.TrimSpace(verified.Email))
	identity := user.Identity{
		Provider:      user.ProviderGoogle,
		Subject:       verified.Subject,
		Email:         &email,
		EmailVerified: true,
	}
	u, err := a.users.CreateRegistered(ctx, user.User{
		Kind:        user.KindRegistered,
		Email:       &email,
		DisplayName: verified.DisplayName,
	}, identity)
	if err != nil {
		return Session{}, err
	}
	return a.issueSession(ctx, u, client)
}

// Upgrade implements Service.
func (a *App) Upgrade(ctx context.Context, userID uuid.UUID, req UpgradeRequest, client ClientContext) (Session, error) {
	current, err := a.users.GetByID(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if current.Kind != user.KindGuest {
		return Session{}, apperr.New(apperr.CodeValidation).
			WithMessage("Only guest accounts can be upgraded.")
	}

	switch req.Method {
	case UpgradeEmail:
		if req.Email == nil {
			return Session{}, apperr.New(apperr.CodeValidation).WithMessage("Email credentials are required.")
		}
		email, err := normalizeEmail(req.Email.Email)
		if err != nil {
			return Session{}, err
		}
		if err := validatePassword(req.Email.Password); err != nil {
			return Session{}, err
		}
		if _, found, err := a.users.FindIdentity(ctx, user.ProviderEmail, email); err != nil {
			return Session{}, err
		} else if found {
			return Session{}, apperr.New(apperr.CodeEmailAlreadyRegistered)
		}
		hash, err := a.passwords.Hash(req.Email.Password)
		if err != nil {
			return Session{}, apperr.Wrap(apperr.CodeInternal, err)
		}
		u, err := a.users.Upgrade(ctx, userID, user.Identity{
			Provider:      user.ProviderEmail,
			Subject:       email,
			Email:         &email,
			PasswordHash:  &hash,
			EmailVerified: true,
		}, nil)
		if err != nil {
			return Session{}, err
		}
		return a.issueSession(ctx, u, client)

	case UpgradeGoogle:
		if req.Google == nil {
			return Session{}, apperr.New(apperr.CodeValidation).WithMessage("Google credentials are required.")
		}
		verified, err := a.google.Verify(ctx, req.Google.IDToken)
		if err != nil {
			return Session{}, err
		}
		if _, found, err := a.users.FindIdentity(ctx, user.ProviderGoogle, verified.Subject); err != nil {
			return Session{}, err
		} else if found {
			return Session{}, apperr.New(apperr.CodeEmailAlreadyRegistered).
				WithMessage("This Google account is already linked to another user.")
		}
		email := strings.ToLower(strings.TrimSpace(verified.Email))
		u, err := a.users.Upgrade(ctx, userID, user.Identity{
			Provider:      user.ProviderGoogle,
			Subject:       verified.Subject,
			Email:         &email,
			EmailVerified: verified.EmailVerified,
		}, verified.DisplayName)
		if err != nil {
			return Session{}, err
		}
		return a.issueSession(ctx, u, client)

	default:
		return Session{}, apperr.New(apperr.CodeValidation).WithMessage("Unsupported upgrade method.")
	}
}

// Refresh implements Service.
func (a *App) Refresh(ctx context.Context, refreshToken string, client ClientContext) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, apperr.New(apperr.CodeSessionExpired)
	}

	session, found, err := a.sessions.FindByTokenHash(ctx, refresh.Hash(refreshToken))
	if err != nil {
		return Session{}, err
	}
	if !found {
		return Session{}, apperr.New(apperr.CodeSessionExpired)
	}

	now := time.Now().UTC()
	if session.RevokedAt != nil {
		if session.RotatedTo != nil {
			_ = a.sessions.RevokeAllForUser(ctx, session.UserID)
		}
		return Session{}, apperr.New(apperr.CodeSessionExpired)
	}
	if !session.IsActive(now) {
		return Session{}, apperr.New(apperr.CodeSessionExpired)
	}

	u, err := a.users.GetByID(ctx, session.UserID)
	if err != nil {
		return Session{}, err
	}

	nextToken, err := refresh.NewToken()
	if err != nil {
		return Session{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	nextSession := user.RefreshSession{
		UserID:    session.UserID,
		TokenHash: refresh.Hash(nextToken),
		ExpiresAt: now.Add(a.refreshTTL),
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
	}
	if _, err := a.sessions.Rotate(ctx, session.ID, nextSession); err != nil {
		return Session{}, err
	}

	accessToken, expiresAt, err := a.tokens.Issue(ctx, u.ID)
	if err != nil {
		return Session{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	profile, err := a.buildProfile(ctx, u)
	if err != nil {
		return Session{}, err
	}
	return Session{
		AccessToken:  accessToken,
		RefreshToken: nextToken,
		ExpiresAt:    expiresAt,
		User:         profile,
	}, nil
}

// Logout implements Service.
func (a *App) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	session, found, err := a.sessions.FindByTokenHash(ctx, refresh.Hash(refreshToken))
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return a.sessions.Revoke(ctx, session.ID)
}

// Session implements Service.
func (a *App) Session(ctx context.Context, userID uuid.UUID) (Profile, error) {
	u, err := a.users.GetByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	return a.buildProfile(ctx, u)
}

func (a *App) issueSession(ctx context.Context, u user.User, client ClientContext) (Session, error) {
	refreshToken, err := refresh.NewToken()
	if err != nil {
		return Session{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	now := time.Now().UTC()
	if _, err := a.sessions.Create(ctx, user.RefreshSession{
		UserID:    u.ID,
		TokenHash: refresh.Hash(refreshToken),
		ExpiresAt: now.Add(a.refreshTTL),
		UserAgent: client.UserAgent,
		IPAddress: client.IPAddress,
	}); err != nil {
		return Session{}, err
	}

	accessToken, expiresAt, err := a.tokens.Issue(ctx, u.ID)
	if err != nil {
		return Session{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	profile, err := a.buildProfile(ctx, u)
	if err != nil {
		return Session{}, err
	}
	return Session{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         profile,
	}, nil
}

func (a *App) buildProfile(ctx context.Context, u user.User) (Profile, error) {
	providers, err := a.users.ListAuthProviders(ctx, u.ID)
	if err != nil {
		return Profile{}, err
	}

	sub, found, err := a.subscriptions.GetByUser(ctx, u.ID)
	if err != nil {
		return Profile{}, err
	}
	if !found {
		sub = subscription.Subscription{
			UserID: u.ID,
			Plan:   subscription.PlanFree,
			Status: subscription.StatusActive,
		}
	}

	entitlements := subscription.FreeEntitlements
	entitlements.AIOperationsPerDay = a.aiDailyLimit

	return Profile{
		ID:           u.ID,
		Kind:         u.Kind,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		AuthMethods:  providers,
		Subscription: sub,
		Entitlements: entitlements,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || !strings.Contains(email, "@") {
		return "", apperr.New(apperr.CodeValidation).WithMessage("A valid email address is required.")
	}
	return email, nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return apperr.New(apperr.CodeValidation).WithMessage("Password must be at least 8 characters.")
	}
	return nil
}

var _ Service = (*App)(nil)
