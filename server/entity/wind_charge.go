package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	// windChargeBurstRadius is the radius of the burst of a wind charge.
	windChargeBurstRadius = 2.4
	// windChargeBurstForce is the knockback force at the centre of the burst. It
	// falls off linearly to 0 at the edge of the burst.
	windChargeBurstForce = 1.2
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
	Drag:       0.01,
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
	return cube.Box(-0.3125, 0, -0.3125, 0.3125, 0.3125, 0.3125)
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
// The burst pushes every entity within windChargeBurstRadius away from the
// impact point. The force falls off linearly with the distance and is added to
// the current velocity of the entity, with an extra upwards push. This is the
// behaviour of the featuremace implementation; there is no special launch for
// the owner of the wind charge.
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

	r := windChargeBurstRadius
	for other := range tx.EntitiesWithin(cube.Box(-r, -r, -r, r, r, r).Translate(impact)) {
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

		delta := other.Position().Sub(impact)
		distance := delta.Len()
		if distance > r {
			continue
		}
		direction := mgl64.Vec3{0, 1, 0}
		if distance > 0 {
			direction = delta.Mul(1 / distance)
		}
		strength := windChargeBurstForce * (1 - distance/r)

		mover, ok := other.(interface {
			SetVelocity(mgl64.Vec3)
			Velocity() mgl64.Vec3
		})
		if !ok {
			continue
		}
		knockback := direction.Mul(strength)
		knockback[1] += strength * 0.5

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
