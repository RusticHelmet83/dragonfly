package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type motionOwner struct {
	world.Entity
	motion   mgl64.Vec3
	onGround bool
}

func (o motionOwner) ThrowMotion() mgl64.Vec3 { return o.motion }
func (o motionOwner) OnGround() bool          { return o.onGround }

func TestPrepareProjectile(t *testing.T) {
	const typ = ProjectileEnderPearl
	t.Cleanup(func() { SetProjectileTuning(typ, ProjectileTuning{}) })

	owner := motionOwner{motion: mgl64.Vec3{0.2, 0.3, -0.1}}
	launch := func() (world.EntitySpawnOpts, ProjectileBehaviourConfig) {
		return world.EntitySpawnOpts{Velocity: mgl64.Vec3{2, 0, 0}}, ProjectileBehaviourConfig{Gravity: 0.03, Drag: 0.01}
	}

	// Default: the motion of the thrower is added, the vertical part only while airborne.
	opts, conf := launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2.2, 0.3, -0.1}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("airborne: got %v, want %v", opts.Velocity, want)
	}
	owner.onGround = true
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2.2, 0, -0.1}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("grounded: got %v, want %v", opts.Velocity, want)
	}

	// Disabled per type.
	SetProjectileTuning(typ, ProjectileTuning{InheritMotion: Bool(false)})
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2, 0, 0}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("disabled: got %v, want %v", opts.Velocity, want)
	}

	// Speed is read by the thrower (see ThrowSpeedOf) and not applied here: the launch velocity only gets the motion
	// of the thrower added. Gravity and drag may be set to 0.
	SetProjectileTuning(typ, ProjectileTuning{Speed: Float(1), Gravity: Float(0), Drag: Float(0.5)})
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2.2, 0, -0.1}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("speed: got %v, want %v", opts.Velocity, want)
	}
	if conf.Gravity != 0 || conf.Drag != 0.5 {
		t.Fatalf("gravity/drag: got %v/%v", conf.Gravity, conf.Drag)
	}

	// The per-owner hook wins over the per-type setting.
	ProjectileInheritMotion = func(world.Entity, string) (bool, bool) { return false, true }
	t.Cleanup(func() { ProjectileInheritMotion = nil })
	SetProjectileTuning(typ, ProjectileTuning{InheritMotion: Bool(true)})
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2, 0, 0}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("hook: got %v, want %v", opts.Velocity, want)
	}

	// The per-owner tuner overrides the global tuning field by field.
	SetProjectileTuning(typ, ProjectileTuning{Speed: Float(1), Gravity: Float(0.2)})
	ProjectileInheritMotion = nil
	ProjectileTuner = func(world.Entity, string) *ProjectileTuning {
		return &ProjectileTuning{Speed: Float(3), InheritMotion: Bool(false)}
	}
	t.Cleanup(func() { ProjectileTuner = nil })
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, typ)
	if want := (mgl64.Vec3{2, 0, 0}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("tuner: got %v, want %v", opts.Velocity, want)
	}
	if conf.Gravity != 0.2 {
		t.Fatalf("tuner kept global gravity: got %v", conf.Gravity)
	}

	// ThrowSpeedOf gives the speed the thrower launches at: the tuner wins over the global tuning, and if no speed
	// is set at all there is none.
	if s, ok := ThrowSpeedOf(owner, typ); !ok || s != 3 {
		t.Fatalf("throw speed with tuner: got %v, %v, want 3, true", s, ok)
	}
	ProjectileTuner = nil
	if s, ok := ThrowSpeedOf(owner, typ); !ok || s != 1 {
		t.Fatalf("throw speed global: got %v, %v, want 1, true", s, ok)
	}
	SetProjectileTuning(typ, ProjectileTuning{})
	if s, ok := ThrowSpeedOf(owner, typ); ok {
		t.Fatalf("throw speed unset: got %v, want none", s)
	}

	// The fishing hook does not inherit by default.
	ProjectileInheritMotion = nil
	opts, conf = launch()
	PrepareProjectile(&opts, &conf, owner, ProjectileFishingHook)
	if want := (mgl64.Vec3{2, 0, 0}); !opts.Velocity.ApproxEqual(want) {
		t.Fatalf("hook default: got %v, want %v", opts.Velocity, want)
	}
}
