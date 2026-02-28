package mod

import (
	"errors"
	"sort"
	"sync"
	"testing"
)

// MockPlugin is a test plugin implementation.
type MockPlugin struct {
	name      string
	version   string
	loadErr   error
	unloadErr error
	loadCalls int
	mu        sync.Mutex
}

func NewMockPlugin(name, version string) *MockPlugin {
	return &MockPlugin{
		name:    name,
		version: version,
	}
}

func (p *MockPlugin) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loadCalls++
	return p.loadErr
}

func (p *MockPlugin) Unload() error {
	return p.unloadErr
}

func (p *MockPlugin) Name() string {
	return p.name
}

func (p *MockPlugin) Version() string {
	return p.version
}

func (p *MockPlugin) SetLoadError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loadErr = err
}

func (p *MockPlugin) SetUnloadError(err error) {
	p.unloadErr = err
}

func (p *MockPlugin) GetLoadCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.loadCalls
}

// MockGenerator is a test generator implementation.
type MockGenerator struct {
	genType string
	genErr  error
}

func NewMockGenerator(genType string) *MockGenerator {
	return &MockGenerator{genType: genType}
}

func (g *MockGenerator) Type() string {
	return g.genType
}

func (g *MockGenerator) Generate(seed int64, params map[string]interface{}) (interface{}, error) {
	if g.genErr != nil {
		return nil, g.genErr
	}
	return map[string]interface{}{"seed": seed, "params": params}, nil
}

func (g *MockGenerator) SetGenerateError(err error) {
	g.genErr = err
}

// TestHookRegistry tests the hook registry functionality.
func TestHookRegistry(t *testing.T) {
	t.Run("RegisterAndTrigger", func(t *testing.T) {
		registry := NewHookRegistry()
		called := false

		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			called = true
			return nil
		})

		if err := registry.Trigger(HookTypeWeaponFire, nil); err != nil {
			t.Fatalf("Trigger failed: %v", err)
		}

		if !called {
			t.Fatal("hook was not called")
		}
	})

	t.Run("MultipleHooks", func(t *testing.T) {
		registry := NewHookRegistry()
		calls := 0

		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			calls++
			return nil
		})
		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			calls++
			return nil
		})

		if err := registry.Trigger(HookTypeWeaponFire, nil); err != nil {
			t.Fatalf("Trigger failed: %v", err)
		}

		if calls != 2 {
			t.Fatalf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("HookError", func(t *testing.T) {
		registry := NewHookRegistry()
		testErr := errors.New("hook error")

		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			return testErr
		})

		err := registry.Trigger(HookTypeWeaponFire, nil)
		if err == nil {
			t.Fatal("expected error from hook")
		}
	})

	t.Run("HookData", func(t *testing.T) {
		registry := NewHookRegistry()
		var received interface{}

		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			received = data
			return nil
		})

		testData := map[string]string{"weapon": "pistol"}
		if err := registry.Trigger(HookTypeWeaponFire, testData); err != nil {
			t.Fatalf("Trigger failed: %v", err)
		}

		dataMap, ok := received.(map[string]string)
		if !ok {
			t.Fatal("received data type mismatch")
		}
		if dataMap["weapon"] != "pistol" {
			t.Fatalf("expected weapon=pistol, got %s", dataMap["weapon"])
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := NewHookRegistry()
		called := false

		registry.Register(HookTypeWeaponFire, func(data interface{}) error {
			called = true
			return nil
		})

		registry.Unregister(HookTypeWeaponFire)

		if err := registry.Trigger(HookTypeWeaponFire, nil); err != nil {
			t.Fatalf("Trigger failed: %v", err)
		}

		if called {
			t.Fatal("hook should not be called after unregister")
		}
	})

	t.Run("List", func(t *testing.T) {
		registry := NewHookRegistry()

		registry.Register(HookTypeWeaponFire, func(data interface{}) error { return nil })
		registry.Register(HookTypeEnemySpawn, func(data interface{}) error { return nil })

		types := registry.List()
		if len(types) != 2 {
			t.Fatalf("expected 2 hook types, got %d", len(types))
		}

		sort.Slice(types, func(i, j int) bool {
			return types[i] < types[j]
		})

		if types[0] != HookTypeEnemySpawn {
			t.Fatalf("expected %s, got %s", HookTypeEnemySpawn, types[0])
		}
		if types[1] != HookTypeWeaponFire {
			t.Fatalf("expected %s, got %s", HookTypeWeaponFire, types[1])
		}
	})

	t.Run("Count", func(t *testing.T) {
		registry := NewHookRegistry()

		registry.Register(HookTypeWeaponFire, func(data interface{}) error { return nil })
		registry.Register(HookTypeWeaponFire, func(data interface{}) error { return nil })
		registry.Register(HookTypeEnemySpawn, func(data interface{}) error { return nil })

		if count := registry.Count(HookTypeWeaponFire); count != 2 {
			t.Fatalf("expected 2 hooks for weapon.fire, got %d", count)
		}
		if count := registry.Count(HookTypeEnemySpawn); count != 1 {
			t.Fatalf("expected 1 hook for enemy.spawn, got %d", count)
		}
		if count := registry.Count(HookTypePlayerDamage); count != 0 {
			t.Fatalf("expected 0 hooks for player.damage, got %d", count)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		registry := NewHookRegistry()

		registry.Register(HookTypeWeaponFire, func(data interface{}) error { return nil })
		registry.Register(HookTypeEnemySpawn, func(data interface{}) error { return nil })

		registry.Clear()

		if len(registry.List()) != 0 {
			t.Fatal("expected no hooks after clear")
		}
	})
}

