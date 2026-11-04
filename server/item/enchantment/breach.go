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

// ArmourPenetration returns the amount (as a fraction of damage, 0-1) that the
// damage reduction of the armour of the victim is lowered by an attack with the
// level of Breach passed. Every level lowers the reduction by 15 percentage
// points, so armour that reduces damage by 50% only reduces it by 35% against
// Breach I. Damage taken = damage * (1 - max(0, reduction - 0.15 * level)).
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
