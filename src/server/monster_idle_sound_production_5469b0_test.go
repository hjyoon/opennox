package server

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Supplement the sealed service comparisons with real clocks, production RNG,
// Audio.EventObj, its delayed queue, registered descriptors and event allocation.
func TestMonsterIdleSoundProduction5469B0(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	for _, entry := range []string{"raw", "ordinary"} {
		for _, seed := range []int{0, 4095} {
			for index := -1; index < 24; index++ {
				for _, mode := range []string{"external", "timer-alias", "absent"} {
					for _, queued := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%d/%s/queued=%t", entry, seed, index, mode, queued), func(t *testing.T) {
							s := &Server{}
							s.SetFrame(100)
							s.SetTickRate(1)
							s.Rand.Logic = prand.New(seed)
							s.Audio.Init(s)
							t.Cleanup(s.Audio.Free)
							// Register descriptors as the loaded sound definitions do.
							for i := 1; i < len(s.Audio.bySound); i++ {
								s.Audio.bySound[i].Field12 = 1
							}
							s.Audio.inAudio = !queued
							// Production audio stores C-owned object pointers in its
							// native event pool; use the same ownership in this fixture.
							unit, freeUnit := alloc.New(Object{})
							t.Cleanup(freeUnit)
							update, freeUpdate := alloc.New(MonsterUpdateData{})
							t.Cleanup(freeUpdate)
							unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
							unit.PosVec = types.Ptf(12.5, -6.25)
							update.Aggression = 0.5
							update.Field131 = math.Float32bits(300)
							update.AIStackInd = int8(index)
							for i := range update.AIStack {
								update.AIStack[i] = AIStackItem{Action: 4, Args: [4]uintptr{1, 2, 3, 4}}
							}
							set, free := alloc.Make([]uint32{}, 5)
							t.Cleanup(free)
							set[4] = 77
							switch mode {
							case "external":
								update.SoundSet122 = unsafe.Pointer(&set[0])
							case "timer-alias":
								update.SoundSet122 = unsafe.Pointer(&update.Field128)
							}
							control := prand.New(seed)
							wantUpdate, wantUnit := *update, *unit
							wantUpdate.Field132 = 100 + uint32(control.IntClamp(20, 60))
							wantID := uint32(77)
							if mode == "timer-alias" {
								wantID = wantUpdate.Field132
							}
							wantEvents := 1
							if mode == "absent" {
								wantEvents = 0
							}
							observed := 0
							s.Audio.OnSound(func(id sound.ID, kind int, obj *Object, pos types.Pointf) {
								observed++
								if uint32(id) != wantID || kind != 0 || obj != unit || pos != unit.PosVec {
									t.Fatal("production audio observer identity/arguments")
								}
							})
							if entry == "ordinary" {
								s.MonsterIdleSound5469B0(unit)
							} else {
								unit.MonsterIdleSoundNative5469B0(s.Frame, s.TickRate, s.Rand.Logic.IntClamp,
									func(id uint32, obj *Object, kind, code uint32) { s.Audio.EventObj(sound.ID(id), obj, int(kind), code) })
							}
							if queued {
								if len(s.Audio.delayedObj) != wantEvents || observed != 0 {
									t.Fatal("production delayed audio queue")
								}
								s.Audio.Reset()
							}
							events := 0
							s.Audio.EachEvent(func(event *AudioEvent) {
								events++
								if uint32(event.Sound) != wantID || event.Obj != unit || event.Pos != unit.PosVec || event.Kind != 0 || event.Code != 0 {
									t.Fatal("production allocated audio event")
								}
							})
							if events != wantEvents || observed != wantEvents || s.Rand.Logic.Index() != control.Index() {
								t.Fatal("production event/RNG counts", events, observed, s.Rand.Logic.Index())
							}
							for _, pair := range []struct {
								got, want unsafe.Pointer
								size      uintptr
							}{
								{unsafe.Pointer(unit), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
								{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
							} {
								if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
									t.Fatal("whole production native record/padding changed")
								}
							}
							if set[4] != 77 {
								t.Fatal("external sound-set modified")
							}
						})
					}
				}
			}
		}
	}
}
