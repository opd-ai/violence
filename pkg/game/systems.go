// Package game provides game initialization and ECS system registration helpers.
package game

import (
	"github.com/opd-ai/violence/pkg/ai"
	"github.com/opd-ai/violence/pkg/animation"
	"github.com/opd-ai/violence/pkg/attackanim"
	"github.com/opd-ai/violence/pkg/attacktrail"
	"github.com/opd-ai/violence/pkg/biome"
	"github.com/opd-ai/violence/pkg/collision"
	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/crosshair"
	"github.com/opd-ai/violence/pkg/damagenumber"
	"github.com/opd-ai/violence/pkg/damagestate"
	"github.com/opd-ai/violence/pkg/dmgfx"
	"github.com/opd-ai/violence/pkg/edgeao"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/entitylabel"
	"github.com/opd-ai/violence/pkg/equipment"
	"github.com/opd-ai/violence/pkg/faction"
	"github.com/opd-ai/violence/pkg/feedback"
	"github.com/opd-ai/violence/pkg/fog"
	"github.com/opd-ai/violence/pkg/hazard"
	"github.com/opd-ai/violence/pkg/healthbar"
	"github.com/opd-ai/violence/pkg/impactburst"
	"github.com/opd-ai/violence/pkg/lighting"
	"github.com/opd-ai/violence/pkg/loot"
	"github.com/opd-ai/violence/pkg/motion"
	"github.com/opd-ai/violence/pkg/outline"
	"github.com/opd-ai/violence/pkg/parallax"
	"github.com/opd-ai/violence/pkg/particle"
	"github.com/opd-ai/violence/pkg/playersprite"
	"github.com/opd-ai/violence/pkg/projectile"
	"github.com/opd-ai/violence/pkg/proximityui"
	"github.com/opd-ai/violence/pkg/rimlight"
	"github.com/opd-ai/violence/pkg/spatial"
	"github.com/opd-ai/violence/pkg/stats"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/opd-ai/violence/pkg/statusfx"
	"github.com/opd-ai/violence/pkg/statustint"
	"github.com/opd-ai/violence/pkg/subsurface"
	"github.com/opd-ai/violence/pkg/telegraph"
	"github.com/opd-ai/violence/pkg/territory"
	"github.com/opd-ai/violence/pkg/trap"
	"github.com/opd-ai/violence/pkg/walltex"
	"github.com/opd-ai/violence/pkg/weapon"
	"github.com/opd-ai/violence/pkg/weaponanim"
	"github.com/opd-ai/violence/pkg/weather"
)

// SystemDependencies holds all system instances needed for wiring.
type SystemDependencies struct {
	Spatial          *spatial.System
	Animation        *animation.AnimationSystem
	Motion           *motion.System
	Status           *status.System
	Combo            *combat.ComboSystem
	LootDrop         *loot.LootDropSystem
	Feedback         *feedback.FeedbackSystem
	Defense          *combat.DefenseSystem
	Lighting         *lighting.LightingSystem
	BossPhase        *combat.BossPhaseSystem
	HazardECS        *hazard.ECSSystem
	Faction          *faction.ReputationSystem
	Stat             *stats.System
	Weather          *weather.System
	Sliding          *collision.SlidingSystem
	Equipment        *equipment.EquipmentSystem
	Positional       *combat.PositionalSystem
	AdaptiveAI       *ai.AdaptiveAISystem
	AO               *lighting.AOSystem
	Projectile       *projectile.System
	BiomeMaterial    *biome.BiomeMaterialSystem
	Trap             *trap.System
	QuestLoot        *loot.QuestLootSystem
	DamageState      *damagestate.System
	Dmgfx            *dmgfx.System
	Territory        *territory.ControlSystem
	Outline          *outline.System
	RimLight         *rimlight.System
	AttackTrail      *attacktrail.System
	Telegraph        *telegraph.System
	LootVisual       *loot.VisualSystem
	HealthBar        *healthbar.System
	Fog              *fog.System
	Parallax         *parallax.System
	WeaponAnim       *weaponanim.System
	WallTex          *walltex.System
	ParticleRenderer *particle.RendererSystem
	DamageNumber     *damagenumber.System
	WeaponVisual     *weapon.VisualSystem
	AttackAnim       *attackanim.System
	StatusFX         *statusfx.System
	StatusTint       *statustint.System
	PlayerSprite     *playersprite.System
	Crosshair        *crosshair.System
	ImpactBurst      *impactburst.System
	EntityLabel      *entitylabel.System
	Particle         *particle.ParticleSystem
	ProximityUI      *proximityui.System
	Subsurface       *subsurface.System
	EdgeAO           *edgeao.System
}

