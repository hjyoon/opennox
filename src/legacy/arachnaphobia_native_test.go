package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/arachnaphobia_original_52dc80_v1.json.gz
var arachnaphobiaOriginal []byte

// Signed input owner identities include -1. The independently checked
// canonical frame represents that same nil owner with DWORD 0xffffffff.
type arachnaphobiaNativeCase struct {
	Input struct {
		Name      string      `json:"name"`
		ID        int32       `json:"id"`
		Second    int         `json:"second"`
		Owner     int         `json:"owner"`
		Caster    int         `json:"caster"`
		Arg       int         `json:"arg"`
		Level     int32       `json:"level"`
		Lookup    uint32      `json:"lookup"`
		Trace     uint32      `json:"trace"`
		Allocate  int         `json:"allocate"`
		Arguments [2][3]int64 `json:"arguments"`
		Mutations []struct {
			Kind   string `json:"kind"`
			At     int    `json:"at"`
			Action string `json:"action"`
			Value  uint32 `json:"value"`
			Object int    `json:"object"`
			Field  int    `json:"field"`
			Arg    int    `json:"arg"`
			Slot   int    `json:"slot"`
		} `json:"mutations"`
	} `json:"input"`
	Before [117]uint32       `json:"before"`
	Events []meteorCastEvent `json:"events"`
	After  [117]uint32       `json:"after"`
	Return uint32            `json:"return_u32"`
}

func TestArachnaphobiaNativeLayout(t *testing.T) {
	var o server.Object
	var arg server.SpellAcceptArg
	var update server.PlayerUpdateData
	var player server.Player
	want := [9]uintptr{
		unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.Y), unsafe.Offsetof(o.UpdateData),
		unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.X), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.Y),
		unsafe.Offsetof(update.Player), unsafe.Offsetof(player.PlayerInd),
	}
	if got := arachnaphobiaLayoutC(); got != want {
		t.Fatalf("independent native C prefixes=%v Go=%v", got, want)
	}
	if unsafe.Sizeof(o) > 0x400 || unsafe.Sizeof(update) > 0x400 || unsafe.Sizeof(player) > 0x2000 {
		t.Fatal("shared native foreign-arena strides do not fit records")
	}
}

func TestArachnaphobiaNativeOriginalMatrix(t *testing.T) {
	var rows []arachnaphobiaNativeCase
	npcVoiceOriginalDecode(t, arachnaphobiaOriginal, "d7ac9805a115f6ebd2418dcd8d198b52f7b2e25d49b5650b7106bcb00f0ead8e", "d0d0fe4dc0354072270cea3bb55fe992e72ce96aea12c4e87a730bd023384ce8", 62, &rows)
	previous := arachnaphobiaCastCall52DC80
	t.Cleanup(func() { arachnaphobiaCastCall52DC80 = previous })
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			// The existing independent C-offset arena supplies complete native
			// Object/argument/update/player records and 64 KiB canaries.
			a := meteorCastArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					for i := range a.raw {
						a.raw[i] = byte(i*31 + 19)
					}
					a.setState(c.Before)
					a.expected = bytes.Clone(a.raw)
					if a.words() != c.Before {
						t.Fatal("C-initialized native frame differs from original input")
					}
					cursor := 0
					counts := make(map[string]int)
					emit := func(e meteorCastEvent) {
						e.At = counts[e.Kind]
						counts[e.Kind]++
						if cursor >= len(c.Events) || !reflect.DeepEqual(e, c.Events[cursor]) {
							var want any = "end"
							if cursor < len(c.Events) {
								want = c.Events[cursor]
							}
							t.Fatalf("event%d got%+v want%+v", cursor, e, want)
						}
						cursor++
					}
					service := func(kind string, args []uint32, result uint32) uint32 {
						a.check(t)
						before := a.hash()
						if kind == "create" {
							o := a.object(int(args[0]))
							o.PosVec = types.Ptf(math.Float32frombits(args[2]), math.Float32frombits(args[3]))
							o.ObjOwner = a.object(int(args[1]))
						}
						for _, m := range c.Input.Mutations {
							if m.Kind != kind || m.At != counts[kind] {
								continue
							}
							state := a.words()
							var index int
							value := m.Value
							switch m.Action {
							case "object":
								index = 11*m.Object + m.Field
								if m.Field == 0 {
									value &= 65535
								}
							case "arg":
								index = 66 + 3*m.Arg + m.Field
							case "cache":
								index = 79 + m.Slot
							case "update-player":
								index = 75 + m.Slot
							case "player-index":
								index, value = 77+m.Slot, value&255
							default:
								t.Fatal(m.Action)
							}
							state[index] = value
							a.setState(state)
						}
						emit(meteorCastEvent{Kind: kind, Args: args, Result: result, Before: before, After: a.hash()})
						a.check(t)
						return result
					}
					const ignored = uint32(0xaabbccdd)
					rt := server.ArachnaphobiaNativeRuntime52DC80{
						LoadCache: func() uint32 {
							value := a.cache[0]
							emit(meteorCastEvent{Kind: "cache-read", Args: []uint32{0}, Result: value})
							return value
						},
						StoreCache: func(value uint32) {
							a.cache[0] = value
							emit(meteorCastEvent{Kind: "cache-write", Args: []uint32{0, value}})
						},
						LookupType: func(name string) uint32 {
							if name != "ArachnaphobiaFocus" {
								t.Fatal(name)
							}
							return service("lookup", []uint32{0}, c.Input.Lookup)
						},
						TraceRay: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
							return service("trace", []uint32{math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y), uint32(flags)}, c.Input.Trace) != 0
						},
						NewObject: func(kind uint32) *server.Object {
							service("allocate", []uint32{kind}, uint32(c.Input.Allocate))
							return a.object(c.Input.Allocate)
						},
						CreateAt: func(o, owner *server.Object, pos types.Pointf) {
							service("create", []uint32{a.owner(o), a.owner(owner), math.Float32bits(pos.X), math.Float32bits(pos.Y), 0}, ignored)
						},
						Inform: func(index uint8, code byte, value int32) {
							service("inform", []uint32{uint32(index), uint32(code), uint32(value)}, ignored)
						},
					}
					arachnaphobiaCastCall52DC80 = func(_ spell.ID, _ *server.Object, owner, caster *server.Object, arg *server.SpellAcceptArg, _ int) int {
						return int(server.ArachnaphobiaNative52DC80(owner, caster, arg, rt))
					}
					got := arachnaphobiaActualCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), a.arguments[c.Input.Arg], c.Input.Level)
					if uint32(got) != c.Return || cursor != len(c.Events) || a.words() != c.After {
						t.Fatalf("return/events/final state got%d/%d want%d/%d", got, cursor, c.Return, len(c.Events))
					}
					a.check(t)
				})
			}
		}
	}
}

