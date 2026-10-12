package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/script"
	"github.com/opennox/libs/script/lua"
	"github.com/opennox/libs/script/lua/mapv0"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/v4/audio"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// This is an explicit AUD input, consumed by the real descriptor parser.
// No sound lookup, event creation, packet enqueue, or position result is hooked.
type luaCompatAudioFile struct{ *bytes.Reader }

func (f luaCompatAudioFile) Skip(n int) {
	if _, err := f.Seek(int64(n), io.SeekCurrent); err != nil {
		panic(err)
	}
}
func (f luaCompatAudioFile) ReadU8() uint8 {
	v, err := f.ReadByte()
	if err != nil {
		panic(err)
	}
	return v
}
func (f luaCompatAudioFile) ReadI8() int8 { return int8(f.ReadU8()) }
func (f luaCompatAudioFile) ReadU16() uint16 {
	var b [2]byte
	if _, err := io.ReadFull(f, b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint16(b[:])
}
func (f luaCompatAudioFile) ReadI16() int16 { return int16(f.ReadU16()) }
func (f luaCompatAudioFile) ReadU32() uint32 {
	var b [4]byte
	if _, err := io.ReadFull(f, b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(b[:])
}
func (f luaCompatAudioFile) ReadI32() int32 { return int32(f.ReadU32()) }
func (f luaCompatAudioFile) ReadString8() (string, error) {
	n, err := f.ReadByte()
	if err != nil {
		return "", err
	}
	b := make([]byte, int(n))
	_, err = io.ReadFull(f, b)
	return string(b), err
}

var luaCompatAudioNames = []struct {
	name string
	id   sound.ID
}{
	{"AnchorCast", sound.SoundAnchorCast},
	{"WalkOnStone", sound.SoundWalkOnStone},
	{"RunOnStone", sound.SoundRunOnStone},
	{"MetalWeaponPickup", sound.SoundMetalWeaponPickup},
	{"MetalWeaponDrop", sound.SoundMetalWeaponDrop},
	{"", 0},
	{"walkonstone", 0},
	{"no-such-audio", 0},
	{"WalkOnStone\x00suffix", 0},
	{"발걸음", 0},
}

func luaCompatAudioDefinitions(t *testing.T, s *Server) {
	t.Helper()
	wire := binary.LittleEndian.AppendUint32(nil, 5)
	for _, item := range luaCompatAudioNames[:5] {
		wire = append(wire, byte(len(item.name)))
		wire = append(wire, item.name...)
		wire = binary.LittleEndian.AppendUint16(wire, 8)
		wire = append(wire, 0)
		wire = binary.LittleEndian.AppendUint16(wire, 40)
		wire = append(wire, 2, 0xa1, 0xa2, 0xa3, 4, 't', 'e', 's', 't', 0)
	}
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("actual AUD parser did not consume the complete sound definitions")
	}
	for _, item := range luaCompatAudioNames[:5] {
		if s.Audio.Field12(item.id) != 1 || s.Audio.MaxDist(item.id) != 600 || s.Audio.Flags(item.id) != 8 {
			t.Fatalf("actual AUD descriptor differs for %s", item.name)
		}
	}
}

type luaCompatAudioPosition struct {
	point types.Pointf
	reads int
}

func (p *luaCompatAudioPosition) Pos() types.Pointf {
	p.reads++
	return p.point
}

// string.char preserves every byte, including NUL and UTF-8; Lua escape
// conventions are not assumed to match Go's quoted string conventions.
func luaCompatAudioString(value string) string {
	parts := make([]string, len(value))
	for i := 0; i < len(value); i++ {
		parts[i] = strconv.Itoa(int(value[i]))
	}
	return "string.char(" + strings.Join(parts, ",") + ")"
}

type luaCompatAudioSnapshot struct {
	sound sound.ID
	x, y  uint32
	obj   *server.Object
	kind  int
	code  uint32
}

func luaCompatAudioEvents(t *testing.T, s *Server) ([]luaCompatAudioSnapshot, []*server.AudioEvent) {
	t.Helper()
	var out []luaCompatAudioSnapshot
	var events []*server.AudioEvent
	s.Audio.EachEvent(func(event *server.AudioEvent) {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(event)) <= math.MaxUint32 {
			t.Fatalf("audio event is not native-width: %p", event)
		}
		out = append(out, luaCompatAudioSnapshot{
			event.Sound, math.Float32bits(event.Pos.X), math.Float32bits(event.Pos.Y), event.Obj, event.Kind, event.Code,
		})
		events = append(events, event)
	})
	return out, events
}

