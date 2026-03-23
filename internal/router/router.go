// @title TorahAi API
// @version 1.0
// @description This is the TorahAi API documentation.
// @termsOfService http://swagger.io/terms/
// @contact.name TaiwoDev
// @contact.email ajasataiwo45@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /torah_ai_backend/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package router

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	// "github.com/swaggo/swag/example/basic/docs"
	"github.com/taiwoajasa245/torah_ai_backend/internal/auth"
	"github.com/taiwoajasa245/torah_ai_backend/internal/chat"
	"github.com/taiwoajasa245/torah_ai_backend/internal/database"
	"github.com/taiwoajasa245/torah_ai_backend/internal/mail"
	authMiddleware "github.com/taiwoajasa245/torah_ai_backend/internal/middleware"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/config"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"

	_ "github.com/taiwoajasa245/torah_ai_backend/docs"
)

type Router struct {
	DB   database.Service
	Mail *mail.Mailer
	Cfg  *config.Config
}

func NewRouter(db database.Service, mail *mail.Mailer, cfg *config.Config) *Router {
	return &Router{
		DB:   db,
		Mail: mail,
		Cfg:  cfg,
	}
}

func (route *Router) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(middleware.RedirectSlashes)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Get home route
	r.Get("/", route.ServerIsWorking)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+route.Cfg.Port+"/swagger/doc.json"), //The url pointing to API definition
	))

	r.Route("/torah_ai_backend/v1", func(r chi.Router) {
		route.loadAuthRoutes(r)
		route.loadChatRoutes(r)
		// route.loadVerseRoutes(r)
	})
	r.Get("/torah_ai_backend/v1", route.ServerIsWorking)

	return r
}

func (route *Router) ServerIsWorking(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Welcome to TorahAi api"
	response.Success(w, resp, "Success")
}

func (route *Router) loadAuthRoutes(router chi.Router) {

	authRepo := auth.NewRepository(route.DB.DB())
	authServie := auth.NewauthService(authRepo, route.Mail)
	authHandler := auth.NewHandler(authServie)

	router.Post("/auth/login", authHandler.LoginHandler)
	router.Post("/auth/register-with-email", authHandler.RegisterHandler)
	router.Post("/auth/forget-password", authHandler.ForgetPasswordHandler)
	router.Post("/auth/reset-password", authHandler.ResetPasswordHandler)

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Get("/auth/me", authHandler.GetUserDetailsHandler)
		// r.Post("/auth/complete-profile", authHandler.CompleteProfileHandler)
		// r.Get("/auth/verify-token", authHandler.VerifyTokenHandler)
		// r.Patch("/auth/update-profile", authHandler.UpdateUserProfileHandler)
	})
}

func (route *Router) loadChatRoutes(router chi.Router) {
	chatRepo := chat.NewRepository(route.DB.DB())
	chatService, err := chat.NewChatService(route.Cfg.GeminiAPIKey, "gemini-2.5-flash", chatRepo)
	if err != nil {
		log.Printf("Chat disabled: %v (set GEMINI_API_KEY to enable)", err)
		return
	}

	chatHandler := chat.NewHandler(chatService)

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Post("/chat", chatHandler.ChatHandler)
		r.Get("/chat", chatHandler.GetAllChatsHandler)
		r.Get("/chat/{id}", chatHandler.GetChatByIDHandler)
		r.Patch("/chat/{id}", chatHandler.UpdateChatHandler)
		r.Delete("/chat/{id}", chatHandler.DeleteChatHandler)
	})
}
