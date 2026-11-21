package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// GAME.EXE's CRT 004031F2 selects 53-bit x87 precision and the
// gameplay loop 0043E2C1 selects round-toward-zero. Binary32 input
// coordinates bound the intermediates inside binary64's exponent range.
// Correct only rounding, without changing the calling thread's FPU.
//
//go:noinline
func pointOnLineAddChop53_57C8A0(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	remainder := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && remainder < 0 || sum < 0 && remainder > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func pointOnLineMulChop53_57C8A0(a, b float64) float64 {
	product := float64(a * b)
	// FMA recovers the exact product residual only; it does not fuse any
	// of the original projection's separate multiply/add instructions.
	remainder := math.FMA(a, b, -product)
	if product > 0 && remainder < 0 || product < 0 && remainder > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}

//go:noinline
func pointOnLineDivChop53_57C8A0(a, b float64) float64 {
	quotient := float64(a / b)
	// The sign of (a - quotient*b)/b identifies a nearest-even quotient
	// rounded away from zero. The bounded residual cannot underflow.
	remainder := math.FMA(-quotient, b, a)
	if remainder != 0 && !math.IsNaN(remainder) &&
		(quotient > 0 && math.Signbit(remainder) != math.Signbit(b) || quotient < 0 && math.Signbit(remainder) == math.Signbit(b)) {
		return math.Nextafter(quotient, 0)
	}
	return quotient
}

func pointOnLineSpill57C8A0(value float64) float32 {
	stored := float32(value)
	// Includes chopped subnormal stores and finite overflow (max-finite,
	// not infinity). Exact infinities and NaNs retain their normal stores.
	if math.Abs(float64(stored)) > math.Abs(value) {
		stored = math.Nextafter32(stored, 0)
	}
	return stored
}

// PointOnTheLineInto projects p onto the line and checks the segment bounds.
func PointOnTheLineInto(line *types.Rectf, a2 types.Pointf, out *types.Pointf) bool {
	p1, p2 := line.Min, line.Max
	// 0057C8AF..0057C909 uses gameplay precision 53 / ToZero, not
	// binary32 coordinate subtraction or unspilled binary64 projection.
	// X retains both its delta and dot product; Y reloads their FST stores
	// and spills its quotient before adding the segment origin.
	dpx := pointOnLineAddChop53_57C8A0(float64(p2.X), -float64(p1.X))
	dpy := pointOnLineAddChop53_57C8A0(float64(p2.Y), -float64(p1.Y))
	storedY := float64(pointOnLineSpill57C8A0(dpy))
	d2 := float64(pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(
		pointOnLineMulChop53_57C8A0(dpy, storedY), pointOnLineMulChop53_57C8A0(dpx, dpx))))
	cy := pointOnLineAddChop53_57C8A0(float64(a2.Y), -float64(p1.Y))
	cx := pointOnLineAddChop53_57C8A0(float64(a2.X), -float64(p1.X))
	dd := pointOnLineAddChop53_57C8A0(pointOnLineMulChop53_57C8A0(cy, storedY), pointOnLineMulChop53_57C8A0(cx, dpx))
	storedDot := float64(pointOnLineSpill57C8A0(dd))
	xQuotient := pointOnLineDivChop53_57C8A0(pointOnLineMulChop53_57C8A0(dpx, dd), d2)
	yQuotient := float64(pointOnLineSpill57C8A0(pointOnLineDivChop53_57C8A0(pointOnLineMulChop53_57C8A0(storedDot, storedY), d2)))
	// The C callers can overlap out with the segment. Read the Y origin
	// after storing X, and read the bounds after both coordinates are stored.
	out.X = pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(xQuotient, float64(line.Min.X)))
	out.Y = pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(yQuotient, float64(line.Min.Y)))
	minX, maxX := projectionBounds(line.Min.X, line.Max.X)
	minY, maxY := projectionBounds(line.Min.Y, line.Max.Y)
	// 0057C95C/69/87 test C0|C3, whereas 0057C978 tests C0 alone.
	return !(minX > out.X) && !(out.X > maxX) && out.Y >= minY && !(out.Y > maxY)
}

func projectionBounds(from, to float32) (min, max float32) {
	if from >= to {
		return to, from
	}
	return from, to
}

// ClampPointToSegment projects p with the supplied segment length.
func ClampPointToSegment(line *types.Rectf, p types.Pointf, length float32, out *types.Pointf) {
	p1, p2 := line.Min, line.Max
	dx := float64(pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(float64(p2.X), -float64(p1.X))))
	dy := pointOnLineAddChop53_57C8A0(float64(p2.Y), -float64(p1.Y))
	cx := pointOnLineAddChop53_57C8A0(float64(p.X), -float64(p1.X))
	cy := pointOnLineAddChop53_57C8A0(float64(p.Y), -float64(p1.Y))
	dot := pointOnLineAddChop53_57C8A0(pointOnLineMulChop53_57C8A0(dy, cy), pointOnLineMulChop53_57C8A0(dx, cx))
	denom := pointOnLineMulChop53_57C8A0(float64(length), float64(length))
	qx := float64(pointOnLineSpill57C8A0(pointOnLineDivChop53_57C8A0(pointOnLineMulChop53_57C8A0(dot, dx), denom)))
	qy := pointOnLineDivChop53_57C8A0(pointOnLineMulChop53_57C8A0(dot, float64(pointOnLineSpill57C8A0(dy))), denom)
	out.X = pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(qx, float64(line.Min.X)))
	out.Y = pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(qy, float64(line.Min.Y)))

	minX, maxX := projectionBounds(line.Min.X, line.Max.X)
	minY, maxY := projectionBounds(line.Min.Y, line.Max.Y)
	// The two original FPU comparisons handle unordered coordinates differently.
	if !(out.X >= minX) {
		out.X = minX
	} else if out.X > maxX {
		out.X = maxX
	}
	if out.Y < minY {
		out.Y = minY
	} else if out.Y > maxY {
		out.Y = maxY
	}
}
