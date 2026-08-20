package server

import (
	"image"

	"github.com/opennox/libs/types"
)

const (
	dropAllItemsSpacingExtra4EDA40 = float32(6)
	dropAllItemsRandomMin4EDA40    = float32(-3)
	dropAllItemsRandomMax4EDA40    = float32(3)
	dropAllItemsSource4EDA40       = `C:\NoxPost\src\Server\Object\pickdrop\drop.c`
	dropAllItemsSourceLineX4EDA40  = int32(823)
	dropAllItemsSourceLineY4EDA40  = int32(824)
	dropAllItemsTraceFlag4EDA40    = uint8(1)
)

type dropAllItemsRay4EDA40 struct {
	Origin      types.Pointf
	Destination types.Pointf
}

type dropAllItemsHooks4EDA40[O any, I comparable] struct {
	loadOwnerArg      func() O
	loadInventoryHead func(O) I
	loadInventoryNext func(I) I
	dropEligible      func(O, I) int32
	loadItemRadius    func(I) float32
	loadOwnerX        func(O) float32
	loadOwnerY        func(O) float32
	ownerPosition     func(O) *types.Pointf

	randomFloat func(min, max float32, source string, line int32) float64
	mapTrace    func(*dropAllItemsRay4EDA40, *types.Pointf, *image.Point, uint8) int32
	drop        func(O, I, *types.Pointf) int32
}

// Keep every x87 arithmetic instruction at an explicit binary64 boundary.
// Nox executes this code with 53-bit precision control. Separate noinline
// helpers prevent contraction and explicit float32 conversion models FSTP.
//
//go:noinline
func dropAllItemsAdd64_4EDA40(a, b float64) float64 { return a + b }

//go:noinline
func dropAllItemsSub64_4EDA40(a, b float64) float64 { return a - b }

//go:noinline
func dropAllItemsMul64_4EDA40(a, b float64) float64 { return a * b }

//go:noinline
func dropAllItemsSpill32_4EDA40(value float64) float32 { return float32(value) }

func dropAllItemsSpacing4EDA40(maxRadius float32) float32 {
	doubled := dropAllItemsAdd64_4EDA40(float64(maxRadius), float64(maxRadius))
	return dropAllItemsSpill32_4EDA40(
		dropAllItemsAdd64_4EDA40(doubled, float64(dropAllItemsSpacingExtra4EDA40)),
	)
}

type dropAllItemsSpiral4EDA40 struct {
	gridSize        int32
	sideLimit       int32
	segmentProgress int32
	direction       int32
	x               int32
	y               int32
	ringHadSuccess  bool
}

func newDropAllItemsSpiral4EDA40() dropAllItemsSpiral4EDA40 {
	return dropAllItemsSpiral4EDA40{
		gridSize:  3,
		sideLimit: 2,
		x:         0,
		y:         -1,
	}
}

func (s *dropAllItemsSpiral4EDA40) move4EDA40(direction int32) {
	switch direction {
	case 0:
		s.x++
	case 1:
		s.y++
	case 2:
		s.x--
	case 3:
		s.y--
	}
}

// advance4EDA40 reproduces the four-entry absolute dispatch table. It returns
// false only after a complete perimeter that placed no item, which selects
// the original live-owner-position fallback.
func (s *dropAllItemsSpiral4EDA40) advance4EDA40() bool {
	if s.segmentProgress != s.sideLimit {
		s.segmentProgress++
		s.move4EDA40(s.direction)
		return true
	}
	if s.direction != 3 && s.sideLimit != 0 {
		s.direction++
		s.segmentProgress = 1
		s.move4EDA40(s.direction)
		return true
	}
	if !s.ringHadSuccess {
		return false
	}
	s.sideLimit += 2
	s.gridSize += 2
	half := s.gridSize / 2
	s.x = 1 - half
	s.y = -half
	s.segmentProgress = 1
	s.direction = 0
	s.ringHadSuccess = false
	return true
}

