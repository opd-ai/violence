package corpse

import (
	"testing"
)

func BenchmarkSpawnCorpse(b *testing.B) {
	sys := NewSystem(1000, "fantasy", 12345)
	corpses := make([]Corpse, 0, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.SpawnCorpse(&corpses, float64(i%100), float64(i%100), int64(i), "enemy", "humanoid", DeathNormal, 64, false)
	}
}

func BenchmarkUpdateCorpses(b *testing.B) {
	sys := NewSystem(1000, "fantasy", 12345)
	corpses := make([]Corpse, 100)
	for i := range corpses {
		corpses[i] = Corpse{
			X:       float64(i * 10),
			Y:       float64(i * 10),
			Seed:    int64(i),
			Opacity: 1.0,
			Age:     0,
			MaxAge:  30.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.UpdateCorpses(&corpses, 0.016)
	}
}

func BenchmarkDetermineDeathType(b *testing.B) {
	damageTypes := []string{
		"fire", "ice", "electric", "acid", "explosion",
		"slash", "crush", "disintegrate", "normal",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetermineDeathType(damageTypes[i%len(damageTypes)])
	}
}
