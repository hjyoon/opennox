package legacy

import (
	"math"
	"sync"
	"testing"

	"github.com/opennox/libs/platform"
)

func preserveLegacyCRTRand315(t *testing.T) *seedCapturePlatform {
	t.Helper()
	legacyCRTRand.Lock()
	state := legacyCRTRand.state
	legacyCRTRand.Unlock()
	old := platform.Get()
	probe := &seedCapturePlatform{Platform: old}
	platform.Set(probe)
	t.Cleanup(func() {
		legacyCRTRand.Lock()
		legacyCRTRand.state = state
		legacyCRTRand.Unlock()
		platform.Set(old)
	})
	return probe
}

func TestPlatformCRTRand315StaticOriginalVectors(t *testing.T) {
	probe := preserveLegacyCRTRand315(t)
	// Independently calculated integer vectors from the statically observed
	// 0040200D multiply/add/shift/mask, not an original executable run.
	for _, tc := range []struct {
		name string
		seed func()
		word uint32
		want []int32
	}{
		{"one", func() { nox_platform_srand(1) }, 1, []int32{41, 18467, 6334, 26500, 19169, 15724}},
		{"prefab", func() { nox_platform_srand(1234) }, 1234, []int32{4068, 213, 12761, 8758, 23056, 7717}},
		{"all bits", func() { nox_platform_srand(0xffffffff) }, 0xffffffff, []int32{35, 29739, 3374, 11141, 31308, 7870}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := len(probe.seeds)
			for repeat := 0; repeat < 2; repeat++ {
				tc.seed()
				for i, want := range tc.want {
					if got := nox_platform_rand(); got != want {
						t.Fatalf("repeat %d draw %d: got %d, want %d", repeat, i, got, want)
					}
				}
			}
			if len(probe.seeds) != before+2 || probe.seeds[before] != int64(tc.word) || probe.seeds[before+1] != int64(tc.word) {
				t.Fatalf("changed platform seed forwarding: %v", probe.seeds[before:])
			}
		})
	}
}

func TestPlatformCRTRand315ConcurrentStream(t *testing.T) {
	preserveLegacyCRTRand315(t)
	nox_platform_srand(1)
	const workers, draws = 8, 512
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range draws {
				if v := nox_platform_rand(); v < 0 || v > 32767 {
					t.Errorf("out-of-range CRT draw: %d", v)
				}
			}
		}()
	}
	wg.Wait()
	// Fixed independent model result after exactly 4096 draws, regardless
	// of the order in which the goroutines acquired the stream lock.
	legacyCRTRand.Lock()
	state := legacyCRTRand.state
	legacyCRTRand.Unlock()
	if state != 0x0047d001 {
		t.Fatalf("lost or repeated concurrent draw: state=%08x", state)
	}
}

func TestPlatformCRTRand315ActualMapgenBounds(t *testing.T) {
	preserveLegacyCRTRand315(t)
	ints, positions := mapgenRandValues315(1234, 512)
	// Each iteration consumes the integer draw followed by the position draw.
	if ints[0] != -376 || positions[0] < -505.34355 || positions[0] > -505.34354 {
		t.Fatalf("wrong first seeded C mapgen values: %d, %.12f", ints[0], positions[0])
	}
	for i := range ints {
		if ints[i] < -500 || ints[i] > 500 || !(positions[i] >= -512 && positions[i] <= 512.00001) {
			t.Fatalf("C mapgen draw %d left its authored bounds: %d, %.12f", i, ints[i], positions[i])
		}
	}
	ints2, positions2 := mapgenRandValues315(1234, 512)
	for i := range ints {
		if ints2[i] != ints[i] || positions2[i] != positions[i] {
			t.Fatalf("C mapgen reseed diverged at draw %d", i)
		}
	}
}

func TestPlatformCRTRand315MapgenIntegerWidths(t *testing.T) {
	preserveLegacyCRTRand315(t)
	ranges := [][2]int32{{-500, 500}, {0, 0}, {127, -3}, {-300000000, 400000000},
		{math.MinInt32, math.MaxInt32}, {2147483640, 2147483647}}
	weighted := [][2]int32{{0, 0}, {-500, 30}, {1000, 5}, {2147483637, 4}}
	digest := uint64(14695981039346656037)
	add := func(value int32) {
		digest ^= uint64(uint32(value))
		digest *= 1099511628211
	}
	for _, seed := range []uint32{1, 1234, math.MaxUint32} {
		mapgenRandValues315(seed, 0)
		for range 64 {
			for _, bounds := range ranges {
				add(mapgenRandInt315(bounds[0], bounds[1]))
			}
			for _, args := range weighted {
				add(mapgenRandWeighted315(args[0], args[1]))
			}
		}
	}
	// Computed with the previous C entries and the same CRT state transitions.
	if want := uint64(0x79a01dceb650cb16); digest != want {
		t.Fatalf("seeded native integer draw digest = %016x, want %016x", digest, want)
	}
}
