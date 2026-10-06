package opennox

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

// aiDependencyHealthFraction546EC4 retains GAME.EXE 00546EC4's unsigned
// WORD quotient in the ordinary gameplay 53-bit/round-toward-zero context.
// Max is read first, and zero Max selects 00583068's 1.0 without reading Cur.
// Both integers and the division residual are inside binary64's range.
func aiDependencyHealthFraction546EC4(health *server.HealthData) float64 {
	maximum := health.Max
	if maximum == 0 {
		return 1
	}
	current := health.Cur
	fraction := float64(current) / float64(maximum)
	// Correct an upward nearest-even rounding locally. FMA obtains only
	// the exact residual sign; it does not contract an original operation
	// or change the thread's floating-point environment.
	if math.FMA(fraction, float64(maximum), -float64(current)) > 0 {
		fraction = math.Nextafter(fraction, 0)
	}
	return fraction
}

func aiDependencyNotHealthy546EC4(unit *server.Object, update *server.MonsterUpdateData) bool {
	fraction := aiDependencyHealthFraction546EC4(unit.HealthData)
	// 00546EF5 reads the entry-cached update only after the quotient.
	// FCOMP tests C0 alone: less/unordered passes, equal/greater fails.
	return !(fraction >= float64(update.ResumeLevel))
}
