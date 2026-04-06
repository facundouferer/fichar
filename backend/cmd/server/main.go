package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/facundouferer/fichar/backend/internal/config"
	"github.com/facundouferer/fichar/backend/internal/handler"
	"github.com/facundouferer/fichar/backend/internal/middleware"
	"github.com/facundouferer/fichar/backend/internal/repository/postgres"
	"github.com/facundouferer/fichar/backend/internal/service"
	"github.com/facundouferer/fichar/backend/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Initialize repositories
	pool := db.GetPool()
	employeeRepo := postgres.NewEmployeeRepo(pool)
	shiftRepo := postgres.NewShiftRepo(pool)
	attendanceRepo := postgres.NewAttendanceRepo(pool)
	logRepo := postgres.NewLogRepo(pool)
	employeeShiftRepo := postgres.NewEmployeeShiftRepo(pool)

	// Initialize services
	employeeSvc := service.NewEmployeeService(employeeRepo)
	shiftSvc := service.NewShiftService(shiftRepo)
	attendanceSvc := service.NewAttendanceService(attendanceRepo, employeeRepo, employeeShiftRepo)
	logSvc := service.NewLogService(logRepo)
	employeeShiftSvc := service.NewEmployeeShiftService(employeeShiftRepo)
	authSvc := service.NewAuthService(employeeRepo, cfg.JWT.Secret)

	// Initialize handlers
	h := handler.NewHandler(employeeSvc, shiftSvc, attendanceSvc, logSvc, employeeShiftSvc)
	authH := handler.NewAuthHandler(authSvc)

	// Create routers
	publicMux := http.NewServeMux()
	adminMux := http.NewServeMux()
	employeeMux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	// Public routes (no auth)
	publicMux.HandleFunc("POST /api/auth/login", authH.Login)
	publicMux.HandleFunc("POST /api/attendance/check", h.CheckAttendance)

	// Protected routes (require auth, no role check)
	protectedMux.HandleFunc("POST /api/auth/change-password", authH.ChangePassword)

	// Admin routes (ADMIN role required)
	adminMux.HandleFunc("POST /api/admin/employees", h.CreateEmployee)
	adminMux.HandleFunc("GET /api/admin/employees", h.ListEmployees)
	adminMux.HandleFunc("PUT /api/admin/employees/{id}", h.UpdateEmployee)
	adminMux.HandleFunc("DELETE /api/admin/employees/{id}", h.DeleteEmployee)
	adminMux.HandleFunc("POST /api/admin/shifts", h.CreateShift)
	adminMux.HandleFunc("GET /api/admin/shifts", h.ListShifts)
	adminMux.HandleFunc("PUT /api/admin/shifts/{id}", h.UpdateShift)
	adminMux.HandleFunc("DELETE /api/admin/shifts/{id}", h.DeleteShift)
	adminMux.HandleFunc("GET /api/admin/logs", h.GetLogs)
	adminMux.HandleFunc("POST /api/admin/employee-shifts", h.AssignShift)
	adminMux.HandleFunc("GET /api/admin/employees/{id}/shifts", h.GetEmployeeShifts)

	// Employee routes (authenticated)
	employeeMux.HandleFunc("GET /api/employees/{id}", h.GetEmployee)
	employeeMux.HandleFunc("GET /api/employees/{id}/attendances", h.GetEmployeeAttendances)

	// Apply auth middleware
	authMiddleware := middleware.AuthMiddleware(authSvc)

	// Rate limiter for public endpoints (5 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(5, time.Minute)

	// Wrap with role check
	adminHandler := authMiddleware(middleware.RequireRole("ADMIN")(adminMux))
	employeeHandler := authMiddleware(employeeMux)
	protectedHandler := authMiddleware(protectedMux)

	// Custom router to handle path matching properly
	router := &Router{
		publicMux:        publicMux,
		protectedHandler: protectedHandler,
		adminHandler:     adminHandler,
		employeeHandler:  employeeHandler,
		healthHandler:    h.Health,
		rateLimiter:      rateLimiter,
	}

	// Apply CORS and logging
	handler := middleware.LoggingMiddleware(middleware.CORSMiddleware(router))

	// Server configuration
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Backend server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
	fmt.Println("Server stopped")
	os.Exit(0)
}

// Router handles path-based routing
type Router struct {
	publicMux        *http.ServeMux
	protectedHandler http.Handler
	adminHandler     http.Handler
	employeeHandler  http.Handler
	healthHandler    func(http.ResponseWriter, *http.Request)
	rateLimiter      *middleware.RateLimiter
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Health check
	if path == "/health" {
		r.healthHandler(w, req)
		return
	}

	// Rate limiting for public endpoints (login and attendance check)
	if req.Method == "POST" && (path == "/api/auth/login" || path == "/api/attendance/check") {
		// Get client IP for rate limiting
		clientIP := getClientIP(req)
		if !r.rateLimiter.Allow(clientIP) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		r.publicMux.ServeHTTP(w, req)
		return
	}

	// Public routes - no auth required
	if req.Method == "POST" && path == "/api/auth/login" {
		r.publicMux.ServeHTTP(w, req)
		return
	}

	// Protected routes - require auth but no role
	if strings.HasPrefix(path, "/api/auth/") && path != "/api/auth/login" {
		r.protectedHandler.ServeHTTP(w, req)
		return
	}

	// Admin routes - require ADMIN role
	if strings.HasPrefix(path, "/api/admin/") {
		r.adminHandler.ServeHTTP(w, req)
		return
	}

	// Employee routes - require auth
	if strings.HasPrefix(path, "/api/employees/") {
		r.employeeHandler.ServeHTTP(w, req)
		return
	}

	// Default - 404
	http.NotFound(w, req)
}

// getClientIP extracts the real client IP from request
func getClientIP(req *http.Request) string {
	// Check for forwarded header (when behind proxy)
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	// Fall back to remote address
	return req.RemoteAddr
}
