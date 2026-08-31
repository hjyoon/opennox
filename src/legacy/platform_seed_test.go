package legacy

import (
	"testing"

	"github.com/opennox/libs/platform"
)

type seedCapturePlatform struct {
	platform.Platform
	seeds []int64
}

func (p *seedCapturePlatform) RandSeed(seed int64) {
	p.seeds = append(p.seeds, seed)
}

func TestFixedSeedWrappersMatchGAMEEXEContract(t *testing.T) {
	old := platform.Get()
	probe := &seedCapturePlatform{Platform: old}
	platform.Set(probe)
	t.Cleanup(func() { platform.Set(old) })

	Sub_4E4DC0()
	Sub_4E4DD0()
	Sub_4E5AC0()
	Sub_4EED30()
	Sub_4EF560()
	Sub_4EF570()
	Sub_4F0630()
	Sub_4F3E20()
	Sub_4FB940()
	Sub_4FB950()

	want := []int64{0x1429, 0x490, 0x13D11, 0x22D6, 0x22D7, 0x7DA, 0x7DB, 0x4E2A, 0x143D, 0x22EA}
	if len(probe.seeds) != len(want) {
		t.Fatalf("seed calls: got %#v, want %#v", probe.seeds, want)
	}
	for i := range want {
		if probe.seeds[i] != want[i] {
			t.Errorf("seed call %d: got %#x, want %#x", i, probe.seeds[i], want[i])
		}
	}
}
