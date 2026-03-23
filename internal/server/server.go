package server

import (
	"context"
	"fmt"
	"log"

	"net/http"

	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/taiwoajasa245/torah_ai_backend/internal/database"
	"github.com/taiwoajasa245/torah_ai_backend/internal/mail"
	"github.com/taiwoajasa245/torah_ai_backend/internal/router"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/config"
)

type Server struct {
	Port    string
	DB      database.Service
	Handler http.Handler
	Cfg     *config.Config
	Mail    *mail.Mailer
	Cancel  context.CancelFunc
}

// NewServer constructs your app server with all dependencies injected.
func NewServer(db database.Service, cfg *config.Config) *Server {
	stats := db.Health()
	mail := mail.NewMail(
		cfg.SmtpFrom,
		"Memory Verse",
		cfg.SmtpPassword,
		cfg.SmtpHost,
		cfg.SmtpPort,
	)

	fmt.Println("Database Health:", stats)

	if stats["status"] != "up" {
		log.Fatal("Database connection failed")
		return &Server{}
	} else {
		log.Println("Database connection successful")
	}

	s := &Server{
		Port: cfg.Port,
		DB:   db,
		Cfg:  cfg,
		Mail: mail,
		// mvService: mvService,
	}

	r := router.NewRouter(db, mail, cfg)
	s.Handler = r.RegisterRoutes()

	return s
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%s", s.Port),
		Handler:      s.Handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
