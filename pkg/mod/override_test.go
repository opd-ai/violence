package mod

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
)

// TestParamRegistry tests the parameter override registry functionality.
func TestParamRegistry(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		reg := NewParamRegistry()
		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage_multiplier",
			Value:         2.0,
			Priority:      0,
			ModName:       "TestMod",
		})

		val, found := reg.Get("weapon", "damage_multiplier")
		if !found {
			t.Fatal("override not found")
		}
		if val.(float64) != 2.0 {
			t.Fatalf("wrong value: got %v, want 2.0", val)
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		reg := NewParamRegistry()
		_, found := reg.Get("weapon", "nonexistent")
		if found {
			t.Fatal("expected not found")
		}

		_, found = reg.Get("nonexistent", "key")
		if found {
			t.Fatal("expected not found for nonexistent generator type")
		}
	})

	t.Run("PriorityResolution", func(t *testing.T) {
		reg := NewParamRegistry()

		// Register low priority first
		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         10,
			Priority:      0,
			ModName:       "BaseMod",
		})

		// Register high priority
		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         50,
			Priority:      10,
			ModName:       "OverrideMod",
		})

		// Register mid priority
		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         25,
			Priority:      5,
			ModName:       "MidMod",
		})

		// Highest priority should win
		val, found := reg.Get("weapon", "damage")
		if !found {
			t.Fatal("override not found")
		}
		if val.(int) != 50 {
			t.Fatalf("wanted highest priority value 50, got %v", val)
		}

		// Verify all overrides are stored and sorted
		overrides := reg.GetOverrides("weapon", "damage")
		if len(overrides) != 3 {
			t.Fatalf("expected 3 overrides, got %d", len(overrides))
		}
		if overrides[0].Priority != 10 {
			t.Fatalf("first override should have priority 10, got %d", overrides[0].Priority)
		}
		if overrides[1].Priority != 5 {
			t.Fatalf("second override should have priority 5, got %d", overrides[1].Priority)
		}
		if overrides[2].Priority != 0 {
			t.Fatalf("third override should have priority 0, got %d", overrides[2].Priority)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		reg := NewParamRegistry()

		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         50,
			ModName:       "Mod1",
		})
		reg.Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "fire_rate",
			Value:         1.5,
			ModName:       "Mod1",
		})
		reg.Register(ParamOverride{
			GeneratorType: "enemy",
			Key:           "health",
			Value:         200,
			ModName:       "Mod1",
		})

		weaponOverrides := reg.GetAll("weapon")
		if len(weaponOverrides) != 2 {
			t.Fatalf("expected 2 weapon overrides, got %d", len(weaponOverrides))
		}
		if weaponOverrides["damage"].(int) != 50 {
			t.Fatalf("wrong weapon damage: %v", weaponOverrides["damage"])
		}
		if weaponOverrides["fire_rate"].(float64) != 1.5 {
			t.Fatalf("wrong weapon fire_rate: %v", weaponOverrides["fire_rate"])
		}

		// Empty type
		textureOverrides := reg.GetAll("texture")
		if len(textureOverrides) != 0 {
			t.Fatalf("expected 0 texture overrides, got %d", len(textureOverrides))
		}
	})

	t.Run("RemoveByMod", func(t *testing.T) {
		reg := NewParamRegistry()

		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "damage", Value: 10, ModName: "Mod1"})
		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "range", Value: 20, ModName: "Mod2"})
		reg.Register(ParamOverride{GeneratorType: "enemy", Key: "health", Value: 100, ModName: "Mod1"})

		removed := reg.RemoveByMod("Mod1")
		if removed != 2 {
			t.Fatalf("expected 2 removed, got %d", removed)
		}

		// Mod1 overrides should be gone
		_, found := reg.Get("weapon", "damage")
		if found {
			t.Fatal("Mod1 weapon damage should be removed")
		}
		_, found = reg.Get("enemy", "health")
		if found {
			t.Fatal("Mod1 enemy health should be removed")
		}

		// Mod2 override should remain
		val, found := reg.Get("weapon", "range")
		if !found {
			t.Fatal("Mod2 weapon range should still exist")
		}
		if val.(int) != 20 {
			t.Fatalf("wrong value: %v", val)
		}
	})

	t.Run("RemoveByModNone", func(t *testing.T) {
		reg := NewParamRegistry()
		removed := reg.RemoveByMod("NonExistentMod")
		if removed != 0 {
			t.Fatalf("expected 0 removed, got %d", removed)
		}
	})

	t.Run("ListTypes", func(t *testing.T) {
		reg := NewParamRegistry()

		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "a", Value: 1, ModName: "M"})
		reg.Register(ParamOverride{GeneratorType: "enemy", Key: "b", Value: 2, ModName: "M"})
		reg.Register(ParamOverride{GeneratorType: "texture", Key: "c", Value: 3, ModName: "M"})

		types := reg.ListTypes()
		sort.Strings(types)
		if len(types) != 3 {
			t.Fatalf("expected 3 types, got %d", len(types))
		}
		expected := []string{"enemy", "texture", "weapon"}
		for i, typ := range types {
			if typ != expected[i] {
				t.Fatalf("expected %q at index %d, got %q", expected[i], i, typ)
			}
		}
	})

	t.Run("Clear", func(t *testing.T) {
		reg := NewParamRegistry()

		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "damage", Value: 10, ModName: "M"})
		reg.Clear()

		types := reg.ListTypes()
		if len(types) != 0 {
			t.Fatalf("expected 0 types after clear, got %d", len(types))
		}
	})

	t.Run("ApplyOverrides", func(t *testing.T) {
		reg := NewParamRegistry()

		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "damage", Value: 50, ModName: "M"})
		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "new_param", Value: "added", ModName: "M"})

		params := map[string]interface{}{
			"damage": 10,
			"range":  100,
		}

		result := reg.ApplyOverrides("weapon", params)

		// Overridden value
		if result["damage"].(int) != 50 {
			t.Fatalf("damage should be overridden to 50, got %v", result["damage"])
		}
		// Existing value preserved
		if result["range"].(int) != 100 {
			t.Fatalf("range should be preserved as 100, got %v", result["range"])
		}
		// New value added
		if result["new_param"].(string) != "added" {
			t.Fatalf("new_param should be added, got %v", result["new_param"])
		}
		// Original params not modified
		if params["damage"].(int) != 10 {
			t.Fatal("original params should not be modified")
		}
	})

	t.Run("ApplyOverridesNoOverrides", func(t *testing.T) {
		reg := NewParamRegistry()

		params := map[string]interface{}{"damage": 10}
		result := reg.ApplyOverrides("weapon", params)

		// Should return same map reference when no overrides
		if result["damage"].(int) != 10 {
			t.Fatalf("expected original value, got %v", result["damage"])
		}
	})

	t.Run("ApplyOverridesNilParams", func(t *testing.T) {
		reg := NewParamRegistry()
		reg.Register(ParamOverride{GeneratorType: "weapon", Key: "damage", Value: 50, ModName: "M"})

		result := reg.ApplyOverrides("weapon", nil)
		if result["damage"].(int) != 50 {
			t.Fatalf("expected 50, got %v", result["damage"])
		}
	})

	t.Run("GetOverridesEmpty", func(t *testing.T) {
		reg := NewParamRegistry()

		overrides := reg.GetOverrides("weapon", "damage")
		if overrides != nil {
			t.Fatalf("expected nil, got %v", overrides)
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		reg := NewParamRegistry()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				reg.Register(ParamOverride{
					GeneratorType: "weapon",
					Key:           "damage",
					Value:         idx,
					Priority:      idx,
					ModName:       "TestMod",
				})
				reg.Get("weapon", "damage")
				reg.GetAll("weapon")
				reg.ListTypes()
			}(i)
		}

		wg.Wait()
	})
}

