package sound

// SpearLunge is a sound played when a player lunges forward with a spear that
// has the Lunge enchantment.
//
// It is sent as a LevelSoundEvent (item.enchant.lunge1, item.enchant.lunge2 or
// item.enchant.lunge3) and NOT as a named sound. That is why playing
// item.spear.lunge with /playsound (or Raw/Custom) is silent: that name is only
// the Java Edition sound event, and Bedrock has no sound definition with it.
// The event is tied to the entity that performs the lunge, so it is sent with
// the player entity type.
type SpearLunge struct {
	// Level is the level of the Lunge enchantment that was used. It selects
	// the variant of the sound that is played (1, 2 or 3). Values outside that
	// range are clamped into it.
	Level int

	sound
}
