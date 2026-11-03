package ui

import (
	"math"
	"sync"

	mu "github.com/egoist/mygo/ui"
)

// This file ports the shape geometry of Flutter's material_new_shapes
// package, which FlClash uses for the dashboard's country badge, the start
// button and the card decoration. The shapes are rounded polygons: a list of
// vertices, each rounded by a circle of a given radius.

// Point is a point in shape space.
type Point struct{ X, Y float64 }

func (p Point) Add(o Point) Point   { return Point{p.X + o.X, p.Y + o.Y} }
func (p Point) Sub(o Point) Point   { return Point{p.X - o.X, p.Y - o.Y} }
func (p Point) Mul(k float64) Point { return Point{p.X * k, p.Y * k} }
func (p Point) Div(k float64) Point { return Point{p.X / k, p.Y / k} }
func (p Point) Dot(o Point) float64 { return p.X*o.X + p.Y*o.Y }
func (p Point) Rotate90() Point     { return Point{-p.Y, p.X} }
func (p Point) Distance() float64   { return math.Hypot(p.X, p.Y) }
func (p Point) Direction() Point    { return p.Div(p.Distance()) }
func (p Point) Angle() float64      { return math.Atan2(p.Y, p.X) }
func (p Point) Rotate(deg float64) Point {
	r := deg * math.Pi / 180
	c, s := math.Cos(r), math.Sin(r)
	return Point{p.X*c - p.Y*s, p.X*s + p.Y*c}
}

// cubic is one cubic Bézier curve.
type cubic struct {
	// The four points: start, first control, second control, end.
	p [4]Point
}

func (c cubic) anchor0() Point { return c.p[0] }
func (c cubic) anchor1() Point { return c.p[3] }

// cornerRounding says how much of a vertex is rounded away.
type cornerRounding struct {
	radius    float64
	smoothing float64
}

var (
	roundUnrounded = cornerRounding{}
	round50        = cornerRounding{radius: 0.5}
	round100       = cornerRounding{radius: 1}
	round476       = cornerRounding{radius: 0.476}
)

const distanceEpsilon = 1e-6

// roundedPolygon is a polygon whose vertices are rounded off.
type roundedPolygon struct {
	vertices []Point
	rounding []cornerRounding
	center   Point
	features [][]cubic
}

// roundedCorner holds the geometry of one rounded vertex.
type roundedCorner struct {
	p0, p1, p2 Point
	rounding   cornerRounding

	d1, d2           Point
	cornerRadius     float64
	smoothing        float64
	cosAngle         float64
	sinAngle         float64
	expectedRoundCut float64
}

func newRoundedCorner(p0, p1, p2 Point, r cornerRounding) roundedCorner {
	c := roundedCorner{p0: p0, p1: p1, p2: p2, rounding: r}
	v01 := p0.Sub(p1)
	v21 := p2.Sub(p1)
	d01 := v01.Distance()
	d21 := v21.Distance()
	if d01 <= 0 || d21 <= 0 {
		c.d1 = Point{}
		c.d2 = Point{}
		c.cosAngle = 0
		c.sinAngle = 0
		return c
	}
	c.d1 = v01.Div(d01)
	c.d2 = v21.Div(d21)
	c.cornerRadius = r.radius
	c.smoothing = r.smoothing
	c.cosAngle = c.d1.Dot(c.d2)
	c.sinAngle = math.Sqrt(math.Max(0, 1-c.cosAngle*c.cosAngle))
	// tan(A/2) = sinA/(1+cosA) = radius/cut
	if c.sinAngle > 1e-3 {
		c.expectedRoundCut = c.cornerRadius * (c.cosAngle + 1) / c.sinAngle
	}
	return c
}

// expectedCut is the cut the smoothing would want, on top of the round cut.
func (c roundedCorner) expectedCut() float64 {
	return (1 + c.smoothing) * c.expectedRoundCut
}