// TestPluginGetGenerators tests auto-registration of plugin generators.
func TestPluginGetGenerators(t *testing.T) {
	t.Run("AutoRegisterOnLoad", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("GenPlugin", "1.0.0")

		weaponGen := NewMockGenerator("weapon")
		enemyGen := NewMockGenerator("enemy")
		plugin.SetGenerators([]Generator{weaponGen, enemyGen})

		if err := manager.LoadPlugin(plugin); err != nil {
			t.Fatalf("LoadPlugin failed: %v", err)
		}

		// Generators should be auto-registered
		if g := manager.Generators().Get("weapon"); g == nil {
			t.Fatal("weapon generator not auto-registered")
		}
		if g := manager.Generators().Get("enemy"); g == nil {
			t.Fatal("enemy generator not auto-registered")
		}

		// Ownership tracked
		if owner := manager.Generators().Owner("weapon"); owner != "GenPlugin" {
			t.Fatalf("weapon generator owner: got %q, want %q", owner, "GenPlugin")
		}
	})

	t.Run("AutoRegisterConflictWarning", func(t *testing.T) {
		manager := NewPluginManager()

		plugin1 := NewMockPlugin("Plugin1", "1.0.0")
		plugin1.SetGenerators([]Generator{NewMockGenerator("weapon")})

		plugin2 := NewMockPlugin("Plugin2", "1.0.0")
		plugin2.SetGenerators([]Generator{NewMockGenerator("weapon")})

		if err := manager.LoadPlugin(plugin1); err != nil {
			t.Fatalf("LoadPlugin 1 failed: %v", err)
		}

		// Second plugin with conflicting generator should load but produce a warning
		if err := manager.LoadPlugin(plugin2); err != nil {
			t.Fatalf("LoadPlugin 2 failed (should succeed with warning): %v", err)
		}

		warnings := manager.Warnings()
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}

		// Generator should still be owned by Plugin1
		if owner := manager.Generators().Owner("weapon"); owner != "Plugin1" {
			t.Fatalf("weapon generator should still be owned by Plugin1, got %q", owner)
		}
	})

	t.Run("NoGenerators", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("EmptyPlugin", "1.0.0")

		if err := manager.LoadPlugin(plugin); err != nil {
			t.Fatalf("LoadPlugin failed: %v", err)
		}

		// No warnings
		if len(manager.Warnings()) != 0 {
			t.Fatalf("expected 0 warnings, got %d", len(manager.Warnings()))
		}
	})

	t.Run("ClearWarnings", func(t *testing.T) {
		manager := NewPluginManager()

		plugin1 := NewMockPlugin("P1", "1.0.0")
		plugin1.SetGenerators([]Generator{NewMockGenerator("weapon")})
		plugin2 := NewMockPlugin("P2", "1.0.0")
		plugin2.SetGenerators([]Generator{NewMockGenerator("weapon")})

		manager.LoadPlugin(plugin1)
		manager.LoadPlugin(plugin2)

		if len(manager.Warnings()) == 0 {
			t.Fatal("should have warnings")
		}

		manager.ClearWarnings()
		if len(manager.Warnings()) != 0 {
			t.Fatal("warnings should be cleared")
		}
	})
}

