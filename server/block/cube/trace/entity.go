package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// defaultEntityGrow is the amount an entity's bounding box is grown by on every
// side when a ray trace is performed against it.
const defaultEntityGrow = 0.3

// EntityResult is the result of a ray trace collision with an entities bounding box.
type EntityResult struct {
	bb   cube.BBox
	pos  mgl64.Vec3
	face cube.Face

	entity world.Entity
}

// BBox returns the entities bounding box that was collided with.
func (r EntityResult) BBox() cube.BBox {
	return r.bb
}

// Position ...
func (r EntityResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face ...
func (r EntityResult) Face() cube.Face {
	return r.face
}

// Entity returns the entity that was collided with.
func (r EntityResult) Entity() world.Entity {
	return r.entity
}

// EntityIntercept performs a ray trace and calculates the point on the entities bounding box's edge nearest to the start position
// that the ray collided with.
// EntityIntercept returns an EntityResult with the entity collided with and with the colliding vector closest to the start position,
// if no colliding point was found, a zero BlockResult is returned ok is false.
func EntityIntercept(e world.Entity, start, end mgl64.Vec3) (result EntityResult, ok bool) {
	return EntityInterceptGrow(e, start, end, defaultEntityGrow)
}

// EntityInterceptGrow is like EntityIntercept, but allows a more forgiving hit
// area: if the ray misses the entity's bounding box grown by the default 0.3,
// it is tested again against the bounding box grown by grow. A grow that is not
// larger than the default behaves exactly like EntityIntercept.
//
// A hit that only happens thanks to the larger box is an "assisted" hit. Its
// position is the point on the ray closest to the centre of the entity (rather
// than a point on the surface of the enlarged box), so whatever reacts to the
// hit happens at the spot where the two actually passed closest. A ray that
// starts inside the enlarged box also counts as a hit, even if it never leaves
// it.
func EntityInterceptGrow(e world.Entity, start, end mgl64.Vec3, grow float64) (result EntityResult, ok bool) {
	box := e.H().Type().BBox(e).Translate(e.Position())

	bb := box.Grow(defaultEntityGrow)
	if r, hit := BBoxIntercept(bb, start, end); hit {
		return EntityResult{bb: bb, pos: r.Position(), face: r.Face(), entity: e}, true
	}
	if grow <= defaultEntityGrow {
		return EntityResult{}, false
	}

	bb = box.Grow(grow)
	r, hit := BBoxIntercept(bb, start, end)
	if !hit && !bboxContains(bb, start) {
		return EntityResult{}, false
	}
	face := cube.FaceUp
	if hit {
		face = r.Face()
	}
	centre := box.Min().Add(box.Max()).Mul(0.5)
	return EntityResult{bb: bb, pos: closestPointOnSegment(start, end, centre), face: face, entity: e}, true
}

// bboxContains reports if the point p lies within bb.
func bboxContains(bb cube.BBox, p mgl64.Vec3) bool {
	min, max := bb.Min(), bb.Max()
	return p[0] >= min[0] && p[0] <= max[0] &&
		p[1] >= min[1] && p[1] <= max[1] &&
		p[2] >= min[2] && p[2] <= max[2]
}

// closestPointOnSegment returns the point on the line segment from start to end
// that is closest to p.
func closestPointOnSegment(start, end, p mgl64.Vec3) mgl64.Vec3 {
	d := end.Sub(start)
	l2 := d.LenSqr()
	if l2 == 0 {
		return start
	}
	t := p.Sub(start).Dot(d) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return start.Add(d.Mul(t))
}
