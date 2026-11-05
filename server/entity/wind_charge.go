package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	// windChargeBurstRadius is the explosion radius of a wind charge. Entities
	// within twice this radius are affected.
	windChargeBurstRadius = 1.2
	// windChargeBurstForce is the knockback multiplier of the burst. The power
	// falls off linearly to 0 at twice the radius.
	windChargeBurstForce = 1.22
)

// NewWindCharge creates a wind charge entity at a position with an owner
// entity.
func NewWindCharge(opts world.EntitySpawnOpts, owner world.Entity) *world.EntityHandle {
	conf := windChargeConf
	conf.Owner = owner.H()
	return opts.New(WindChargeType, conf)
}

var windChargeConf = ProjectileBehaviourConfig{
	Gravity:    0,
	Drag:       0,
	Damage:     -1,
	Collidable: true,
	Hit:        windChargeHit,
}

// WindChargeType is a world.EntityType implementation for wind charges.
var WindChargeType windChargeType

type windChargeType struct{}

func (windChargeType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (windChargeType) EncodeEntity() string { return "minecraft:wind_charge_projectile" }
func (windChargeType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.15625, -0.15, -0.15625, 0.15625, 0.1625, 0.15625)
}
func (windChargeType) DecodeNBT(_ map[string]any, data *world.EntityData) {
	data.Data = windChargeConf.New()
}
func (windChargeType) EncodeNBT(*world.EntityData) map[string]any { return nil }

// windChargeHit is called when a wind charge hits a block or an entity.
func windChargeHit(e *Ent, tx *world.Tx, target trace.Result) {
	switch r := target.(type) {
	case trace.BlockResult:
		windChargeBurst(e, tx, blastOrigin(r), nil)
	case trace.EntityResult:
		other := r.Entity()
		if proj, ok := other.(*Ent); ok {
			if b, ok := proj.Behaviour().(*ProjectileBehaviour); ok && b.conf.Collidable {
				// A wind charge that hits another collidable projectile, such as
				// an ender pearl, cancels out: the other projectile lands
				// immediately (so a pearl teleports its owner) and no burst
				// happens.
				if !proj.H().Closed() {
					b.TriggerHit(proj, tx, target)
					_ = proj.Close()
				}
				return
			}
		}
		windChargeBurst(e, tx, r.Position(), other)
	}
}

// HandleWindChargeCollision should be called by the Hit function of any
// collidable projectile that is not a wind charge (such as a custom ender
// pearl) with the target it hit. If the target is a wind charge, the wind
// charge bursts at the point of collision and is consumed, and true is
// returned. The projectile itself should then still resolve its own hit, so an
// ender pearl still teleports its owner.
func HandleWindChargeCollision(tx *world.Tx, target trace.Result) bool {
	r, ok := target.(trace.EntityResult)
	if !ok || r.Entity().H().Type() != WindChargeType {
		return false
	}
	wind, ok := r.Entity().(*Ent)
	if !ok || wind.H().Closed() {
		return false
	}
	windChargeBurst(wind, tx, r.Position(), nil)
	_ = wind.Close()
	return true
}

