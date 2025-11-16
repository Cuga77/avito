package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"avito/internal/config"
	"avito/internal/handler"
	"avito/internal/repository/postgres"
	"avito/internal/service"
	"avito/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	appLogger.Info("Starting application",
		"app", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.Server.Port,
	)

	dbConfig := postgres.Config{
		URL:            cfg.Database.URL,
		MaxConnections: cfg.Database.MaxConnections,
		MaxIdle:        cfg.Database.MaxIdle,
		ConnLifetime:   cfg.Database.ConnLifetime,
	}

	db, err := postgres.NewDB(dbConfig)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	appLogger.Info("Successfully connected to database")

	teamRepo := postgres.NewTeamRepository(db.DB)
	userRepo := postgres.NewUserRepository(db.DB)
	prRepo := postgres.NewPullRequestRepository(db.DB)
	taskRepo := postgres.NewTaskRepository(db.DB)
	appLogger.Info("Repository layer initialized")

	teamService := service.NewTeamService(db, teamRepo, userRepo)
	prService := service.NewPRService(db, prRepo, userRepo, teamRepo)
	userService := service.NewUserService(userRepo, prRepo, prService, teamRepo, taskRepo, appLogger)
	statsService := service.NewStatsService(prRepo, userRepo, teamRepo, appLogger)

	taskWorker := service.NewTaskWorker(taskRepo, userRepo, prService, appLogger)
	appLogger.Info("Service layer initialized")

	h := handler.NewHandler(teamService, userService, prService, appLogger.Logger)
	statsHandler := handler.NewStatsHandler(statsService, appLogger)
	appLogger.Info("Handler layer initialized")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggingMiddleware(appLogger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"pr-reviewer"}`)); err != nil {
			appLogger.Error("Failed to write health response", "error", err)
		}
	})

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", h.CreateTeam)
		r.Get("/get", h.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.SetIsActive)
		r.Get("/getReview", h.GetPRsByReviewer)
		r.Post("/batchDeactivate", h.BatchDeactivate)
		r.Get("/{user_id}", h.GetUser)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.CreatePR)
		r.Post("/merge", h.MergePR)
		r.Post("/reassign", h.ReassignReviewer)
	})

	r.Route("/stats", func(r chi.Router) {
		r.Get("/team", statsHandler.GetTeamStats)
		r.Get("/user", statsHandler.GetUserStats)
		r.Get("/global", statsHandler.GetGlobalStats)
		r.Get("/workload", statsHandler.GetWorkloadStats)
		r.Get("/health", statsHandler.GetHealthStats)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go taskWorker.Run(ctx)

	go func() {
		appLogger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	<-ctx.Done()

	appLogger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server stopped gracefully")
}

func loggingMiddleware(logger *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
