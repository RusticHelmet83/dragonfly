package sound

// WindCharge is a sound played when a wind charge bursts.
type WindCharge struct{ sound }

// MaceSmash is a sound played when a mace smash attack lands.
type MaceSmash struct {
	// Ground specifies if the attacked entity was standing on the ground. If
	// false, the sound of a smash on an airborne entity is played.
	Ground bool
	// Heavy specifies if the smash was a heavy one, which is the case when the
	// attacker fell more than 5 blocks.
	Heavy bool

	sound
}
