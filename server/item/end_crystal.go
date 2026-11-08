package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndCrystal is an item that can be placed on obsidian or bedrock to spawn an End crystal entity.
type EndCrystal struct{}

// endCrystalSupport represents a block that supports an End crystal.
type endCrystalSupport interface {
	SupportsEndCrystal() bool
}

// UseOnBlock places an End crystal on top of the clicked block if it is
// obsidian or bedrock, the block above it is air and no entities intersect the
// space the crystal is placed in. The face clicked is ignored, and neighbouring
// blocks never matter: only the support block and the space above it do.
func (e EndCrystal) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	support, ok := tx.Block(pos).(endCrystalSupport)
	if !ok || !support.SupportsEndCrystal() {
		return false
	}

	above := pos.Side(cube.FaceUp)
	if above.OutOfBounds(tx.Range()) {
		return false
	}
	// Like vanilla, only the block directly above the support block has to be
	// empty. The crystal is allowed to overlap solid blocks higher up.
	if tx.Block(above) != air() {
		return false
	}

	// Vanilla anchors the entity check at the block above the support block and
	// spans two blocks upwards. Any intersecting entity blocks placement,
	// including the placing player.
	box := cube.Box(0, 0, 0, 1, 2, 1).Translate(above.Vec3())
	for entity := range tx.EntitiesWithin(box.Grow(2)) {
		if entity.H().Type().BBox(entity).Translate(entity.Position()).IntersectsWith(box) {
			return false
		}
	}

	opts := world.EntitySpawnOpts{Position: pos.Side(cube.FaceUp).Vec3Middle()}
	tx.AddEntity(tx.World().EntityRegistry().Config().EndCrystal(opts))
	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (EndCrystal) EncodeItem() (name string, meta int16) {
	return "minecraft:end_crystal", 0
}
