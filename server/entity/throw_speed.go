package entity

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// ThrowSpeedOf returns the throw speed that is set for the type of projectile passed (one of the Projectile*
// constants) when it is thrown by owner: the tuning of the type (see SetProjectileTuning), overridden field by
// field by ProjectileTuner. If ok is false no speed is set, and the thrower should use the vanilla speed. For arrows
// the speed is a multiplier of the launch velocity instead.
//
// Items and anything else that launches a projectile read the speed with this function and throw at it directly,
// so the projectile starts at exactly that speed. PrepareProjectile does not rescale the launch velocity.
func ThrowSpeedOf(owner world.Entity, entityType string) (speed float64, ok bool) {
	t := ProjectileTuningOf(entityType)
	if owner != nil && ProjectileTuner != nil {
		if o := ProjectileTuner(owner, entityType); o != nil {
			t = t.overriddenBy(*o)
		}
	}
	if t.Speed == nil {
		return 0, false
	}
	return *t.Speed, true
}

func init() {
	// Lets the items of the item package, which can't import this package, read the configured throw speed.
	item.ThrowSpeed = ThrowSpeedOf
}
