package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/NSACodeGov/CodeGov/pkg/models"
)

// Effect represents the policy effect
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Rule represents a single policy rule
type Rule struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Effect            Effect           `json:"effect"`
	Routes            []string         `json:"routes"`
	Methods           []string         `json:"methods"`
	RequiredClearance models.Clearance `json:"required_clearance"`
	AllowedLayers     []models.Layer   `json:"allowed_layers,omitempty"`
	AllowedDevices    []uint16         `json:"allowed_devices,omitempty"`
	DeniedDevices     []uint16         `json:"denied_devices,omitempty"`
	Priority          int              `json:"priority"` // Higher priority wins in conflicts
}

// Policy represents a collection of policy rules
type Policy struct {
	Version string  `json:"version"`
	Rules   []*Rule `json:"rules"`
}

// Context represents the request context for policy evaluation
type Context struct {
	Route       string
	Method      string
	DeviceID    uint16
	Layer       models.Layer
	Clearance   models.Clearance
	RequestID   string
	SourceIP    string
	TokenID     uint16
	TokenOffset models.TokenOffset
}

// Decision represents a policy decision
type Decision struct {
	Effect   Effect
	Reason   string
	RuleID   string
	RuleName string
}

// Engine is the policy engine
type Engine struct {
	mu       sync.RWMutex
	policy   *Policy
	registry *models.DeviceRegistry
}

// NewEngine creates a new policy engine
func NewEngine(registry *models.DeviceRegistry) *Engine {
	return &Engine{
		policy: &Policy{
			Version: "1.0",
			Rules:   make([]*Rule, 0),
		},
		registry: registry,
	}
}

// LoadFromFile loads policy from a JSON file
func (e *Engine) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("failed to parse policy file: %w", err)
	}

	if err := e.Validate(&policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.policy = &policy

	return nil
}

// LoadFromJSON loads policy from JSON bytes
func (e *Engine) LoadFromJSON(data []byte) error {
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("failed to parse policy JSON: %w", err)
	}

	if err := e.Validate(&policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.policy = &policy

	return nil
}

