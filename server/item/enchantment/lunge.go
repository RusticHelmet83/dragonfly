package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Lunge is an enchantment for spears that propels the wielder forward
// horizontally whenever they perform a jab attack, at the cost of hunger and an
// additional point of durability.
var Lunge lunge

type lunge struct{}

// Name ...
func (lunge) Name() string {
	return "Lunge"
}

// MaxLevel ...
func (lunge) MaxLevel() int {
	return 3
}

// Cost ...
func (lunge) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*8
	return minCost, minCost + 20
}

// Rarity ...
func (lunge) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Boost returns the instant horizontal momentum boost, in blocks per tick, that
// Lunge gives the wielder at the level passed.
func (lunge) Boost(level int) float64 {
	return 0.458 * float64(level)
}

// Exhaustion returns the exhaustion the wielder gains for a single lunge at the
// level passed. Every 4 exhaustion drains one saturation/hunger point.
func (lunge) Exhaustion(level int) float64 {
	return 4 * float64(level)
}

// CompatibleWithEnchantment ...
func (lunge) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (lunge) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && t.ToolType() == item.TypeSpear
}
