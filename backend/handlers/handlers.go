package handlers

import (
	"backend/database"
	"backend/models"
	"backend/services"
	"backend/websocket"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler contains all the dependencies needed for HTTP handlers
type Handler struct {
	db               *database.DB
	hub              *websocket.Hub
	anomalyDetector  *services.AnomalyDetector
}

// New creates a new handler instance
func New(db *database.DB, hub *websocket.Hub, anomalyDetector *services.AnomalyDetector) *Handler {
	return &Handler{
		db:              db,
		hub:             hub,
		anomalyDetector: anomalyDetector,
	}
}

// GetEvents retrieves recent events with pagination
func (h *Handler) GetEvents(c *gin.Context) {
	limit := 50 // default
	offset := 0 // default
	machineID := c.Query("machine_id")

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	events, err := h.db.GetRecentEvents(limit, offset, machineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(events),
		},
	})
}

// GetEventStats retrieves event statistics
func (h *Handler) GetEventStats(c *gin.Context) {
	machineID := c.Query("machine_id")
	sinceParam := c.DefaultQuery("since", "24h")

	// Parse since parameter
	var since time.Time
	switch sinceParam {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		if duration, err := time.ParseDuration(sinceParam); err == nil {
			since = time.Now().Add(-duration)
		} else {
			since = time.Now().Add(-24 * time.Hour)
		}
	}

	stats, err := h.db.GetEventStats(machineID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve event statistics",
			"details": err.Error(),
		})
		return
	}

	// Add real-time statistics from anomaly detector
	if machineID != "" {
		if machineStats := h.anomalyDetector.GetMachineStats(machineID); machineStats != nil {
			stats.AvgTemperature = machineStats["avg_temperature"].(float64)
			stats.AvgConveyorSpeed = machineStats["avg_conveyor_speed"].(float64)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"period": gin.H{
			"since":    since.Format(time.RFC3339),
			"duration": sinceParam,
		},
	})
}

// GetAlerts retrieves unacknowledged alerts
func (h *Handler) GetAlerts(c *gin.Context) {
	alerts, err := h.db.GetUnacknowledgedAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// AcknowledgeAlert acknowledges a specific alert
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	alertIDParam := c.Param("id")
	alertID, err := strconv.Atoi(alertIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	err = h.db.AcknowledgeAlert(alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to acknowledge alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Alert acknowledged successfully",
		"alert_id": alertID,
	})
}

// GetProcessParameters retrieves all process parameters
func (h *Handler) GetProcessParameters(c *gin.Context) {
	params, err := h.db.GetProcessParameters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve process parameters",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"parameters": params,
		"count":      len(params),
	})
}

// UpdateProcessParameter updates a specific process parameter
func (h *Handler) UpdateProcessParameter(c *gin.Context) {
	var updateRequest struct {
		ParameterName  string `json:"parameter_name" binding:"required"`
		ParameterValue string `json:"parameter_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.db.UpdateProcessParameter(updateRequest.ParameterName, updateRequest.ParameterValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update process parameter",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Process parameter updated successfully",
		"parameter_name":  updateRequest.ParameterName,
		"parameter_value": updateRequest.ParameterValue,
	})
}

// GetMachines retrieves all machines
func (h *Handler) GetMachines(c *gin.Context) {
	machines, err := h.db.GetMachines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve machines",
			"details": err.Error(),
		})
		return
	}

	// Enhance with real-time statistics
	for i := range machines {
		if stats := h.anomalyDetector.GetMachineStats(machines[i].MachineID); stats != nil {
			machines[i].Config["real_time_stats"] = stats
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"machines": machines,
		"count":    len(machines),
	})
}

// GetSystemHealth returns overall system health information
func (h *Handler) GetSystemHealth(c *gin.Context) {
	stats, _ := h.db.GetEventStats("", time.Now().Add(-1*time.Hour))

	health := gin.H{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"websocket": gin.H{
			"connected_clients": h.hub.GetClientCount(),
		},
		"database": gin.H{
			"status": "connected",
		},
		"recent_activity": gin.H{
			"total_events_1h":  stats.TotalEvents,
			"fault_events_1h":  stats.FaultEvents,
			"warning_events_1h": stats.WarningEvents,
			"uptime_percent":   stats.UptimePercent,
		},
		"thresholds": h.anomalyDetector.GetThresholds(),
	}

	// Determine overall health status
	if stats.UptimePercent < 95.0 {
		health["status"] = "degraded"
	}
	if stats.UptimePercent < 90.0 {
		health["status"] = "unhealthy"
	}

	c.JSON(http.StatusOK, health)
}

// UpdateAnomalyThresholds updates anomaly detection thresholds
func (h *Handler) UpdateAnomalyThresholds(c *gin.Context) {
	var thresholds models.AnomalyThresholds
	if err := c.ShouldBindJSON(&thresholds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid threshold data",
			"details": err.Error(),
		})
		return
	}

	// Validate thresholds
	if thresholds.ConveyorSpeedMin < 0 || thresholds.ConveyorSpeedMax <= thresholds.ConveyorSpeedMin {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid conveyor speed thresholds",
		})
		return
	}

	if thresholds.TemperatureMax <= thresholds.TemperatureMin {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid temperature thresholds",
		})
		return
	}

	h.anomalyDetector.UpdateThresholds(&thresholds)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Anomaly thresholds updated successfully",
		"thresholds": thresholds,
	})
}

// GetAnomalyThresholds retrieves current anomaly detection thresholds
func (h *Handler) GetAnomalyThresholds(c *gin.Context) {
	thresholds := h.anomalyDetector.GetThresholds()
	c.JSON(http.StatusOK, gin.H{
		"thresholds": thresholds,
	})
}

// WebSocketEndpoint handles WebSocket connections
func (h *Handler) WebSocketEndpoint(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}