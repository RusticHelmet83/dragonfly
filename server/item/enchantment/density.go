package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Density is an enchantment for maces that increases the damage of a smash
// attack for every block fallen.
var Density density

type density struct{}

// Name ...
func (density) Name() string {
	return "Density"
}

// MaxLevel ...
func (density) MaxLevel() int {
	return 5
}

// Cost ...
func (density) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*8
	return minCost, minCost + 20
}

// Rarity ...
func (density) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// PerBlockAddend returns the additional damage dealt by a smash attack for
// every block the attacker fell.
func (density) PerBlockAddend(level int) float64 {
	return 0.5 * float64(level)
}

// CompatibleWithEnchantment ...
func (density) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Breach
}

// CompatibleWithItem ...
func (density) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Mace)
	return ok
}