// TestGeneratorRegistry tests the generator registry functionality.
func TestGeneratorRegistry(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")

		if err := registry.Register(gen); err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		retrieved := registry.Get("weapon")
		if retrieved == nil {
			t.Fatal("generator not found")
		}
		if retrieved.Type() != "weapon" {
			t.Fatalf("wrong generator type: got %s, want weapon", retrieved.Type())
		}
	})

	t.Run("RegisterDuplicate", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen1 := NewMockGenerator("weapon")
		gen2 := NewMockGenerator("weapon")

		if err := registry.Register(gen1); err != nil {
			t.Fatalf("first Register failed: %v", err)
		}

		err := registry.Register(gen2)
		if err == nil {
			t.Fatal("expected error when registering duplicate type")
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		registry := NewGeneratorRegistry()

		gen := registry.Get("nonexistent")
		if gen != nil {
			t.Fatal("expected nil for nonexistent generator")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")

		if err := registry.Register(gen); err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		registry.Unregister("weapon")

		if registry.Get("weapon") != nil {
			t.Fatal("generator should be unregistered")
		}
	})

	t.Run("List", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen1 := NewMockGenerator("weapon")
		gen2 := NewMockGenerator("enemy")

		registry.Register(gen1)
		registry.Register(gen2)

		types := registry.List()
		if len(types) != 2 {
			t.Fatalf("expected 2 generator types, got %d", len(types))
		}

		sort.Strings(types)
		if types[0] != "enemy" || types[1] != "weapon" {
			t.Fatalf("unexpected generator types: %v", types)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen1 := NewMockGenerator("weapon")
		gen2 := NewMockGenerator("enemy")

		registry.Register(gen1)
		registry.Register(gen2)

		registry.Clear()

		if len(registry.List()) != 0 {
			t.Fatal("expected no generators after clear")
		}
	})

	t.Run("Generate", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		gen := NewMockGenerator("weapon")

		registry.Register(gen)

		result, err := gen.Generate(12345, map[string]interface{}{"damage": 10})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result type mismatch")
		}
		if resultMap["seed"].(int64) != 12345 {
			t.Fatalf("wrong seed: got %v, want 12345", resultMap["seed"])
		}
	})
}

