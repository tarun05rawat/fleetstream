package services

import (
	"backend/models"
	"fmt"
	"log"
	"sync"
	"time"
)

// AnomalyDetector handles fault detection and anomaly analysis
type AnomalyDetector struct {
	thresholds    *models.AnomalyThresholds
	slidingWindow map[string]*SlidingWindow
	mutex         sync.RWMutex
	alertCallback func(*models.Alert)
}

// SlidingWindow maintains recent events for a machine
type SlidingWindow struct {
	events   []*models.SensorEvent
	maxSize  int
	position int
	full     bool
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(alertCallback func(*models.Alert)) *AnomalyDetector {
	return &AnomalyDetector{
		thresholds: &models.AnomalyThresholds{
			ConveyorSpeedMin: 0.1,
			ConveyorSpeedMax: 3.5,
			TemperatureMin:   15.0,
			TemperatureMax:   85.0,
			RobotAngleMin:    0.0,
			RobotAngleMax:    180.0,
		},
		slidingWindow: make(map[string]*SlidingWindow),
		alertCallback: alertCallback,
	}
}

// NewSlidingWindow creates a new sliding window
func NewSlidingWindow(maxSize int) *SlidingWindow {
	return &SlidingWindow{
		events:  make([]*models.SensorEvent, maxSize),
		maxSize: maxSize,
	}
}

// Add adds an event to the sliding window
func (sw *SlidingWindow) Add(event *models.SensorEvent) {
	sw.events[sw.position] = event
	sw.position = (sw.position + 1) % sw.maxSize
	if !sw.full && sw.position == 0 {
		sw.full = true
	}
}

// GetEvents returns all events in the window
func (sw *SlidingWindow) GetEvents() []*models.SensorEvent {
	if !sw.full {
		return sw.events[:sw.position]
	}

	result := make([]*models.SensorEvent, sw.maxSize)
	for i := 0; i < sw.maxSize; i++ {
		idx := (sw.position + i) % sw.maxSize
		result[i] = sw.events[idx]
	}
	return result
}

// GetRecentEvents returns the N most recent events
func (sw *SlidingWindow) GetRecentEvents(n int) []*models.SensorEvent {
	events := sw.GetEvents()
	if n >= len(events) {
		return events
	}
	return events[len(events)-n:]
}

// AnalyzeEvent analyzes a sensor event for anomalies
func (ad *AnomalyDetector) AnalyzeEvent(event *models.SensorEvent) {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()

	// Get or create sliding window for this machine
	window, exists := ad.slidingWindow[event.MachineID]
	if !exists {
		window = NewSlidingWindow(50) // Keep last 50 events
		ad.slidingWindow[event.MachineID] = window
	}

	// Add event to sliding window
	window.Add(event)

	// Perform anomaly detection
	ad.detectThresholdViolations(event)
	ad.detectTrendAnomalies(event, window)
	ad.detectPatternAnomalies(event, window)
}

// detectThresholdViolations detects simple threshold violations
func (ad *AnomalyDetector) detectThresholdViolations(event *models.SensorEvent) {
	var alerts []*models.Alert

	// Check conveyor speed
	if event.ConveyorSpeed < ad.thresholds.ConveyorSpeedMin {
		alerts = append(alerts, &models.Alert{
			AlertType: "conveyor_speed_low",
			Severity:  "high",
			Message:   fmt.Sprintf("Conveyor speed below minimum threshold: %.2f m/s (min: %.2f)", event.ConveyorSpeed, ad.thresholds.ConveyorSpeedMin),
		})
	} else if event.ConveyorSpeed > ad.thresholds.ConveyorSpeedMax {
		alerts = append(alerts, &models.Alert{
			AlertType: "conveyor_speed_high",
			Severity:  "high",
			Message:   fmt.Sprintf("Conveyor speed above maximum threshold: %.2f m/s (max: %.2f)", event.ConveyorSpeed, ad.thresholds.ConveyorSpeedMax),
		})
	}

	// Check temperature
	if event.Temperature < ad.thresholds.TemperatureMin {
		alerts = append(alerts, &models.Alert{
			AlertType: "temperature_low",
			Severity:  "medium",
			Message:   fmt.Sprintf("Temperature below minimum threshold: %.1f°C (min: %.1f)", event.Temperature, ad.thresholds.TemperatureMin),
		})
	} else if event.Temperature > ad.thresholds.TemperatureMax {
		alerts = append(alerts, &models.Alert{
			AlertType: "temperature_high",
			Severity:  "high",
			Message:   fmt.Sprintf("Temperature above maximum threshold: %.1f°C (max: %.1f)", event.Temperature, ad.thresholds.TemperatureMax),
		})
	}

	// Check robot arm angle
	if event.RobotArmAngle < ad.thresholds.RobotAngleMin || event.RobotArmAngle > ad.thresholds.RobotAngleMax {
		alerts = append(alerts, &models.Alert{
			AlertType: "robot_angle_invalid",
			Severity:  "medium",
			Message:   fmt.Sprintf("Robot arm angle out of valid range: %.1f° (range: %.1f-%.1f)", event.RobotArmAngle, ad.thresholds.RobotAngleMin, ad.thresholds.RobotAngleMax),
		})
	}

	// Process status-based alerts
	if event.Status == "fault" {
		severity := "high"
		message := fmt.Sprintf("Machine fault detected: %s", event.EventType)

		if faultDesc, ok := event.AdditionalData["description"].(string); ok {
			message = fmt.Sprintf("Machine fault: %s", faultDesc)
		}

		alerts = append(alerts, &models.Alert{
			AlertType: event.EventType,
			Severity:  severity,
			Message:   message,
		})
	} else if event.Status == "warning" {
		alerts = append(alerts, &models.Alert{
			AlertType: event.EventType,
			Severity:  "medium",
			Message:   fmt.Sprintf("Warning condition: %s", event.EventType),
		})
	}

	// Send alerts
	for _, alert := range alerts {
		if ad.alertCallback != nil {
			ad.alertCallback(alert)
		}
	}
}

// detectTrendAnomalies detects anomalies based on trends
func (ad *AnomalyDetector) detectTrendAnomalies(event *models.SensorEvent, window *SlidingWindow) {
	recentEvents := window.GetRecentEvents(10)
	if len(recentEvents) < 5 {
		return // Not enough data
	}

	// Check for rapid temperature rise
	if ad.detectRapidTemperatureChange(recentEvents) {
		alert := &models.Alert{
			AlertType: "rapid_temperature_change",
			Severity:  "medium",
			Message:   fmt.Sprintf("Rapid temperature change detected on machine %s", event.MachineID),
		}
		if ad.alertCallback != nil {
			ad.alertCallback(alert)
		}
	}

	// Check for conveyor speed instability
	if ad.detectSpeedInstability(recentEvents) {
		alert := &models.Alert{
			AlertType: "speed_instability",
			Severity:  "medium",
			Message:   fmt.Sprintf("Conveyor speed instability detected on machine %s", event.MachineID),
		}
		if ad.alertCallback != nil {
			ad.alertCallback(alert)
		}
	}
}

// detectPatternAnomalies detects pattern-based anomalies
func (ad *AnomalyDetector) detectPatternAnomalies(event *models.SensorEvent, window *SlidingWindow) {
	recentEvents := window.GetRecentEvents(20)
	if len(recentEvents) < 10 {
		return
	}

	// Check for repeated faults
	faultCount := 0
	for _, e := range recentEvents {
		if e.Status == "fault" {
			faultCount++
		}
	}

	if faultCount >= 3 {
		alert := &models.Alert{
			AlertType: "repeated_faults",
			Severity:  "high",
			Message:   fmt.Sprintf("Multiple faults detected in recent history (%d faults in last 20 events)", faultCount),
		}
		if ad.alertCallback != nil {
			ad.alertCallback(alert)
		}
	}
}

// detectRapidTemperatureChange checks for rapid temperature changes
func (ad *AnomalyDetector) detectRapidTemperatureChange(events []*models.SensorEvent) bool {
	if len(events) < 5 {
		return false
	}

	// Calculate temperature change rate over last 5 events
	tempChange := events[len(events)-1].Temperature - events[len(events)-5].Temperature
	timeSpan := events[len(events)-1].Timestamp.Sub(events[len(events)-5].Timestamp).Seconds()

	if timeSpan > 0 {
		changeRate := tempChange / timeSpan
		return changeRate > 2.0 || changeRate < -2.0 // 2°C per second threshold
	}

	return false
}

// detectSpeedInstability checks for unstable conveyor speed
func (ad *AnomalyDetector) detectSpeedInstability(events []*models.SensorEvent) bool {
	if len(events) < 5 {
		return false
	}

	// Calculate standard deviation of recent speeds
	var sum, sumSquares float64
	n := float64(len(events))

	for _, event := range events {
		sum += event.ConveyorSpeed
		sumSquares += event.ConveyorSpeed * event.ConveyorSpeed
	}

	mean := sum / n
	variance := (sumSquares / n) - (mean * mean)
	stdDev := variance * variance // Simplified calculation

	// If standard deviation is high, speed is unstable
	return stdDev > 0.5
}

// UpdateThresholds updates the anomaly detection thresholds
func (ad *AnomalyDetector) UpdateThresholds(thresholds *models.AnomalyThresholds) {
	ad.mutex.Lock()
	defer ad.mutex.Unlock()
	ad.thresholds = thresholds
	log.Printf("Updated anomaly detection thresholds: %+v", thresholds)
}

// GetThresholds returns current thresholds
func (ad *AnomalyDetector) GetThresholds() *models.AnomalyThresholds {
	ad.mutex.RLock()
	defer ad.mutex.RUnlock()
	return ad.thresholds
}

// GetMachineStats returns statistics for a specific machine
func (ad *AnomalyDetector) GetMachineStats(machineID string) map[string]interface{} {
	ad.mutex.RLock()
	defer ad.mutex.RUnlock()

	window, exists := ad.slidingWindow[machineID]
	if !exists {
		return nil
	}

	events := window.GetEvents()
	if len(events) == 0 {
		return nil
	}

	var avgTemp, avgSpeed, avgAngle float64
	faultCount := 0

	for _, event := range events {
		avgTemp += event.Temperature
		avgSpeed += event.ConveyorSpeed
		avgAngle += event.RobotArmAngle
		if event.Status == "fault" {
			faultCount++
		}
	}

	n := float64(len(events))
	return map[string]interface{}{
		"event_count":          len(events),
		"avg_temperature":      avgTemp / n,
		"avg_conveyor_speed":   avgSpeed / n,
		"avg_robot_arm_angle":  avgAngle / n,
		"fault_rate":           float64(faultCount) / n,
		"last_event_time":      events[len(events)-1].Timestamp,
	}
}