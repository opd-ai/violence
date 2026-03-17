package automap

import (
	"testing"
)

func TestNewCollapsibleMinimap(t *testing.T) {
	baseMap := NewMap(50, 50)
	cfg := DefaultCollapsibleConfig()

	cm := NewCollapsibleMinimap(baseMap, cfg)

	if cm == nil {
		t.Fatal("NewCollapsibleMinimap returned nil")
	}
	if cm.baseMap != baseMap {
		t.Error("base map not assigned")
	}
	if cm.GetState() != StateCompact {
		t.Errorf("expected initial state StateCompact, got %v", cm.GetState())
	}
}

func TestDefaultCollapsibleConfig(t *testing.T) {
	cfg := DefaultCollapsibleConfig()

	if cfg.ExpandedWidth <= 0 {
		t.Error("ExpandedWidth should be positive")
	}
	if cfg.CompactWidth <= 0 {
		t.Error("CompactWidth should be positive")
	}
	if cfg.ExpandedWidth <= cfg.CompactWidth {
		t.Error("ExpandedWidth should be larger than CompactWidth")
	}
	if cfg.TransitionSpeed <= 0 {
		t.Error("TransitionSpeed should be positive")
	}
	if cfg.ExpandedOpacity <= 0 || cfg.ExpandedOpacity > 1 {
		t.Error("ExpandedOpacity should be in (0, 1]")
	}
}

func TestCollapsibleMinimap_SetState(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(baseMap, cfg)

	tests := []struct {
		name   string
		state  MinimapState
		expect MinimapState
	}{
		{"set expanded", StateExpanded, StateExpanded},
		{"set compact", StateCompact, StateCompact},
		{"set hidden", StateHidden, StateHidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cm.SetState(tc.state)
			if cm.targetState != tc.expect {
				t.Errorf("expected target state %v, got %v", tc.expect, cm.targetState)
			}
		})
	}
}

func TestCollapsibleMinimap_ToggleExpand(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Start at compact
	if cm.targetState != StateCompact {
		t.Fatalf("expected initial state compact, got %v", cm.targetState)
	}

	// Toggle to expanded
	cm.ToggleExpand()
	if cm.targetState != StateExpanded {
		t.Errorf("after toggle, expected expanded, got %v", cm.targetState)
	}

	// Toggle back to compact
	cm.ToggleExpand()
	if cm.targetState != StateCompact {
		t.Errorf("after second toggle, expected compact, got %v", cm.targetState)
	}
}

func TestCollapsibleMinimap_UpdateTransition(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.TransitionSpeed = 10.0 // Fast for testing
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Set to expanded and update
	cm.SetState(StateExpanded)

	// Run multiple update cycles
	for i := 0; i < 20; i++ {
		cm.Update(0.016, 5.0, 5.0)
	}

	// Transition should have progressed toward 1.0
	if cm.transition < 0.5 {
		t.Errorf("transition should progress toward 1.0, got %v", cm.transition)
	}
}

func TestCollapsibleMinimap_AutoHide(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.AutoHideEnabled = true
	cfg.AutoHideDelay = 1.0 // 1 second delay for testing
	cfg.TransitionSpeed = 5.0
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Force expand
	cm.ForceExpand()

	// Simulate being idle for longer than auto-hide delay
	for i := 0; i < 100; i++ {
		cm.Update(0.02, 5.0, 5.0) // No movement, 2 seconds total
	}

	// Should have triggered auto-hide to compact
	if cm.targetState != StateCompact {
		t.Errorf("expected auto-hide to compact after idle, target state: %v", cm.targetState)
	}
}

func TestCollapsibleMinimap_AutoShowOnMove(t *testing.T) {
	baseMap := NewMap(20, 20)
	cfg := DefaultCollapsibleConfig()
	cfg.AutoShowOnMove = true
	cfg.TransitionSpeed = 5.0
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Start at compact
	cm.SetState(StateCompact)
	cm.Update(0.016, 5.0, 5.0)

	// Move to a new unrevealed area
	cm.lastRevealedX = 5
	cm.lastRevealedY = 5
	// Mark area at 10,10 as not revealed
	baseMap.Revealed[10][10] = false

	// Update with new position (entering unrevealed area)
	cm.Update(0.016, 10.5, 10.5)

	// Should have triggered expansion
	if cm.targetState != StateExpanded {
		t.Errorf("expected auto-show on new area, target state: %v", cm.targetState)
	}
}

func TestCollapsibleMinimap_ForceExpand(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(baseMap, cfg)

	cm.ForceExpand()

	if cm.targetState != StateExpanded {
		t.Errorf("ForceExpand should set state to expanded, got %v", cm.targetState)
	}
	if cm.idleTime != 0 {
		t.Error("ForceExpand should reset idle time")
	}
}

func TestCollapsibleMinimap_ForceCompact(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(baseMap, cfg)

	cm.ForceExpand()
	cm.ForceCompact()

	if cm.targetState != StateCompact {
		t.Errorf("ForceCompact should set state to compact, got %v", cm.targetState)
	}
}

func TestCollapsibleMinimap_IsExpanded(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.TransitionSpeed = 100.0 // Instant for testing
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Initially not expanded
	if cm.IsExpanded() {
		t.Error("should not be expanded initially")
	}

	// Force expand and update until fully expanded
	cm.ForceExpand()
	for i := 0; i < 50; i++ {
		cm.Update(0.1, 5.0, 5.0)
	}

	if !cm.IsExpanded() {
		t.Errorf("should be expanded after transition, transition=%v, state=%v", cm.transition, cm.currentState)
	}
}

