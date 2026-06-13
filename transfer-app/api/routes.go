package api

import (
	"net/http"
	"transfer-app/internal/auth"
	"transfer-app/internal/notification"
	"transfer-app/internal/transfer"
	"transfer-app/internal/user"
	"transfer-app/internal/wallet"
	"transfer-app/internal/webhook"
	"transfer-app/pkg/otp"
	"transfer-app/pkg/paystack"
	"transfer-app/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() http.Handler {
	// Clients
	paystackClient := paystack.NewClient()
	otpClient := otp.NewClient()

	// Services
	userSvc := user.NewService()
	notifSvc := notification.NewService(otpClient)
	walletSvc := wallet.NewService(paystackClient)
	authSvc := auth.NewService(otpClient)
	transferSvc := transfer.NewService(walletSvc, userSvc, notifSvc, paystackClient)
	webhookSvc := webhook.NewService(paystackClient, walletSvc, userSvc, notifSvc)

	// Handlers
	authHandler := auth.NewHandler(authSvc)
	userHandler := user.NewHandler(userSvc)
	walletHandler := wallet.NewHandler(walletSvc)
	transferHandler := transfer.NewHandler(transferSvc)
	webhookHandler := webhook.NewHandler(webhookSvc)

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "NairaTransfer API is running", map[string]string{"status": "ok"})
	})

	// Webhook (no auth)
	r.Post("/webhook/paystack", webhookHandler.Paystack)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/send-otp", authHandler.SendOTP)
			r.Post("/verify-otp", authHandler.VerifyOTP)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh-token", authHandler.RefreshToken)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.JWTMiddleware)

			// User
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userHandler.GetProfile)
				r.Post("/set-pin", userHandler.SetPin)
				r.Put("/change-password", userHandler.ChangePassword)
			})

			// Wallet
			r.Route("/wallet", func(r chi.Router) {
				r.Get("/balance", walletHandler.GetBalance)
				r.Post("/fund", walletHandler.FundWallet)
				r.Get("/banks", walletHandler.ListBanks)
				r.Post("/resolve-account", walletHandler.ResolveAccount)
			})

			// Beneficiaries
			r.Route("/beneficiaries", func(r chi.Router) {
				r.Get("/", walletHandler.ListBeneficiaries)
				r.Post("/", walletHandler.AddBeneficiary)
				r.Delete("/{id}", walletHandler.DeleteBeneficiary)
			})

			// Transfers
			r.Route("/transfers", func(r chi.Router) {
				r.Post("/p2p", transferHandler.P2PTransfer)
				r.Post("/bank", transferHandler.BankTransfer)
				r.Get("/", transferHandler.GetTransactions)
				r.Get("/single", transferHandler.GetTransaction)
			})
		})
	})

	return r
}
