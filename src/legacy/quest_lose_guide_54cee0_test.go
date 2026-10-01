package legacy

import (
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestQuestLoseGuide54CEE0NativeCGoRoundTrip(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := GetServer
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: s} }
	t.Cleanup(func() { GetServer = oldServer })
	u, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	p, freePlayer := alloc.New(server.Player{})
	t.Cleanup(freePlayer)
	u.UpdateData, update.Player = unsafe.Pointer(update), p
	p.Info().SetPlayerClass(2)
	p.PlayerInd = 255
	for i := range p.SpellLvl {
		p.SpellLvl[i] = uint32(i) | 0x80000000
	}
	for _, id := range []int{2, 3, 39, 40} {
		p.BeastScrollLvl[id] = 1
	}
	p.BeastScrollLvl[9], p.BeastScrollLvl[10], p.BeastScrollLvl[11] = 2, 0x80000001, math.MaxUint32
	s.Rand.Logic = prand.New(81)
	wantRNG := prand.New(81)
	selected := []int{2, 3, 39, 40}[wantRNG.IntClamp(0, 3)]
	wantPlayer := *p
	wantPlayer.BeastScrollLvl[selected] = 0
	packets := 0
	s.NetSendPacketXxx = func(index int, data []byte, related *server.Object, remove, sequence int) int {
		packets++
		if *p != wantPlayer || index != 255 || !reflect.DeepEqual(data, []byte{0xf0, 0x13, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
			t.Fatal("wrong packet/fields")
		}
		return -257
	}
	questLoseGuideCEntry54CEE0(u)
	if packets != 1 || s.Rand.Logic.Index() != wantRNG.Index() {
		t.Fatal("wrong packets or RNG")
	}
	for i := range p.BeastScrollLvl {
		p.BeastScrollLvl[i] = 2
	}
	p.BeastScrollLvl[40] = 1
	selected, packets = 40, 0
	wantPlayer = *p
	wantPlayer.BeastScrollLvl[40] = 0
	s.Rand.Logic = prand.New(0)
	questLoseGuideCEntry54CEE0(u)
	if packets != 1 || s.Rand.Logic.Index() != 1 {
		t.Fatal("last guide did not use a one-candidate RNG call")
	}
	s.Rand.Logic = nil
	s.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { t.Fatal("unexpected packet"); return 0 }
	for _, class := range []player.Class{0, 1, 128, 255} {
		p.Info().SetPlayerClass(class)
		before := *p
		questLoseGuideCEntry54CEE0(u)
		if *p != before {
			t.Fatal("non-Conjurer changed fields")
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(update), unsafe.Pointer(p)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
