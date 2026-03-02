package mod

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestDefaultPermissions(t *testing.T) {
	perms := DefaultPermissions()

	if !perms.AllowFileRead {
		t.Error("expected AllowFileRead to be true by default")
	}

	if perms.AllowFileWrite {
		t.Error("expected AllowFileWrite to be false by default")
	}

	if perms.AllowEntitySpawn {
		t.Error("expected AllowEntitySpawn to be false by default")
	}

	if !perms.AllowAssetLoad {
		t.Error("expected AllowAssetLoad to be true by default")
	}

	if perms.AllowUIModify {
		t.Error("expected AllowUIModify to be false by default")
	}
}

func TestNewModAPI(t *testing.T) {
	perms := DefaultPermissions()
	api := NewModAPI("test_mod", perms)

	if api == nil {
		t.Fatal("expected non-nil ModAPI")
	}

	if api.modName != "test_mod" {
		t.Errorf("expected modName 'test_mod', got %s", api.modName)
	}

	if api.GetModName() != "test_mod" {
		t.Errorf("expected GetModName to return 'test_mod', got %s", api.GetModName())
	}

	if len(api.eventHandlers) != 0 {
		t.Error("expected empty eventHandlers map")
	}
}

func TestModAPI_RegisterEventHandler(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())

	handlerCalled := false
	handler := func(data EventData) error {
		handlerCalled = true
		return nil
	}

	err := api.RegisterEventHandler(EventTypeWeaponFire, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify handler was registered
	if len(api.eventHandlers[EventTypeWeaponFire]) != 1 {
		t.Error("expected 1 handler for weapon.fire")
	}

	// Trigger event
	err = api.TriggerEvent(EventTypeWeaponFire, EventData{Type: EventTypeWeaponFire})
	if err != nil {
		t.Fatalf("unexpected error triggering event: %v", err)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestModAPI_RegisterEventHandler_NilHandler(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())

	err := api.RegisterEventHandler(EventTypeWeaponFire, nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}

	if err.Error() != "handler cannot be nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestModAPI_TriggerEvent_NoHandlers(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())

	// Triggering event with no handlers should succeed
	err := api.TriggerEvent(EventTypeWeaponFire, EventData{Type: EventTypeWeaponFire})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModAPI_TriggerEvent_MultipleHandlers(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())

	callCount := 0

	handler1 := func(data EventData) error {
		callCount++
		return nil
	}

	handler2 := func(data EventData) error {
		callCount++
		return nil
	}

	api.RegisterEventHandler(EventTypeWeaponFire, handler1)
	api.RegisterEventHandler(EventTypeWeaponFire, handler2)

	err := api.TriggerEvent(EventTypeWeaponFire, EventData{Type: EventTypeWeaponFire})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 handler calls, got %d", callCount)
	}
}

func TestModAPI_TriggerEvent_HandlerError(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())

	handler := func(data EventData) error {
		return fmt.Errorf("handler error")
	}

	api.RegisterEventHandler(EventTypeWeaponFire, handler)

	err := api.TriggerEvent(EventTypeWeaponFire, EventData{Type: EventTypeWeaponFire})
	if err == nil {
		t.Fatal("expected error from handler")
	}
}

func TestModAPI_SpawnEntity_PermissionDenied(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = false
	api := NewModAPI("test_mod", perms)

	_, err := api.SpawnEntity("enemy", 10.0, 20.0)
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if err.Error() != "permission denied: entity spawn not allowed for mod test_mod" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_SpawnEntity_PermissionGranted(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)

	// Should not get permission error (will get "not implemented" error instead)
	_, err := api.SpawnEntity("enemy", 10.0, 20.0)
	if err != nil && err.Error() == "permission denied: entity spawn not allowed for mod test_mod" {
		t.Fatal("unexpected permission denied error")
	}
}

