package policy

import (
	"encoding/json"
	"testing"

	"github.com/NSACodeGov/CodeGov/pkg/models"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine(nil)

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	if engine.policy == nil {
		t.Fatal("expected non-nil policy")
	}
}

func TestValidate(t *testing.T) {
	engine := NewEngine(nil)

	tests := []struct {
		name    string
		policy  *Policy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: &Policy{
				Version: "1.0",
				Rules: []*Rule{
					{
						ID:       "rule1",
						Name:     "Test Rule",
						Effect:   EffectAllow,
						Routes:   []string{"/test"},
						Methods:  []string{"GET"},
						Priority: 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			policy: &Policy{
				Rules: []*Rule{
					{
						ID:     "rule1",
						Effect: EffectAllow,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing rule ID",
			policy: &Policy{
				Version: "1.0",
				Rules: []*Rule{
					{
						Name:   "Test",
						Effect: EffectAllow,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid effect",
			policy: &Policy{
				Version: "1.0",
				Rules: []*Rule{
					{
						ID:     "rule1",
						Effect: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate rule IDs",
			policy: &Policy{
				Version: "1.0",
				Rules: []*Rule{
					{
						ID:     "rule1",
						Effect: EffectAllow,
					},
					{
						ID:     "rule1",
						Effect: EffectDeny,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid layer",
			policy: &Policy{
				Version: "1.0",
				Rules: []*Rule{
					{
						ID:            "rule1",
						Effect:        EffectAllow,
						AllowedLayers: []models.Layer{"invalid"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.Validate(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	engine := NewEngine(nil)

	policy := &Policy{
		Version: "1.0",
		Rules: []*Rule{
			{
				ID:       "allow-public",
				Name:     "Allow public",
				Effect:   EffectAllow,
				Routes:   []string{"/public"},
				Methods:  []string{"GET"},
				Priority: 100,
			},
			{
				ID:                "allow-with-clearance",
				Name:              "Allow with clearance",
				Effect:            EffectAllow,
				Routes:            []string{"/protected"},
				Methods:           []string{"GET"},
				RequiredClearance: models.ClearanceLevel5,
				Priority:          50,
			},
			{
				ID:             "allow-device",
				Name:           "Allow specific device",
				Effect:         EffectAllow,
				Routes:         []string{"/device/*"},
				Methods:        []string{"*"},
				AllowedDevices: []uint16{1, 2},
				Priority:       60,
			},
			{
				ID:       "deny-default",
				Name:     "Deny all",
				Effect:   EffectDeny,
				Routes:   []string{"*"},
				Methods:  []string{"*"},
				Priority: 0,
			},
		},
	}

	engine.LoadFromJSON(mustMarshal(policy))

	tests := []struct {
		name           string
		ctx            *Context
		expectedEffect Effect
	}{
		{
			name: "allow public",
			ctx: &Context{
				Route:  "/public",
				Method: "GET",
			},
			expectedEffect: EffectAllow,
		},
		{
			name: "deny public POST",
			ctx: &Context{
				Route:  "/public",
				Method: "POST",
			},
			expectedEffect: EffectDeny,
		},
		{
			name: "allow with sufficient clearance",
			ctx: &Context{
				Route:     "/protected",
				Method:    "GET",
				Clearance: models.ClearanceLevel7,
			},
			expectedEffect: EffectAllow,
		},
		{
			name: "deny with insufficient clearance",
			ctx: &Context{
				Route:     "/protected",
				Method:    "GET",
				Clearance: models.ClearanceLevel3,
			},
			expectedEffect: EffectDeny,
		},
		{
			name: "allow specific device",
			ctx: &Context{
				Route:    "/device/status",
				Method:   "GET",
				DeviceID: 1,
			},
			expectedEffect: EffectAllow,
		},
		{
			name: "deny other device",
			ctx: &Context{
				Route:    "/device/status",
				Method:   "GET",
				DeviceID: 99,
			},
			expectedEffect: EffectDeny,
		},
		{
			name: "deny by default",
			ctx: &Context{
				Route:  "/unknown",
				Method: "GET",
			},
			expectedEffect: EffectDeny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.Evaluate(tt.ctx)
			if decision.Effect != tt.expectedEffect {
				t.Errorf("expected effect %s, got %s (reason: %s)", tt.expectedEffect, decision.Effect, decision.Reason)
			}
		})
	}
}

func TestCheckConflict(t *testing.T) {
	rule1 := &Rule{
		ID:       "rule1",
		Effect:   EffectAllow,
		Routes:   []string{"/test"},
		Methods:  []string{"GET"},
		Priority: 10,
	}

	rule2 := &Rule{
		ID:       "rule2",
		Effect:   EffectDeny,
		Routes:   []string{"/test"},
		Methods:  []string{"GET"},
		Priority: 10, // Same priority
	}

	conflict := checkConflict(rule1, rule2)
	if conflict == "" {
		t.Error("expected conflict between rules with different effects on same route/method/priority")
	}

	// Different priority should not conflict
	rule2.Priority = 20
	conflict = checkConflict(rule1, rule2)
	if conflict != "" {
		t.Error("expected no conflict when priorities differ")
	}
}

func TestMatchesRoute(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		route    string
		matches  bool
	}{
		{"exact match", []string{"/test"}, "/test", true},
		{"no match", []string{"/test"}, "/other", false},
		{"wildcard all", []string{"*"}, "/anything", true},
		{"prefix match", []string{"/api/*"}, "/api/users", true},
		{"prefix no match", []string{"/api/*"}, "/other/users", false},
		{"empty patterns", []string{}, "/anything", true},
		{"multiple patterns", []string{"/a", "/b", "/c"}, "/b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if matches := matchesRoute(tt.patterns, tt.route); matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestMatchesMethod(t *testing.T) {
	tests := []struct {
		name    string
		methods []string
		method  string
		matches bool
	}{
		{"exact match", []string{"GET"}, "GET", true},
		{"no match", []string{"GET"}, "POST", false},
		{"wildcard", []string{"*"}, "DELETE", true},
		{"empty methods", []string{}, "GET", true},
		{"multiple methods", []string{"GET", "POST", "PUT"}, "POST", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if matches := matchesMethod(tt.methods, tt.method); matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func mustMarshal(p *Policy) []byte {
	data, _ := json.Marshal(p)
	return data
}
