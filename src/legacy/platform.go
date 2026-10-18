package legacy

/*
#include <stdbool.h>
*/
import "C"
import (
	"sync"

	"github.com/opennox/libs/platform"
)

var (
	PlatformTicks          func() uint64
	Nox_ticks_reset_416D40 func()
)

// GAME.EXE 0040200D consumes a uint32 CRT state and returns only bits 16..30.
// Its two live consumers are the classic generator's integer/float helpers;
// an arbitrary Go RandInt value can put prefab coordinates outside the map.
// The original state belongs to the game thread. Serialize that stream here
// without tying it to whichever OS thread executes a CGo callback.
var legacyCRTRand = struct {
	sync.Mutex
	state uint32
}{state: 1}

//export nox_platform_rand
func nox_platform_rand() int32 {
	legacyCRTRand.Lock()
	legacyCRTRand.state = legacyCRTRand.state*214013 + 2531011
	v := int32((legacyCRTRand.state >> 16) & 0x7fff)
	legacyCRTRand.Unlock()
	return v
}

//export nox_platform_srand
func nox_platform_srand(seed C.uint) {
	legacyCRTRand.Lock()
	legacyCRTRand.state = uint32(seed)
	legacyCRTRand.Unlock()
	// Retain the existing platform hook and its exact unsigned seed forwarding.
	platform.RandSeed(int64(seed))
}

//export nox_platform_get_ticks
func nox_platform_get_ticks() C.uint { return C.uint(PlatformTicks()) }

//export nox_ticks_reset_416D40
func nox_ticks_reset_416D40() { Nox_ticks_reset_416D40() }
