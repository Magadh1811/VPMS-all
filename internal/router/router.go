package router

import (
	"database/sql"
	"time"

	"Backend-Go/internal/config"
	"Backend-Go/internal/handlers"
	"Backend-Go/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(db *sql.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS for local frontend
	c := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(c))

	h := handler.New(db, cfg)

	// Health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// DB check
	r.GET("/dbcheck", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "failed", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "message": "Connected to database"})
	})

	api := r.Group("/")

	// Auth
	auth := api.Group("/auth")
	{
		auth.POST("/signup", h.Signup)
		auth.POST("/login", h.Login)
	}

	// User
	user := api.Group("/")
	user.Use(middleware.AuthJWT(cfg))
	{
		user.GET("/user/me", h.Me)
		user.POST("/vehicles", h.AddVehicle)
		user.POST("/parking/book", h.BookSpot)
		user.POST("/parking/release/:spotId", h.Release)
		user.GET("/parking/history", h.UserHistory)
	}

	// Admin
	admin := api.Group("/")
	admin.Use(middleware.AuthJWT(cfg), middleware.RequireRole("admin"))
	{
		admin.POST("/parking-lots", h.CreateLot)
		admin.POST("/parking-spots", h.CreateSpot)
		admin.DELETE("/parking-spots/:id", h.DeleteSpot)
		admin.GET("/parking/occupancy", h.Occupancy)
		admin.GET("/parking/reports", h.Reports)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "route not found"}})
	})

	return r
}