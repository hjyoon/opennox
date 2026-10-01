package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type questLivesReportServer4D9D60 struct {
	Server
	native *server.Server
}

func (s *questLivesReportServer4D9D60) S() *server.Server { return s.native }

func TestQuestLivesReport4D9D60RealCEntryHighNativePointers(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := GetServer
	GetServer = func() Server { return &questLivesReportServer4D9D60{native: s} }
	t.Cleanup(func() { GetServer = oldServer })
	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeUpdate)
	unit.NetCode, unit.UpdateData = 0x89ab9234, unsafe.Pointer(update)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(update)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("C-owned native pointer=%p must exceed 4 GiB", pointer)
			}
		}
		if unsafe.Offsetof(update.ExtraLives) == 320 {
			t.Fatal("native ExtraLives must not use the PE32 offset")
		}
	}
	t.Logf("C entry unit=%p update=%p life-offset=%d", unit, update, unsafe.Offsetof(update.ExtraLives))
	values := []uint32{0x100, 0x101, 0x89abcdef, math.MaxUint32}
	for value := uint32(0); value < 256; value++ {
		values = append(values, value)
	}
	for _, lives := range values {
		update.ExtraLives = lives
		for _, recipient := range []int32{math.MinInt32, -1, 31, 255, math.MaxInt32} {
			for _, result := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
				calls := 0
				s.NetSendPacketXxx = func(ind int, packet []byte, related *server.Object, remove, sequence int) int {
					calls++
					want := []byte{0xf0, 4, byte(uint64(lives) % 256), 0x34, 0x92}
					if ind != int(recipient) || !reflect.DeepEqual(packet, want) || related != nil || remove != 1 || sequence != 0 {
						t.Fatalf("native send=%d/%x/%p/%d/%d want=%d/%x/nil/1/0", ind, packet, related, remove, sequence, recipient, want)
					}
					return int(result)
				}
				beforeUnit, beforeUpdate := *unit, *update
				if got := questLivesReportCEntry4D9D60(recipient, unit); got != result || calls != 1 || *unit != beforeUnit || *update != beforeUpdate {
					t.Fatalf("C entry lives=%08x return=%d want=%d calls=%d or sender changed state", lives, got, result, calls)
				}
			}
		}
	}
}
