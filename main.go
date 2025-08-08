package main

import (
	"context"
	_ "efectz/docs"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/swaggo/echo-swagger"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api/v1

type Config struct {
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

type Subscription struct {
	ID          string    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      string    `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SubscriptionService struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewSubscriptionService(db *pgxpool.Pool, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{db: db, logger: logger}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body Subscription true "Subscription data"
// @Success 201 {object} Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (s *SubscriptionService) CreateSubscription(c echo.Context) error {
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		s.logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
		time.Now(), time.Now()).Scan(&sub.ID)
	if err != nil {
		s.logger.Error("failed to create subscription", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create subscription"})
	}

	s.logger.Info("subscription created", "id", sub.ID)
	return c.JSON(http.StatusCreated, sub)
}

// GetSubscription godoc
// @Summary Get a subscription
// @Description Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} Subscription
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (s *SubscriptionService) GetSubscription(c echo.Context) error {
	id := c.Param("id")
	var sub Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE id = $1`
	err := s.db.QueryRow(context.Background(), query, id).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		s.logger.Error("failed to get subscription", "id", id, "error", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "subscription not found"})
	}

	s.logger.Info("subscription retrieved", "id", id)
	return c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Update a subscription
// @Description Update subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body Subscription true "Subscription data"
// @Success 200 {object} Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (s *SubscriptionService) UpdateSubscription(c echo.Context) error {
	id := c.Param("id")
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		s.logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	query := `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5, 
              updated_at = $6 WHERE id = $7 RETURNING id`
	err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
		time.Now(), id).Scan(&sub.ID)
	if err != nil {
		s.logger.Error("failed to update subscription", "id", id, "error", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "subscription not found"})
	}

	s.logger.Info("subscription updated", "id", id)
	return c.JSON(http.StatusOK, sub)
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (s *SubscriptionService) DeleteSubscription(c echo.Context) error {
	id := c.Param("id")
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		s.logger.Error("failed to delete subscription", "id", id, "error", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "subscription not found"})
	}
	if result.RowsAffected() == 0 {
		s.logger.Error("subscription not found", "id", id)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "subscription not found"})
	}

	s.logger.Info("subscription deleted", "id", id)
	return c.NoContent(http.StatusNoContent)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description List all subscriptions with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Service name filter"
// @Success 200 {array} Subscription
// @Router /subscriptions [get]
func (s *SubscriptionService) ListSubscriptions(c echo.Context) error {
	userID := c.QueryParam("user_id")
	serviceName := c.QueryParam("service_name")
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	if userID != "" {
		query += " AND user_id = $1"
		args = append(args, userID)
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", len(args)+1)
		args = append(args, serviceName)
	}

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		s.logger.Error("failed to list subscriptions", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list subscriptions"})
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate,
			&sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			s.logger.Error("failed to scan subscription", "error", err)
			continue
		}
		subscriptions = append(subscriptions, sub)
	}

	s.logger.Info("subscriptions listed", "count", len(subscriptions))
	return c.JSON(http.StatusOK, subscriptions)
}

// CalculateTotal godoc
// @Summary Calculate total subscription cost
// @Description Calculate total cost for subscriptions in a period
// @Tags subscriptions
// @Produce json
// @Param start_date query string true "Start date (MM-YYYY)"
// @Param end_date query string true "End date (MM-YYYY)"
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Service name filter"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Router /subscriptions/total [get]
func (s *SubscriptionService) CalculateTotal(c echo.Context) error {
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")
	userID := c.QueryParam("user_id")
	serviceName := c.QueryParam("service_name")

	query := `SELECT SUM(price) FROM subscriptions WHERE start_date <= $1 AND (end_date IS NULL OR end_date >= $2)`
	args := []interface{}{endDate, startDate}
	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", len(args)+1)
		args = append(args, userID)
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", len(args)+1)
		args = append(args, serviceName)
	}

	var total int
	err := s.db.QueryRow(context.Background(), query, args...).Scan(&total)
	if err != nil {
		s.logger.Error("failed to calculate total", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to calculate total"})
	}

	s.logger.Info("total calculated", "total", total)
	return c.JSON(http.StatusOK, map[string]int{"total": total})
}

func main() {
	// Load configuration
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		slog.Error("Error reading config.yaml", "error", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(configFile, &cfg); err != nil {
		slog.Error("Error unmarshaling config", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Initialize database
	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${remote_ip} | ${method} | ${uri} | ${status} | ${error}\n",
	}))
	e.Use(middleware.Recover())

	// Initialize service
	service := NewSubscriptionService(db, logger)

	// Routes
	v1 := e.Group("/api/v1")
	v1.POST("/subscriptions", service.CreateSubscription)
	v1.GET("/subscriptions/:id", service.GetSubscription)
	v1.PUT("/subscriptions/:id", service.UpdateSubscription)
	v1.DELETE("/subscriptions/:id", service.DeleteSubscription)
	v1.GET("/subscriptions", service.ListSubscriptions)
	v1.GET("/subscriptions/total", service.CalculateTotal)
	e.GET("/swagger/*", func(c echo.Context) error {
		err := echoSwagger.WrapHandler(c)
		if err != nil {
			c.Logger().Error("Swagger error: ", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Swagger failed")
		}
		return nil
	})

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
}