// TestPluginManager tests the plugin manager functionality.
func TestPluginManager(t *testing.T) {
	t.Run("LoadPlugin", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")

		if err := manager.LoadPlugin(plugin); err != nil {
			t.Fatalf("LoadPlugin failed: %v", err)
		}

		if plugin.GetLoadCalls() != 1 {
			t.Fatalf("expected 1 Load call, got %d", plugin.GetLoadCalls())
		}

		plugins := manager.ListPlugins()
		if len(plugins) != 1 {
			t.Fatalf("expected 1 plugin, got %d", len(plugins))
		}
		if plugins[0] != "TestPlugin" {
			t.Fatalf("wrong plugin name: got %s", plugins[0])
		}
	})

	t.Run("LoadPluginError", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")
		plugin.SetLoadError(errors.New("load error"))

		err := manager.LoadPlugin(plugin)
		if err == nil {
			t.Fatal("expected error from LoadPlugin")
		}

		if len(manager.ListPlugins()) != 0 {
			t.Fatal("plugin should not be loaded on error")
		}
	})

	t.Run("LoadPluginDuplicate", func(t *testing.T) {
		manager := NewPluginManager()
		plugin1 := NewMockPlugin("TestPlugin", "1.0.0")
		plugin2 := NewMockPlugin("TestPlugin", "2.0.0")

		if err := manager.LoadPlugin(plugin1); err != nil {
			t.Fatalf("first LoadPlugin failed: %v", err)
		}

		err := manager.LoadPlugin(plugin2)
		if err == nil {
			t.Fatal("expected error when loading duplicate plugin")
		}
	})

	t.Run("UnloadPlugin", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")

		manager.LoadPlugin(plugin)

		if err := manager.UnloadPlugin("TestPlugin"); err != nil {
			t.Fatalf("UnloadPlugin failed: %v", err)
		}

		if len(manager.ListPlugins()) != 0 {
			t.Fatal("plugin should be unloaded")
		}
	})

	t.Run("UnloadPluginError", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")
		plugin.SetUnloadError(errors.New("unload error"))

		manager.LoadPlugin(plugin)

		err := manager.UnloadPlugin("TestPlugin")
		if err == nil {
			t.Fatal("expected error from UnloadPlugin")
		}
	})

	t.Run("UnloadPluginNotFound", func(t *testing.T) {
		manager := NewPluginManager()

		err := manager.UnloadPlugin("NonexistentPlugin")
		if err == nil {
			t.Fatal("expected error when unloading nonexistent plugin")
		}
	})

	t.Run("GetPlugin", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")

		manager.LoadPlugin(plugin)

		retrieved, err := manager.GetPlugin("TestPlugin")
		if err != nil {
			t.Fatalf("GetPlugin failed: %v", err)
		}
		if retrieved.Name() != "TestPlugin" {
			t.Fatalf("wrong plugin: got %s", retrieved.Name())
		}
	})

	t.Run("GetPluginNotFound", func(t *testing.T) {
		manager := NewPluginManager()

		_, err := manager.GetPlugin("NonexistentPlugin")
		if err == nil {
			t.Fatal("expected error for nonexistent plugin")
		}
	})

	t.Run("HooksAccess", func(t *testing.T) {
		manager := NewPluginManager()
		hooks := manager.Hooks()

		if hooks == nil {
			t.Fatal("Hooks() returned nil")
		}

		hooks.Register(HookTypeWeaponFire, func(data interface{}) error {
			return nil
		})

		if hooks.Count(HookTypeWeaponFire) != 1 {
			t.Fatal("hook not registered")
		}
	})

	t.Run("GeneratorsAccess", func(t *testing.T) {
		manager := NewPluginManager()
		generators := manager.Generators()

		if generators == nil {
			t.Fatal("Generators() returned nil")
		}

		gen := NewMockGenerator("weapon")
		generators.Register(gen)

		if generators.Get("weapon") == nil {
			t.Fatal("generator not registered")
		}
	})

	t.Run("UnloadAll", func(t *testing.T) {
		manager := NewPluginManager()
		plugin1 := NewMockPlugin("Plugin1", "1.0.0")
		plugin2 := NewMockPlugin("Plugin2", "1.0.0")

		manager.LoadPlugin(plugin1)
		manager.LoadPlugin(plugin2)

		manager.Hooks().Register(HookTypeWeaponFire, func(data interface{}) error {
			return nil
		})
		gen := NewMockGenerator("weapon")
		manager.Generators().Register(gen)

		if err := manager.UnloadAll(); err != nil {
			t.Fatalf("UnloadAll failed: %v", err)
		}

		if len(manager.ListPlugins()) != 0 {
			t.Fatal("plugins should be unloaded")
		}
		if len(manager.Hooks().List()) != 0 {
			t.Fatal("hooks should be cleared")
		}
		if len(manager.Generators().List()) != 0 {
			t.Fatal("generators should be cleared")
		}
	})

	t.Run("UnloadAllWithError", func(t *testing.T) {
		manager := NewPluginManager()
		plugin := NewMockPlugin("TestPlugin", "1.0.0")
		plugin.SetUnloadError(errors.New("unload error"))

		manager.LoadPlugin(plugin)

		err := manager.UnloadAll()
		if err == nil {
			t.Fatal("expected error from UnloadAll")
		}

		// Plugins should still be cleared despite error
		if len(manager.ListPlugins()) != 0 {
			t.Fatal("plugins should be cleared even on error")
		}
	})
}

// TestConcurrency tests thread-safety of registries.
func TestConcurrency(t *testing.T) {
	t.Run("HookRegistryConcurrent", func(t *testing.T) {
		registry := NewHookRegistry()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				registry.Register(HookTypeWeaponFire, func(data interface{}) error {
					return nil
				})
				registry.Trigger(HookTypeWeaponFire, nil)
				registry.List()
				registry.Count(HookTypeWeaponFire)
			}()
		}

		wg.Wait()
	})

	t.Run("GeneratorRegistryConcurrent", func(t *testing.T) {
		registry := NewGeneratorRegistry()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				gen := NewMockGenerator("weapon")
				registry.Register(gen)
				registry.Get("weapon")
				registry.List()
			}(i)
		}

		wg.Wait()
	})

	t.Run("PluginManagerConcurrent", func(t *testing.T) {
		manager := NewPluginManager()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				plugin := NewMockPlugin("TestPlugin", "1.0.0")
				manager.LoadPlugin(plugin)
				manager.GetPlugin("TestPlugin")
				manager.ListPlugins()
			}(i)
		}

		wg.Wait()
	})
}