func TestCollapsibleMinimap_SetBaseMap(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(baseMap, cfg)

	newMap := NewMap(20, 20)
	cm.SetBaseMap(newMap)

	if cm.GetBaseMap() != newMap {
		t.Error("SetBaseMap did not update base map")
	}
	if cm.lastRevealedX != -1 || cm.lastRevealedY != -1 {
		t.Error("SetBaseMap should reset revealed tracking")
	}
}

func TestStateName(t *testing.T) {
	tests := []struct {
		state    MinimapState
		expected string
	}{
		{StateExpanded, "expanded"},
		{StateCompact, "compact"},
		{StateHidden, "hidden"},
		{MinimapState(99), "unknown"},
	}

	for _, tc := range tests {
		name := stateName(tc.state)
		if name != tc.expected {
			t.Errorf("stateName(%d) = %q, expected %q", tc.state, name, tc.expected)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test lerp32
	if lerp32(0, 10, 0.5) != 5 {
		t.Error("lerp32(0, 10, 0.5) should be 5")
	}
	if lerp32(0, 10, 0) != 0 {
		t.Error("lerp32(0, 10, 0) should be 0")
	}
	if lerp32(0, 10, 1) != 10 {
		t.Error("lerp32(0, 10, 1) should be 10")
	}

	// Test clampFloat64
	if clampFloat64(0.5, 0, 1) != 0.5 {
		t.Error("clampFloat64(0.5, 0, 1) should be 0.5")
	}
	if clampFloat64(-1, 0, 1) != 0 {
		t.Error("clampFloat64(-1, 0, 1) should be 0")
	}
	if clampFloat64(2, 0, 1) != 1 {
		t.Error("clampFloat64(2, 0, 1) should be 1")
	}

	// Test easeOutCubic
	if easeOutCubic(0) != 0 {
		t.Error("easeOutCubic(0) should be 0")
	}
	if easeOutCubic(1) != 1 {
		t.Error("easeOutCubic(1) should be 1")
	}
	// Ease out should be above linear midpoint
	if easeOutCubic(0.5) <= 0.5 {
		t.Error("easeOutCubic(0.5) should be above 0.5")
	}
}

func TestCollapsibleMinimap_HiddenState(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.TransitionSpeed = 10.0
	cfg.AutoHideEnabled = false // Disable auto-hide interference
	cfg.AutoShowOnMove = false  // Disable auto-show interference
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Set to hidden state
	cm.SetState(StateHidden)

	// Update several times to complete transition
	for i := 0; i < 100; i++ {
		cm.Update(0.05, 5.0, 5.0)
	}

	// Should be hidden
	if cm.GetState() != StateHidden {
		t.Errorf("expected hidden state, got %v (transition: %v)", cm.GetState(), cm.transition)
	}

	// Transition should be at hidden target (-0.5)
	if cm.transition > -0.2 {
		t.Errorf("transition should be <= -0.2 for hidden state, got %v", cm.transition)
	}
}

func TestCollapsibleMinimap_StateTransitions(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.TransitionSpeed = 5.0
	cm := NewCollapsibleMinimap(baseMap, cfg)

	states := []MinimapState{StateExpanded, StateCompact, StateHidden, StateExpanded}

	for _, state := range states {
		cm.SetState(state)
		for i := 0; i < 50; i++ {
			cm.Update(0.02, 5.0, 5.0)
		}
		// Should eventually reach target state
		if cm.GetState() != state {
			t.Errorf("expected state %v after transitions, got %v", stateName(state), stateName(cm.GetState()))
		}
	}
}

func TestCollapsibleMinimap_GetConfig(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.ExpandedWidth = 300
	cm := NewCollapsibleMinimap(baseMap, cfg)

	if cm.config.ExpandedWidth != 300 {
		t.Errorf("expected ExpandedWidth 300, got %v", cm.config.ExpandedWidth)
	}
}

func TestCollapsibleMinimap_NilBaseMap(t *testing.T) {
	cfg := DefaultCollapsibleConfig()
	cm := NewCollapsibleMinimap(nil, cfg)

	// Should not panic with nil base map
	cm.Update(0.016, 0, 0)

	// GetBaseMap should return nil
	if cm.GetBaseMap() != nil {
		t.Error("expected nil base map")
	}
}

func TestCollapsibleMinimap_AutoShowDisabled(t *testing.T) {
	baseMap := NewMap(20, 20)
	cfg := DefaultCollapsibleConfig()
	cfg.AutoShowOnMove = false
	cfg.TransitionSpeed = 5.0
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Start compact
	cm.SetState(StateCompact)
	cm.Update(0.016, 5.0, 5.0)

	// Move to new unrevealed area
	cm.lastRevealedX = 5
	cm.lastRevealedY = 5
	baseMap.Revealed[10][10] = false

	// Update with new position
	cm.Update(0.016, 10.5, 10.5)

	// Should stay compact since auto-show is disabled
	if cm.targetState != StateCompact {
		t.Errorf("with AutoShowOnMove=false, should stay compact, got %v", cm.targetState)
	}
}

func TestCollapsibleMinimap_AutoHideDisabled(t *testing.T) {
	baseMap := NewMap(10, 10)
	cfg := DefaultCollapsibleConfig()
	cfg.AutoHideEnabled = false
	cm := NewCollapsibleMinimap(baseMap, cfg)

	// Force expand
	cm.ForceExpand()

	// Simulate being idle for a long time
	for i := 0; i < 500; i++ {
		cm.Update(0.02, 5.0, 5.0) // 10 seconds total
	}

	// Should still be expanded since auto-hide is disabled
	if cm.targetState != StateExpanded {
		t.Errorf("with AutoHideEnabled=false, should stay expanded, got %v", cm.targetState)
	}
}