func TestArachnaphobiaNativePublicPointers(t *testing.T) {
	a := meteorCastArenaNew(t, 64)
	previous := arachnaphobiaCastCall52DC80
	t.Cleanup(func() { arachnaphobiaCastCall52DC80 = previous })
	for route := 0; route < 3; route++ {
		for _, value := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/id%d", route, value), func(t *testing.T) {
				calls := 0
				arachnaphobiaCastCall52DC80 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
					calls++
					wantLevel := int32(0) // The original C signature has no sixth argument.
					if route == 2 {
						wantLevel = value
					}
					if int32(id) != value || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.arguments[1] || int32(level) != wantLevel {
						t.Fatal("native C owners/argument/signed ID or existing Go level changed")
					}
					return -12345
				}
				before := bytes.Clone(a.raw)
				got := arachnaphobiaActualCall(route, value, a.objects[0], a.objects[1], a.objects[2], a.arguments[1], value)
				if got != -12345 || calls != 1 || !bytes.Equal(a.raw, before) {
					t.Fatalf("return/calls=%d/%d or foreign bytes changed", got, calls)
				}
			})
		}
	}
}

// Use the production default binding and real MapTraceRay/type allocation
// services. No return value, blocked trace or missing-type result is supplied.
func TestArachnaphobiaNativeDefaultBindings(t *testing.T) {
	cache := memmap.PtrUint32(0x5D4594, 2487812)
	oldCache, oldServer, oldCast := *cache, GetServer, arachnaphobiaCastCall52DC80
	t.Cleanup(func() { *cache, GetServer, arachnaphobiaCastCall52DC80 = oldCache, oldServer, oldCast })
	arachnaphobiaCastCall52DC80 = castArachnaphobiaNative52DC80
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := meteorCastArenaNew(t, placement)
			initialize := func() {
				for i := range a.raw {
					a.raw[i] = byte(i*31 + 19)
				}
				var state [117]uint32
				for i := 0; i < 6; i++ {
					for _, field := range []int{7, 8, 9, 10} {
						state[11*i+field] = math.MaxUint32
					}
				}
				state[66], state[69] = math.MaxUint32, math.MaxUint32
				state[75], state[76] = math.MaxUint32, math.MaxUint32
				a.patch(a.raw, state)
			}
			for _, kind := range []uint32{100, 0xfedc0064} {
				t.Run(fmt.Sprintf("route%d/place%d/blocked%x", route, placement, kind), func(t *testing.T) {
					initialize()
					*cache = kind
					GetServer = func() Server { return &arachnaphobiaCastLegacyServer52DC80{srv: new(server.Server)} }
					a.objects[2].PosVec = types.Ptf(-100, -200)
					a.arguments[0].Pos = types.Ptf(30, 40)
					before := bytes.Clone(a.raw)
					got := arachnaphobiaActualCall(route, math.MinInt32, a.objects[5], a.objects[0], a.objects[2], a.arguments[0], math.MaxInt32)
					if got != 0 || *cache != kind || !bytes.Equal(a.raw, before) {
						t.Fatalf("actual default blocked cast result=%d or state changed", got)
					}
				})
			}
			t.Run(fmt.Sprintf("route%d/place%d/open-missing-type", route, placement), func(t *testing.T) {
				initialize()
				srv, caster, update := monsterLookAtFixture5125A0(t)
				srv.Map.Init()
				t.Cleanup(srv.Map.Free)
				if srv.Walls.InitHigh() == 0 {
					t.Fatal("cannot initialize actual native wall map")
				}
				t.Cleanup(srv.Walls.Free)
				caster.PosVec = types.Ptf(300, 300)
				a.arguments[0].Pos = types.Ptf(330, 300)
				*cache = math.MaxUint32
				GetServer = func() Server { return &arachnaphobiaCastLegacyServer52DC80{srv: srv} }
				before := bytes.Clone(a.raw)
				beforeCaster, beforeUpdate := *caster, *update
				got := arachnaphobiaActualCall(route, math.MaxInt32, a.objects[5], a.objects[0], caster, a.arguments[0], math.MinInt32)
				if got != 1 || *cache != math.MaxUint32 || !bytes.Equal(a.raw, before) || *caster != beforeCaster || *update != beforeUpdate {
					t.Fatalf("actual open map/missing-type result=%d or state changed", got)
				}
			})
		}
	}
}