// fromVertices builds the shape: every vertex rounded by its own rounding.
func (p *roundedPolygon) fromVertices(vertices []Point, rounding []cornerRounding, center Point) {
	p.vertices = vertices
	p.rounding = rounding
	p.center = center

	n := len(vertices)
	corners := make([]roundedCorner, n)
	for i := 0; i < n; i++ {
		r := roundUnrounded
		if rounding != nil {
			r = rounding[i]
		}
		corners[i] = newRoundedCorner(vertices[(i+n-1)%n], vertices[i], vertices[(i+1)%n], r)
	}

	// How much of each side's cut the sides can actually hold: rounding
	// first, smoothing with what is left.
	type cutAdjust struct{ round, smooth float64 }
	adjusts := make([]cutAdjust, n)
	for i := 0; i < n; i++ {
		expectedRound := corners[i].expectedRoundCut + corners[(i+1)%n].expectedRoundCut
		expectedCut := corners[i].expectedCut() + corners[(i+1)%n].expectedCut()
		side := vertices[i].Sub(vertices[(i+1)%n]).Distance()
		switch {
		case expectedRound > side:
			adjusts[i] = cutAdjust{side / expectedRound, 0}
		case expectedCut > side:
			adjusts[i] = cutAdjust{1, (side - expectedRound) / (expectedCut - expectedRound)}
		default:
			adjusts[i] = cutAdjust{1, 1}
		}
	}

	curves := make([][]cubic, n)
	for i := 0; i < n; i++ {
		allowed := [2]float64{}
		for delta := 0; delta <= 1; delta++ {
			a := adjusts[(i+n-1+delta)%n]
			allowed[delta] = corners[i].expectedRoundCut*a.round +
				(corners[i].expectedCut()-corners[i].expectedRoundCut)*a.smooth
		}
		curves[i] = corners[i].cubics(allowed[0], allowed[1])
	}

	// The corners, and the straight edges between them.
	features := make([][]cubic, 0, n)
	for i := 0; i < n; i++ {
		features = append(features, curves[i])
		next := curves[(i+1)%n]
		features = append(features, []cubic{straightLine(curves[i][len(curves[i])-1].anchor1(), next[0].anchor0())})
	}
	p.features = features
}

// cubics returns the three curves of one rounded corner: the flank, the arc
// and the other flank.
func (c roundedCorner) cubics(allowedCut0, allowedCut1 float64) []cubic {
	allowedCut := math.Min(allowedCut0, allowedCut1)
	if c.expectedRoundCut < distanceEpsilon || allowedCut < distanceEpsilon ||
		c.cornerRadius < distanceEpsilon {
		return []cubic{straightLine(c.p1, c.p1)}
	}

	actualRoundCut := math.Min(allowedCut, c.expectedRoundCut)
	actualSmoothing0 := c.actualSmoothing(allowedCut0)
	actualSmoothing1 := c.actualSmoothing(allowedCut1)
	actualR := c.cornerRadius * actualRoundCut / c.expectedRoundCut
	centerDistance := math.Hypot(actualR, actualRoundCut)
	center := c.p1.Add(c.d1.Add(c.d2).Mul(0.5).Direction().Mul(centerDistance))

	intersection0 := c.p1.Add(c.d1.Mul(actualRoundCut))
	intersection2 := c.p1.Add(c.d2.Mul(actualRoundCut))

	flanking0 := c.flankingCurve(actualRoundCut, actualSmoothing0, c.p1, c.p0, intersection0, intersection2, center, actualR)
	flanking2 := c.flankingCurve(actualRoundCut, actualSmoothing1, c.p1, c.p2, intersection2, intersection0, center, actualR)

	return []cubic{flanking0, circularArc(center, flanking0.anchor1(), flanking2.anchor0()), flanking2}
}

// actualSmoothing is how much of the smoothing the room allows.
func (c roundedCorner) actualSmoothing(allowedCut float64) float64 {
	switch {
	case allowedCut > c.expectedCut():
		return c.smoothing
	case allowedCut > c.expectedRoundCut:
		return c.smoothing * (allowedCut - c.expectedRoundCut) / (c.expectedCut() - c.expectedRoundCut)
	default:
		return 0
	}
}

// flankingCurve connects the straight side with the round arc.
func (c roundedCorner) flankingCurve(actualRoundCut, actualSmoothing float64,
	corner, sideStart, intersection, otherIntersection, center Point, actualR float64) cubic {

	sideDirection := sideStart.Sub(corner).Direction()
	curveStart := corner.Add(sideDirection.Mul(actualRoundCut * (1 + actualSmoothing)))

	// The part of the circle the smoothing leaves out.
	mid := intersection.Add(otherIntersection).Mul(0.5)
	p := interpolate(intersection, mid, actualSmoothing)

	curveEnd := center.Add(Point{p.X - center.X, p.Y - center.Y}.Direction().Mul(actualR))

	circleTangent := curveEnd.Sub(center).Rotate90()
	anchorEnd := lineIntersection(sideStart, sideDirection, curveEnd, circleTangent)
	end := intersection
	if anchorEnd != nil {
		end = *anchorEnd
	}
	anchorStart := curveStart.Add(end.Mul(2)).Mul(1.0 / 3.0)

	return cubic{[4]Point{curveStart, anchorStart, end, curveEnd}}
}

// interpolate moves from a to b by t.
func interpolate(a, b Point, t float64) Point {
	return Point{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t}
}

