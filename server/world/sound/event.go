package sound

// Event is a Sound identified by a name registered in the client's built-in
// LevelSoundEvent list (e.g. "mace.smash_ground", "wind_charge.burst").
// Unlike Raw, which plays through the PlaySound packet and is always
// distance-attenuated, Event goes through the LevelSoundEvent packet, which
// can be played at full volume regardless of distance when sent through
// Player.PlaySound.
type Event struct {
	// Name is the registered LevelSoundEvent identifier.
	Name string

	sound
}
