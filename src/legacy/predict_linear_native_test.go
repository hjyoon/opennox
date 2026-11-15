package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/predict_linear_original_523530_v1.json.gz
var predictLinearOriginal []byte

type predictLinearObjectState struct {
	Type      uint16    `json:"type"`
	Floats    [5]uint32 `json:"floats"`
	Direction uint16    `json:"direction"`
	Update    int       `json:"update"`
}

type predictLinearState struct {
	Objects      []predictLinearObjectState `json:"objects"`
	UpdatePlayer [3]int                     `json:"update_player"`
	PlayerIndex  [3]byte                    `json:"player_index"`
}

type predictLinearEvent struct {
	Kind        string             `json:"kind"`
	At          int                `json:"at"`
	Args        []uint32           `json:"args"`
	ReturnOwner *int               `json:"return_owner"`
	Return      *uint32            `json:"return_u32"`
	Packet      *string            `json:"packet_hex"`
	Before      predictLinearState `json:"before"`
	After       predictLinearState `json:"after"`
}

type predictLinearCase struct {
	Input struct {
		Name          string `json:"name"`
		Target        int    `json:"target"`
		NativeContext *struct {
			ClassBits uint32  `json:"class_bits"`
			NetCode   uint32  `json:"net_code_field"`
			Extent    uint32  `json:"extent"`
			Active    [3]byte `json:"active_players"`
		} `json:"native_context"`
		Mutations []struct {
			Stage  string          `json:"stage"`
			At     int             `json:"at"`
			Action string          `json:"action"`
			Value  json.RawMessage `json:"value"`
		} `json:"mutations"`
	} `json:"input"`
	Before predictLinearState   `json:"before"`
	Events []predictLinearEvent `json:"events"`
	After  predictLinearState   `json:"after"`
	Return uint32               `json:"return_u32"`
}

func predictLinearCases(t *testing.T) []predictLinearCase {
	t.Helper()
	var rows []predictLinearCase
	npcVoiceOriginalDecode(t, predictLinearOriginal, "aeed38dc906ce27c739b99fea5c4692f074009181dffedd72df80940f026ea91", "81e4d079f8ba0d350f4abb7ecb4cd203fb3f5b552d7c60c8ca0f05d7bf283f10", 683, &rows)
	return rows
}

type predictLinearArena struct {
	raw      []byte
	objects  [4]*server.Object
	updates  [3]*server.PlayerUpdateData
	players  [3]*server.Player
	layout   [18]uintptr
	expected []byte
}

func predictLinearArenaNew(t *testing.T, placement int) *predictLinearArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x10000)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("native foreign owners must be above4GiB")
	}
	v := &predictLinearArena{raw: raw, layout: predictLinearLayoutC()}
	for i := range v.objects {
		v.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x800+placement]))
	}
	for i := range v.updates {
		v.updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x4000+i*0x1000+placement]))
		v.players[i] = (*server.Player)(unsafe.Pointer(&raw[0x8000+i*0x1800+placement]))
	}
	return v
}

func (v *predictLinearArena) offset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(unsafe.Pointer(&v.raw[0])))
}

func (v *predictLinearArena) objectID(p *server.Object) int {
	for i, obj := range v.objects {
		if obj == p {
			return i
		}
	}
	panic("unmapped native prediction object")
}

func (v *predictLinearArena) object(index int) *server.Object {
	if index < 0 {
		return nil
	}
	return v.objects[index]
}

