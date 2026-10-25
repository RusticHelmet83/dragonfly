package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// PressurePlate is a model for pressure plate blocks.
type PressurePlate struct {
	// Powered is whether the pressure plate is in its pressed down state.
	Powered bool
}

// BBox returns no collision box: pressure plates are non-solid, entities rest on
// whatever is beneath them.
func (p PressurePlate) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return nil
}

// FaceSolid always returns false.
func (PressurePlate) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}