func TestLuaCompatAudioEffectNativeQueueAndPackets(t *testing.T) {
	for _, position := range []string{"nil", "coordinates", "object", "destroyed-object", "custom-positioner"} {
		for _, item := range luaCompatAudioNames {
			for _, timing := range []string{"immediate", "delayed", "blocked-now", "blocked-on-replay", "unblocked-on-replay"} {
				for _, loaded := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/name-%x/%s/loaded-%t", position, item.name, timing, loaded), func(t *testing.T) {
						s := scriptNSRoamFlagServer515C80(t)
						noxflags.ResetGame()
						noxflags.SetGame(noxflags.GameHost)
						s.SetFrame(37)
						s.Audio.Init(s.Server)
						t.Cleanup(s.Audio.Free)
						if loaded {
							luaCompatAudioDefinitions(t, s)
						}
						if got := sound.ByName(item.name); got != item.id {
							t.Fatalf("actual sound-name lookup %q = %d, want %d", item.name, got, item.id)
						}

						unit, freeUnit := alloc.New(server.Object{})
						listener, freeListener := alloc.New(server.Object{})
						update, freeUpdate := alloc.New(server.PlayerUpdateData{})
						player, freePlayer := alloc.New(server.Player{})
						id, freeID := alloc.CString("native_audio_source")
						t.Cleanup(freeID)
						t.Cleanup(freePlayer)
						t.Cleanup(freeUpdate)
						t.Cleanup(freeListener)
						t.Cleanup(freeUnit)
						*unit = server.Object{
							ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled,
							IDPtr: unsafe.Pointer(id), PosVec: types.Ptf(-640.75, 19.5), Extent: 0xe1234567,
						}
						if position == "destroyed-object" {
							unit.ObjFlags |= object.FlagDestroyed
						}
						*player = server.Player{PlayerInd: 17, Pos3632Vec: types.Ptf(17.25, 20)}
						*update = server.PlayerUpdateData{Player: player}
						*listener = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
						s.Objs.List = unit
						t.Cleanup(func() { s.Objs.List = nil })
						list := netlist.New()
						list.Init()
						t.Cleanup(list.Free)
						oldNetList := s.NetList
						s.NetList = list
						t.Cleanup(func() { s.NetList = oldNetList })
						for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(listener), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(id)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
								t.Fatalf("native audio input address is below 4 GiB: %p", p)
							}
						}
						var game script.Game = noxScriptImpl{s}
						vm := lua.NewVM(nil, game, "")
						vm.InitAPI("Nox", mapv0.New)
						t.Cleanup(func() { _ = vm.Close() })
						if _, err := vm.Exec(`source = Nox.Object("native_audio_source"); assert(type(source) == "userdata")`); err != nil {
							t.Fatal(err)
						}
						point := types.Ptf(13.5, -21.25)
						var pos script.Positioner = point
						luaPos := "13.5, -21.25"
						probe := &luaCompatAudioPosition{point: types.Ptf(1234.75, -47.5)}
						switch position {
						case "nil":
							pos, luaPos = nil, "nil"
						case "object", "destroyed-object":
							pos, luaPos, point = scrObject{asObjectS(unit)}, "source", unit.PosVec
						case "custom-positioner":
							pos, point = probe, probe.point
						}
						before := make([][]byte, 4)
						for i, p := range []struct {
							ptr  unsafe.Pointer
							size uintptr
						}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(listener), unsafe.Sizeof(*listener)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(player), unsafe.Sizeof(*player)}} {
							before[i] = scriptNSRoamFlagBytes515C80(p.ptr, p.size)
						}

						var reference []luaCompatAudioSnapshot
						for _, route := range []string{"namespace-control", "legacy-game", "real-Lua"} {
							if route == "real-Lua" && position == "custom-positioner" {
								continue // arbitrary Go implementations have no Lua constructor
							}
							s.Audio.Free()
							s.Audio.Init(s.Server)
							if loaded {
								luaCompatAudioDefinitions(t, s)
							}
							noxflags.UnsetGame(noxflags.GameFlag20)
							if timing == "immediate" || timing == "blocked-now" {
								s.Audio.Reset()
							}
							if timing == "blocked-now" || timing == "unblocked-on-replay" {
								noxflags.SetGame(noxflags.GameFlag20)
							}
							probe.reads = 0
							for repeat := 0; repeat < 2; repeat++ {
								switch route {
								case "namespace-control":
									s.NoxScript().AudioEvent(audio.Name(item.name), pos)
								case "legacy-game":
									game.AudioEffect(item.name, pos)
								case "real-Lua":
									code := "Nox.Audio.Effect(" + luaCompatAudioString(item.name) + ", " + luaPos + ")"
									if _, err := vm.Exec(code); err != nil {
										t.Errorf("real Lua Audio.Effect: %v", err)
									}
								}
							}
							got, _ := luaCompatAudioEvents(t, s)
							if timing != "immediate" && timing != "blocked-now" {
								if len(got) != 0 {
									t.Errorf("%s produced audio before next-frame replay", route)
								}
								if timing == "blocked-on-replay" {
									noxflags.SetGame(noxflags.GameFlag20)
								} else {
									noxflags.UnsetGame(noxflags.GameFlag20)
								}
								s.Audio.Reset()
							}
							got, events := luaCompatAudioEvents(t, s)
							wantCount := 0
							if loaded && item.id != 0 && pos != nil && timing != "blocked-now" && timing != "blocked-on-replay" {
								wantCount = 2
							}
							if len(got) != wantCount {
								t.Errorf("%s event count=%d, want %d", route, len(got), wantCount)
							}
							for _, event := range got {
								want := luaCompatAudioSnapshot{item.id, math.Float32bits(point.X), math.Float32bits(point.Y), nil, 0, 0}
								if event != want {
									t.Errorf("%s event=%+v, want %+v", route, event, want)
								}
							}
							if position == "custom-positioner" && probe.reads != 2 {
								t.Errorf("%s position reads=%d, want 2 even for disabled or unknown sound", route, probe.reads)
							}
							if route == "namespace-control" {
								reference = got
							} else if !fmtAudioSnapshotsEqual(reference, got) {
								t.Errorf("%s differs from actual native namespace", route)
							}

							for _, event := range events {
								if !s.Server.SendAudioEvent501FD0(listener, event, 50, server.AudioEventPacketRuntime501FD0{
									WindowWidth: func() int32 { return 640 },
								}) {
									t.Errorf("%s real audio packet enqueue failed", route)
								}
							}
							packet := list.CopyPacketsA(ntype.PlayerInd(player.PlayerInd), netlist.Kind1)
							var wantPacket []byte
							distance := int32(math.Trunc(float64(point.X - player.Pos3632Vec.X)))
							pan := byte(distance * 50 / 320)
							packed := uint16(item.id) | uint16(50<<10)
							for i := 0; i < wantCount; i++ {
								wantPacket = append(wantPacket, 0xa6, pan, byte(packed), byte(packed>>8))
							}
							if !bytes.Equal(packet, wantPacket) {
								t.Errorf("%s wire=%x, want %x", route, packet, wantPacket)
							}
							for ind := 0; ind < 32; ind++ {
								if ind != int(player.PlayerInd) && len(list.CopyPacketsA(ntype.PlayerInd(ind), netlist.Kind1)) != 0 {
									t.Errorf("%s sent audio to unexpected recipient %d", route, ind)
								}
							}
							for i, p := range []struct {
								ptr  unsafe.Pointer
								size uintptr
							}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(listener), unsafe.Sizeof(*listener)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(player), unsafe.Sizeof(*player)}} {
								if !bytes.Equal(before[i], scriptNSRoamFlagBytes515C80(p.ptr, p.size)) {
									t.Errorf("%s changed unrelated native record %d", route, i)
								}
							}
							if s.Frame() != 37 || alloc.GoString(id) != "native_audio_source" {
								t.Errorf("%s changed game frame or native object name", route)
							}
						}
					})
				}
			}
		}
	}
}

func fmtAudioSnapshotsEqual(a, b []luaCompatAudioSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
