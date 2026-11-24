package item

import "github.com/df-mc/dragonfly/server/world"

// ThrowSpeed, if set, returns the throw speed that is configured for the type of projectile passed (one of the
// entity.Projectile* names, such as "minecraft:snowball") when it is thrown by owner. It is set by the entity
// package, which owns the projectile tuning and can't be imported here. If ok is false, no speed is configured and
// the item throws at its vanilla speed. For arrows the speed is a multiplier of the launch velocity of the bow or
// crossbow instead.
var ThrowSpeed func(owner world.Entity, entityType string) (speed float64, ok bool)

// throwSpeed returns the speed a projectile of the type passed is thrown at by owner: the configured speed, see
// ThrowSpeed, or vanilla if none is configured. Items multiply their throw direction with it, so the projectile is
// launched at that speed right away.
func throwSpeed(owner world.Entity, entityType string, vanilla float64) float64 {
	if ThrowSpeed != nil {
		if s, ok := ThrowSpeed(owner, entityType); ok {
			return s
		}
	}
	return vanilla
}
