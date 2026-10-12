package opennox

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/script"
	"github.com/opennox/libs/script/lua"
	"github.com/opennox/libs/script/lua/mapv0"
	"github.com/opennox/libs/types"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func luaCompatSayInvoke(f func()) (panicked any) {
	defer func() { panicked = recover() }()
	f()
	return nil
}

// Preserve every input byte while keeping each string.char call below Lua's
// compiler register limit. No duration, text, payload or expected value changes.
func luaCompatSayString(value string) string {
	var parts []string
	for len(value) > 64 {
		parts = append(parts, luaCompatAudioString(value[:64]))
		value = value[64:]
	}
	parts = append(parts, luaCompatAudioString(value))
	return "(" + strings.Join(parts, " .. ") + ")"
}

// The real native namespace is a separate lifecycle control. Expectations below
// are literal MSG_TEXT_MESSAGE fields/payloads; no sender, enqueue, string,
// duration conversion, iterator or recipient result is hooked. Long inputs use
// the existing native sender's bounded policy, not PE32 stack-overflow behavior.
// These are real packet queues, not remote transport/client bubble-render tests.
func TestLuaCompatSayNativePackets(t *testing.T) {
	messages := []struct {
		name, input  string
		flags, count byte
		body         []byte
	}{
		{"empty", "", 2, 1, []byte{0}},
		{"ascii", "Hello", 2, 6, []byte{'H', 'e', 'l', 'l', 'o', 0}},
		{"latin1", "é", 2, 2, []byte{0xe9, 0}},
		{"wide-surrogate", "漢🙂", 4, 4, []byte{0x22, 0x6f, 0x3d, 0xd8, 0x42, 0xde, 0, 0}},
		{"korean", "한글", 4, 3, []byte{0x5c, 0xd5, 0x00, 0xae, 0, 0}},
		{"embedded-nul", "ab\x00ignored漢", 2, 3, []byte{'a', 'b', 0}},
		{"latin1-nul", "é\x00漢", 2, 2, []byte{0xe9, 0}},
		{"bounded-narrow", strings.Repeat("a", 300), 2, 255, append(bytes.Repeat([]byte{'a'}, 254), 0)},
		{"bounded-wide", strings.Repeat("漢", 254), 4, 254, append(bytes.Repeat([]byte{0x22, 0x6f}, 253), 0, 0)},
		{"bounded-surrogate", strings.Repeat("漢", 252) + "🙂", 4, 253, append(bytes.Repeat([]byte{0x22, 0x6f}, 252), 0, 0)},
	}
	durations := []struct {
		name    string
		value   script.Duration
		lua     string
		wire    uint16
		liveLua bool
	}{
		{"omitted", script.Duration{}, "", 0, true},
		{"frames-zero", script.Frames(0), "Nox.Frames(0)", 0, true},
		{"frames-one", script.Frames(1), "Nox.Frames(1)", 1, true},
		{"frames-max", script.Frames(65535), "Nox.Frames(65535)", 65535, true},
		{"frames-wrap", script.Frames(65536), "Nox.Frames(65536)", 0, true},
		{"frames-negative", script.Frames(-1), "Nox.Frames(-1)", 65535, true},
		{"seconds", script.Time(1500 * time.Millisecond), "Nox.Sec(1.5)", 90, true},
		// The old Lua API has no Infinite constructor. Keep the Go interface
		// case instead of substituting an omitted/zero Lua duration.
		{"infinite-go", ns4.Infinite(), "", 0, false},
	}
	classes := []object.Class{object.ClassSimple, object.ClassMonster, object.ClassPlayer,
		object.ClassMonster | object.ClassClientPersist, object.ClassSimple | object.ClassImmobile}
	for _, class := range classes {
		for _, flags := range []object.Flags{object.FlagEnabled, object.FlagEnabled | object.FlagDead, object.FlagEnabled | object.FlagDestroyed} {
			for _, message := range messages {
				for _, duration := range durations {
					for _, fanout := range []int{0, 1, 3} {
						for _, coop := range []bool{false, true} {
							t.Run(fmt.Sprintf("class-%08x/flags-%08x/%s/%s/recipients-%d/coop-%t",
								uint32(class), uint32(flags), message.name, duration.name, fanout, coop), func(t *testing.T) {
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
									t.Fatal("cannot allocate server-associated speaker")
								}
								id, freeID := alloc.CString("native_say_target")
								t.Cleanup(freeID)
								unit.ObjClass, unit.ObjFlags = class, flags
								unit.NetCode, unit.Extent, unit.Field38 = 0x1234, 0x2345, 0x12345678
								unit.PosVec, unit.IDPtr = types.Ptf(-640.75, 19.5), unsafe.Pointer(id)
								s.Objs.List = unit
								t.Cleanup(func() { s.Objs.List, s.Objs.DeletedList = nil, nil })
								var recipientUnits []*server.Object
								var updates []*server.PlayerUpdateData
								indices := []ntype.PlayerInd{7, 17, 31}
								for i, ind := range indices {
									listener := s.Objs.NewObject(&server.ObjectType{})
									update, freeUpdate := alloc.New(server.PlayerUpdateData{})
									t.Cleanup(freeUpdate)
									if listener == nil {
										t.Fatal("cannot allocate actual recipient unit")
									}
									player := s.Players.ByIndRaw(ind)
									if player == nil {
										t.Fatal("actual player slot absent")
									}
									player.Active, player.PlayerInd, player.PlayerUnit = 0, byte(ind), listener
									if i < fanout {
										player.Active = 1
									}
									update.Player = player
									listener.ObjClass, listener.UpdateData = object.ClassPlayer, unsafe.Pointer(update)
									recipientUnits, updates = append(recipientUnits, listener), append(updates, update)
								}
								// A real active player without a unit must not receive this packet.
								emptyUnitPlayer := s.Players.ByIndRaw(21)
								emptyUnitPlayer.Active, emptyUnitPlayer.PlayerInd = 1, 21
								list := netlist.New()
								list.Init()
								t.Cleanup(list.Free)
								oldList := s.NetList
								s.NetList = list
								t.Cleanup(func() { s.NetList = oldList })
								type nativeRecord struct {
									ptr    unsafe.Pointer
									size   uintptr
									before []byte
								}
								records := []nativeRecord{{unsafe.Pointer(unit), unsafe.Sizeof(*unit), nil}}
								for i, listener := range recipientUnits {
									records = append(records, nativeRecord{unsafe.Pointer(listener), unsafe.Sizeof(*listener), nil},
										nativeRecord{unsafe.Pointer(updates[i]), unsafe.Sizeof(*updates[i]), nil})
								}
								for ind := ntype.PlayerInd(0); ind < 32; ind++ {
									player := s.Players.ByIndRaw(ind)
									records = append(records, nativeRecord{unsafe.Pointer(player), unsafe.Sizeof(*player), nil})
								}
								for i := range records {
									if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(records[i].ptr) <= math.MaxUint32 {
										t.Fatalf("native chat input address is not above 4 GiB: %p", records[i].ptr)
									}
									records[i].before = scriptNSRoamFlagBytes515C80(records[i].ptr, records[i].size)
								}
								var game script.Game = noxScriptImpl{s}
								target, ok := game.ObjectByID("native_say_target").(script.Chatty)
								if !ok {
									t.Fatal("actual ObjectByID did not expose Chatty")
								}
								initialGame, initialEngine := noxflags.GetGame(), noxflags.GetEngine()
								code := uint16(0x1234)
								if class.HasAny(object.ClassClientPersist | object.ClassImmobile) {
									code = 0xa345
								}
								wantOne := []byte{0xa8, byte(code), byte(code >> 8), message.flags, 0x80, 0xfd, 0x13, 0x00,
									message.count, byte(duration.wire), byte(duration.wire >> 8)}
								wantOne = append(wantOne, message.body...)
								wantTwo := append(append([]byte(nil), wantOne...), wantOne...)
								var reference [32][]byte
								for _, route := range []string{"native-control", "legacy-interface", "real-Lua"} {
									if route == "real-Lua" && !duration.liveLua {
										continue
									}
									list.ResetAll()
									var vm *lua.VM
									if route == "real-Lua" {
										vm = lua.NewVM(nil, game, "")
										vm.InitAPI("Nox", mapv0.New)
										if _, err := vm.Exec("target = Nox.Object('native_say_target'); assert(type(target) == 'userdata')"); err != nil {
											t.Fatal(err)
										}
									}
									for repeat := 0; repeat < 2; repeat++ {
										invoke := func() {
											switch route {
											case "native-control":
												asObjectS(unit).ChatStrTimer(message.input, duration.value)
											case "legacy-interface":
												target.Say(message.input, duration.value)
											case "real-Lua":
												arg := ""
												if duration.lua != "" {
													arg = ", " + duration.lua
												}
												if _, err := vm.Exec("target:Say(" + luaCompatSayString(message.input) + arg + ")"); err != nil {
													t.Errorf("%s execution failed: %v", route, err)
												}
											}
										}
										if p := luaCompatSayInvoke(invoke); p != nil {
											t.Errorf("%s panicked: %v", route, p)
										}
									}
									if vm != nil {
										if err := vm.Close(); err != nil {
											t.Fatal(err)
										}
									}
									for ind := ntype.PlayerInd(0); ind < 32; ind++ {
										got := list.CopyPacketsA(ind, netlist.Kind1)
										var want []byte
										for i, recipient := range indices {
											if i < fanout && ind == recipient {
												want = wantTwo
											}
										}
										if !bytes.Equal(got, want) {
											t.Errorf("%s recipient %d packet=%x, want %x", route, ind, got, want)
										}
										if route == "native-control" {
											reference[ind] = append([]byte(nil), got...)
										} else if !bytes.Equal(got, reference[ind]) {
											t.Errorf("%s recipient %d differs from actual native namespace", route, ind)
										}
										if list.ByInd(ind, netlist.Kind2).Size() != 0 {
											t.Errorf("%s emitted on another queue", route)
										}
									}
									for i, record := range records {
										if !bytes.Equal(record.before, scriptNSRoamFlagBytes515C80(record.ptr, record.size)) {
											t.Errorf("%s modified native record %d", route, i)
										}
									}
									if s.Frame() != 37 || s.TickRate() != 60 || noxflags.GetGame() != initialGame ||
										noxflags.GetEngine() != initialEngine || s.Objs.List != unit || s.Objs.DeletedList != nil {
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

func TestLuaCompatSayEmptyObject(t *testing.T) {
	var target script.Chatty = scrObject{}
	if p := luaCompatSayInvoke(func() { target.Say("empty target", script.Frames(1)) }); p != nil {
		t.Fatalf("empty Chatty panicked: %v", p)
	}
}
