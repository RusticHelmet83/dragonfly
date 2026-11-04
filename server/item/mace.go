package item

import "math"

// Mace is a heavy melee weapon that deals massive damage when the attacker
// smashes into an entity from a height. A smash attack happens when the
// attacker has fallen more than MaceSmashMinFall blocks. The smash logic itself
// lives in player.Player.AttackEntity.
type Mace struct{}

const (
	// MaceSmashMinFall is the minimum fall distance in blocks for an attack with
	// a mace to be a smash attack.
	MaceSmashMinFall = 1.5
	// MaceHeavySmashFall is the fall distance in blocks above which a smash
	// attack is a heavy one, which has a different sound.
	MaceHeavySmashFall = 5.0
	// MaceSmashRadius is the radius around the entity hit by a smash attack in
	// which other entities are knocked back.
	MaceSmashRadius = 2.5
)

// MaxCount ...
func (Mace) MaxCount() int {
	return 1
}

// AttackDamage ...
func (Mace) AttackDamage() float64 {
	return 6
}

// IsMace marks the item as a mace.
func (Mace) IsMace() bool {
	return true
}

// DurabilityInfo ...
func (Mace) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    501,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 1,
		BreakDurability:  1,
	}
}

// EnchantmentValue ...
func (Mace) EnchantmentValue() int {
	return 15
}

// EncodeItem ...
func (Mace) EncodeItem() (name string, meta int16) {
	return "minecraft:mace", 0
}

// MaceSmashBonus returns the bonus damage a smash attack with a mace deals for
// the fall distance passed. The first 3 blocks deal 4 damage each, the next 5
// blocks 2 damage each and every block after that 1 damage. perBlockAddend is
// added for every block fallen and is used by the Density enchantment.
func MaceSmashBonus(fallDistance, perBlockAddend float64) float64 {
	remaining := math.Max(fallDistance, 0)
	bonus := 0.0
	for _, tier := range [...]struct{ blocks, damage float64 }{{3, 4}, {5, 2}, {math.Inf(1), 1}} {
		n := math.Min(remaining, tier.blocks)
		bonus += n * (tier.damage + perBlockAddend)
		remaining -= n
		if remaining <= 0 {
			break
		}
	}
	return bonus
}