// lineIntersection returns where the two lines meet, or nil.
func lineIntersection(p0, d0, p1, d1 Point) *Point {
	rotatedD1 := d1.Rotate90()
	den := d0.Dot(rotatedD1)
	if math.Abs(den) < distanceEpsilon {
		return nil
	}
	num := p1.Sub(p0).Dot(rotatedD1)
	if math.Abs(den) < distanceEpsilon*math.Abs(num) {
		return nil
	}
	k := num / den
	pt := p0.Add(d0.Mul(k))
	return &pt
}

// straightLine is a degenerate cubic, a straight segment.
func straightLine(a, b Point) cubic {
	return cubic{[4]Point{a, a, b, b}}
}

// circularArc is the arc of a circle between two points.
func circularArc(center, p0, p1 Point) cubic {
	d0 := Point{p0.X - center.X, p0.Y - center.Y}.Direction()
	d1 := Point{p1.X - center.X, p1.Y - center.Y}.Direction()
	rotated0 := d0.Rotate90()
	rotated1 := d1.Rotate90()
	clockwise := rotated0.Dot(Point{p1.X - center.X, p1.Y - center.Y}) >= 0
	cosa := d0.Dot(d1)
	if cosa > 0.999 {
		return straightLine(p0, p1)
	}
	k := Point{p0.X - center.X, p0.Y - center.Y}.Distance() * 4 / 3 *
		(math.Sqrt(2*(1-cosa)) - math.Sqrt(1-cosa*cosa)) / (1 - cosa) *
		map[bool]float64{true: 1, false: -1}[clockwise]
	return cubic{[4]Point{
		p0,
		Point{p0.X + rotated0.X*k, p0.Y + rotated0.Y*k},
		Point{p1.X - rotated1.X*k, p1.Y - rotated1.Y*k},
		p1,
	}}
}

// normalized scales the shape into the unit square, centred.
func (p *roundedPolygon) normalized() {
	bounds := p.calculateBounds()
	width := bounds[2] - bounds[0]
	height := bounds[3] - bounds[1]
	side := math.Max(width, height)
	if side <= 0 {
		return
	}
	offsetX := (side-width)/2 - bounds[0]
	offsetY := (side-height)/2 - bounds[1]
	for i := range p.vertices {
		p.vertices[i] = Point{
			(p.vertices[i].X + offsetX) / side,
			(p.vertices[i].Y + offsetY) / side,
		}
	}
	for i := range p.features {
		for j := range p.features[i] {
			for k := range p.features[i][j].p {
				p.features[i][j].p[k] = Point{
					(p.features[i][j].p[k].X + offsetX) / side,
					(p.features[i][j].p[k].Y + offsetY) / side,
				}
			}
		}
	}
	p.center = Point{(p.center.X + offsetX) / side, (p.center.Y + offsetY) / side}
}

// calculateBounds is the bounding box of the shape's curves.
func (p *roundedPolygon) calculateBounds() [4]float64 {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, feature := range p.features {
		for _, c := range feature {
			for _, pt := range c.p {
				minX = math.Min(minX, pt.X)
				minY = math.Min(minY, pt.Y)
				maxX = math.Max(maxX, pt.X)
				maxY = math.Max(maxY, pt.Y)
			}
		}
	}
	return [4]float64{minX, minY, maxX, maxY}
}

// ---- the shapes FlClash uses ----------------------------------------------

// starPolygon is the star shape of the cookie figures: vertices alternate
// between the outer and the inner radius.
func starPolygon(numVertices int, innerRadius float64, rounding cornerRounding) *roundedPolygon {
	vertices := make([]Point, 0, numVertices*2)
	roundings := make([]cornerRounding, 0, numVertices*2)
	for i := 0; i < numVertices; i++ {
		a := math.Pi / float64(numVertices) * 2 * float64(i)
		vertices = append(vertices, Point{math.Cos(a), math.Sin(a)})
		roundings = append(roundings, rounding)
		b := math.Pi / float64(numVertices) * float64(2*i+1)
		vertices = append(vertices, Point{math.Cos(b) * innerRadius, math.Sin(b) * innerRadius})
		roundings = append(roundings, rounding)
	}
	p := &roundedPolygon{}
	p.fromVertices(vertices, roundings, Point{0, 0})
	return p
}

// customPolygon repeats points around a centre, mirroring them when asked.
func customPolygon(points []Point, roundings []cornerRounding, reps int, center Point, mirroring bool) *roundedPolygon {
	actual := repeatPoints(points, roundings, reps, center, mirroring)
	p := &roundedPolygon{}
	p.fromVertices(actual.points, actual.roundings, center)
	return p
}

type pointSet struct {
	points    []Point
	roundings []cornerRounding
}