// windChargeBurst performs the burst of the wind charge e at the impact point
// passed. The entity direct, if non-nil, is the entity that was hit directly
// and takes damage.
//
// This mirrors Java's ServerExplosion.hurtEntities with a radius of 1.2 and a
// knockback multiplier of 1.22: every entity within twice the radius of the
// impact (measured from its feet) is pushed from the impact towards its eye
// position with a power of (1 - dist) * exposure * 1.22, where dist is the
// distance divided by twice the radius and exposure is the fraction of its
// bounding box with a clear line to the impact. There is no additional upwards
// push and the owner gets no special treatment.
func windChargeBurst(e *Ent, tx *world.Tx, impact mgl64.Vec3, direct world.Entity) {
	tx.PlaySound(impact, sound.WindCharge{})
	tx.AddParticle(impact, particle.WindCharge{})

	var owner world.Entity
	if b, ok := e.Behaviour().(*ProjectileBehaviour); ok {
		owner, _ = b.Owner().Entity(tx)
	}

	if living, ok := direct.(Living); ok && !living.Dead() {
		living.Hurt(1, ProjectileDamageSource{Projectile: e, Owner: owner})
	}

	doubleRadius := windChargeBurstRadius * 2
	search := cube.Box(-doubleRadius-1, -doubleRadius-1, -doubleRadius-1, doubleRadius+1, doubleRadius+1, doubleRadius+1).Translate(impact)
	for other := range tx.EntitiesWithin(search) {
		if other.H().Type() == WindChargeType {
			continue
		}
		living, ok := other.(Living)
		if !ok || living.Dead() {
			continue
		}
		if fs, ok := other.(world.FakeSpectator); ok && fs.IsFakeSpectating() {
			continue
		}
		if g, ok := other.(interface{ GameMode() world.GameMode }); ok && !g.GameMode().HasCollision() {
			continue
		}
		// Creative players that are flying take no knockback.
		if f, ok := other.(interface{ Flying() bool }); ok && f.Flying() {
			if g, ok := other.(interface{ GameMode() world.GameMode }); ok && g.GameMode().CreativeInventory() {
				continue
			}
		}

		dist := other.Position().Sub(impact).Len() / doubleRadius
		if dist > 1 {
			continue
		}
		mover, ok := other.(interface {
			SetVelocity(mgl64.Vec3)
			Velocity() mgl64.Vec3
		})
		if !ok {
			continue
		}

		origin := other.Position()
		if eyed, ok := other.(interface{ EyeHeight() float64 }); ok {
			origin[1] += eyed.EyeHeight()
		}
		delta := origin.Sub(impact)
		if delta.LenSqr() == 0 {
			continue
		}
		direction := delta.Normalize()

		box := other.H().Type().BBox(other).Translate(other.Position())
		power := (1 - dist) * windChargeSeenPercent(tx, impact, box) * windChargeBurstForce
		knockback := direction.Mul(power)

		// Fall damage is negated above the height of the burst, but the fall
		// distance keeps counting so a mace smash is still possible.
		if neg, ok := other.(interface{ NegateFallDamageAbove(y float64) }); ok {
			neg.NegateFallDamageAbove(min(impact[1], other.Position()[1]) - 0.01)
		}
		if reset, ok := other.(interface{ ResetFallDistance() }); ok {
			reset.ResetFallDistance()
		}
		mover.SetVelocity(mover.Velocity().Add(knockback))
	}
}

// windChargeSeenPercent is a port of Java's ServerExplosion.getSeenPercent. It
// returns the fraction of sample points in the box that have an unobstructed
// line to the centre of the burst.
func windChargeSeenPercent(tx *world.Tx, center mgl64.Vec3, box cube.BBox) float64 {
	size := box.Max().Sub(box.Min())
	xs, ys, zs := 1/(size[0]*2+1), 1/(size[1]*2+1), 1/(size[2]*2+1)
	xOffset := (1 - math.Floor(1/xs)*xs) / 2
	zOffset := (1 - math.Floor(1/zs)*zs) / 2

	var hits, count int
	for xx := 0.0; xx <= 1.0; xx += xs {
		for yy := 0.0; yy <= 1.0; yy += ys {
			for zz := 0.0; zz <= 1.0; zz += zs {
				from := mgl64.Vec3{
					box.Min()[0] + xx*size[0] + xOffset,
					box.Min()[1] + yy*size[1],
					box.Min()[2] + zz*size[2] + zOffset,
				}
				clear := true
				trace.TraverseBlocks(from, center, func(pos cube.Pos) bool {
					if _, ok := trace.BlockIntercept(pos, tx, tx.Block(pos), from, center); ok {
						clear = false
						return false
					}
					return true
				})
				if clear {
					hits++
				}
				count++
			}
		}
	}
	return float64(hits) / float64(count)
}

// blastOrigin returns the point a wind charge bursts at when it hits a block:
// the impact point pushed slightly out of the face that was hit.
func blastOrigin(r trace.BlockResult) mgl64.Vec3 {
	var normal mgl64.Vec3
	switch r.Face() {
	case cube.FaceDown:
		normal = mgl64.Vec3{0, -1, 0}
	case cube.FaceUp:
		normal = mgl64.Vec3{0, 1, 0}
	case cube.FaceNorth:
		normal = mgl64.Vec3{0, 0, -1}
	case cube.FaceSouth:
		normal = mgl64.Vec3{0, 0, 1}
	case cube.FaceWest:
		normal = mgl64.Vec3{-1, 0, 0}
	case cube.FaceEast:
		normal = mgl64.Vec3{1, 0, 0}
	}
	return r.Position().Add(normal.Mul(0.25))
}
