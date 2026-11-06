package entity

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
)

type (
	// AttackDamageSource is used for damage caused by other entities, for
	// example when a player attacks another player.
	// TODO(2): Give the Warden's sonic boom a dedicated unblockable source when Warden support is added.
	AttackDamageSource struct {
		// Attacker holds the attacking entity. The entity may be a player or
		// any other entity.
		Attacker world.Entity
		// ArmourPenetration is the amount (0-1) by which the damage reduction of
		// the victim's armour is lowered when the damage is reduced by armour. It
		// is set by the Breach enchantment of the mace.
		ArmourPenetration float64
		// AttributeSwap is true for the follow-up hit of an attribute swap that
		// happens after the weapon hit it upgrades. It lands inside the attack
		// immunity of the victim on purpose, so that only the extra damage is
		// dealt. Handlers should not discard it as an immunity-window hit.
		AttributeSwap bool
	}

	// VoidDamageSource is used for damage caused by an entity being in the
	// void.
	VoidDamageSource struct{}

	// SuffocationDamageSource is used for damage caused by an entity
	// suffocating in a block.
	SuffocationDamageSource struct{}

	// DrowningDamageSource is used for damage caused by an entity drowning in
	// water.
	DrowningDamageSource struct{}

	// FallDamageSource is used for damage caused by falling.
	FallDamageSource struct{}

	// GlideDamageSource is used for damage caused by gliding into a block.
	GlideDamageSource struct{}

	// LightningDamageSource is used for damage caused by being struck by
	// lightning.
	LightningDamageSource struct{}

	// ProjectileDamageSource is used for damage caused by a projectile.
	ProjectileDamageSource struct {
		// Projectile and Owner are the world.Entity that dealt the damage and
		// the one that fired the projectile respectively.
		Projectile, Owner world.Entity
		// Piercing is true if the projectile bypasses shields.
		Piercing bool
	}

	// ExplosionDamageSource is used for damage caused by an explosion.
	ExplosionDamageSource struct {
		// Source is the source of the explosion that dealt the damage.
		Source world.ExplosionSource
	}
)

func (FallDamageSource) ReducedByArmour() bool     { return false }
func (FallDamageSource) ReducedByResistance() bool { return true }
func (FallDamageSource) Fire() bool                { return false }
func (FallDamageSource) AffectedByEnchantment(e item.EnchantmentType) bool {
	return e == enchantment.FeatherFalling
}
func (FallDamageSource) IgnoreTotem() bool              { return false }
func (GlideDamageSource) ReducedByArmour() bool         { return false }
func (GlideDamageSource) ReducedByResistance() bool     { return true }
func (GlideDamageSource) Fire() bool                    { return false }
func (GlideDamageSource) IgnoreTotem() bool             { return false }
func (LightningDamageSource) ReducedByArmour() bool     { return true }
func (LightningDamageSource) ReducedByResistance() bool { return true }
func (LightningDamageSource) Fire() bool                { return false }
func (LightningDamageSource) IgnoreTotem() bool         { return false }
func (a AttackDamageSource) ArmourPenetrationFraction() float64 {
	return a.ArmourPenetration
}

// AttributeSwapHit returns true if the damage is the follow-up hit of an
// attribute swap, see AttackDamageSource.AttributeSwap.
func (a AttackDamageSource) AttributeSwapHit() bool       { return a.AttributeSwap }
func (AttackDamageSource) ReducedByArmour() bool          { return true }
func (AttackDamageSource) ReducedByResistance() bool      { return true }
func (AttackDamageSource) Fire() bool                     { return false }
func (AttackDamageSource) IgnoreTotem() bool              { return false }
func (VoidDamageSource) ReducedByResistance() bool        { return false }
func (VoidDamageSource) ReducedByArmour() bool            { return false }
func (VoidDamageSource) Fire() bool                       { return false }
func (VoidDamageSource) IgnoreTotem() bool                { return true }
func (SuffocationDamageSource) ReducedByResistance() bool { return false }
func (SuffocationDamageSource) ReducedByArmour() bool     { return false }
func (SuffocationDamageSource) Fire() bool                { return false }
func (SuffocationDamageSource) IgnoreTotem() bool         { return false }
func (DrowningDamageSource) ReducedByResistance() bool    { return false }
func (DrowningDamageSource) ReducedByArmour() bool        { return false }
func (DrowningDamageSource) Fire() bool                   { return false }
func (DrowningDamageSource) IgnoreTotem() bool            { return false }
func (ProjectileDamageSource) ReducedByResistance() bool  { return true }
func (ProjectileDamageSource) ReducedByArmour() bool      { return true }
func (ProjectileDamageSource) Fire() bool                 { return false }
func (ProjectileDamageSource) AffectedByEnchantment(e item.EnchantmentType) bool {
	return e == enchantment.ProjectileProtection
}
func (ProjectileDamageSource) IgnoreTotem() bool        { return false }
func (ExplosionDamageSource) ReducedByResistance() bool { return true }
func (ExplosionDamageSource) ReducedByArmour() bool     { return true }
func (ExplosionDamageSource) Fire() bool                { return false }
func (ExplosionDamageSource) AffectedByEnchantment(e item.EnchantmentType) bool {
	return e == enchantment.BlastProtection
}
func (ExplosionDamageSource) IgnoreTotem() bool { return false }

// ShieldBlockInfo returns the position of an attack if its attacker is available.
func (s AttackDamageSource) ShieldBlockInfo() (world.ShieldBlockInfo, bool) {
	if s.Attacker == nil {
		return world.ShieldBlockInfo{}, false
	}
	return world.ShieldBlockInfo{Origin: s.Attacker.Position()}, true
}

// ShieldBlockInfo returns the position of a projectile or its owner.
func (s ProjectileDamageSource) ShieldBlockInfo() (world.ShieldBlockInfo, bool) {
	if s.Piercing {
		return world.ShieldBlockInfo{}, false
	}
	if s.Projectile != nil {
		return world.ShieldBlockInfo{Origin: s.Projectile.Position(), BlockWhenImmune: true, BlockZeroDamage: true}, true
	}
	if s.Owner != nil {
		return world.ShieldBlockInfo{Origin: s.Owner.Position(), BlockWhenImmune: true, BlockZeroDamage: true}, true
	}
	return world.ShieldBlockInfo{}, false
}

// ShieldBlockInfo returns the origin and source of a shield-blockable explosion.
func (s ExplosionDamageSource) ShieldBlockInfo() (world.ShieldBlockInfo, bool) {
	if s.Source == nil {
		return world.ShieldBlockInfo{}, false
	}
	if source, ok := s.Source.(world.ShieldBlockInfoSource); ok {
		return source.ShieldBlockInfo()
	}
	return world.ShieldBlockInfo{Origin: s.Source.Position(), BlockWhenImmune: true}, true
}
