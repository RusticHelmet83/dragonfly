package entity

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type endCrystalBehaviour struct {
	showBase      bool
	beamTarget    cube.Pos
	hasBeamTarget bool
	explosionSize float64
}

func (b endCrystalBehaviour) Apply(data *world.EntityData) {
	data.Data = b
}

// Tick continuously generates fire at the End crystal's position while in the
// End, if the block at that position is air and the block below it is not air.
func (endCrystalBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if tx.World().Dimension() == world.End {
		pos := cube.PosFromVec3(e.Position())
		if _, air := tx.Block(pos.Side(cube.FaceDown)).(block.Air); !air {
			if _, air := tx.Block(pos).(block.Air); air {
				fire := block.Fire{}
				tx.SetBlock(pos, fire, nil)
				tx.ScheduleBlockUpdate(pos, fire, time.Duration(30+rand.IntN(10))*time.Second/20)
			}
		}
	}
	return nil
}

// Explode makes the End crystal explode itself when hit by another explosion,
// causing a chain reaction.
func (b endCrystalBehaviour) Explode(e *Ent, src world.ExplosionSource, impact float64) {
	if impact <= 0 {
		return
	}
	explodeEndCrystal(e, b.explosionSize, explosionSourceOwner(src))
}

// Hurt makes the End crystal explode when damaged by any source, even by
// damage that deals no health. Void damage removes it without an explosion.
func (b endCrystalBehaviour) Hurt(e *Ent, damage float64, src world.DamageSource) (float64, bool) {
	damage = max(damage, 0)
	if _, ok := src.(VoidDamageSource); ok {
		_ = e.Close()
		return damage, true
	}
	explodeEndCrystal(e, b.explosionSize, damageSourceOwner(src))
	return damage, true
}

func (endCrystalBehaviour) Immobile() bool {
	return true
}

func (b endCrystalBehaviour) ShowBase() bool {
	return b.showBase
}

func (b endCrystalBehaviour) BeamTarget() (cube.Pos, bool) {
	return b.beamTarget, b.hasBeamTarget
}

// EndCrystalExploded, if set, is called every time an End crystal actually
// detonates (regardless of whether the explosion damages/kills anyone),
// letting callers tally a precise "crystals exploded" stat per owner.
var EndCrystalExploded func(owner world.Entity)

// explodeEndCrystal closes the End crystal and creates a non-incendiary
// explosion at its base, if the crystal was not closed yet.
func explodeEndCrystal(e *Ent, explosionSize float64, owner world.Entity) {
	if _, ok := e.H().Entity(e.tx); !ok {
		return
	}
	_ = e.Close()
	if EndCrystalExploded != nil {
		EndCrystalExploded(owner)
	}
	block.ExplosionConfig{
		SuppressUnderwaterImpact: true,
	}.Explode(e.tx, world.EntityExplosionSource{
		Entity:        e,
		ExplosionSize: explosionSize,
		Owner:         owner,
	})
}

func damageSourceOwner(src world.DamageSource) world.Entity {
	switch s := src.(type) {
	case AttackDamageSource:
		return s.Attacker
	case ProjectileDamageSource:
		return s.Owner
	default:
		return nil
	}
}

func explosionSourceOwner(src world.ExplosionSource) world.Entity {
	switch s := src.(type) {
	case world.EntityExplosionSource:
		return s.Owner
	case world.BlockExplosionSource:
		return s.Owner
	default:
		return nil
	}
}