// TestGeneratorRegistryOwnership tests generator ownership tracking.
func TestGeneratorRegistryOwnership(t *testing.T) {
	t.Run("RegisterFrom", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")

		if err := reg.RegisterFrom(gen, "PluginA"); err != nil {
			t.Fatalf("RegisterFrom failed: %v", err)
		}

		if owner := reg.Owner("weapon"); owner != "PluginA" {
			t.Fatalf("wrong owner: got %q, want %q", owner, "PluginA")
		}
	})

	t.Run("RegisterFromConflict", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		gen1 := NewMockGenerator("weapon")
		gen2 := NewMockGenerator("weapon")

		if err := reg.RegisterFrom(gen1, "PluginA"); err != nil {
			t.Fatalf("first RegisterFrom failed: %v", err)
		}

		err := reg.RegisterFrom(gen2, "PluginB")
		if err == nil {
			t.Fatal("expected conflict error")
		}
	})

	t.Run("ForceRegister", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		gen1 := NewMockGenerator("weapon")
		gen2 := NewMockGenerator("weapon")

		reg.RegisterFrom(gen1, "PluginA")
		prev := reg.ForceRegister(gen2, "PluginB")

		if prev != "PluginA" {
			t.Fatalf("previous owner: got %q, want %q", prev, "PluginA")
		}
		if owner := reg.Owner("weapon"); owner != "PluginB" {
			t.Fatalf("new owner: got %q, want %q", owner, "PluginB")
		}
	})

	t.Run("ForceRegisterNoPrevious", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")

		prev := reg.ForceRegister(gen, "PluginA")
		if prev != "" {
			t.Fatalf("expected empty previous owner, got %q", prev)
		}
	})

	t.Run("OwnerNotFound", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		if owner := reg.Owner("nonexistent"); owner != "" {
			t.Fatalf("expected empty owner, got %q", owner)
		}
	})

	t.Run("ClearResetsOwnership", func(t *testing.T) {
		reg := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")
		reg.RegisterFrom(gen, "PluginA")

		reg.Clear()
		if owner := reg.Owner("weapon"); owner != "" {
			t.Fatalf("expected empty owner after clear, got %q", owner)
		}
	})
}

