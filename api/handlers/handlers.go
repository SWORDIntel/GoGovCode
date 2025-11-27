package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NSACodeGov/CodeGov/api/middleware"
	"github.com/NSACodeGov/CodeGov/internal/logging"
	"github.com/NSACodeGov/CodeGov/pkg/models"
)

// PublicHandler handles public endpoints (no auth required)
func PublicHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message": "This is a public endpoint",
			"access":  "unrestricted",
		}

		json.NewEncoder(w).Encode(response)
	}
}

// RestrictedHandler handles restricted endpoints (clearance required)
func RestrictedHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearance, hasClearance := middleware.GetClearance(r.Context())
		device, hasDevice := middleware.GetDevice(r.Context())

		response := map[string]interface{}{
			"message": "This is a restricted endpoint",
			"access":  "granted",
		}

		if hasClearance {
			response["clearance"] = clearance.String()
		}

		if hasDevice {
			response["device"] = map[string]interface{}{
				"id":    device.ID,
				"name":  device.Name,
				"layer": device.Layer,
				"class": device.Class,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// DeviceOnlyHandler handles device-only endpoints
func DeviceOnlyHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device, hasDevice := middleware.GetDevice(r.Context())

		if !hasDevice {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  "forbidden",
				"reason": "device registration required",
			})
			return
		}

		response := map[string]interface{}{
			"message": "This is a device-only endpoint",
			"device": map[string]interface{}{
				"id":           device.ID,
				"name":         device.Name,
				"layer":        device.Layer,
				"class":        device.Class,
				"status_token": fmt.Sprintf("0x%04X", device.GetStatusToken()),
				"config_token": fmt.Sprintf("0x%04X", device.GetConfigToken()),
				"data_token":   fmt.Sprintf("0x%04X", device.GetDataToken()),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// DeviceStatusHandler handles device status requests
func DeviceStatusHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device, hasDevice := middleware.GetDevice(r.Context())

		if !hasDevice {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "device not found in context",
			})
			return
		}

		clearance, _ := middleware.GetClearance(r.Context())

		response := map[string]interface{}{
			"device_id": device.ID,
			"name":      device.Name,
			"layer":     device.Layer,
			"class":     device.Class,
			"clearance": clearance.String(),
			"status":    "operational",
			"tokens": map[string]string{
				"status": fmt.Sprintf("0x%04X", device.GetStatusToken()),
				"config": fmt.Sprintf("0x%04X", device.GetConfigToken()),
				"data":   fmt.Sprintf("0x%04X", device.GetDataToken()),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// HighSecurityHandler requires high clearance
func HighSecurityHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearance, hasClearance := middleware.GetClearance(r.Context())

		if !hasClearance {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "clearance required",
			})
			return
		}

		// Require at least level 7
		if !clearance.IsHigherOrEqual(models.ClearanceLevel7) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "insufficient clearance",
				"required":          models.ClearanceLevel7.String(),
				"provided":          clearance.String(),
			})
			return
		}

		response := map[string]interface{}{
			"message":   "Access granted to high security endpoint",
			"clearance": clearance.String(),
			"level":     clearance.Level(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