// Validate validates a policy
func (e *Engine) Validate(policy *Policy) error {
	if policy.Version == "" {
		return fmt.Errorf("policy version is required")
	}

	ruleIDs := make(map[string]bool)
	conflicts := make([]string, 0)

	for i, rule := range policy.Rules {
		// Check required fields
		if rule.ID == "" {
			return fmt.Errorf("rule %d: ID is required", i)
		}
		if rule.Effect != EffectAllow && rule.Effect != EffectDeny {
			return fmt.Errorf("rule %s: invalid effect '%s'", rule.ID, rule.Effect)
		}

		// Check for duplicate IDs
		if ruleIDs[rule.ID] {
			return fmt.Errorf("rule %s: duplicate rule ID", rule.ID)
		}
		ruleIDs[rule.ID] = true

		// Validate clearance
		if !models.ValidateClearance(rule.RequiredClearance) && rule.RequiredClearance != 0 {
			return fmt.Errorf("rule %s: invalid clearance level", rule.ID)
		}

		// Validate layers
		for _, layer := range rule.AllowedLayers {
			if layer != models.LayerData && layer != models.LayerTransport &&
				layer != models.LayerControl && layer != models.LayerApplication {
				return fmt.Errorf("rule %s: invalid layer '%s'", rule.ID, layer)
			}
		}

		// Validate devices if registry is available
		if e.registry != nil {
			for _, deviceID := range rule.AllowedDevices {
				if _, err := e.registry.GetDevice(deviceID); err != nil {
					return fmt.Errorf("rule %s: unknown device %d", rule.ID, deviceID)
				}
			}
			for _, deviceID := range rule.DeniedDevices {
				if _, err := e.registry.GetDevice(deviceID); err != nil {
					return fmt.Errorf("rule %s: unknown device %d", rule.ID, deviceID)
				}
			}
		}

		// Check for conflicts with other rules
		for j := i + 1; j < len(policy.Rules); j++ {
			other := policy.Rules[j]
			if conflict := checkConflict(rule, other); conflict != "" {
				conflicts = append(conflicts, fmt.Sprintf("%s vs %s: %s", rule.ID, other.ID, conflict))
			}
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("policy conflicts detected:\n  %s", strings.Join(conflicts, "\n  "))
	}

	return nil
}

// checkConflict checks if two rules conflict
func checkConflict(r1, r2 *Rule) string {
	// Different effects on same route/method/device combination
	if r1.Effect != r2.Effect && r1.Priority == r2.Priority {
		// Check if they apply to the same routes
		for _, route1 := range r1.Routes {
			for _, route2 := range r2.Routes {
				if route1 == route2 {
					// Check if they apply to the same methods
					for _, method1 := range r1.Methods {
						for _, method2 := range r2.Methods {
							if method1 == method2 || method1 == "*" || method2 == "*" {
								return fmt.Sprintf("conflicting effects on route %s method %s with same priority", route1, method1)
							}
						}
					}
				}
			}
		}
	}

	return ""
}

// Evaluate evaluates a request context against the policy
func (e *Engine) Evaluate(ctx *Context) *Decision {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Default deny
	decision := &Decision{
		Effect: EffectDeny,
		Reason: "no matching policy rule",
	}

	var matchedRule *Rule
	highestPriority := -1

	// Find matching rules
	for _, rule := range e.policy.Rules {
		if e.ruleMatches(rule, ctx) {
			// Higher priority wins
			if rule.Priority > highestPriority {
				matchedRule = rule
				highestPriority = rule.Priority
			}
		}
	}

	if matchedRule != nil {
		decision.Effect = matchedRule.Effect
		decision.RuleID = matchedRule.ID
		decision.RuleName = matchedRule.Name

		if matchedRule.Effect == EffectAllow {
			decision.Reason = fmt.Sprintf("allowed by rule '%s'", matchedRule.Name)
		} else {
			decision.Reason = fmt.Sprintf("denied by rule '%s'", matchedRule.Name)
		}
	}

	return decision
}

// ruleMatches checks if a rule matches the context
func (e *Engine) ruleMatches(rule *Rule, ctx *Context) bool {
	// Check route
	if !matchesRoute(rule.Routes, ctx.Route) {
		return false
	}

	// Check method
	if !matchesMethod(rule.Methods, ctx.Method) {
		return false
	}

	// Check clearance
	if rule.RequiredClearance > 0 && !ctx.Clearance.IsHigherOrEqual(rule.RequiredClearance) {
		return false
	}

	// Check allowed layers
	if len(rule.AllowedLayers) > 0 && !containsLayer(rule.AllowedLayers, ctx.Layer) {
		return false
	}

	// Check denied devices (takes precedence)
	if containsDevice(rule.DeniedDevices, ctx.DeviceID) {
		return true // Match for deny
	}

	// Check allowed devices
	if len(rule.AllowedDevices) > 0 && !containsDevice(rule.AllowedDevices, ctx.DeviceID) {
		return false
	}

	return true
}

// matchesRoute checks if a route matches any pattern
func matchesRoute(patterns []string, route string) bool {
	if len(patterns) == 0 {
		return true
	}

	for _, pattern := range patterns {
		if pattern == "*" || pattern == route {
			return true
		}
		// Simple prefix matching
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(route, prefix) {
				return true
			}
		}
	}

	return false
}

// matchesMethod checks if a method matches
func matchesMethod(methods []string, method string) bool {
	if len(methods) == 0 {
		return true
	}

	for _, m := range methods {
		if m == "*" || m == method {
			return true
		}
	}

	return false
}

// containsLayer checks if a layer is in the list
func containsLayer(layers []models.Layer, layer models.Layer) bool {
	for _, l := range layers {
		if l == layer {
			return true
		}
	}
	return false
}

// containsDevice checks if a device is in the list
func containsDevice(devices []uint16, deviceID uint16) bool {
	for _, d := range devices {
		if d == deviceID {
			return true
		}
	}
	return false
}

// GetPolicy returns a copy of the current policy
func (e *Engine) GetPolicy() *Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy
	policyCopy := &Policy{
		Version: e.policy.Version,
		Rules:   make([]*Rule, len(e.policy.Rules)),
	}
	copy(policyCopy.Rules, e.policy.Rules)

	return policyCopy
}