// TestLoaderLoadAllMods tests loading mods from the mods directory.
func TestLoaderLoadAllMods(t *testing.T) {
	t.Run("EmptyDirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		if err := os.Mkdir(modsDir, 0o755); err != nil {
			t.Fatalf("failed to create mods dir: %v", err)
		}

		loader := NewLoaderWithDir(modsDir)
		count, err := loader.LoadAllMods()
		if err != nil {
			t.Fatalf("LoadAllMods failed: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected 0 mods loaded, got %d", count)
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		loader := NewLoaderWithDir("/nonexistent/mods/dir")
		count, err := loader.LoadAllMods()
		if err != nil {
			t.Fatalf("LoadAllMods should succeed for nonexistent dir, got: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected 0 mods, got %d", count)
		}
	})

	t.Run("MultipleValidMods", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		if err := os.Mkdir(modsDir, 0o755); err != nil {
			t.Fatalf("failed to create mods dir: %v", err)
		}

		// Create mod A
		modADir := filepath.Join(modsDir, "amod")
		os.Mkdir(modADir, 0o755)
		os.WriteFile(filepath.Join(modADir, "mod.json"), []byte(`{"name": "AMod", "version": "1.0.0"}`), 0o644)

		// Create mod B
		modBDir := filepath.Join(modsDir, "bmod")
		os.Mkdir(modBDir, 0o755)
		os.WriteFile(filepath.Join(modBDir, "mod.json"), []byte(`{"name": "BMod", "version": "2.0.0"}`), 0o644)

		loader := NewLoaderWithDir(modsDir)
		count, err := loader.LoadAllMods()
		if err != nil {
			t.Fatalf("LoadAllMods failed: %v", err)
		}
		if count != 2 {
			t.Fatalf("expected 2 mods loaded, got %d", count)
		}

		mods := loader.ListMods()
		if len(mods) != 2 {
			t.Fatalf("expected 2 mods listed, got %d", len(mods))
		}

		// Should be loaded in alphabetical order
		if mods[0].Name != "AMod" {
			t.Fatalf("first mod should be AMod, got %s", mods[0].Name)
		}
		if mods[1].Name != "BMod" {
			t.Fatalf("second mod should be BMod, got %s", mods[1].Name)
		}
	})

	t.Run("SkipInvalidMod", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		os.Mkdir(modsDir, 0o755)

		// Valid mod
		validDir := filepath.Join(modsDir, "amod")
		os.Mkdir(validDir, 0o755)
		os.WriteFile(filepath.Join(validDir, "mod.json"), []byte(`{"name": "ValidMod", "version": "1.0.0"}`), 0o644)

		// Invalid mod (bad JSON)
		invalidDir := filepath.Join(modsDir, "bmod")
		os.Mkdir(invalidDir, 0o755)
		os.WriteFile(filepath.Join(invalidDir, "mod.json"), []byte(`invalid json`), 0o644)

		loader := NewLoaderWithDir(modsDir)
		count, err := loader.LoadAllMods()
		if err != nil {
			t.Fatalf("LoadAllMods failed: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 valid mod loaded, got %d", count)
		}

		// Should have a warning about the invalid mod
		warnings := loader.GetWarnings()
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
	})

	t.Run("SkipDirsWithoutManifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		os.Mkdir(modsDir, 0o755)

		// Dir without mod.json
		emptyDir := filepath.Join(modsDir, "nomod")
		os.Mkdir(emptyDir, 0o755)

		// Dir with mod.json
		modDir := filepath.Join(modsDir, "validmod")
		os.Mkdir(modDir, 0o755)
		os.WriteFile(filepath.Join(modDir, "mod.json"), []byte(`{"name": "ValidMod", "version": "1.0.0"}`), 0o644)

		loader := NewLoaderWithDir(modsDir)
		count, _ := loader.LoadAllMods()
		if count != 1 {
			t.Fatalf("expected 1 mod loaded, got %d", count)
		}
	})

	t.Run("SkipFiles", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		os.Mkdir(modsDir, 0o755)

		// File (not a directory)
		os.WriteFile(filepath.Join(modsDir, "readme.txt"), []byte("not a mod"), 0o644)

		// Valid mod dir
		modDir := filepath.Join(modsDir, "validmod")
		os.Mkdir(modDir, 0o755)
		os.WriteFile(filepath.Join(modDir, "mod.json"), []byte(`{"name": "ValidMod", "version": "1.0.0"}`), 0o644)

		loader := NewLoaderWithDir(modsDir)
		count, _ := loader.LoadAllMods()
		if count != 1 {
			t.Fatalf("expected 1 mod (files should be skipped), got %d", count)
		}
	})

	t.Run("ConflictingModsWarning", func(t *testing.T) {
		tmpDir := t.TempDir()
		modsDir := filepath.Join(tmpDir, "mods")
		os.Mkdir(modsDir, 0o755)

		// Mod A
		modADir := filepath.Join(modsDir, "amod")
		os.Mkdir(modADir, 0o755)
		os.WriteFile(filepath.Join(modADir, "mod.json"), []byte(`{"name": "ModA", "version": "1.0.0"}`), 0o644)

		// Mod B (conflicts with A)
		modBDir := filepath.Join(modsDir, "bmod")
		os.Mkdir(modBDir, 0o755)
		os.WriteFile(filepath.Join(modBDir, "mod.json"), []byte(`{"name": "ModB", "version": "1.0.0"}`), 0o644)

		loader := NewLoaderWithDir(modsDir)
		loader.AddConflict("ModA", "ModB")

		count, _ := loader.LoadAllMods()
		if count != 1 {
			t.Fatalf("expected 1 mod loaded (second should conflict), got %d", count)
		}

		warnings := loader.GetWarnings()
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning about conflict, got %d: %v", len(warnings), warnings)
		}
	})
}

