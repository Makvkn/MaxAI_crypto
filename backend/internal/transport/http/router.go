// Package http assembles the API surface: the middleware chain from §154 and
// the versioned route tree from §90.
package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/maxaicrypto/backend/internal/app/config"
	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
	"github.com/maxaicrypto/backend/internal/transport/http/handlers"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// RouterDeps carries everything the route tree needs. Application services are
// added here as each vertical slice lands.
type RouterDeps struct {
	Config  *config.Config
	Logger  *slog.Logger
	Health  *handlers.HealthHandler
	Auth      *handlers.AuthHandler
	Wallets      *handlers.WalletsHandler
	Portfolio    *handlers.PortfolioHandler
	Performance  *handlers.PerformanceHandler
	Transactions *handlers.TransactionsHandler
	AI           *handlers.AIHandler
	Tokens       appauth.TokenIssuer
	RateLimiter  *redis.RateLimiter
}

// NewRouter builds the HTTP handler for the API process.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Middleware order follows §154: identity and recovery wrap everything so
	// that a panic anywhere below still produces a correlated error envelope.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(deps.Logger))
	r.Use(middleware.Recovery)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.Config.HTTP.AllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", middleware.HeaderRequestID},
		ExposedHeaders:   []string{middleware.HeaderRequestID},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.BodyLimit(deps.Config.HTTP.MaxRequestBytes))
	if deps.Tokens != nil {
		r.Use(middleware.OptionalAuthenticate(deps.Tokens))
	}
	r.Use(middleware.RateLimit(deps.RateLimiter, deps.Config.Limits))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, r, apperr.New(apperr.CodeNotFound))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, r, apperr.New(apperr.CodeNotFound).
			WithMessage("This method is not allowed for the requested resource."))
	})

	// Operational endpoints live outside the versioned contract.
	r.Get("/health", deps.Health.Live)
	r.Get("/ready", deps.Health.Ready)

	r.Route("/api/v1", func(v1 chi.Router) {
		mountV1(v1, deps)
	})

	return r
}

// mountV1 declares the versioned route tree. Groups are registered here as the
// corresponding application services are implemented; every route below is
// defined in the OpenAPI contract first (§15 Principle, §136).
func mountV1(r chi.Router, deps RouterDeps) {
	mountAuth(r, deps)
	mountWallets(r, deps)

	r.Route("/ai", func(ai chi.Router) {
		if deps.AI == nil || deps.Tokens == nil {
			ai.Get("/usage", notImplemented)
			ai.Post("/scenarios", notImplemented)
			ai.Route("/conversations", func(conversations chi.Router) {
				conversations.Get("/", notImplemented)
				conversations.Post("/", notImplemented)
				conversations.Get("/{conversationID}", notImplemented)
				conversations.Get("/{conversationID}/messages", notImplemented)
				conversations.Post("/{conversationID}/messages", notImplemented)
			})
			return
		}

		ai.With(middleware.Authenticate(deps.Tokens)).Group(func(secured chi.Router) {
			secured.Get("/usage", deps.AI.GetUsage)
			secured.Post("/scenarios", deps.AI.SimulateScenario)
			secured.Route("/conversations", func(conversations chi.Router) {
				conversations.Get("/", deps.AI.ListConversations)
				conversations.Post("/", deps.AI.CreateConversation)
				conversations.Get("/{conversationID}", deps.AI.GetConversation)
				conversations.Get("/{conversationID}/messages", deps.AI.ListMessages)
				conversations.Post("/{conversationID}/messages", deps.AI.SendMessage)
			})
		})
	})
}

func mountWallets(r chi.Router, deps RouterDeps) {
	r.Route("/wallets", func(wallets chi.Router) {
		if deps.Wallets == nil || deps.Tokens == nil {
			wallets.Get("/", notImplemented)
			wallets.Post("/", notImplemented)
			wallets.Route("/{walletID}", func(wallet chi.Router) {
				wallet.Get("/", notImplemented)
				wallet.Delete("/", notImplemented)
				wallet.Post("/sync", notImplemented)
				wallet.Get("/portfolio", notImplemented)
				wallet.Get("/performance", notImplemented)
				wallet.Get("/transactions", notImplemented)
				wallet.Get("/transactions/{transactionID}", notImplemented)
			})
			return
		}

		wallets.With(middleware.Authenticate(deps.Tokens)).Group(func(secured chi.Router) {
			secured.Get("/", deps.Wallets.List)
			secured.Post("/", deps.Wallets.Create)
			secured.Route("/{walletID}", func(wallet chi.Router) {
				wallet.Get("/", deps.Wallets.Get)
				wallet.Delete("/", deps.Wallets.Delete)
				wallet.Post("/sync", deps.Wallets.RequestSync)
				if deps.Portfolio != nil {
					wallet.Get("/portfolio", deps.Portfolio.Get)
				} else {
					wallet.Get("/portfolio", notImplemented)
				}
				if deps.Performance != nil {
					wallet.Get("/performance", deps.Performance.Get)
				} else {
					wallet.Get("/performance", notImplemented)
				}
				if deps.Transactions != nil {
					wallet.Get("/transactions", deps.Transactions.List)
					wallet.Get("/transactions/{transactionID}", deps.Transactions.Get)
				} else {
					wallet.Get("/transactions", notImplemented)
					wallet.Get("/transactions/{transactionID}", notImplemented)
				}
			})
		})
	})
}

func mountAuth(r chi.Router, deps RouterDeps) {
	r.Route("/auth", func(auth chi.Router) {
		if deps.Auth == nil || deps.Tokens == nil {
			auth.Post("/guest", notImplemented)
			auth.Post("/email/register", notImplemented)
			auth.Post("/email/login", notImplemented)
			auth.Post("/google", notImplemented)
			auth.Post("/refresh", notImplemented)
			auth.Post("/upgrade", notImplemented)
			auth.Post("/logout", notImplemented)
			auth.Get("/session", notImplemented)
			return
		}

		auth.Post("/guest", deps.Auth.CreateGuest)
		auth.Post("/email/register", deps.Auth.RegisterEmail)
		auth.Post("/email/login", deps.Auth.LoginEmail)
		auth.Post("/google", deps.Auth.LoginGoogle)
		auth.Post("/refresh", deps.Auth.Refresh)

		auth.With(middleware.Authenticate(deps.Tokens)).Group(func(secured chi.Router) {
			secured.Post("/upgrade", deps.Auth.Upgrade)
			secured.Post("/logout", deps.Auth.Logout)
			secured.Get("/session", deps.Auth.Session)
		})
	})
}

// notImplemented answers routes whose vertical slice has not been built yet.
// Returning a documented error keeps the frontend on the real contract instead
// of silently receiving a 404 that looks like a routing bug.
func notImplemented(w http.ResponseWriter, r *http.Request) {
	response.Error(w, r, apperr.New(apperr.CodeNotImplemented))
}
