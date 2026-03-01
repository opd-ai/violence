package feedback

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func BenchmarkUpdate(b *testing.B) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	// Populate with typical gameplay load
	for i := 0; i < 20; i++ {
		fs.SpawnDamageNumber(float64(i), float64(i), 50, false)
		fs.SpawnImpactEffect(float64(i), float64(i), ImpactHit)
	}
	fs.AddScreenShake(5.0)
	fs.AddHitFlash(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Update(w)
	}
}

func BenchmarkSpawnDamageNumber(b *testing.B) {
	fs := NewFeedbackSystem(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.SpawnDamageNumber(10.0, 20.0, 50, false)
	}
}

func BenchmarkSpawnImpactEffect(b *testing.B) {
	fs := NewFeedbackSystem(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.SpawnImpactEffect(10.0, 20.0, ImpactHit)
	}
}

func BenchmarkScreenShake(b *testing.B) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	fs.AddScreenShake(10.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Update(w)
	}
}

func BenchmarkHitFlash(b *testing.B) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	fs.AddHitFlash(0.8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.Update(w)
	}
}
