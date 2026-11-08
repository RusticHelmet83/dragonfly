package item

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WindChargeThrowSpeed is the launch speed of a thrown wind charge in blocks
// per tick. Vanilla uses 1.5. The Versai fork uses 2.6, slightly above the
// custom ender pearl throw speed (practice/handler: Rotation.Vec3() * 2.35).
// Wind charges have no gravity and no drag, so this is also their constant
// speed for the whole flight. Because the pearl is thrown first, the charge
// starts behind it and has to make up the head start; the pearl slows down
// from drag and falls from gravity, so a slightly faster charge can catch it.
// Simulated: at 2.6 a charge thrown up to ~8 ticks after a pearl can catch it
// at 30 degrees upward, and about 5 ticks after at 20 degrees. 2.35 only
// manages that for shallow angles with a gap of ~3 ticks or less.
var WindChargeThrowSpeed = 2.6

// WindChargeSpread scales the random aiming spread of a thrown wind charge.
// 1.0 is the vanilla Java uncertainty (0.0172275 per axis, triangle
// distributed). Set it to 0 for perfectly deterministic throws, which makes
// pearl catching consistent.
var WindChargeSpread = 1.0

// WindCharge is a throwable item that bursts on impact, knocking back every
// entity close to it and launching the thrower if it stands on top of the burst.
type WindCharge struct{}

// MaxCount ...
func (WindCharge) MaxCount() int {
	return 64
}

// Cooldown ...
func (WindCharge) Cooldown() time.Duration {
	return time.Millisecond * 500
}

// Use ...
func (WindCharge) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().WindCharge
	// Java spawns the charge at the x/z of the thrower, at eye height, and
	// shoots it with a power of 1.5 and an uncertainty of 1.0. The power is
	// WindChargeThrowSpeed here and the uncertainty is scaled by WindChargeSpread.
	pos := eyePosition(user)
	rot := user.Rotation()
	// Triangle-distributed spread of 0.0172275 * uncertainty per axis.
	spread := func() float64 { return 0.0172275 * WindChargeSpread * (rand.Float64() - rand.Float64()) }
	dir := rot.Vec3().Normalize().Add(mgl64.Vec3{spread(), spread(), spread()}).Normalize()
	vel := dir.Mul(WindChargeThrowSpeed)
	// The known movement of the thrower is added, ignoring vertical movement
	// while the thrower is on the ground.
	if mover, ok := user.(interface{ Velocity() mgl64.Vec3 }); ok {
		m := mover.Velocity()
		if g, ok := user.(interface{ OnGround() bool }); ok && g.OnGround() {
			m[1] = 0
		}
		vel = vel.Add(m)
	}
	opts := world.EntitySpawnOpts{Position: pos, Velocity: vel, Rotation: rot}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(pos, sound.Raw{Name: "random.bow", Volume: 0.5, Pitch: 0.33 + rand.Float32()*0.17})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}
