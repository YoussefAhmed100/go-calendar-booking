package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"calendar-booking/internal/common/config"
	"calendar-booking/internal/common/database"
	"calendar-booking/internal/common/response"
	"calendar-booking/internal/common/validator"
	"calendar-booking/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

// TEMPORARY: only to verify validation works. Removed in Step 3.
type pingRequest struct {
	Name  string `json:"name" binding:"required,min=3,max=50"`
	Email string `json:"email" binding:"required,email"`
}

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	authModule := auth.New(db)

	if err := authModule.Migrate(); err != nil {
		log.Fatalf("failed to migrate auth module: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	validator.Setup()

	r := gin.Default()
	// Needed later so rate limiting by client IP can't be spoofed via headers.
	_ = r.SetTrustedProxies(nil)

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "up"})
	})

	r.POST("/api/ping", func(c *gin.Context) {
		var req pingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BindError(c, err)
			return
		}
		response.JSON(c, http.StatusOK, gin.H{"message": "pong", "name": req.Name})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server running on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
}
