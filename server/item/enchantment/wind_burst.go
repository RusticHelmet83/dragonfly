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

// Radius returns the radius in blocks of the explosion of Wind Burst. It is 3.5 for every level.
func (windBurst) Radius(level int) float64 {
	return 3.5
}

// KnockbackMultiplier returns the knockback multiplier of the explosion of Wind Burst at the level passed: 1.2, 1.75
// and 2.2 for levels I, II and III, like in Java Edition. The attacker is launched with about this speed (in blocks
// per tick) straight up.
func (windBurst) KnockbackMultiplier(level int) float64 {
	switch {
	case level <= 1:
		return 1.2
	case level == 2:
		return 1.75
	case level == 3:
		return 2.2
	}
	return 1.5 + 0.35*float64(level-1)
}

// LaunchHeight returns the approximate height in blocks the attacker is launched into the air with the level of
// Wind Burst passed.
//
// Deprecated: the launch is now the knockback of an explosion, see KnockbackMultiplier. The value is only
// informative.
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