// RegisterECSSystems registers all ECS systems with the World in the correct order.
// It also wires up inter-system dependencies.
func RegisterECSSystems(world *engine.World, deps *SystemDependencies) {
	// Core simulation systems (order matters for dependency resolution)
	world.AddSystem(deps.Spatial)
	world.AddSystem(deps.Animation)
	world.AddSystem(deps.Motion)
	world.AddSystem(deps.Status)
	world.AddSystem(deps.Combo)
	world.AddSystem(deps.LootDrop)
	world.AddSystem(deps.Feedback)
	world.AddSystem(deps.Defense)
	world.AddSystem(deps.Lighting)
	world.AddSystem(deps.BossPhase)
	world.AddSystem(deps.HazardECS)
	world.AddSystem(deps.Faction)
	world.AddSystem(deps.Stat)

	// Weather system (adds components to world)
	world.AddSystem(deps.Weather)
	deps.Weather.AddWeatherToWorld(world)

	// Collision and navigation systems
	world.AddSystem(deps.Sliding)
	world.AddSystem(deps.Equipment)
	world.AddSystem(deps.Positional)
	world.AddSystem(deps.AdaptiveAI)
	world.AddSystem(deps.AO)

	// Connect AO system to spatial index
	deps.AO.SetSpatialGrid(deps.Spatial.GetGrid())

	// Projectile system with dependencies
	world.AddSystem(deps.Projectile)
	deps.Projectile.SetSpatialGrid(deps.Spatial.GetGrid())
	deps.Projectile.SetParticleSpawner(deps.Particle)
	deps.Projectile.SetFeedbackProvider(deps.Feedback)

	// Environment and combat systems
	world.AddSystem(deps.BiomeMaterial)
	world.AddSystem(deps.Trap)
	world.AddSystem(deps.QuestLoot)
	world.AddSystem(deps.DamageState)

	// Damage visual effects with dependencies
	world.AddSystem(deps.Dmgfx)
	deps.Dmgfx.SetParticleSpawner(deps.Particle)
	deps.Dmgfx.SetFeedbackProvider(deps.Feedback)
	deps.Projectile.SetDamageVisualProvider(deps.Dmgfx)

	// Territory and visual systems
	world.AddSystem(deps.Territory)
	world.AddSystem(deps.Outline)
	world.AddSystem(deps.RimLight)
	world.AddSystem(deps.AttackTrail)
	world.AddSystem(deps.Telegraph)
	world.AddSystem(deps.LootVisual)
	world.AddSystem(deps.HealthBar)
	world.AddSystem(deps.Fog)
	world.AddSystem(deps.Parallax)
	world.AddSystem(deps.WeaponAnim)
	world.AddSystem(deps.WallTex)
	world.AddSystem(deps.ParticleRenderer)
	world.AddSystem(deps.DamageNumber)
	world.AddSystem(deps.WeaponVisual)
	world.AddSystem(deps.AttackAnim)
	world.AddSystem(deps.StatusFX)
	world.AddSystem(deps.StatusTint)
	world.AddSystem(deps.PlayerSprite)
	world.AddSystem(deps.Crosshair)
	world.AddSystem(deps.ImpactBurst)
	world.AddSystem(deps.EntityLabel)
	world.AddSystem(deps.ProximityUI)
	world.AddSystem(deps.Subsurface)
	world.AddSystem(deps.EdgeAO)
}

// ConnectSlidingSystem wires the sliding system to the spatial index.
// Must be called after spatial system is initialized.
func ConnectSlidingSystem(sliding *collision.SlidingSystem, spatial *spatial.System) {
	sliding.SetSpatialIndex(spatial.GetGrid())
}
