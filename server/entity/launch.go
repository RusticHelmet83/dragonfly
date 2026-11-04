package entity

import "math"

const (
	// playerGravity and playerDrag are the per-tick vertical movement constants
	// of a Bedrock player: v' = (v - gravity) * drag.
	playerGravity = 0.08
	playerDrag    = 0.98
)

// LaunchVelocityForHeight returns the initial upwards velocity (in blocks per
// tick) a player needs for its apex to be exactly height blocks above its
// starting position, using the Bedrock player movement model. It is used to
// translate "launch the player N blocks into the air" (wind charges, Wind
// Burst) into an actual velocity, instead of relying on hand-tuned constants.
func LaunchVelocityForHeight(height float64) float64 {
	if height <= 0 {
		return 0
	}
	lo, hi := 0.0, 10.0
	for i := 0; i < 64; i++ {
		mid := (lo + hi) / 2
		if apexHeight(mid) < height {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// apexHeight simulates the vertical movement of a player launched with the
// velocity v0 and returns the highest point reached.
func apexHeight(v0 float64) float64 {
	var y, v float64 = 0, v0
	for i := 0; i < 1000 && v > 0; i++ {
		y += v
		v = (v - playerGravity) * playerDrag
	}
	return math.Max(y, 0)
}
