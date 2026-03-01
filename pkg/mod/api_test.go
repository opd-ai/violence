package mod

import (
	"fmt"
	"testing"
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
