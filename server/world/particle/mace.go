package particle

import "github.com/df-mc/dragonfly/server/world"

// WindCharge is a particle shown when a wind charge bursts, or when a mace with
// the Wind Burst enchantment smashes an entity.
type WindCharge struct{ particle }

// SmashAttack is a particle shown when a mace smash attack lands on the ground.
// It shows a ring of dust coloured like the Block passed.
type SmashAttack struct {
	particle

	// Block is the block the dust is taken from, usually the block the
	// attacked entity is standing on.
	Block world.Block
}
