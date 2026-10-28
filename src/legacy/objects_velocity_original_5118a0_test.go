package legacy

import (
	"math"
	"runtime"
	"testing"

	"github.com/opennox/opennox/v1/internal/porttest"
)

func TestObjectsVelocityOriginalNativeC5118A0(t *testing.T) {
	porttest.RunVelocityOriginal5118A0(t, "../server/testdata/objects_velocity_original_5118a0.json", func(s *porttest.VelocityScene5118A0) int32 {
		old := objectsVelocityRuntime5118A0
		objectsVelocityRuntime5118A0 = s.Runtime
		defer func() { objectsVelocityRuntime5118A0 = old }()
		return ObjectsVelocityCProbe5118A0(math.Float32frombits(s.Input.Step))
	})
}

func TestObjectsVelocityOriginalX87NaNLoad5118A0(t *testing.T) {
	if runtime.GOARCH != "386" && runtime.GOARCH != "amd64" {
		t.Skip("requires x86 FLD m32fp")
	}
	for _, word := range []uint32{0x7f812345, 0xff812345, 0x7f832345, 0x7fc12345} {
		got := velocityLoadNaNProbe5118A0(word)
		if uint32(got) != word|0x400000 || (got>>32)&1 != uint64((word>>22&1)^1) {
			t.Errorf("FLD %08x = %08x status=%04x", word, uint32(got), got>>32)
		}
	}
}
