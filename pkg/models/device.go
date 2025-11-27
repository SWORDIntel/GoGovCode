package models

import (
	"fmt"
)

// Clearance represents a DSMIL clearance level
type Clearance uint32

const (
	// Clearance levels from 0x02020202 to 0x09090909
	ClearanceLevel2 Clearance = 0x02020202
	ClearanceLevel3 Clearance = 0x03030303
	ClearanceLevel4 Clearance = 0x04040404
	ClearanceLevel5 Clearance = 0x05050505
	ClearanceLevel6 Clearance = 0x06060606
	ClearanceLevel7 Clearance = 0x07070707
	ClearanceLevel8 Clearance = 0x08080808
	ClearanceLevel9 Clearance = 0x09090909
)

// Layer represents a DSMIL layer
type Layer string

const (
	LayerData        Layer = "data"
	LayerTransport   Layer = "transport"
	LayerControl     Layer = "control"
	LayerApplication Layer = "application"
)

// DeviceClass represents the type of device
type DeviceClass string

const (
	DeviceClassSensor    DeviceClass = "sensor"
	DeviceClassActuator  DeviceClass = "actuator"
	DeviceClassGateway   DeviceClass = "gateway"
	DeviceClassController DeviceClass = "controller"
)

// TokenOffset represents the token type offset
type TokenOffset int

const (
	TokenOffsetStatus TokenOffset = 0
	TokenOffsetConfig TokenOffset = 1
	TokenOffsetData   TokenOffset = 2
)

// Device represents a DSMIL device
type Device struct {
	ID        uint16      `json:"device_id"`
	Layer     Layer       `json:"layer"`
	Class     DeviceClass `json:"class"`
	Clearance Clearance   `json:"clearance"`
	Name      string      `json:"name"`
	TokenBase uint16      `json:"token_base"`
}

// ComputeToken calculates the token ID for a device
// Formula: 0x8000 + (device_id * 3) + offset
func (d *Device) ComputeToken(offset TokenOffset) uint16 {
	return 0x8000 + (d.ID * 3) + uint16(offset)
}

// GetStatusToken returns the STATUS token for this device
func (d *Device) GetStatusToken() uint16 {
	return d.ComputeToken(TokenOffsetStatus)
}

// GetConfigToken returns the CONFIG token for this device
func (d *Device) GetConfigToken() uint16 {
	return d.ComputeToken(TokenOffsetConfig)
}

// GetDataToken returns the DATA token for this device
func (d *Device) GetDataToken() uint16 {
	return d.ComputeToken(TokenOffsetData)
}

// DeviceRegistry manages device information
type DeviceRegistry struct {
	devices map[uint16]*Device
	tokens  map[uint16]*Device // Maps token ID to device
}

// NewDeviceRegistry creates a new device registry
func NewDeviceRegistry() *DeviceRegistry {
	return &DeviceRegistry{
		devices: make(map[uint16]*Device),
		tokens:  make(map[uint16]*Device),
	}
}

// Register adds a device to the registry
func (r *DeviceRegistry) Register(device *Device) error {
	if _, exists := r.devices[device.ID]; exists {
		return fmt.Errorf("device %d already registered", device.ID)
	}

	device.TokenBase = 0x8000 + (device.ID * 3)
	r.devices[device.ID] = device

	// Register all token types
	r.tokens[device.GetStatusToken()] = device
	r.tokens[device.GetConfigToken()] = device
	r.tokens[device.GetDataToken()] = device

	return nil
}

// GetDevice retrieves a device by ID
func (r *DeviceRegistry) GetDevice(deviceID uint16) (*Device, error) {
	device, ok := r.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}
	return device, nil
}

// GetDeviceByToken retrieves a device by token ID
func (r *DeviceRegistry) GetDeviceByToken(tokenID uint16) (*Device, TokenOffset, error) {
	device, ok := r.tokens[tokenID]
	if !ok {
		return nil, 0, fmt.Errorf("token %d not found", tokenID)
	}

	// Determine offset
	offset := TokenOffset((tokenID - device.TokenBase) % 3)
	return device, offset, nil
}

// ListDevices returns all registered devices
func (r *DeviceRegistry) ListDevices() []*Device {
	devices := make([]*Device, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	return devices
}

// ClearanceLevel returns the numeric level from a clearance value
func (c Clearance) Level() int {
	// Extract the level from the repeating byte pattern
	return int((c >> 24) & 0xFF)
}

// String returns a string representation of the clearance
func (c Clearance) String() string {
	return fmt.Sprintf("0x%08X (Level %d)", uint32(c), c.Level())
}

// IsHigherThan checks if this clearance is higher than another
func (c Clearance) IsHigherThan(other Clearance) bool {
	return c > other
}

// IsHigherOrEqual checks if this clearance is higher or equal to another
func (c Clearance) IsHigherOrEqual(other Clearance) bool {
	return c >= other
}

// ValidateClearance checks if a clearance value is valid
func ValidateClearance(c Clearance) bool {
	// Must be between level 2 and level 9
	level := c.Level()
	return level >= 2 && level <= 9
}

// CanAccessLayer checks if data flow is allowed from source to target layer
// DSMIL enforces upward-only data flows (lower → higher)
func CanAccessLayer(sourceLayer, targetLayer Layer) bool {
	layerOrder := map[Layer]int{
		LayerData:        1,
		LayerTransport:   2,
		LayerControl:     3,
		LayerApplication: 4,
	}

	sourceLevel, sourceOk := layerOrder[sourceLayer]
	targetLevel, targetOk := layerOrder[targetLayer]

	if !sourceOk || !targetOk {
		return false
	}

	// Allow same layer or upward (lower → higher)
	return sourceLevel <= targetLevel
}
