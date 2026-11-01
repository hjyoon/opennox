package legacy

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type monsterIdleProductionServer5469B0 struct {
	Server
	srv *server.Server
}

func (s *monsterIdleProductionServer5469B0) S() *server.Server { return s.srv }

// Audio remains the original reference's declared world service. The actual
// public C entry obtains frame/FPS and prand through its production bindings.
// The server-package companion tests exercise real Audio.EventObj and replay.
func TestMonsterIdleSoundProductionPublicC5469B0(t *testing.T) {
	if monsterIdleRuntimeOverride5469B0 != nil {
		t.Fatal("clean production bindings required")
	}
	oldServer := GetServer
	t.Cleanup(func() { GetServer = oldServer })
	for _, seed := range []int{0, 4095} {
		for index := -1; index < 24; index++ {
			for _, mode := range []string{"external", "timer-alias", "absent"} {
				t.Run(fmt.Sprintf("%d/%d/%s", seed, index, mode), func(t *testing.T) {
					s := &server.Server{}
					GetServer = func() Server { return &monsterIdleProductionServer5469B0{srv: s} }
					s.SetFrame(100)
					s.SetTickRate(1)
					s.Rand.Logic = prand.New(seed)
					unit, freeUnit := alloc.New(server.Object{})
					t.Cleanup(freeUnit)
					update, freeUpdate := alloc.New(server.MonsterUpdateData{})
					t.Cleanup(freeUpdate)
					set, freeSet := alloc.Make([]uint32{}, 5)
					t.Cleanup(freeSet)
					unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
					update.Aggression, update.Field131, update.AIStackInd = 0.5, math.Float32bits(300), int8(index)
					for i := range update.AIStack {
						update.AIStack[i] = server.AIStackItem{Action: 4, Args: [4]uintptr{1, 2, 3, 4}}
					}
					set[4] = 77
					switch mode {
					case "external":
						update.SoundSet122 = unsafe.Pointer(&set[0])
					case "timer-alias":
						update.SoundSet122 = unsafe.Pointer(&update.Field128)
					}
					if unsafe.Sizeof(uintptr(0)) == 8 {
						for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(&set[0])} {
							if uintptr(ptr) <= math.MaxUint32 {
								t.Fatal("actual native allocation above 4 GiB required")
							}
						}
					}
					control := prand.New(seed)
					wantUnit, wantUpdate := *unit, *update
					wantUpdate.Field132 = 100 + uint32(control.IntClamp(20, 60))
					wantID := uint32(77)
					if mode == "timer-alias" {
						wantID = wantUpdate.Field132
					}
					wantCalls := 1
					if mode == "absent" {
						wantCalls = 0
					}
					calls := 0
					runtime := monsterIdleSoundRuntime5469B0()
					runtime.Audio = func(id uint32, obj *server.Object, kind, code uint32) {
						calls++
						if id != wantID || obj != unit || kind != 0 || code != 0 || update.Field132 != wantUpdate.Field132 {
							t.Fatal("public C audio service arguments/order")
						}
					}
					monsterIdleRuntimeOverride5469B0 = &runtime
					defer func() { monsterIdleRuntimeOverride5469B0 = nil }()
					idleOriginalC5469B0("Idle", unit)
					if calls != wantCalls || s.Rand.Logic.Index() != control.Index() {
						t.Fatal("production C service/RNG counts")
					}
					for _, pair := range []struct {
						got, want unsafe.Pointer
						size      uintptr
					}{
						{unsafe.Pointer(unit), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
						{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
					} {
						if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
							t.Fatal("whole C-owned production records/padding")
						}
					}
					if set[4] != 77 {
						t.Fatal("sound-set modified")
					}
				})
			}
		}
	}
}