func (v *predictLinearArena) state() predictLinearState {
	s := predictLinearState{Objects: make([]predictLinearObjectState, 4)}
	for i, o := range v.objects {
		u := -1
		for j, update := range v.updates {
			if o.UpdateData == unsafe.Pointer(update) {
				u = j
			}
		}
		if u < 0 {
			panic("unmapped native update link")
		}
		s.Objects[i] = predictLinearObjectState{o.TypeInd, [5]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.Float28), math.Float32bits(o.VelVec.X), math.Float32bits(o.VelVec.Y)}, uint16(o.Direction1), u}
	}
	for i, u := range v.updates {
		owner := -1
		for j, pl := range v.players {
			if u.Player == pl {
				owner = j
			}
		}
		if owner < 0 {
			panic("unmapped native player link")
		}
		s.UpdatePlayer[i], s.PlayerIndex[i] = owner, v.players[i].PlayerInd
	}
	return s
}

// patchState writes only independently observed C field locations. All other
// bytes retain their canaries, including every unused pointer and scalar.
func (v *predictLinearArena) patchState(raw []byte, s predictLinearState) {
	putPointer := func(at int, p unsafe.Pointer) {
		if v.layout[12] == 8 {
			binary.LittleEndian.PutUint64(raw[at:at+8], uint64(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint32(raw[at:at+4], uint32(uintptr(p)))
		}
	}
	for i, o := range v.objects {
		at, state := v.offset(unsafe.Pointer(o)), s.Objects[i]
		binary.LittleEndian.PutUint16(raw[at+int(v.layout[1]):], state.Type)
		for j, offset := range []uintptr{v.layout[2], v.layout[3], v.layout[4], v.layout[5], v.layout[6]} {
			binary.LittleEndian.PutUint32(raw[at+int(offset):], state.Floats[j])
		}
		binary.LittleEndian.PutUint16(raw[at+int(v.layout[7]):], state.Direction)
		putPointer(at+int(v.layout[8]), unsafe.Pointer(v.updates[state.Update]))
	}
	for i, u := range v.updates {
		putPointer(v.offset(unsafe.Pointer(u))+int(v.layout[10]), unsafe.Pointer(v.players[s.UpdatePlayer[i]]))
		raw[v.offset(unsafe.Pointer(v.players[i]))+int(v.layout[11])] = s.PlayerIndex[i]
	}
}

func (v *predictLinearArena) prepare(t *testing.T, c predictLinearCase) {
	t.Helper()
	for i := range v.raw {
		v.raw[i] = byte(i*31 + 19)
	}
	v.patchState(v.raw, c.Before)
	if state := v.state(); !reflect.DeepEqual(state, c.Before) {
		t.Fatalf("native input state got%+v want%+v", state, c.Before)
	}
	v.expected = bytes.Clone(v.raw)
}

func (v *predictLinearArena) check(t *testing.T) {
	t.Helper()
	if bytes.Equal(v.raw, v.expected) {
		return
	}
	for i := range v.raw {
		if v.raw[i] != v.expected[i] {
			t.Fatalf("undeclared64KiB arena write at%x: got%02x want%02x", i, v.raw[i], v.expected[i])
		}
	}
}

func (v *predictLinearArena) mutate(t *testing.T, c predictLinearCase, kind string, at int) {
	t.Helper()
	for _, m := range c.Input.Mutations {
		if m.Stage != kind || m.At != at {
			continue
		}
		if m.Action == "target-fields" {
			o := v.objects[c.Input.Target]
			o.TypeInd, o.Direction1 = 0xfeab, 0xfe80
			o.PosVec, o.Float28, o.VelVec = types.Ptf(65537.75, -1.75), -1.96875, types.Ptf(7.96875, -8.0625)
			continue
		}
		var pair [2]int
		if err := json.Unmarshal(m.Value, &pair); err != nil {
			t.Fatal(err)
		}
		switch m.Action {
		case "unit-update":
			v.objects[pair[0]].UpdateData = unsafe.Pointer(v.updates[pair[1]])
		case "update-player":
			v.updates[pair[0]].Player = v.players[pair[1]]
		case "player-index":
			v.players[pair[0]].PlayerInd = byte(pair[1])
		default:
			t.Fatal("unknown declared mutation", m.Action)
		}
	}
}

func TestPredictLinearNativeLayout(t *testing.T) {
	var o server.Object
	var u server.PlayerUpdateData
	var pl server.Player
	var point types.Pointf
	want := [18]uintptr{unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Offsetof(o.TypeInd), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(point.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(point.Y), unsafe.Offsetof(o.Float28), unsafe.Offsetof(o.VelVec) + unsafe.Offsetof(point.X), unsafe.Offsetof(o.VelVec) + unsafe.Offsetof(point.Y), unsafe.Offsetof(o.Direction1), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(u.QuestWarpGate) + unsafe.Sizeof(u.QuestWarpGate), unsafe.Offsetof(u.Player), unsafe.Offsetof(pl.PlayerInd), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(o.TypeInd), unsafe.Sizeof(pl.PlayerInd), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.NetCode), unsafe.Offsetof(o.Extent)}
	if got := predictLinearLayoutC(); got != want {
		t.Fatalf("independent C/native shared-prefix layout got%v want%v", got, want)
	}
}

func TestPredictLinearNativeOriginalMatrix(t *testing.T) {
	rows := predictLinearCases(t)
	old := predictLinearRuntimeForCall523530
	t.Cleanup(func() { predictLinearRuntimeForCall523530 = old })
	checks := 0
	for route := 0; route < 4; route++ {
		for _, placement := range []int{0, 64, 512} {
			v := predictLinearArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					v.prepare(t, c)
					cursor := 0
					take := func(kind string, args []uint32, packet []byte) predictLinearEvent {
						if cursor == len(c.Events) {
							t.Fatal("unexpected extra service", kind)
						}
						e := c.Events[cursor]
						cursor++
						if kind != e.Kind || !slices.Equal(args, e.Args) || !reflect.DeepEqual(v.state(), e.Before) {
							t.Fatalf("event%d kind/args/state: got%s%v want%+v", cursor, kind, args, e)
						}
						v.check(t)
						if e.Packet != nil {
							want, err := hex.DecodeString(*e.Packet)
							if err != nil || !bytes.Equal(packet, want) {
								t.Fatalf("exact original packet got%x want%x err%v", packet, want, err)
							}
						} else if packet != nil {
							t.Fatal("packet supplied to non-send service")
						}
						v.mutate(t, c, kind, e.At)
						v.patchState(v.expected, e.After)
						if !reflect.DeepEqual(v.state(), e.After) {
							t.Fatal("declared service mutation mismatch", kind)
						}
						v.check(t)
						return e
					}
					predictLinearRuntimeForCall523530 = func() server.PredictLinearNativeRuntime523530 {
						return server.PredictLinearNativeRuntime523530{
							NetCode: func(obj *server.Object) uint32 {
								e := take("net-code", []uint32{uint32(v.objectID(obj))}, nil)
								if e.Return == nil {
									t.Fatal("missing declared net-code result")
								}
								return *e.Return
							},
							First: func() *server.Object {
								e := take("first", nil, nil)
								if e.ReturnOwner == nil {
									t.Fatal("missing first native owner")
								}
								return v.object(*e.ReturnOwner)
							},
							Next: func(unit *server.Object) *server.Object {
								e := take("next", []uint32{uint32(v.objectID(unit))}, nil)
								if e.ReturnOwner == nil {
									t.Fatal("missing next native owner")
								}
								return v.object(*e.ReturnOwner)
							},
							Send: func(index uint8, packet []byte) {
								take("send", []uint32{uint32(index), 1, uint32(len(packet))}, packet)
							},
						}
					}
					if result := predictLinearActualCall(route, v.objects[c.Input.Target]); uint32(result) != c.Return {
						t.Fatalf("int result%#x want%#x", uint32(result), c.Return)
					}
					if cursor != len(c.Events) || !reflect.DeepEqual(v.state(), c.After) {
						t.Fatal("missing service or final state", cursor, len(c.Events))
					}
					v.check(t)
					checks++
					runtime.KeepAlive(v)
				})
			}
		}
	}
	if checks != 8196 {
		t.Fatal("missing route/placement/original cases", checks)
	}
	t.Log("683 unchanged original cases x4 public-C/typed-C/production-Go/matching-caller-expression routes x3 native foreign placements; whole64KiB arena, full target/unit/update/player identities, exact packets and live mutations")
}

