package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/script"
	"github.com/opennox/libs/script/lua"
	"github.com/opennox/libs/script/lua/mapv0"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Seed bubbles through the actual native allocator/list constructor. The input
// is explicit, not a fake sender/decoder/list result. Pixel rendering and remote
// UDP transport are not exercised by this controlled lifecycle fixture.
func luaCompatMuteSeed(t *testing.T, code uint16, text string) unsafe.Pointer {
	t.Helper()
	header := []byte{0xa8, byte(code), byte(code >> 8), 4, 100, 0, 200, 0, 16, 90, 0}
	legacy.Nox_xxx_createTextBubble_48D880(header, text)
	value := legacy.Nox_xxx_netCode2ChatBubble_48D850(int(code))
	if value == 0 {
		t.Fatal("actual native bubble constructor did not link a record")
	}
	ptr := unsafe.Pointer(uintptr(value))
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
		t.Fatalf("actual native bubble is not above 4 GiB: %p", ptr)
	}
	raw := scriptNSRoamFlagBytes515C80(ptr, legacy.ChatBubbleSize())
	if binary.LittleEndian.Uint32(raw[640:644]) != 127 || binary.LittleEndian.Uint32(raw[656:660]) != uint32(code) {
		t.Fatalf("native bubble frame/code differ: frame=%d code=%x", binary.LittleEndian.Uint32(raw[640:644]), binary.LittleEndian.Uint32(raw[656:660]))
	}
	return ptr
}

