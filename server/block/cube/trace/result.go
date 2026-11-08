package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"math"
)

// Result represents the result of a ray trace collision with a bounding box.
type Result interface {
	// BBox returns the bounding box collided with.
	BBox() cube.BBox
	// Position returns where the ray first collided with the bounding box.
	Position() mgl64.Vec3
	// Face returns the face of the bounding box that was collided on.
	Face() cube.Face
}

type EntityFilter func(iter.Seq[world.Entity]) iter.Seq[world.Entity]

// Perform performs a ray trace between start and end, checking if any blocks or entities collided with the
// ray. The physics.BBox that's passed is used for checking if any entity within the bounding box collided
// with the ray.
func Perform(start, end mgl64.Vec3, tx *world.Tx, box cube.BBox, filter EntityFilter) (hit Result, ok bool) {
	return PerformGrow(start, end, tx, box, filter, nil)
}

// PerformGrow is like Perform, but grow, if non-nil, is called for every entity
// the ray may hit and returns how far that entity's bounding box should be
// grown for the hit test (see EntityInterceptGrow). Returning a value not
// larger than the default of 0.3 keeps the normal behaviour for that entity.
// Entities are only considered if they are within the box passed grown by its
// own size, so grow values above roughly 1.0 may not take full effect.
func PerformGrow(start, end mgl64.Vec3, tx *world.Tx, box cube.BBox, filter EntityFilter, grow func(world.Entity) float64) (hit Result, ok bool) {
	// Check if there's any blocks that we may collide with.
	TraverseBlocks(start, end, func(pos cube.Pos) (cont bool) {
		b := tx.Block(pos)

		// Check if we collide with the block's model.
		if result, ok := BlockIntercept(pos, tx, b, start, end); ok {
			hit = result
			end = hit.Position()
			return false
		}
		return true
	})

	// Now check for any entities that we may collide with.
	dist := math.MaxFloat64
	bb := box.Translate(start).Extend(end.Sub(start))
	entities := tx.EntitiesWithin(bb.Grow(8.0))
	if filter != nil {
		entities = filter(entities)
	}
	for entity := range entities {
		if !entity.H().Type().BBox(entity).Translate(entity.Position()).IntersectsWith(bb) {
			continue
		}
		// Check if we collide with the entities bounding box.
		g := defaultEntityGrow
		if grow != nil {
			if v := grow(entity); v > g {
				g = v
			}
		}
		result, ok := EntityInterceptGrow(entity, start, end, g)
		if !ok {
			continue
		}

		if distance := start.Sub(result.Position()).LenSqr(); distance < dist {
			dist = distance
			hit = result
		}
	}

	return hit, hit != nil
}
