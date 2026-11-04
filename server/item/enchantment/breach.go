package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Breach is an enchantment for maces that reduces the effectiveness of the
// armour of the entity that is attacked.
var Breach breach

type breach struct{}

// Name ...
func (breach) Name() string {
	return "Breach"
}

// MaxLevel ...
func (breach) MaxLevel() int {
	return 4
}

// Cost ...
func (breach) Cost(level int) (int, int) {
	minCost := 15 + (level-1)*9
	return minCost, minCost + 50
}

// Rarity ...
func (breach) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// ArmourPenetration returns the fraction (0-1) of the armour points of the
// victim that is ignored by an attack with the level of Breach passed. Every
// level ignores 15% of the armour.
func (breach) ArmourPenetration(level int) float64 {
	return min(0.15*float64(level), 1)
}

// CompatibleWithEnchantment ...
func (breach) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Density
}

// CompatibleWithItem ...
func (breach) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Mace)
	return ok
}
