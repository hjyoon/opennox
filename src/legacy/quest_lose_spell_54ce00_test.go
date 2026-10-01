package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestLoseSpell54CE00NativeCGoRoundTrip(t *testing.T) {
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
	p.PlayerInd = 255
	candidates := []int{1, 4, 128, 136}
	for _, class := range []player.Class{1, 2} {
		p.Info().SetPlayerClass(class)
		for _, id := range []int{1, 4, 9, 27, 34, 41, 128, 136} {
			p.SpellLvl[id] = uint32(id) | 0x80000000
		}
		s.Rand.Logic = prand.New(81)
		wantRNG := prand.New(81)
		selected := candidates[wantRNG.IntClamp(0, len(candidates)-1)]
		wantPlayer := *p
		wantPlayer.SpellLvl[selected] = 0
		packets := 0
		s.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
			packets++
			if *p != wantPlayer || recipient != 255 || !reflect.DeepEqual(data, []byte{0xf0, 0x11, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
				t.Fatalf("packet = %d/%x/%p/%d/%d", recipient, data, related, remove, sequence)
			}
			return -257
		}
		questLoseSpellCEntry54CE00(u)
		if packets != 1 || s.Rand.Logic.Index() != wantRNG.Index() {
			t.Fatalf("class %d packets = %d", class, packets)
		}
	}
	p.Info().SetPlayerClass(player.Class(255))
	s.Rand.Logic = nil
	s.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { t.Fatal("unexpected packet"); return 0 }
	before := *p
	questLoseSpellCEntry54CE00(u)
	if *p != before {
		t.Fatal("non-Wizard/Conjurer changed fields")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(update), unsafe.Pointer(p)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