// repeatPoints spreads the points around the centre.
func repeatPoints(points []Point, roundings []cornerRounding, reps int, center Point, mirroring bool) pointSet {
	var out pointSet
	if mirroring {
		measures := make([]struct {
			angle    float64
			distance float64
		}, len(points))
		for i, p := range points {
			off := p.Sub(center)
			measures[i].angle = off.Angle()
			measures[i].distance = off.Distance()
		}
		actualReps := reps * 2
		sectionAngle := 2 * math.Pi / float64(actualReps)
		for r := 0; r < actualReps; r++ {
			for index := range points {
				i := index
				if r%2 == 1 {
					i = len(points) - 1 - index
				}
				if i > 0 || r%2 == 0 {
					var a float64
					if r%2 == 0 {
						a = sectionAngle*float64(r) + measures[i].angle
					} else {
						a = sectionAngle*float64(r) + sectionAngle - measures[i].angle + 2*measures[0].angle
					}
					final := Point{math.Cos(a), math.Sin(a)}.Mul(measures[i].distance).Add(center)
					out.points = append(out.points, final)
					out.roundings = append(out.roundings, roundings[i])
				}
			}
		}
		return out
	}
	np := len(points)
	for i := 0; i < np*reps; i++ {
		rot := float64(i/np) * 360 / float64(reps)
		out.points = append(out.points, points[i%np].Sub(center).Rotate(rot).Add(center))
		out.roundings = append(out.roundings, roundings[i%np])
	}
	return out
}

// The three shapes FlClash's dashboard uses.
var (
	// shapeCookie9 is the nine-sided cookie the start button grows into.
	shapeCookie9 = func() *roundedPolygon {
		p := starPolygon(9, 0.8, round50)
		p.normalized()
		return p
	}()

	// shapeCookie12 is the twelve-sided cookie behind the card decoration.
	shapeCookie12 = func() *roundedPolygon {
		p := starPolygon(12, 0.8, round50)
		p.normalized()
		return p
	}()

	// shapeClover4 is the four-leaf clover the country badge sits on.
	shapeClover4 = func() *roundedPolygon {
		p := customPolygon(
			[]Point{{0.5, 0.074}, {0.725, -0.099}},
			[]cornerRounding{roundUnrounded, round476},
			4, Point{0.5, 0.5}, true,
		)
		p.normalized()
		return p
	}()
)

// morphCache holds the shapes the start button morphs through, so that the
// polygon is built once per step rather than in every frame.
var morphCache sync.Map

// morphPolygon is the shape of the start button at a step of its morph: a
// nine-pointed star whose inner radius grows as the core starts, which turns
// a circle into a cookie-9.
func morphPolygon(step int) *roundedPolygon {
	if v, ok := morphCache.Load(step); ok {
		return v.(*roundedPolygon)
	}
	progress := float32(step) / float32(morphSteps)
	inner := 1 - 0.2*float64(progress)
	p := starPolygon(9, inner, round50)
	p.normalized()
	morphCache.Store(step, p)
	return p
}

// morphSteps is how many shapes the button morphs through.
const morphSteps = 24

// morphShape is the path of the start button at a progress of its morph.
func morphShape(progress float32, x, y, size float32) *mu.Path {
	step := int(math.Round(float64(progress) * morphSteps))
	if step < 0 {
		step = 0
	}
	if step > morphSteps {
		step = morphSteps
	}
	return morphPolygon(step).path(x, y, size, morphTurn(progress))
}

// path builds the mygo path of a shape, scaled into the box. The features
// run one after another around the outline, so they make one closed path.
func (p *roundedPolygon) path(x, y, size float32, rotation float32) *mu.Path {
	path := &mu.Path{}
	cos, sin := math.Cos(float64(rotation)), math.Sin(float64(rotation))
	// Rotate about the shape's centre, then scale into the box.
	transform := func(pt Point) [2]float32 {
		dx := pt.X - 0.5
		dy := pt.Y - 0.5
		rx := dx*cos - dy*sin
		ry := dx*sin + dy*cos
		return [2]float32{x + float32(rx+0.5)*size, y + float32(ry+0.5)*size}
	}
	first := true
	for _, feature := range p.features {
		for _, c := range feature {
			start := transform(c.p[0])
			if first {
				path.MoveTo(start[0], start[1])
				first = false
			}
			c1 := transform(c.p[1])
			c2 := transform(c.p[2])
			end := transform(c.p[3])
			path.CubeTo(c1[0], c1[1], c2[0], c2[1], end[0], end[1])
		}
	}
	path.Close()
	return path
}

// drawShape draws a shape into a box.
func (p *roundedPolygon) draw(pt *mu.Painter, x, y, size float32, rotation float32, color mu.Color) {
	pt.FillPath(p.path(x, y, size, rotation), color)
}
