package biome

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewBiomeMaterialSystem(t *testing.T) {
	sys := NewBiomeMaterialSystem("fantasy")

	if sys == nil {
		t.Fatal("NewBiomeMaterialSystem returned nil")
	}

	if sys.currentGenre != "fantasy" {
		t.Errorf("currentGenre = %v, want fantasy", sys.currentGenre)
	}
}

func TestBiomeMaterialSystemSetGenre(t *testing.T) {
	sys := NewBiomeMaterialSystem("fantasy")
	sys.SetGenre("scifi")

	if sys.currentGenre != "scifi" {
		t.Errorf("genre not updated, got %v want scifi", sys.currentGenre)
	}
}

func TestBiomeMaterialSystemUpdate(t *testing.T) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("fantasy")

	// Should not panic on empty world
	sys.Update(w)

	if sys.gameTime <= 0 {
		t.Error("gameTime not incremented")
	}
}

func TestSpawnMaterialsAtPosition(t *testing.T) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("fantasy")

	spawnCount := 0
	sys.SetMaterialSpawnCallback(func(materialID string, amount int, x, y float64) {
		spawnCount++
		t.Logf("Spawned: %s x%d at (%.2f, %.2f)", materialID, amount, x, y)
	})

	sys.SpawnMaterialsAtPosition(w, BiomeForestDungeon, 2, 10.0, 20.0, 12345)

	t.Logf("Total spawned: %d materials", spawnCount)

	// Verify entities were created
	matType := reflect.TypeOf((*MaterialComponent)(nil))
	entities := w.Query(matType)

	if len(entities) == 0 {
		t.Error("no material entities created")
	}

	// Verify components
	for _, e := range entities {
		matComp, ok := w.GetComponent(e, matType)
		if !ok {
			t.Error("entity missing MaterialComponent")
			continue
		}

		mc := matComp.(*MaterialComponent)
		if mc.MaterialID == "" {
			t.Error("material has empty ID")
		}
		if mc.Amount <= 0 {
			t.Error("material has invalid amount")
		}
		if mc.BiomeType != BiomeForestDungeon {
			t.Error("material has wrong biome type")
		}

		t.Logf("Material entity %d: %s x%d from %s", e, mc.MaterialID, mc.Amount, mc.BiomeType.String())
	}
}

func TestSpawnMaterialsDeterminism(t *testing.T) {
	seed := uint64(99999)
	biome := BiomeDeepCavern

	w1 := engine.NewWorld()
	sys1 := NewBiomeMaterialSystem("fantasy")
	sys1.SpawnMaterialsAtPosition(w1, biome, 2, 0, 0, seed)

	w2 := engine.NewWorld()
	sys2 := NewBiomeMaterialSystem("fantasy")
	sys2.SpawnMaterialsAtPosition(w2, biome, 2, 0, 0, seed)

	matType := reflect.TypeOf((*MaterialComponent)(nil))
	entities1 := w1.Query(matType)
	entities2 := w2.Query(matType)

	if len(entities1) != len(entities2) {
		t.Errorf("entity counts differ: %d vs %d", len(entities1), len(entities2))
	}

	// Build material ID sets for comparison (order may vary due to map iteration)
	mats1 := make(map[string]int)
	mats2 := make(map[string]int)

	for _, e := range entities1 {
		comp, _ := w1.GetComponent(e, matType)
		mc := comp.(*MaterialComponent)
		mats1[mc.MaterialID] += mc.Amount
	}

	for _, e := range entities2 {
		comp, _ := w2.GetComponent(e, matType)
		mc := comp.(*MaterialComponent)
		mats2[mc.MaterialID] += mc.Amount
	}

	// Verify same materials and amounts
	if len(mats1) != len(mats2) {
		t.Errorf("material type counts differ: %d vs %d", len(mats1), len(mats2))
	}

	for matID, amt1 := range mats1 {
		amt2, exists := mats2[matID]
		if !exists {
			t.Errorf("material %s exists in first spawn but not second", matID)
		} else if amt1 != amt2 {
			t.Errorf("amounts differ for %s: %d vs %d", matID, amt1, amt2)
		}
	}
}

func TestMaterialComponentType(t *testing.T) {
	comp := &MaterialComponent{
		MaterialID: "test_mat",
		Amount:     5,
		BiomeType:  BiomeForestDungeon,
	}

	if comp.Type() != "MaterialComponent" {
		t.Errorf("MaterialComponent.Type() = %v, want MaterialComponent", comp.Type())
	}
}

func TestGetMaterialCount(t *testing.T) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("fantasy")

	if sys.GetMaterialCount() != 0 {
		t.Error("initial material count should be 0")
	}

	sys.SpawnMaterialsAtPosition(w, BiomeForestDungeon, 2, 0, 0, 54321)
	sys.Update(w)

	count := sys.GetMaterialCount()
	if count == 0 {
		t.Error("material count should be > 0 after spawning")
	}

	t.Logf("Material count: %d", count)
}

func TestMultipleBiomeSpawns(t *testing.T) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("scifi")

	biomes := []BiomeType{
		BiomeAbandonedLab,
		BiomeAlienHive,
		BiomeDeepCavern,
	}

	for i, biome := range biomes {
		x := float64(i * 10)
		y := float64(i * 5)
		seed := uint64(1000 + i)
		sys.SpawnMaterialsAtPosition(w, biome, 2, x, y, seed)
	}

	matType := reflect.TypeOf((*MaterialComponent)(nil))
	entities := w.Query(matType)

	if len(entities) == 0 {
		t.Fatal("no materials spawned from multiple biomes")
	}

	// Count materials per biome
	biomeCounts := make(map[BiomeType]int)
	for _, e := range entities {
		comp, _ := w.GetComponent(e, matType)
		mc := comp.(*MaterialComponent)
		biomeCounts[mc.BiomeType]++
	}

	for biome, count := range biomeCounts {
		t.Logf("%s: %d materials", biome.String(), count)
	}
}

func BenchmarkSpawnMaterialsAtPosition(b *testing.B) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.SpawnMaterialsAtPosition(w, BiomeForestDungeon, 2, 0, 0, uint64(i))
	}
}

func BenchmarkSystemUpdate(b *testing.B) {
	w := engine.NewWorld()
	sys := NewBiomeMaterialSystem("fantasy")

	// Spawn some materials first
	for i := 0; i < 100; i++ {
		sys.SpawnMaterialsAtPosition(w, BiomeDeepCavern, 2, float64(i), 0, uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}
