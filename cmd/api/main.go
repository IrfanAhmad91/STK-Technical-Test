package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stk/menu-tree-api/config"
	_ "github.com/stk/menu-tree-api/docs" // Import generated docs
	"github.com/stk/menu-tree-api/internal/handler"
	"github.com/stk/menu-tree-api/internal/repository"
	"github.com/stk/menu-tree-api/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Hierarchical Menu Tree API
// @version 1.0
// @description RESTful API for managing hierarchical menu structures with unlimited depth
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://github.com/stk/menu-tree-api
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load database configuration
	dbConfig := config.LoadDatabaseConfig()

	// Establish database connection
	db, err := config.NewDatabaseConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL database!")

	// Initialize repository
	menuRepo := repository.NewPostgreSQLRepository(db)

	// Initialize service
	menuService := service.NewMenuService(menuRepo)

	// Initialize handler
	menuHandler := handler.NewMenuHandler(menuService)

	// Create Gin router (Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1)
	router := gin.Default()

	// Add CORS middleware for frontend integration (Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, specify exact origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add request logging middleware (Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1)
	router.Use(requestLoggerMiddleware())

	// Panic recovery middleware with structured error responses (Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1)
	router.Use(gin.CustomRecovery(panicRecoveryMiddleware))

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"message": "Menu Tree API is running",
			})
		})

		// Menu endpoints (Requirements 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1)
		menus := v1.Group("/menus")
		{
			menus.GET("", menuHandler.GetAll)              // GET /api/v1/menus
			menus.GET("/:id", menuHandler.GetByID)         // GET /api/v1/menus/:id
			menus.POST("", menuHandler.Create)             // POST /api/v1/menus
			menus.PUT("/:id", menuHandler.Update)          // PUT /api/v1/menus/:id
			menus.DELETE("/:id", menuHandler.Delete)       // DELETE /api/v1/menus/:id
			menus.PUT("/:id/reorder", menuHandler.Reorder) // PUT /api/v1/menus/:id/reorder
			menus.PUT("/:id/move", menuHandler.Move)       // PUT /api/v1/menus/:id/move
		}
	}

	// Register Swagger UI handler (Requirement 6.1, 6.2, 6.3, 6.4, 6.5)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	fmt.Printf("\nServer starting on port %s...\n", port)
	fmt.Printf("API endpoints: http://localhost:%s/api/v1/menus\n", port)
	fmt.Printf("Swagger documentation: http://localhost:%s/swagger/index.html\n", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// requestLoggerMiddleware logs request details including request ID, timestamp, and duration
// Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Record start time
		startTime := time.Now()

		// Log incoming request
		log.Printf("[%s] --> %s %s", requestID, c.Request.Method, c.Request.URL.Path)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log completed request with status and duration
		log.Printf("[%s] <-- %s %s | Status: %d | Duration: %v",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// panicRecoveryMiddleware handles panics and returns structured error responses
// Requirement 1.1, 1.3, 1.4, 1.5, 1.6, 3.1, 4.1
func panicRecoveryMiddleware(c *gin.Context, err interface{}) {
	// Get request ID from context
	requestID, exists := c.Get("request_id")
	if !exists {
		requestID = "unknown"
	}

	// Log panic with stack trace
	log.Printf("[%s] PANIC RECOVERED: %v", requestID, err)

	// Return structured error response
	c.JSON(500, gin.H{
		"code":       "INTERNAL_ERROR",
		"message":    "An internal server error occurred",
		"request_id": requestID,
	})
}
