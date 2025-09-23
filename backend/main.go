package main

import (
	"backend/config"
	"backend/database"
	"backend/handlers"
	"backend/kafka"
	"backend/models"
	"backend/services"
	"backend/websocket"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting FactoryFlow Backend Server on port %s", cfg.Server.Port)

	// Initialize database
	db, err := database.New(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	log.Println("WebSocket hub started")

	// Initialize anomaly detector with alert callback
	alertCallback := func(alert *models.Alert) {
		// Store alert in database
		if err := db.InsertAlert(alert); err != nil {
			log.Printf("Failed to store alert: %v", err)
		} else {
			log.Printf("Alert created: %s - %s", alert.AlertType, alert.Message)
		}

		// Broadcast alert to WebSocket clients
		wsHub.BroadcastAlert(alert)
	}

	anomalyDetector := services.NewAnomalyDetector(alertCallback)

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topics)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer consumer.Stop()

	log.Printf("Kafka consumer initialized, topics: %v", cfg.Kafka.Topics)

	// Start Kafka consumer
	consumer.Start()

	// Process events from Kafka
	go func() {
		for {
			select {
			case event := <-consumer.EventChannel():
				if event != nil {
					// Store event in database
					dbEvent, err := db.InsertEvent(event)
					if err != nil {
						log.Printf("Failed to store event: %v", err)
						continue
					}

					// Analyze for anomalies
					anomalyDetector.AnalyzeEvent(event)

					// Broadcast to WebSocket clients
					wsHub.BroadcastEvent(event)

					log.Printf("Event processed: ID=%d, Machine=%s, Status=%s",
						dbEvent.ID, event.MachineID, event.Status)
				}

			case err := <-consumer.ErrorChannel():
				log.Printf("Kafka consumer error: %v", err)
			}
		}
	}()

	// Periodic statistics broadcast
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats, err := db.GetEventStats("", time.Now().Add(-1*time.Hour))
			if err != nil {
				log.Printf("Failed to get stats: %v", err)
				continue
			}

			wsHub.BroadcastStats(map[string]interface{}{
				"system_stats":      stats,
				"connected_clients": wsHub.GetClientCount(),
				"timestamp":         time.Now(),
			})
		}
	}()

	// Initialize HTTP handlers
	handler := handlers.New(db, wsHub, anomalyDetector)

	// Setup Gin router
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.AllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Events
		api.GET("/events", handler.GetEvents)
		api.GET("/events/stats", handler.GetEventStats)

		// Alerts
		api.GET("/alerts", handler.GetAlerts)
		api.PUT("/alerts/:id/acknowledge", handler.AcknowledgeAlert)

		// Process parameters
		api.GET("/parameters", handler.GetProcessParameters)
		api.PUT("/parameters", handler.UpdateProcessParameter)

		// Machines
		api.GET("/machines", handler.GetMachines)

		// System health
		api.GET("/system/health", handler.GetSystemHealth)

		// Anomaly detection
		api.GET("/anomaly/thresholds", handler.GetAnomalyThresholds)
		api.PUT("/anomaly/thresholds", handler.UpdateAnomalyThresholds)
	}

	// WebSocket endpoint
	router.GET("/ws", handler.WebSocketEndpoint)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}