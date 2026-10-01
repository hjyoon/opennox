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

func TestQuestLoseAbility54CFB0NativeCGoRoundTrip(t *testing.T) {
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
	for i := range p.SpellLvl {
		p.SpellLvl[i] = uint32(i) | 0x80000000
	}
	s.Rand.Logic = prand.New(81)
	wantRNG := prand.New(81)
	selected := wantRNG.IntClamp(1, 5)
	wantLevels := p.SpellLvl
	wantLevels[selected] = 0
	packets := 0
	s.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
		packets++
		if p.SpellLvl != wantLevels || recipient != 255 || !reflect.DeepEqual(data, []byte{0xf0, 0x12, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
			t.Fatalf("packet = %d/%x/%p/%d/%d", recipient, data, related, remove, sequence)
		}
		return 0x123456e1
	}
	if got := questLoseAbilityCEntry54CFB0(u); got != -31 || packets != 1 || s.Rand.Logic.Index() != wantRNG.Index() {
		t.Fatalf("result/packets = %d/%d", got, packets)
	}
	p.Info().SetPlayerClass(player.Class(255))
	s.Rand.Logic = nil
	if got := questLoseAbilityCEntry54CFB0(u); got != -1 || packets != 1 {
		t.Fatalf("class result/packets = %d/%d", got, packets)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(update), unsafe.Pointer(p)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned pointer %p must exceed 4 GiB", pointer)
			}
		}
	}
}
