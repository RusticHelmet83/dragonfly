package particle

// Raw is a Particle that spawns an arbitrary named vanilla particle effect
// by its full identifier, e.g. "minecraft:wind_explosion_emitter".
type Raw struct {
	particle
	// Name is the full particle effect identifier.
	Name string
}
