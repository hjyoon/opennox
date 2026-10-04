package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Exercise the real object allocator/free boundary rather than calling the
// Sentry update on a destroyed object: normal ticks skip destroyed updates.
func TestSentryGlobeFreeObject510FD0UnlinksBeforeRelease(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, position := range []string{"only", "head", "middle", "tail", "outside", "reset"} {
			t.Run(fmt.Sprintf("enabled=%t/%s", enabled, position), func(t *testing.T) {
				s := New(nil, nil, strman.New())
				t.Cleanup(s.Close)
				if !s.Objs.Init(4) {
					t.Fatal("object allocator initialization failed")
				}
				t.Cleanup(s.Objs.FreeObjects)
				s.NetList.Init()
				s.Players.list = make([]Player, common.MaxPlayers)
				const index = ntype.PlayerInd(2)
				player := &s.Players.list[index]
				player.Active, player.PlayerInd = 1, byte(index)
				player.Field10, player.Field12 = 20, 20
				player.Pos3632Vec = types.Ptf(100, 100)

				var linked []*Object
				for i := 0; i < 3; i++ {
					data, freeData := alloc.New(SentryUpdateData{Field4: 0x3f800000})
					t.Cleanup(freeData)
					obj := s.Objs.alloc.NewObject()
					*obj = Object{serverHandle: s.handle, NetCode: uint32(i + 1),
						ObjClass:   object.ClassSimple | object.ClassImmobile,
						UpdateData: unsafe.Pointer(data), PosVec: types.Ptf(90, 90), Pos39: types.Ptf(110, 110)}
					if enabled {
						obj.ObjFlags |= object.FlagEnabled
					}
					s.Objs.Alive++
					s.sentryGlobe510E60.insert(obj)
					linked = append([]*Object{obj}, linked...)
				}
				target := linked[0]
				want := linked[1:]
				switch position {
				case "only":
					s.SentryGlobeReset510E50()
					target.ObjFlags &^= object.FlagMarked
					s.sentryGlobe510E60.insert(target)
					want = nil
				case "middle":
					target, want = linked[1], []*Object{linked[0], linked[2]}
				case "tail":
					target, want = linked[2], linked[:2]
				case "outside":
					target = s.Objs.alloc.NewObject()
					*target = Object{serverHandle: s.handle, NetCode: 4, ObjFlags: object.FlagMarked}
					s.Objs.Alive++
					want = linked
				case "reset":
					s.SentryGlobeReset510E50()
					want = nil
				}
				target.ObjFlags |= object.FlagDestroyed
				code, alive := target.NetCode, s.Objs.Alive
				if got := s.Objs.FreeObject(target); got != alive-1 || target.NetCode != code {
					t.Fatalf("free accounting/code = %d/%d, want %d/%d", got, target.NetCode, alive-1, code)
				}

				node := s.sentryGlobe510E60.head
				var previous *Object
				for i, expected := range want {
					if node != expected || node == target {
						t.Fatalf("live Sentry node %d = %p, want %p; released = %p", i, node, expected, target)
					}
					if node.Field125 != previous || !node.Flags().Has(object.FlagMarked) {
						t.Fatalf("live Sentry node %d has a broken previous link/marker", i)
					}
					previous, node = node, node.InvNextItem
				}
				if node != nil {
					t.Fatalf("Sentry head/tail retained released storage %p", node)
				}
				// The normal network sender must not walk allocator poison after
				// deletion, and must still deliver every surviving enabled ray.
				s.SentryGlobeSendToPlayer511100(int(index))
				wantPackets := 0
				if enabled {
					wantPackets = len(want) * 9
				}
				if packets := s.NetList.CopyPacketsA(index, netlist.Kind1); len(packets) != wantPackets {
					t.Fatalf("surviving ray bytes = %d, want %d", len(packets), wantPackets)
				}
			})
		}
	}
}