func TestModAPI_LoadTexture_PermissionDenied(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = false
	api := NewModAPI("test_mod", perms)

	_, err := api.LoadTexture("texture.png")
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if err.Error() != "permission denied: asset load not allowed for mod test_mod" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_LoadTexture_PermissionGranted(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	// Should not get permission error (will get "not implemented" error instead)
	_, err := api.LoadTexture("texture.png")
	if err != nil && err.Error() == "permission denied: asset load not allowed for mod test_mod" {
		t.Fatal("unexpected permission denied error")
	}
}

func TestModAPI_PlaySound_PermissionDenied(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = false
	api := NewModAPI("test_mod", perms)

	err := api.PlaySound(SoundID(123))
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if err.Error() != "permission denied: asset load not allowed for mod test_mod" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_PlaySound_PermissionGranted(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	// Should not get permission error (will get "not implemented" error instead)
	err := api.PlaySound(SoundID(123))
	if err != nil && err.Error() == "permission denied: asset load not allowed for mod test_mod" {
		t.Fatal("unexpected permission denied error")
	}
}

func TestModAPI_ShowNotification_PermissionDenied(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowUIModify = false
	api := NewModAPI("test_mod", perms)

	err := api.ShowNotification("test message")
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if err.Error() != "permission denied: UI modify not allowed for mod test_mod" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_ShowNotification_PermissionGranted(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowUIModify = true
	api := NewModAPI("test_mod", perms)

	// Should not get permission error (will get "not implemented" error instead)
	err := api.ShowNotification("test message")
	if err != nil && err.Error() == "permission denied: UI modify not allowed for mod test_mod" {
		t.Fatal("unexpected permission denied error")
	}
}

func TestModAPI_GetPermissions(t *testing.T) {
	perms := ModPermissions{
		AllowFileRead:    true,
		AllowFileWrite:   true,
		AllowEntitySpawn: true,
		AllowAssetLoad:   false,
		AllowUIModify:    true,
	}

	api := NewModAPI("test_mod", perms)

	returnedPerms := api.GetPermissions()

	if returnedPerms.AllowFileRead != perms.AllowFileRead {
		t.Error("AllowFileRead mismatch")
	}

	if returnedPerms.AllowFileWrite != perms.AllowFileWrite {
		t.Error("AllowFileWrite mismatch")
	}

	if returnedPerms.AllowEntitySpawn != perms.AllowEntitySpawn {
		t.Error("AllowEntitySpawn mismatch")
	}

	if returnedPerms.AllowAssetLoad != perms.AllowAssetLoad {
		t.Error("AllowAssetLoad mismatch")
	}

	if returnedPerms.AllowUIModify != perms.AllowUIModify {
		t.Error("AllowUIModify mismatch")
	}
}

func TestEventData_Structure(t *testing.T) {
	data := EventData{
		Type: EventTypeWeaponFire,
		Params: map[string]interface{}{
			"weapon_id": 42,
			"damage":    100,
		},
	}

	if data.Type != EventTypeWeaponFire {
		t.Errorf("expected Type %s, got %s", EventTypeWeaponFire, data.Type)
	}

	if data.Params["weapon_id"] != 42 {
		t.Error("expected weapon_id to be 42")
	}

	if data.Params["damage"] != 100 {
		t.Error("expected damage to be 100")
	}
}

func TestEventType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"weapon_fire", EventTypeWeaponFire, "weapon.fire"},
		{"enemy_spawn", EventTypeEnemySpawn, "enemy.spawn"},
		{"enemy_killed", EventTypeEnemyKilled, "enemy.killed"},
		{"player_damage", EventTypePlayerDamage, "player.damage"},
		{"player_heal", EventTypePlayerHeal, "player.heal"},
		{"level_generate", EventTypeLevelGenerate, "level.generate"},
		{"level_complete", EventTypeLevelComplete, "level.complete"},
		{"item_pickup", EventTypeItemPickup, "item.pickup"},
		{"door_open", EventTypeDoorOpen, "door.open"},
		{"door_close", EventTypeDoorClose, "door.close"},
		{"genre_set", EventTypeGenreSet, "genre.set"},
		{"mod_load", EventTypeModLoad, "mod.load"},
		{"mod_unload", EventTypeModUnload, "mod.unload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

func TestModAPI_BindGameSystems(t *testing.T) {
	api := NewModAPI("test_mod", DefaultPermissions())
	world := engine.NewWorld()
	// Audio and sprite gen not initialized to avoid Ebiten display requirements
	hudMsg := ""
	hudMsgTime := 0

	api.BindGameSystems(world, nil, nil, &hudMsg, &hudMsgTime)

	if api.world != world {
		t.Error("world not bound correctly")
	}
	if api.hudMessage != &hudMsg {
		t.Error("hudMessage not bound correctly")
	}
	if api.hudMessageTime != &hudMsgTime {
		t.Error("hudMessageTime not bound correctly")
	}
}

func TestModAPI_SpawnEntity_NotBound(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)

	_, err := api.SpawnEntity("enemy", 10.0, 20.0)
	if err == nil {
		t.Fatal("expected error for unbound world")
	}
	if err.Error() != "mod API not bound to game world" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_SpawnEntity_Enemy(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)
	world := engine.NewWorld()
	api.BindGameSystems(world, nil, nil, nil, nil)

	entityID, err := api.SpawnEntity("enemy", 10.0, 20.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := engine.Entity(entityID)

	// Check Position component
	pos, ok := world.GetComponent(e, reflect.TypeOf(&engine.Position{}))
	if !ok {
		t.Fatal("expected Position component")
	}
	if p := pos.(*engine.Position); p.X != 10.0 || p.Y != 20.0 {
		t.Errorf("expected position (10, 20), got (%f, %f)", p.X, p.Y)
	}

	// Check Health component
	_, ok = world.GetComponent(e, reflect.TypeOf(&engine.Health{}))
	if !ok {
		t.Fatal("expected Health component for enemy")
	}

	// Check Velocity component
	_, ok = world.GetComponent(e, reflect.TypeOf(&engine.Velocity{}))
	if !ok {
		t.Fatal("expected Velocity component for enemy")
	}
}

func TestModAPI_SpawnEntity_Prop(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)
	world := engine.NewWorld()
	api.BindGameSystems(world, nil, nil, nil, nil)

	entityID, err := api.SpawnEntity("prop", 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := engine.Entity(entityID)

	// Check Position component
	pos, ok := world.GetComponent(e, reflect.TypeOf(&engine.Position{}))
	if !ok {
		t.Fatal("expected Position component")
	}
	if p := pos.(*engine.Position); p.X != 5.0 || p.Y != 15.0 {
		t.Errorf("expected position (5, 15), got (%f, %f)", p.X, p.Y)
	}

	// Props should not have Health or Velocity
	_, ok = world.GetComponent(e, reflect.TypeOf(&engine.Health{}))
	if ok {
		t.Error("prop should not have Health component")
	}
}

func TestModAPI_SpawnEntity_Pickup(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)
	world := engine.NewWorld()
	api.BindGameSystems(world, nil, nil, nil, nil)

	entityID, err := api.SpawnEntity("pickup", 3.0, 7.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := engine.Entity(entityID)

	// Check Position component
	pos, ok := world.GetComponent(e, reflect.TypeOf(&engine.Position{}))
	if !ok {
		t.Fatal("expected Position component")
	}
	if p := pos.(*engine.Position); p.X != 3.0 || p.Y != 7.0 {
		t.Errorf("expected position (3, 7), got (%f, %f)", p.X, p.Y)
	}
}

func TestModAPI_SpawnEntity_Projectile(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)
	world := engine.NewWorld()
	api.BindGameSystems(world, nil, nil, nil, nil)

	entityID, err := api.SpawnEntity("projectile", 8.0, 12.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := engine.Entity(entityID)

	// Check Position component
	pos, ok := world.GetComponent(e, reflect.TypeOf(&engine.Position{}))
	if !ok {
		t.Fatal("expected Position component")
	}
	if p := pos.(*engine.Position); p.X != 8.0 || p.Y != 12.0 {
		t.Errorf("expected position (8, 12), got (%f, %f)", p.X, p.Y)
	}

	// Check Velocity component
	_, ok = world.GetComponent(e, reflect.TypeOf(&engine.Velocity{}))
	if !ok {
		t.Fatal("expected Velocity component for projectile")
	}
}

func TestModAPI_SpawnEntity_UnknownType(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowEntitySpawn = true
	api := NewModAPI("test_mod", perms)
	world := engine.NewWorld()
	api.BindGameSystems(world, nil, nil, nil, nil)

	_, err := api.SpawnEntity("unknown", 1.0, 2.0)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
	if err.Error() != "unknown entity type: unknown (must be enemy, prop, pickup, or projectile)" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_LoadTexture_NotBound(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	_, err := api.LoadTexture("enemy:goblin:1234")
	if err == nil {
		t.Fatal("expected error for unbound sprite generator")
	}
	if err.Error() != "mod API not bound to sprite generator" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_LoadTexture_Success(t *testing.T) {
	// Test the texture ID generation logic by checking hash consistency
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	// Create mock sprite generator
	type mockSpriteGen struct{}
	var mockGen mockSpriteGen
	api.BindGameSystems(nil, nil, &mockGen, nil, nil)

	texID, err := api.LoadTexture("enemy:goblin:1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if texID == 0 {
		t.Error("expected non-zero texture ID")
	}

	// Same path should produce same ID (deterministic hashing)
	texID2, err := api.LoadTexture("enemy:goblin:1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if texID != texID2 {
		t.Error("expected same texture ID for same path")
	}

	// Different path should produce different ID
	texID3, err := api.LoadTexture("enemy:orc:5678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if texID == texID3 {
		t.Error("expected different texture ID for different path")
	}
}

func TestModAPI_PlaySound_NotBound(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	err := api.PlaySound(SoundID(123))
	if err == nil {
		t.Fatal("expected error for unbound audio engine")
	}
	if err.Error() != "mod API not bound to audio engine" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_PlaySound_Success(t *testing.T) {
	// Test the permission and binding logic with mock audio engine
	perms := DefaultPermissions()
	perms.AllowAssetLoad = true
	api := NewModAPI("test_mod", perms)

	// Create mock audio engine that implements AudioEngine interface
	mock := &mockAudioEngine{}
	api.BindGameSystems(nil, mock, nil, nil, nil)

	err := api.PlaySound(SoundID(123))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.called {
		t.Error("expected PlaySFX to be called")
	}

	expectedSFX := "mod_sfx_123"
	if mock.lastSFX != expectedSFX {
		t.Errorf("expected SFX name '%s', got '%s'", expectedSFX, mock.lastSFX)
	}
}

// mockAudioEngine implements AudioEngine for testing
type mockAudioEngine struct {
	called       bool
	lastSFX      string
	lastX, lastY float64
}

func (m *mockAudioEngine) PlaySFX(name string, x, y float64) error {
	m.called = true
	m.lastSFX = name
	m.lastX = x
	m.lastY = y
	return nil
}

func TestModAPI_ShowNotification_NotBound(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowUIModify = true
	api := NewModAPI("test_mod", perms)

	err := api.ShowNotification("test message")
	if err == nil {
		t.Fatal("expected error for unbound HUD system")
	}
	if err.Error() != "mod API not bound to HUD system" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestModAPI_ShowNotification_Success(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowUIModify = true
	api := NewModAPI("test_mod", perms)
	hudMsg := ""
	hudMsgTime := 0
	api.BindGameSystems(nil, nil, nil, &hudMsg, &hudMsgTime)

	err := api.ShowNotification("Test Notification")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hudMsg != "Test Notification" {
		t.Errorf("expected hudMsg to be 'Test Notification', got '%s'", hudMsg)
	}

	if hudMsgTime != 180 {
		t.Errorf("expected hudMsgTime to be 180, got %d", hudMsgTime)
	}
}

func TestModAPI_ShowNotification_Overwrite(t *testing.T) {
	perms := DefaultPermissions()
	perms.AllowUIModify = true
	api := NewModAPI("test_mod", perms)
	hudMsg := "Old Message"
	hudMsgTime := 50
	api.BindGameSystems(nil, nil, nil, &hudMsg, &hudMsgTime)

	err := api.ShowNotification("New Message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hudMsg != "New Message" {
		t.Errorf("expected hudMsg to be 'New Message', got '%s'", hudMsg)
	}

	if hudMsgTime != 180 {
		t.Errorf("expected hudMsgTime to be reset to 180, got %d", hudMsgTime)
	}
}