func TestLuaCompatMuteNativeLifecycle(t *testing.T) {
	if old := legacy.Get_nox_alloc_chat_1197364(); old != nil {
		t.Fatalf("native chat allocator already initialized: %p", old)
	}
	handles.Init()
	t.Cleanup(handles.Release)
	classes := []object.Class{object.ClassSimple, object.ClassMonster, object.ClassPlayer,
		object.ClassMonster | object.ClassClientPersist, object.ClassSimple | object.ClassImmobile}
	codes := []struct {
		name            string
		net, extent     uint32
		dynamic, static uint16
	}{
		{"ordinary", 0x1234, 0x2345, 0x1234, 0xa345},
		{"zero", 0, 0, 0, 0x8000},
		{"maximum", 0x7fff, 0x7fff, 0x7fff, 0xffff},
		{"reserved-global", 0x5eac, 0x5ead, 0x5eac, 0xdead},
		{"invalid-net", 0x8000, 0x2345, 0, 0},
		{"invalid-extent", 0x1234, 0x8000, 0, 0},
	}
	for _, class := range classes {
		for _, flags := range []object.Flags{object.FlagEnabled, object.FlagEnabled | object.FlagDead, object.FlagEnabled | object.FlagDestroyed} {
			for _, code := range codes {
				for _, fanout := range []int{0, 1, 3} {
					for _, position := range []string{"absent", "head", "middle", "tail"} {
						for _, connected := range []bool{false, true} {
							for _, coop := range []bool{false, true} {
								t.Run(fmt.Sprintf("class-%08x/flags-%08x/%s/recipients-%d/%s/connected-%t/coop-%t",
									uint32(class), uint32(flags), code.name, fanout, position, connected, coop), func(t *testing.T) {
									s := scriptNSRoamFlagServer515C80(t)
									noxflags.ResetGame()
									noxflags.SetGame(noxflags.GameHost)
									if coop {
										noxflags.SetGame(noxflags.GameModeCoop)
									}
									s.SetFrame(37)
									s.SetTickRate(60)
									s.Audio.Init(s.Server)
									t.Cleanup(s.Audio.Free)
									if !s.Objs.Init(8) {
										t.Fatal("cannot initialize actual native object pool")
									}
									t.Cleanup(s.Objs.FreeObjects)
									unit := s.Objs.NewObject(&server.ObjectType{})
									if unit == nil {
										t.Fatal("cannot allocate actual server-associated speaker")
									}
									id, freeID := alloc.CString("native_mute_target")
									t.Cleanup(freeID)
									unit.IDPtr, unit.ObjClass, unit.ObjFlags = unsafe.Pointer(id), class, flags
									unit.NetCode, unit.Extent, unit.Field38 = code.net, code.extent, 0x12345678
									unit.PosVec = types.Ptf(-640.75, 19.5)
									s.Objs.List = unit
									t.Cleanup(func() { s.Objs.List, s.Objs.DeletedList = nil, nil })
									indices := []ntype.PlayerInd{7, 17, 31}
									var listeners []*server.Object
									var updates []*server.PlayerUpdateData
									for i, ind := range indices {
										listener := s.Objs.NewObject(&server.ObjectType{})
										update, freeUpdate := alloc.New(server.PlayerUpdateData{})
										t.Cleanup(freeUpdate)
										if listener == nil {
											t.Fatal("cannot allocate actual recipient unit")
										}
										player := s.Players.ByIndRaw(ind)
										player.Active, player.PlayerInd, player.PlayerUnit = 0, byte(ind), listener
										if i < fanout {
											player.Active = 1
										}
										update.Player = player
										listener.ObjClass, listener.UpdateData = object.ClassPlayer, unsafe.Pointer(update)
										listeners, updates = append(listeners, listener), append(updates, update)
									}
									empty := s.Players.ByIndRaw(21)
									empty.Active, empty.PlayerInd = 1, 21
									list := s.NetList
									list.Init()
									t.Cleanup(list.Free)
									list.ResetAll()
									type record struct {
										ptr    unsafe.Pointer
										size   uintptr
										before []byte
									}
									records := []record{{ptr: unsafe.Pointer(unit), size: unsafe.Sizeof(*unit)}}
									for i, listener := range listeners {
										records = append(records, record{ptr: unsafe.Pointer(listener), size: unsafe.Sizeof(*listener)},
											record{ptr: unsafe.Pointer(updates[i]), size: unsafe.Sizeof(*updates[i])})
									}
									for ind := ntype.PlayerInd(0); ind < 32; ind++ {
										player := s.Players.ByIndRaw(ind)
										records = append(records, record{ptr: unsafe.Pointer(player), size: unsafe.Sizeof(*player)})
									}
									for i := range records {
										if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(records[i].ptr) <= math.MaxUint32 {
											t.Fatalf("native mute input address is not above 4 GiB: %p", records[i].ptr)
										}
										records[i].before = scriptNSRoamFlagBytes515C80(records[i].ptr, records[i].size)
									}
									pool := alloc.NewClass("lua-mute-native-chat", legacy.ChatBubbleSize(), 8)
									handle := pool.UPtr()
									legacy.Set_nox_alloc_chat_1197364(handle)
									t.Cleanup(func() {
										legacy.ResetChatBubbles()
										legacy.Set_nox_alloc_chat_1197364(nil)
										pool.Free()
									})
									oldConnected := *memmap.PtrUint32(0x5d4594, 815764)
									*memmap.PtrUint32(0x5d4594, 815764) = uint32(bool2int(connected))
									t.Cleanup(func() { *memmap.PtrUint32(0x5d4594, 815764) = oldConnected })
									var game script.Game = noxScriptImpl{s}
									target, ok := game.ObjectByID("native_mute_target").(script.Chatty)
									if !ok {
										t.Fatal("actual ObjectByID did not expose Chatty")
									}
									wantCode := code.dynamic
									if class.HasAny(object.ClassClientPersist | object.ClassImmobile) {
										wantCode = code.static
									}
									wantOne := []byte{0xca, byte(wantCode), byte(wantCode >> 8)}
									wantTwo := append(append([]byte(nil), wantOne...), wantOne...)
									initialGame, initialEngine := noxflags.GetGame(), noxflags.GetEngine()
									for _, route := range []string{"native-control", "legacy-interface", "real-Lua"} {
										list.ResetAll()
										legacy.ResetChatBubbles()
										order := []uint16{0x5678, wantCode ^ 0x8000}
										switch position {
										case "head":
											order = append([]uint16{wantCode}, order...)
										case "middle":
											order = []uint16{order[0], wantCode, order[1]}
										case "tail":
											order = append(order, wantCode)
										}
										pointers := make(map[uint16]unsafe.Pointer)
										bubbles := make(map[uint16][]byte)
										for _, netCode := range order {
											ptr := luaCompatMuteSeed(t, netCode, "한글🙂 native bubble")
											pointers[netCode] = ptr
										}
										for netCode, ptr := range pointers {
											bubbles[netCode] = scriptNSRoamFlagBytes515C80(ptr, legacy.ChatBubbleSize())
										}
										var vm *lua.VM
										if route == "real-Lua" {
											vm = lua.NewVM(nil, game, "")
											vm.InitAPI("Nox", mapv0.New)
											if _, err := vm.Exec("target = Nox.Object('native_mute_target'); assert(type(target) == 'userdata')"); err != nil {
												t.Fatal(err)
											}
										}
										for repeat := 0; repeat < 2; repeat++ {
											if panicked := luaCompatSayInvoke(func() {
												switch route {
												case "native-control":
													asObjectS(unit).DestroyChat()
												case "legacy-interface":
													target.Mute()
												case "real-Lua":
													if _, err := vm.Exec("target:Mute()"); err != nil {
														t.Errorf("%s execution failed: %v", route, err)
													}
												}
											}); panicked != nil {
												t.Errorf("%s panicked: %v", route, panicked)
											}
										}
										if vm != nil {
											if err := vm.Close(); err != nil {
												t.Fatal(err)
											}
										}
										for netCode, before := range bubbles {
											if legacy.Nox_xxx_netCode2ChatBubble_48D850(int(netCode)) != int(uintptr(pointers[netCode])) ||
												!bytes.Equal(before, scriptNSRoamFlagBytes515C80(pointers[netCode], legacy.ChatBubbleSize())) {
												t.Errorf("%s changed a client bubble before packet delivery", route)
											}
										}
										var delivery []byte
										for ind := ntype.PlayerInd(0); ind < 32; ind++ {
											got := list.CopyPacketsA(ind, 1)
											var want []byte
											for i, recipient := range indices {
												if i < fanout && ind == recipient {
													want = wantTwo
												}
											}
											if !bytes.Equal(got, want) {
												t.Errorf("%s recipient %d packet=%x, want %x", route, ind, got, want)
											}
											if ind == 7 {
												delivery = got
											}
											if list.ByInd(ind, 2).Size() != 0 {
												t.Errorf("%s emitted on another queue", route)
											}
										}
										for len(delivery) >= 3 {
											if used := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(7, 0xca, delivery[:3]); used != 3 {
												t.Errorf("%s actual native client consumed %d bytes, want 3", route, used)
											}
											delivery = delivery[3:]
											for netCode, before := range bubbles {
												gone := connected && (netCode == wantCode || wantCode == 0xdead)
												got := legacy.Nox_xxx_netCode2ChatBubble_48D850(int(netCode))
												if gone {
													if got != 0 {
														t.Errorf("%s client kept removed bubble %x at %x", route, netCode, got)
													}
												} else {
													if got != int(uintptr(pointers[netCode])) {
														t.Errorf("%s client changed unrelated/disconnected bubble %x", route, netCode)
														continue
													}
													// Only the last two native pointer slots are list links.
													// Payload, expiry, drawable and all scalar bytes stay exact.
													prefix := int(legacy.ChatBubbleSize() - 2*unsafe.Sizeof(uintptr(0)))
													if !bytes.Equal(before[:prefix], scriptNSRoamFlagBytes515C80(pointers[netCode], uintptr(prefix))) {
														t.Errorf("%s client modified surviving bubble payload %x", route, netCode)
													}
													if !connected && !bytes.Equal(before, scriptNSRoamFlagBytes515C80(pointers[netCode], legacy.ChatBubbleSize())) {
														t.Errorf("%s disconnected client changed bubble links", route)
													}
												}
											}
										}
										for i, record := range records {
											if !bytes.Equal(record.before, scriptNSRoamFlagBytes515C80(record.ptr, record.size)) {
												t.Errorf("%s modified native server record %d", route, i)
											}
										}
										if s.Frame() != 37 || s.TickRate() != 60 || noxflags.GetGame() != initialGame ||
											noxflags.GetEngine() != initialEngine || s.Objs.List != unit || s.Objs.DeletedList != nil ||
											nox_client_isConnected() != connected {
											t.Errorf("%s modified unrelated runtime state", route)
										}
										if events, _ := luaCompatAudioEvents(t, s); len(events) != 0 {
											t.Errorf("%s emitted unexpected audio", route)
										}
									}
								})
							}
						}
					}
				}
			}
		}
	}
}

func TestLuaCompatMuteEmptyObject(t *testing.T) {
	var target script.Chatty = scrObject{}
	if panicked := luaCompatSayInvoke(target.Mute); panicked != nil {
		t.Fatalf("empty Chatty panicked: %v", panicked)
	}
}