// TestLoaderWarnings tests the warning system.
func TestLoaderWarnings(t *testing.T) {
	t.Run("CombinesWarnings", func(t *testing.T) {
		loader := NewLoader()
		loader.EnableUnsafePlugins = true // Enable unsafe plugins for testing

		// Add a plugin manager warning
		plugin1 := NewMockPlugin("P1", "1.0.0")
		plugin1.SetGenerators([]Generator{NewMockGenerator("weapon")})
		plugin2 := NewMockPlugin("P2", "1.0.0")
		plugin2.SetGenerators([]Generator{NewMockGenerator("weapon")})

		loader.RegisterPlugin(plugin1)
		loader.RegisterPlugin(plugin2)

		warnings := loader.GetWarnings()
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning (generator conflict), got %d: %v", len(warnings), warnings)
		}
	})

	t.Run("ClearWarnings", func(t *testing.T) {
		loader := NewLoader()
		loader.EnableUnsafePlugins = true // Enable unsafe plugins for testing
		plugin1 := NewMockPlugin("P1", "1.0.0")
		plugin1.SetGenerators([]Generator{NewMockGenerator("weapon")})
		plugin2 := NewMockPlugin("P2", "1.0.0")
		plugin2.SetGenerators([]Generator{NewMockGenerator("weapon")})

		loader.RegisterPlugin(plugin1)
		loader.RegisterPlugin(plugin2)

		loader.ClearWarnings()
		if len(loader.GetWarnings()) != 0 {
			t.Fatal("warnings should be cleared")
		}
	})
}

