package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"microservice/internal/cache"
	"microservice/internal/config"
	"microservice/internal/handler"
)

func main() {
	cfg := config.Load()

	redisCache, err := cache.New(cfg.RedisAddr, cfg.CacheTTL)
	if err != nil {
		slog.Error("redis connection failed", "addr", cfg.RedisAddr, "err", err)
		os.Exit(1)
	}

	h := handler.New(redisCache)

	r := gin.New()
	r.Use(gin.Recovery(), requestLogger())

	r.GET("/health", h.Health)

	api := r.Group("/api", gin.BasicAuth(gin.Accounts{cfg.AuthUser: cfg.AuthPass}))
	{
		api.GET("/items", h.ListItems)
		api.GET("/items/:id", h.GetItem)
		api.POST("/items", h.CreateItem)
		api.DELETE("/items/:id", h.DeleteItem)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("server stopped")
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
		)
	}
}
