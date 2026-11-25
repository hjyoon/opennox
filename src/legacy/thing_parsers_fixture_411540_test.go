package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

var (
	thingFixtureSave411540       = cfixture.ThingSave411540
	thingFixtureState411540      = cfixture.ThingState411540
	thingFixtureSeed411540       = cfixture.ThingSeed411540
	thingFixturePointer411540    = cfixture.ThingPointer411540
	thingFixtureLayout411540     = cfixture.ThingLayout411540
	thingFixtureSignedChar411540 = cfixture.ThingSignedChar411540
	thingFixtureRGB411540        = cfixture.ThingRGB411540
)

func thingFixtureMemfile411540(input []byte, start int) (*nox_memfile, func()) {
	f, free := cfixture.ThingMemfile411540(input, start)
	return (*nox_memfile)(f), free
}

func thingFixtureInput411540(f *nox_memfile, n int) []byte {
	return cfixture.ThingInput411540(unsafe.Pointer(f), n)
}

func thingFixtureOffset411540(f *nox_memfile) int {
	return cfixture.ThingOffset411540(unsafe.Pointer(f))
}

func thingFixtureCall411540(entry uint32, reference bool, f *nox_memfile, p unsafe.Pointer) int32 {
	return cfixture.ThingCall411540(entry, reference, unsafe.Pointer(f), p)
}

func thingFixtureAudioReal415660(f *nox_memfile, p unsafe.Pointer) int32 {
	return cfixture.ThingAudioReal415660(unsafe.Pointer(f), p)
}

func thingFixtureGoMemfile411540(b []byte) *nox_memfile {
	return (*nox_memfile)(cfixture.ThingGoMemfile411540(b))
}

func thingFixtureScratch411540() ([]byte, func()) {
	p, free := alloc.Malloc(256*1024 + 32)
	b := unsafe.Slice((*byte)(p), 256*1024+32)
	for i := range b {
		b[i] = byte(0xA7 + i%13)
	}
	return b, free
}