//go:embed testdata/predict_linear_default_523530_v1.json.gz
var predictLinearDefaultOriginal []byte

func TestPredictLinearNativeDefaultNetwork(t *testing.T) {
	const child = "OPENNOX_PREDICT_LINEAR_DEFAULT_CHILD"
	if os.Getenv(child) != "1" {
		program, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(program, "-test.run=^TestPredictLinearNativeDefaultNetwork$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual default prediction composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var rows []predictLinearCase
	npcVoiceOriginalDecode(t, predictLinearDefaultOriginal, "f4d05b629ea4dec73a2b5626976a3db8095fed70b2aa45d2f65405adcf16d3de", "ca4c75c13196bf35d63814bc4467effc502190ed360d38ac979ad66055792dc9", 48, &rows)
	s, first := generatorDamageCompositionServer(t)
	players := [3]*server.Player{first, s.Players.NewRaw(1002), s.Players.NewRaw(1003)}
	for _, pl := range players {
		if pl == nil {
			t.Fatal("actual player allocation")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(pl)) <= math.MaxUint32 {
			t.Fatal("actual players must be native high owners")
		}
	}
	v := predictLinearArenaNew(t, 64)
	v.players = players
	t.Cleanup(func() {
		for _, pl := range players {
			pl.PlayerUnit = nil
		}
	})
	compositions := 0
	for route := 0; route < 4; route++ {
		for _, c := range rows {
			if c.Input.NativeContext == nil || c.Input.Target != 0 || len(c.Input.Mutations) != 0 {
				t.Fatal("invalid original default context")
			}
			for i := range v.raw {
				v.raw[i] = byte(i*31 + 19)
			}
			for i, o := range v.objects {
				state := c.Before.Objects[i]
				*o = server.Object{
					TypeInd: state.Type, PosVec: types.Ptf(math.Float32frombits(state.Floats[0]), math.Float32frombits(state.Floats[1])),
					Float28: math.Float32frombits(state.Floats[2]), VelVec: types.Ptf(math.Float32frombits(state.Floats[3]), math.Float32frombits(state.Floats[4])),
					UpdateData: unsafe.Pointer(v.updates[state.Update]),
				}
				o.Direction1 = 0xfe80
				if uint16(o.Direction1) != state.Direction {
					t.Fatal("default direction input")
				}
				classBits, code, extent := uint32(4), uint32(0), uint32(0)
				if i == 0 {
					classBits, code, extent = c.Input.NativeContext.ClassBits, c.Input.NativeContext.NetCode, c.Input.NativeContext.Extent
				}
				predictLinearContextC(o, classBits, code, extent)
			}
			for i, u := range v.updates {
				u.Player = players[c.Before.UpdatePlayer[i]]
				players[i].PlayerUnit = v.objects[i+1]
				players[i].Active = c.Input.NativeContext.Active[i]
			}
			if !reflect.DeepEqual(v.state(), c.Before) {
				t.Fatal("actual default native input differs from original selected fields")
			}
			if c.Events[0].Return == nil || uint32(s.GetUnitNetCode(v.objects[0])) != *c.Events[0].Return {
				t.Fatal("actual default net-code differs from executed original helper")
			}
			var packet []byte
			for _, event := range c.Events {
				if event.Kind == "send" {
					want, err := hex.DecodeString(*event.Packet)
					if err != nil {
						t.Fatal(err)
					}
					if packet == nil {
						packet = want
					} else if !bytes.Equal(packet, want) {
						t.Fatal("original snapshot changed between receivers")
					}
				}
			}
			if len(packet) != 14 {
				t.Fatal("missing original default packet")
			}
			before := bytes.Clone(v.raw)
			if result := predictLinearActualCall(route, v.objects[0]); uint32(result) != c.Return {
				t.Fatal("actual default result", result, c.Return)
			}
			if !bytes.Equal(v.raw, before) || !reflect.DeepEqual(v.state(), c.After) {
				t.Fatal("default native64KiB owner/state changed")
			}
			v.objects[0].PosVec, v.objects[0].TypeInd = types.Pointf{}, 0
			for i, pl := range players {
				var want []byte
				if c.Input.NativeContext.Active[i] != 0 {
					want = packet
				}
				got := s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1)
				if !bytes.Equal(got, want) {
					t.Fatalf("%s/route%d/player%d actual queue%x want%x", c.Input.Name, route, i, got, want)
				}
				other := s.NetList.ByInd(ntype.PlayerInd(pl.PlayerInd), netlist.Kind2)
				if other.Size() != 0 || other.Count() != 0 {
					t.Fatal("prediction sent to wrong queue", other.Size(), other.Count())
				}
			}
			compositions++
		}
	}
	if compositions != 192 {
		t.Fatal("missing actual default compositions", compositions)
	}
	t.Log("actual default prediction compositions=192; no runtime override, C-initialized context, actual net-code/first/next and NetList, inactive-player gap, native high unit/update/player identities and copied14byte packet")
}

func TestPredictLinearNativeNestedPacketOwnership(t *testing.T) {
	old := predictLinearRuntimeForCall523530
	t.Cleanup(func() { predictLinearRuntimeForCall523530 = old })
	var c *predictLinearCase
	for _, row := range predictLinearCases(t) {
		if row.Input.Target == 0 && len(row.Events) == 4 && row.Events[2].Kind == "send" {
			copy := row
			c = &copy
			break
		}
	}
	if c == nil || c.Events[2].Packet == nil || c.Events[0].Return == nil {
		t.Fatal("missing single-receiver original case")
	}
	wire, err := hex.DecodeString(*c.Events[2].Packet)
	if err != nil {
		t.Fatal(err)
	}
	for route := 0; route < 4; route++ {
		v := predictLinearArenaNew(t, 64)
		v.prepare(t, *c)
		depth, sends := 0, 0
		var outer []byte
		predictLinearRuntimeForCall523530 = func() server.PredictLinearNativeRuntime523530 {
			return server.PredictLinearNativeRuntime523530{
				NetCode: func(obj *server.Object) uint32 {
					if obj != v.objects[0] {
						t.Fatal("nested target lost native identity")
					}
					return *c.Events[0].Return
				},
				First: func() *server.Object { return v.object(*c.Events[1].ReturnOwner) },
				Next: func(unit *server.Object) *server.Object {
					if unit != v.object(*c.Events[1].ReturnOwner) {
						t.Fatal("nested recipient lost native identity")
					}
					return nil
				},
				Send: func(index uint8, packet []byte) {
					sends++
					if uint32(index) != c.Events[2].Args[0] || !bytes.Equal(packet, wire) {
						t.Fatal("nested original packet/recipient", index, packet)
					}
					if depth == 0 {
						outer = packet
						outer[0] = 0x5a
						depth++
						if result := predictLinearActualCall(route, v.objects[0]); result != 0 {
							t.Fatal("nested result", result)
						}
						depth--
						if outer[0] != 0x5a || !bytes.Equal(outer[1:], wire[1:]) {
							t.Fatal("nested call overwrote outer packet storage")
						}
					} else if unsafe.SliceData(packet) == unsafe.SliceData(outer) {
						t.Fatal("nested packet shares outer storage")
					}
				},
			}
		}
		if result := predictLinearActualCall(route, v.objects[0]); result != 0 || sends != 2 {
			t.Fatal("outer result/send count", result, sends)
		}
		v.check(t)
	}
}
