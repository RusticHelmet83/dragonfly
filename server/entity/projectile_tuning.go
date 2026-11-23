package entity

import (
	"sync"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Entity type names of the projectiles that can be tuned with SetProjectileTuning.
const (
	ProjectileSnowball           = "minecraft:snowball"
	ProjectileEgg                = "minecraft:egg"
	ProjectileEnderPearl         = "minecraft:ender_pearl"
	ProjectileSplashPotion       = "minecraft:splash_potion"
	ProjectileLingeringPotion    = "minecraft:lingering_potion"
	ProjectileBottleOfEnchanting = "minecraft:xp_bottle"
	ProjectileWindCharge         = "minecraft:wind_charge_projectile"
	ProjectileArrow              = "minecraft:arrow"
	ProjectileFishingHook        = "minecraft:fishing_hook"
)

// ProjectileTuning holds overrides for how a type of projectile is launched and how it flies. A nil field leaves
// the default of the projectile untouched.
type ProjectileTuning struct {
	// Speed is the launch speed in blocks per tick. It does not include the motion of the thrower, see
	// InheritMotion. For arrows it is a multiplier of the launch velocity instead (1 is the default).
	Speed *float64
	// Gravity is the downwards acceleration of the projectile in blocks per tick squared. 0 disables gravity.
	Gravity *float64
	// Drag is the fraction of its velocity the projectile loses every tick. 0 disables drag.
	Drag *float64
	// InheritMotion decides if the movement of the thrower is added to the launch velocity of the projectile, like
	// Java Edition does for everything that is thrown or shot: the horizontal movement always, the vertical
	// movement only while the thrower is not on the ground. It defaults to true for every projectile except the
	// fishing hook, which does not inherit the movement of the thrower in Java Edition.
	InheritMotion *bool
}

// Float returns a pointer to v, for the fields of ProjectileTuning.
func Float(v float64) *float64 { return &v }

// Bool returns a pointer to v, for the fields of ProjectileTuning.
func Bool(v bool) *bool { return &v }

var projectileTuning = struct {
	sync.RWMutex
	m map[string]ProjectileTuning
}{m: map[string]ProjectileTuning{}}

// SetProjectileTuning sets the tuning of the type of projectile passed (one of the Projectile* constants),
// replacing the previous tuning of the type. Passing the zero ProjectileTuning restores all defaults. It applies
// to projectiles that are created afterwards. It is safe to call at any time.
func SetProjectileTuning(entityType string, t ProjectileTuning) {
	projectileTuning.Lock()
	defer projectileTuning.Unlock()
	projectileTuning.m[entityType] = t
}

// ProjectileTuningOf returns the tuning currently set for the type of projectile passed.
func ProjectileTuningOf(entityType string) ProjectileTuning {
	projectileTuning.RLock()
	defer projectileTuning.RUnlock()
	return projectileTuning.m[entityType]
}

// ProjectileInheritMotion, if set, is called when a projectile is created, and overrides whether the motion of the
// owner is added to the launch velocity of the projectile (see ProjectileTuning.InheritMotion) for this owner. If ok
// is false the tuning of the type decides. It can be used to turn inheriting off for a game mode, for example.
var ProjectileInheritMotion func(owner world.Entity, entityType string) (inherit, ok bool)

// throwMotionSource is implemented by owners that know how fast they are moving. Players implement it.
type throwMotionSource interface {
	// ThrowMotion returns the movement of the owner in blocks per tick.
	ThrowMotion() mgl64.Vec3
	// OnGround reports if the owner is standing on the ground.
	OnGround() bool
}

// PrepareProjectile applies the tuning of the type of projectile, ProjectileModifier and the motion of the
// owner to the spawn options and config of a projectile that is about to be created, in that order. The speed
// of the tuning is applied before the motion of the owner is added, so it only changes the throw speed.
// Projectile constructors, including ones outside of this package, should call it with the entity type name.
func PrepareProjectile(opts *world.EntitySpawnOpts, conf *ProjectileBehaviourConfig, owner world.Entity, entityType string) {
	t := ProjectileTuningOf(entityType)
	if t.Speed != nil {
		scaleLaunch(opts, *t.Speed, entityType)
	}
	if t.Gravity != nil {
		conf.Gravity = *t.Gravity
	}
	if t.Drag != nil {
		conf.Drag = *t.Drag
	}
	if owner == nil {
		return
	}

	if ProjectileModifier != nil {
		if speed, gravity, drag, ok := ProjectileModifier(owner, entityType); ok {
			if speed > 0 {
				scaleLaunch(opts, speed, entityType)
			}
			if gravity > 0 {
				conf.Gravity = gravity
			}
			if drag > 0 {
				conf.Drag = drag
			}
		}
	}

	inherit := entityType != ProjectileFishingHook
	if t.InheritMotion != nil {
		inherit = *t.InheritMotion
	}
	if ProjectileInheritMotion != nil {
		if v, ok := ProjectileInheritMotion(owner, entityType); ok {
			inherit = v
		}
	}
	if !inherit {
		return
	}
	if src, ok := owner.(throwMotionSource); ok {
		m := src.ThrowMotion()
		if src.OnGround() {
			m[1] = 0
		}
		opts.Velocity = opts.Velocity.Add(m)
	}
}

// scaleLaunch changes the launch velocity of a projectile to the speed passed: a multiplier for arrows, and the
// exact speed for everything else.
func scaleLaunch(opts *world.EntitySpawnOpts, speed float64, entityType string) {
	if entityType == ProjectileArrow {
		opts.Velocity = opts.Velocity.Mul(speed)
		return
	}
	if l := opts.Velocity.Len(); l > 0 {
		opts.Velocity = opts.Velocity.Mul(speed / l)
	}
}
