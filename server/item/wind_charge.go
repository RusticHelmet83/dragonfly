package item

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

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
	pos := eyePosition(user)
	rot := user.Rotation()
	opts := world.EntitySpawnOpts{Position: pos, Velocity: rot.Vec3().Mul(2), Rotation: rot}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(pos, sound.Raw{Name: "random.bow", Volume: 0.5, Pitch: 0.33 + rand.Float32()*0.17})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}
