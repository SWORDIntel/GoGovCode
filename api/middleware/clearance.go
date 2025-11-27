package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NSACodeGov/CodeGov/internal/audit"
	"github.com/NSACodeGov/CodeGov/internal/logging"
	"github.com/NSACodeGov/CodeGov/internal/policy"
	"github.com/NSACodeGov/CodeGov/pkg/models"
)

// Context keys for clearance data
type clearanceKey string

const (
	ClearanceKey clearanceKey = "clearance"
	DeviceKey    clearanceKey = "device"
)

// ClearanceConfig holds configuration for clearance middleware
type ClearanceConfig struct {
	PolicyEngine   *policy.Engine
	AuditLogger    *audit.Logger
	Logger         *logging.Logger
	DeviceRegistry *models.DeviceRegistry
	Enabled        bool
}

// Clearance middleware extracts and validates clearance information
func Clearance(config *ClearanceConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Extract clearance data from headers
			deviceIDStr := r.Header.Get("X-Device-ID")
			layerStr := r.Header.Get("X-Layer")
			clearanceStr := r.Header.Get("X-Clearance")
			tokenIDStr := r.Header.Get("X-Token-ID")

			// Parse device ID
			var deviceID uint16
			if deviceIDStr != "" {
				id, err := strconv.ParseUint(deviceIDStr, 10, 16)
				if err != nil {
					config.Logger.WarnContext(r.Context(), "invalid device ID", map[string]interface{}{
						"device_id": deviceIDStr,
						"error":     err.Error(),
					})
					respondUnauthorized(w, r, config, "invalid device ID")
					return
				}
				deviceID = uint16(id)
			}

			// Parse clearance
			var clearance models.Clearance
			if clearanceStr != "" {
				// Support both hex (0x03030303) and decimal formats
				clearanceStr = strings.TrimPrefix(clearanceStr, "0x")
				clearanceStr = strings.TrimPrefix(clearanceStr, "0X")

				c, err := strconv.ParseUint(clearanceStr, 16, 32)
				if err != nil {
					config.Logger.WarnContext(r.Context(), "invalid clearance", map[string]interface{}{
						"clearance": clearanceStr,
						"error":      err.Error(),
					})
					respondUnauthorized(w, r, config, "invalid clearance format")
					return
				}
				clearance = models.Clearance(c)

				if !models.ValidateClearance(clearance) {
					respondUnauthorized(w, r, config, "invalid clearance level")
					return
				}
			}

			// Parse layer
			layer := models.Layer(layerStr)
			if layerStr != "" {
				// Validate layer
				validLayers := map[models.Layer]bool{
					models.LayerData:        true,
					models.LayerTransport:   true,
					models.LayerControl:     true,
					models.LayerApplication: true,
				}
				if !validLayers[layer] {
					respondUnauthorized(w, r, config, "invalid layer")
					return
				}
			}

			// Parse token ID (optional)
			var tokenID uint16
			var tokenOffset models.TokenOffset
			if tokenIDStr != "" {
				id, err := strconv.ParseUint(tokenIDStr, 10, 16)
				if err != nil {
					respondUnauthorized(w, r, config, "invalid token ID")
					return
				}
				tokenID = uint16(id)

				// Look up device by token
				if config.DeviceRegistry != nil {
					device, offset, err := config.DeviceRegistry.GetDeviceByToken(tokenID)
					if err == nil {
						deviceID = device.ID
						layer = device.Layer
						clearance = device.Clearance
						tokenOffset = offset
					}
				}
			}

			// Get device info if registry is available
			var device *models.Device
			if deviceID > 0 && config.DeviceRegistry != nil {
				var err error
				device, err = config.DeviceRegistry.GetDevice(deviceID)
				if err != nil {
					config.Logger.WarnContext(r.Context(), "device not found", map[string]interface{}{
						"device_id": deviceID,
					})
					respondUnauthorized(w, r, config, "device not registered")
					return
				}

				// Use device's clearance if not explicitly provided
				if clearance == 0 {
					clearance = device.Clearance
				}
				if layer == "" {
					layer = device.Layer
				}
			}

			// Add clearance info to context
			ctx := r.Context()
			if clearance > 0 {
				ctx = context.WithValue(ctx, ClearanceKey, clearance)
			}
			if device != nil {
				ctx = context.WithValue(ctx, DeviceKey, device)
			}
			if deviceID > 0 {
				ctx = logging.WithDeviceID(ctx, fmt.Sprintf("%d", deviceID))
			}
			if layer != "" {
				ctx = logging.WithLayer(ctx, string(layer))
			}

			// Evaluate policy
			if config.PolicyEngine != nil {
				policyCtx := &policy.Context{
					Route:       r.URL.Path,
					Method:      r.Method,
					DeviceID:    deviceID,
					Layer:       layer,
					Clearance:   clearance,
					RequestID:   logging.GetRequestID(ctx),
					SourceIP:    r.RemoteAddr,
					TokenID:     tokenID,
					TokenOffset: tokenOffset,
				}

				decision := config.PolicyEngine.Evaluate(policyCtx)

				// Log audit event
				if config.AuditLogger != nil {
					auditEvent := &audit.AuditEvent{
						Actor:      fmt.Sprintf("device-%d", deviceID),
						Clearance:  clearance,
						DeviceID:   deviceID,
						Layer:      layer,
						Action:     r.URL.Path,
						Method:     r.Method,
						Resource:   r.URL.String(),
						RequestID:  logging.GetRequestID(ctx),
						SourceIP:   r.RemoteAddr,
						StatusCode: 0, // Will be set later
					}

					if decision.Effect == policy.EffectAllow {
						auditEvent.Decision = audit.DecisionAllow
						auditEvent.Reason = decision.Reason
					} else {
						auditEvent.Decision = audit.DecisionDeny
						auditEvent.Reason = decision.Reason
						auditEvent.StatusCode = http.StatusForbidden
					}

					config.AuditLogger.Log(auditEvent)
				}

				// Enforce policy decision
				if decision.Effect == policy.EffectDeny {
					config.Logger.WarnContext(ctx, "access denied by policy", map[string]interface{}{
						"rule":      decision.RuleID,
						"reason":    decision.Reason,
						"device_id": deviceID,
						"clearance": clearance,
						"route":     r.URL.Path,
					})

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":  "access denied",
						"reason": decision.Reason,
					})
					return
				}
			}

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// respondUnauthorized sends an unauthorized response
func respondUnauthorized(w http.ResponseWriter, r *http.Request, config *ClearanceConfig, reason string) {
	if config.AuditLogger != nil {
		event := &audit.AuditEvent{
			Actor:      "unknown",
			Action:     r.URL.Path,
			Method:     r.Method,
			Resource:   r.URL.String(),
			Decision:   audit.DecisionDeny,
			Reason:     reason,
			RequestID:  logging.GetRequestID(r.Context()),
			SourceIP:   r.RemoteAddr,
			StatusCode: http.StatusUnauthorized,
		}
		config.AuditLogger.Log(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  "unauthorized",
		"reason": reason,
	})
}

// GetClearance retrieves clearance from context
func GetClearance(ctx context.Context) (models.Clearance, bool) {
	clearance, ok := ctx.Value(ClearanceKey).(models.Clearance)
	return clearance, ok
}

// GetDevice retrieves device from context
func GetDevice(ctx context.Context) (*models.Device, bool) {
	device, ok := ctx.Value(DeviceKey).(*models.Device)
	return device, ok
}
