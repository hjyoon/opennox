package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestAnimationSelectorOriginalGo533790(t *testing.T) {
	animationRunOriginal533790(t, "Go", func(f *animationFixture533790) uint32 {
		return uint32(server.MonsterAnimationNative533790(&f.unit.Value, f.runtime))
	})
}
