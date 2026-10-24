package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// The Online/not-Chat/server-option gate remains in the existing server
// caller. These are the original updater's roster, CTF, audio and marker
// service boundaries; every object/update/player pointer stays native-width.
type playerCamperRuntime4D7CC0 struct {
	frame     func() uint32
	firstUnit func() *Object
	nextUnit  func(*Object) *Object
	isCTF     func() bool
	audio     func(uint32, *Object)
	markerOn  func(*Object)
	markerOff func(*Object)
}

// playerCamperDistance4D7CC0 retains the original unspilled FSUB/FMUL/FADD
// sequence in the established gameplay precision53/chop context. The shared
// arithmetic primitives correct each operation locally without changing the
// native thread's floating-point environment or fusing the distance sum.
func playerCamperDistance4D7CC0(x, y float32, previousX, previousY uint32) float64 {
	dy := logicRandomFloatAddChop53_416030(float64(y), -float64(math.Float32frombits(previousY)))
	dx := logicRandomFloatAddChop53_416030(float64(x), -float64(math.Float32frombits(previousX)))
	xSquared := logicRandomFloatMulChop53_416030(dx, dx)
	ySquared := logicRandomFloatMulChop53_416030(dy, dy)
	return logicRandomFloatAddChop53_416030(xSquared, ySquared)
}

func (s *playerCamper) update4D7CC0(r playerCamperRuntime4D7CC0) {
	if r.frame()-s.LastCheck > s.StartTime {
		for it := r.firstUnit(); it != nil; it = r.nextUnit(it) {
			if it.Flags().HasAny(object.FlagDestroyed | object.FlagDead) {
				continue
			}
			update := it.UpdateDataPlayer()
			if update.Player.Field3680&1 == 0 {
				distance := playerCamperDistance4D7CC0(it.PosVec.X, it.PosVec.Y, update.Field62, update.Field63)
				// Original FCOMP tests C0 only: unordered also refreshes.
				if !(distance >= float64(float32(s.RadiusSq))) {
					update.Field64 = r.frame()
				}
				update.Field62 = math.Float32bits(it.PosVec.X)
				update.Field63 = math.Float32bits(it.PosVec.Y)
			}
		}
		s.LastCheck = r.frame()
	}
	for it := r.firstUnit(); it != nil; it = r.nextUnit(it) {
		if it.Flags().HasAny(object.FlagDestroyed | object.FlagDead) {
			continue
		}
		update := it.UpdateDataPlayer()
		if update.Player.Field3680&1 == 0 {
			if s.LastCheck-update.Field64 > s.FadeTime {
				if update.IsCamping != 0 {
					r.markerOff(it)
					update.IsCamping = 0
				}
			} else if update.IsCamping == 0 {
				if !r.isCTF() || update.Player.WeaponEquip&1 != 0 {
					r.audio(774, it)
				}
				r.markerOn(it)
				update.IsCamping = 1
			}
		}
	}
}
