package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// WindBurst is an enchantment for maces that launches the attacker back up into
// the air after a smash attack.
var WindBurst windBurst

type windBurst struct{}

// Name ...
func (windBurst) Name() string {
	return "Wind Burst"
}

// MaxLevel ...
func (windBurst) MaxLevel() int {
	return 3
}

// Cost ...
func (windBurst) Cost(level int) (int, int) {
	minCost := 15 + (level-1)*9
	return minCost, minCost + 50
}

// Rarity ...
func (windBurst) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// LaunchHeight returns the height in blocks the attacker is launched into the
// air with the level of Wind Burst passed. Level I launches roughly 8 blocks,
// level II 16 and level III 24.
func (windBurst) LaunchHeight(level int) float64 {
	return 8 * float64(level)
}

// CompatibleWithEnchantment ...
func (windBurst) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (windBurst) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Mace)
	return ok
}