// TestLoaderOverrides tests override convenience methods on Loader.
func TestLoaderOverrides(t *testing.T) {
	t.Run("RegisterAndGetOverrides", func(t *testing.T) {
		loader := NewLoader()

		loader.RegisterOverride(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         100,
			Priority:      5,
			ModName:       "DamageMod",
		})
		loader.RegisterOverride(ParamOverride{
			GeneratorType: "weapon",
			Key:           "fire_rate",
			Value:         2.0,
			Priority:      0,
			ModName:       "RateMod",
		})

		overrides := loader.GetOverrides("weapon")
		if len(overrides) != 2 {
			t.Fatalf("expected 2 overrides, got %d", len(overrides))
		}
		if overrides["damage"].(int) != 100 {
			t.Fatalf("wrong damage override: %v", overrides["damage"])
		}
		if overrides["fire_rate"].(float64) != 2.0 {
			t.Fatalf("wrong fire_rate override: %v", overrides["fire_rate"])
		}
	})

	t.Run("OverrideAppliedToGeneration", func(t *testing.T) {
		loader := NewLoader()
		loader.EnableUnsafePlugins = true // Enable unsafe plugins for testing

		// Register a generator
		gen := NewMockGenerator("weapon")
		plugin := NewMockPlugin("TestPlugin", "1.0.0")
		plugin.SetGenerators([]Generator{gen})
		loader.RegisterPlugin(plugin)

		// Register an override
		loader.RegisterOverride(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage_multiplier",
			Value:         3.0,
			ModName:       "TestMod",
		})

		// Apply overrides to params and generate
		params := map[string]interface{}{"base_damage": 10}
		merged := loader.PluginManager().Overrides().ApplyOverrides("weapon", params)

		retrieved := loader.PluginManager().Generators().Get("weapon")
		result, err := retrieved.Generate(42, merged)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		genParams := resultMap["params"].(map[string]interface{})

		if genParams["damage_multiplier"].(float64) != 3.0 {
			t.Fatalf("override not applied: %v", genParams["damage_multiplier"])
		}
		if genParams["base_damage"].(int) != 10 {
			t.Fatalf("original param lost: %v", genParams["base_damage"])
		}
	})

	t.Run("OverridePrecedence", func(t *testing.T) {
		loader := NewLoader()

		// Low priority override
		loader.RegisterOverride(ParamOverride{
			GeneratorType: "enemy",
			Key:           "health",
			Value:         50,
			Priority:      0,
			ModName:       "LowMod",
		})

		// High priority override
		loader.RegisterOverride(ParamOverride{
			GeneratorType: "enemy",
			Key:           "health",
			Value:         200,
			Priority:      10,
			ModName:       "HighMod",
		})

		overrides := loader.GetOverrides("enemy")
		if overrides["health"].(int) != 200 {
			t.Fatalf("high priority override should win, got %v", overrides["health"])
		}
	})
}

// TestPluginManagerOverrides tests the full override + generator integration.
func TestPluginManagerOverrides(t *testing.T) {
	t.Run("OverridesAccess", func(t *testing.T) {
		manager := NewPluginManager()
		overrides := manager.Overrides()
		if overrides == nil {
			t.Fatal("Overrides() returned nil")
		}
	})

	t.Run("UnloadAllClearsOverrides", func(t *testing.T) {
		manager := NewPluginManager()

		manager.Overrides().Register(ParamOverride{
			GeneratorType: "weapon",
			Key:           "damage",
			Value:         100,
			ModName:       "TestMod",
		})

		plugin := NewMockPlugin("TestPlugin", "1.0.0")
		manager.LoadPlugin(plugin)
		manager.UnloadAll()

		types := manager.Overrides().ListTypes()
		if len(types) != 0 {
			t.Fatalf("overrides should be cleared after UnloadAll, got %d types", len(types))
		}
	})
}
