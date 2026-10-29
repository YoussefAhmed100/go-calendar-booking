package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embeds the IANA timezone database so time.LoadLocation works in minimal Docker images

	"github.com/gin-gonic/gin"

	"calendar-booking/internal/common/config"
	"calendar-booking/internal/common/database"
	"calendar-booking/internal/common/validator"
	"calendar-booking/internal/modules/appointment"
	"calendar-booking/internal/modules/auth"
	googlemodule "calendar-booking/internal/modules/google"
)

func main() {
	cfg := config.Load()

	// Database
	db := database.Connect(cfg)

	// Validator
	validator.Setup()

	// Auth module
	authModule := auth.New(db, cfg)

	if err := authModule.Migrate(); err != nil {
		log.Fatalf("auth module migration failed: %v", err)
	}

	// Google module
	googleModule := googlemodule.New(
		db,
		cfg,
		authModule.AuthenticationMiddleware(),
	)

	if err := googleModule.Migrate(); err != nil {
		log.Fatalf("failed to migrate Google module: %v", err)
	}

	// Appointment module
	appointmentModule, err := appointment.New(
		db,
		cfg,
		googleModule,
		authModule.AuthenticationMiddleware(),
	)
	if err != nil {
		log.Fatalf("failed to create appointment module: %v", err)
	}

	if err := appointmentModule.Migrate(); err != nil {
		log.Fatalf("failed to migrate appointment module: %v", err)
	}

	// HTTP router
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := router.Group("/api")

	authModule.RegisterRoutes(api)
	googleModule.RegisterRoutes(api)
	appointmentModule.RegisterRoutes(api)

	// HTTP server
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second, // above the 15s Google timeout so 502 responses can be written
		IdleTimeout:       60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("server starting on :%s", cfg.AppPort)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}