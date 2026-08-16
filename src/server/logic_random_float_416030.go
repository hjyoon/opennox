package server

import (
	"math"

	"github.com/opennox/libs/prand"
)

const (
	logicRandomFloatTableScale416030 = float32(0x1.0002p-15)
	logicRandomFloatTableMax416030   = 0x7fff
)

// Keep the x87 53-bit arithmetic boundaries of GAME.EXE 00416030 explicit.
//
//go:noinline
func logicRandomFloatAdd64_416030(a, b float64) float64 { return a + b }

//go:noinline
func logicRandomFloatSub64_416030(a, b float64) float64 { return a - b }

//go:noinline
func logicRandomFloatMul64_416030(a, b float64) float64 { return a * b }

// logicRandomFloat416030 reproduces GAME.EXE 00416030 with the public prand
// state. Int(0, 0x7fff) exposes one complete original table dword because all
// 4096 sealed entries are in that range; it also advances and wraps the index
// exactly once. The original C3-only x87 comparison treats both zero and
// unordered ranges as the no-step path and returns the binary32 max bound.
func logicRandomFloat416030(random *prand.Rand, min, max float32) float64 {
	delta := logicRandomFloatSub64_416030(float64(max), float64(min))
	if delta == 0 || math.IsNaN(delta) {
		return float64(max)
	}
	value := int32(random.Int(0, logicRandomFloatTableMax416030))
	scaled := logicRandomFloatMul64_416030(
		float64(value),
		float64(logicRandomFloatTableScale416030),
	)
	return logicRandomFloatAdd64_416030(
		logicRandomFloatMul64_416030(delta, scaled),
		float64(min),
	)
}