func dropAllItemsCandidate4EDA40[O any, I comparable](
	hooks dropAllItemsHooks4EDA40[O, I],
	owner O,
	spacing float32,
	spiral *dropAllItemsSpiral4EDA40,
	ray *dropAllItemsRay4EDA40,
) {
	xScaled := dropAllItemsMul64_4EDA40(float64(spiral.x), float64(spacing))
	xBase := dropAllItemsSpill32_4EDA40(
		dropAllItemsAdd64_4EDA40(xScaled, float64(hooks.loadOwnerX(owner))),
	)
	yScaled := dropAllItemsMul64_4EDA40(float64(spiral.y), float64(spacing))
	yBase := dropAllItemsSpill32_4EDA40(
		dropAllItemsSub64_4EDA40(float64(hooks.loadOwnerY(owner)), yScaled),
	)

	xRandom := hooks.randomFloat(
		dropAllItemsRandomMin4EDA40,
		dropAllItemsRandomMax4EDA40,
		dropAllItemsSource4EDA40,
		dropAllItemsSourceLineX4EDA40,
	)
	ray.Destination.X = dropAllItemsSpill32_4EDA40(
		dropAllItemsAdd64_4EDA40(xRandom, float64(xBase)),
	)
	yRandom := hooks.randomFloat(
		dropAllItemsRandomMin4EDA40,
		dropAllItemsRandomMax4EDA40,
		dropAllItemsSource4EDA40,
		dropAllItemsSourceLineY4EDA40,
	)
	ray.Destination.Y = dropAllItemsSpill32_4EDA40(
		dropAllItemsAdd64_4EDA40(yRandom, float64(yBase)),
	)
}

func dropAllItemsFallback4EDA40[O any, I comparable](
	hooks dropAllItemsHooks4EDA40[O, I],
	owner O,
	current I,
) int32 {
	var zero I
	var result int32
	for {
		next := hooks.loadInventoryNext(current)
		result = hooks.dropEligible(owner, current)
		if result != 0 {
			result = hooks.drop(owner, current, hooks.ownerPosition(owner))
		}
		if next == zero {
			return result
		}
		current = next
	}
}

// dropAllItems4EDA40 preserves GAME.EXE 004EDA40. The first inventory pass
// computes an ordered strict maximum radius among eligible items. The second
// pass caches each successor before eligibility, trace, or drop side effects.
// Eligible items consume traceable candidates from a persistent local ray and
// square-perimeter spiral; the primary drop result is ignored. A perimeter
// with no successful trace switches the current item and cached-successor
// suffix to the exact live owner-position pointer. Normal exhaustion returns
// zero; fallback preserves the final predicate or drop result in EAX.
func dropAllItems4EDA40[O any, I comparable](hooks dropAllItemsHooks4EDA40[O, I]) int32 {
	var zero I
	owner := hooks.loadOwnerArg()
	maxRadius := float32(0)
	for item := hooks.loadInventoryHead(owner); item != zero; {
		if hooks.dropEligible(owner, item) != 0 {
			radius := hooks.loadItemRadius(item)
			if radius > maxRadius {
				maxRadius = radius
			}
		}
		item = hooks.loadInventoryNext(item)
	}

	current := hooks.loadInventoryHead(owner)
	ray := dropAllItemsRay4EDA40{
		Origin: types.Pointf{
			X: hooks.loadOwnerX(owner),
			Y: hooks.loadOwnerY(owner),
		},
	}
	spacing := dropAllItemsSpacing4EDA40(maxRadius)
	if current == zero {
		return 0
	}

	spiral := newDropAllItemsSpiral4EDA40()
	for {
		next := hooks.loadInventoryNext(current)
		if hooks.dropEligible(owner, current) != 0 {
			for {
				dropAllItemsCandidate4EDA40(hooks, owner, spacing, &spiral, &ray)
				succeeded := hooks.mapTrace(&ray, nil, nil, dropAllItemsTraceFlag4EDA40) != 0
				if succeeded {
					_ = hooks.drop(owner, current, &ray.Destination)
					spiral.ringHadSuccess = true
				}
				if !spiral.advance4EDA40() {
					return dropAllItemsFallback4EDA40(hooks, owner, current)
				}
				if succeeded {
					break
				}
			}
		}
		if next == zero {
			return 0
		}
		current = next
	}
}
