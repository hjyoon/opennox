package server

import "github.com/opennox/libs/prand"

// GAME.EXE 00416090 uses the other stream and always consumes one table word,
// including zero and unordered ranges. It shares the binary32 scale and
// retained 53-bit, round-toward-zero instruction boundaries of 00416030, but
// not that function's early return. Do not clamp or swap the input bounds.
func otherRandomFloat416090(random *prand.Rand, min, max float32) float64 {
	delta := logicRandomFloatAddChop53_416030(float64(max), -float64(min))
	value := int32(random.Int(0, logicRandomFloatTableMax416030))
	scaled := logicRandomFloatMulChop53_416030(
		float64(value),
		float64(logicRandomFloatTableScale416030),
	)
	return logicRandomFloatAddChop53_416030(
		logicRandomFloatMulChop53_416030(delta, scaled),
		float64(min),
	)
}
