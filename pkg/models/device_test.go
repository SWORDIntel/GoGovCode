package models

import (
	"testing"
)

func TestComputeToken(t *testing.T) {
	device := &Device{
		ID: 1,
	}

	tests := []struct {
		name     string
		offset   TokenOffset
		expected uint16
	}{
		{
			name:     "status token",
			offset:   TokenOffsetStatus,
			expected: 0x8003, // 0x8000 + (1 * 3) + 0
		},
		{
			name:     "config token",
			offset:   TokenOffsetConfig,
			expected: 0x8004, // 0x8000 + (1 * 3) + 1
		},
		{
			name:     "data token",
			offset:   TokenOffsetData,
			expected: 0x8005, // 0x8000 + (1 * 3) + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := device.ComputeToken(tt.offset)
			if token != tt.expected {
				t.Errorf("expected token 0x%04X, got 0x%04X", tt.expected, token)
			}
		})
	}
}

func TestClearanceLevel(t *testing.T) {
	tests := []struct {
		clearance Clearance
		level     int
	}{
		{ClearanceLevel2, 2},
		{ClearanceLevel3, 3},
		{ClearanceLevel5, 5},
		{ClearanceLevel9, 9},
	}

	for _, tt := range tests {
		t.Run(tt.clearance.String(), func(t *testing.T) {
			if level := tt.clearance.Level(); level != tt.level {
				t.Errorf("expected level %d, got %d", tt.level, level)
			}
		})
	}
}

func TestClearanceComparison(t *testing.T) {
	if !ClearanceLevel5.IsHigherThan(ClearanceLevel3) {
		t.Error("level 5 should be higher than level 3")
	}

	if ClearanceLevel3.IsHigherThan(ClearanceLevel5) {
		t.Error("level 3 should not be higher than level 5")
	}

	if !ClearanceLevel5.IsHigherOrEqual(ClearanceLevel5) {
		t.Error("level 5 should be higher or equal to level 5")
	}
}

func TestValidateClearance(t *testing.T) {
	tests := []struct {
		name      string
		clearance Clearance
		valid     bool
	}{
		{"valid level 2", ClearanceLevel2, true},
		{"valid level 9", ClearanceLevel9, true},
		{"invalid level 0", 0, false},
		{"invalid level 1", 0x01010101, false},
		{"invalid level 10", 0x0A0A0A0A, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if valid := ValidateClearance(tt.clearance); valid != tt.valid {
				t.Errorf("expected valid=%v, got %v", tt.valid, valid)
			}
		})
	}
}

func TestCanAccessLayer(t *testing.T) {
	tests := []struct {
		name   string
		source Layer
		target Layer
		can    bool
	}{
		{"data to data", LayerData, LayerData, true},
		{"data to transport", LayerData, LayerTransport, true},
		{"data to control", LayerData, LayerControl, true},
		{"data to application", LayerData, LayerApplication, true},
		{"transport to data", LayerTransport, LayerData, false},
		{"transport to transport", LayerTransport, LayerTransport, true},
		{"transport to control", LayerTransport, LayerControl, true},
		{"control to data", LayerControl, LayerData, false},
		{"control to transport", LayerControl, LayerTransport, false},
		{"control to control", LayerControl, LayerControl, true},
		{"control to application", LayerControl, LayerApplication, true},
		{"application to data", LayerApplication, LayerData, false},
		{"application to application", LayerApplication, LayerApplication, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if can := CanAccessLayer(tt.source, tt.target); can != tt.can {
				t.Errorf("expected %v, got %v", tt.can, can)
			}
		})
	}
}

func TestDeviceRegistry(t *testing.T) {
	registry := NewDeviceRegistry()

	device1 := &Device{
		ID:        1,
		Name:      "test-device-1",
		Layer:     LayerData,
		Class:     DeviceClassSensor,
		Clearance: ClearanceLevel3,
	}

	device2 := &Device{
		ID:        2,
		Name:      "test-device-2",
		Layer:     LayerControl,
		Class:     DeviceClassController,
		Clearance: ClearanceLevel7,
	}

	// Register device 1
	if err := registry.Register(device1); err != nil {
		t.Fatalf("failed to register device 1: %v", err)
	}

	// Try to register device 1 again (should fail)
	if err := registry.Register(device1); err == nil {
		t.Error("expected error when registering duplicate device")
	}

	// Register device 2
	if err := registry.Register(device2); err != nil {
		t.Fatalf("failed to register device 2: %v", err)
	}

	// Get device 1
	retrieved, err := registry.GetDevice(1)
	if err != nil {
		t.Fatalf("failed to get device 1: %v", err)
	}
	if retrieved.Name != "test-device-1" {
		t.Errorf("expected device name 'test-device-1', got %s", retrieved.Name)
	}

	// Get non-existent device
	if _, err := registry.GetDevice(999); err == nil {
		t.Error("expected error when getting non-existent device")
	}

	// Get device by token
	statusToken := device1.GetStatusToken()
	retrievedByToken, offset, err := registry.GetDeviceByToken(statusToken)
	if err != nil {
		t.Fatalf("failed to get device by token: %v", err)
	}
	if retrievedByToken.ID != 1 {
		t.Errorf("expected device ID 1, got %d", retrievedByToken.ID)
	}
	if offset != TokenOffsetStatus {
		t.Errorf("expected offset STATUS, got %v", offset)
	}

	// List devices
	devices := registry.ListDevices()
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}
