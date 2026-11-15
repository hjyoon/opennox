package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/pull_push_original_52bfa0_v1.json.gz
var pullPushNativeOriginal []byte

type pullPushNativeMutation struct {
	Kind, Action string
	Slot, Field  int
	Value        int64
}
type pullPushNativeInput struct {
	Name                  string
	Cast                  int
	ID                    int32
	Second, Owner, Caster int
	Arg                   bool
	Level                 int32
	Coefficient           uint64 `json:"coefficient_bits"`
	Objects               [4][2]uint32
	Argument              [3]int64
	RadialResult          uint32 `json:"radial_result"`
	AudioIDResult         uint32 `json:"audio_id_result"`
	AudioResult           uint32 `json:"audio_result"`
	Mutations             []pullPushNativeMutation
}
type pullPushNativeEvent struct {
	Kind          string
	At            int
	Args          json.RawMessage
	Result        uint64
	Before, After string
}
type pullPushNativeCase struct {
	Input         pullPushNativeInput
	Before, After [11]uint32
	Events        []pullPushNativeEvent
	Return        uint32 `json:"return_u32"`
}
type pullPushNativeArena struct {
	raw, expected []byte
	objects       [4]*server.Object
	argument      *server.SpellAcceptArg
}

var pullPushPositionOffset = unsafe.Offsetof(server.Object{}.PosVec)
var pullPushArgumentPositionOffset = unsafe.Offsetof(server.SpellAcceptArg{}.Pos)

func pullPushNativeArenaNew(t *testing.T, placement int) *pullPushNativeArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("owner must exceed4GiB")
	}
	a := &pullPushNativeArena{raw: raw}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x800+placement]))
	}
	a.argument = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x5000+placement]))
	return a
}
func (a *pullPushNativeArena) object(n int) *server.Object {
	if n == -1 {
		return nil
	}
	return a.objects[n]
}
func (a *pullPushNativeArena) role(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if p == o {
			return uint32(i)
		}
	}
	panic("unowned native object")
}
func (a *pullPushNativeArena) offset(p unsafe.Pointer) int {
	n := int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0])))
	if n < 0 || n >= len(a.raw) {
		panic("pointer outside foreign owner")
	}
	return n
}
func (a *pullPushNativeArena) patch(b []byte, m pullPushNativeMutation) {
	if m.Action == "object" {
		poisonCastNativePut(b, a.offset(unsafe.Pointer(a.objects[m.Slot]))+int(pullPushPositionOffset)+4*m.Field, 4, uint64(m.Value))
	} else if m.Field == 0 {
		poisonCastNativePut(b, a.offset(unsafe.Pointer(a.argument)), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(a.object(int(m.Value))))))
	} else {
		poisonCastNativePut(b, a.offset(unsafe.Pointer(a.argument))+int(pullPushArgumentPositionOffset)+4*(m.Field-1), 4, uint64(m.Value))
	}
}
func (a *pullPushNativeArena) initialize(in pullPushNativeInput) {
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	for i, row := range in.Objects {
		for j, v := range row {
			a.patch(a.raw, pullPushNativeMutation{Action: "object", Slot: i, Field: j, Value: int64(v)})
		}
	}
	for j, v := range in.Argument {
		a.patch(a.raw, pullPushNativeMutation{Action: "argument", Field: j, Value: v})
	}
	a.expected = bytes.Clone(a.raw)
}
func (a *pullPushNativeArena) words() (out [11]uint32) {
	for i, o := range a.objects {
		out[2*i], out[2*i+1] = math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)
	}
	off := a.offset(unsafe.Pointer(a.argument)) + int(pullPushArgumentPositionOffset)
	out[8], out[9], out[10] = a.role(a.argument.Obj), binary.LittleEndian.Uint32(a.raw[off:]), binary.LittleEndian.Uint32(a.raw[off+4:])
	return
}
func (a *pullPushNativeArena) state() string {
	v := a.words()
	var b [44]byte
	for i, x := range v {
		binary.LittleEndian.PutUint32(b[4*i:], x)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b[:]))
}
func (a *pullPushNativeArena) check(t *testing.T) {
	t.Helper()
	if !bytes.Equal(a.raw, a.expected) {
		t.Fatal("unexpected write anywhere in64KiB native owner")
	}
}
func (a *pullPushNativeArena) mutate(in pullPushNativeInput, kind string) {
	for _, m := range in.Mutations {
		if m.Kind == kind {
			a.patch(a.raw, m)
			a.patch(a.expected, m)
		}
	}
}
func pullPushNativeRows(t *testing.T) []pullPushNativeCase {
	t.Helper()
	var rows []pullPushNativeCase
	npcVoiceOriginalDecode(t, pullPushNativeOriginal, "d5304b3b03c61e9ec103ee4df9b034d04cee2265ceb042ab1053d33ca50272d7", "cc1f7ed74c0e41f1fc729f0555b642c86cabbfa16cd303283dece417ac670b89", 566, &rows)
	return rows
}
func TestPullPushNativeLayout(t *testing.T) {
	var arg server.SpellAcceptArg
	want := [7]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(arg), unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.X), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.Y), pullPushPositionOffset, pullPushPositionOffset + 4}
	if got := pullPushNativeLayoutC(); got != want {
		t.Fatal("independent native C layout", got, want)
	}
}
func TestPullPushNativeOriginal(t *testing.T) {
	rows := pullPushNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := pullPushNativeArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					a.initialize(c.Input)
					if a.words() != c.Before {
						t.Fatal("canonical input differs")
					}
					cursor := 0
					observe := func(kind string, args []any, result uint64) {
						a.check(t)
						before := a.state()
						a.mutate(c.Input, kind)
						data, err := json.Marshal(args)
						if err != nil {
							t.Fatal(err)
						}
						e := pullPushNativeEvent{Kind: kind, Args: data, Result: result, Before: before, After: a.state()}
						if cursor >= len(c.Events) || !reflect.DeepEqual(e, c.Events[cursor]) {
							t.Fatalf("original service%d %s differs: %+v; want %+v", cursor, kind, e, c.Events[cursor])
						}
						cursor++
					}
					oldPull, oldPush := pullCastCall52BFA0, pushCastCall52C000
					defer func() { pullCastCall52BFA0, pushCastCall52C000 = oldPull, oldPush }()
					dispatch := func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						wantArg := a.argument
						if !c.Input.Arg {
							wantArg = nil
						}
						if int32(id) != c.Input.ID || second != a.object(c.Input.Second) || owner != a.object(c.Input.Owner) || caster != a.object(c.Input.Caster) || arg != wantArg || int32(level) != c.Input.Level {
							t.Fatal("native public pointer/signed scalar identity differs")
						}
						rt := server.PullPushNativeRuntime52BFA0{
							Balance: func(key string) float64 {
								observe("balance", []any{key}, c.Input.Coefficient)
								return math.Float64frombits(c.Input.Coefficient)
							},
							PushUnits: func(origin types.Pointf, outer, inner, force float32, source *server.Object, callback, callbackArg int) {
								observe("radial", []any{a.role(caster), math.Float32bits(origin.X), math.Float32bits(origin.Y), math.Float32bits(outer), math.Float32bits(inner), math.Float32bits(force), uint32(0), uint32(callback), uint32(callbackArg)}, uint64(c.Input.RadialResult))
								if source != nil {
									t.Fatal("original radial source must be null")
								}
							},
							CastSound: func(id int32) sound.ID {
								observe("audio-id", []any{uint32(id), uint32(0)}, uint64(c.Input.AudioIDResult))
								return sound.ID(c.Input.AudioIDResult)
							},
							Audio: func(id sound.ID, owner *server.Object, kind int, code uint32) {
								observe("audio", []any{uint32(id), a.role(owner), uint32(kind), code}, uint64(c.Input.AudioResult))
							},
						}
						if c.Input.Cast == 0 {
							return int(server.PullCastNative52BFA0(int32(id), owner, caster, int32(level), rt))
						}
						return int(server.PushCastNative52C000(int32(id), owner, caster, int32(level), rt))
					}
					pullCastCall52BFA0, pushCastCall52C000 = dispatch, dispatch
					arg := a.argument
					if !c.Input.Arg {
						arg = nil
					}
					got := pullPushNativeCall(route, c.Input.Cast, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), arg, c.Input.Level)
					if uint32(got) != c.Return || cursor != len(c.Events) || a.words() != c.After {
						t.Fatal("complete original return/events/final state differs")
					}
					a.check(t)
				})
			}
		}
	}
}
func TestPullPushNativePublicPointers(t *testing.T) {
	a := pullPushNativeArenaNew(t, 64)
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	oldPull, oldPush := pullCastCall52BFA0, pushCastCall52C000
	t.Cleanup(func() { pullCastCall52BFA0, pushCastCall52C000 = oldPull, oldPush })
	for cast := 0; cast < 2; cast++ {
		for route := 0; route < 3; route++ {
			for _, v := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
				t.Run(fmt.Sprintf("cast%d/route%d/id%d", cast, route, v), func(t *testing.T) {
					calls := 0
					dispatch := func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						calls++
						if int32(id) != v || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.argument || level != int(v) {
							t.Fatal("complete native arguments changed")
						}
						return -12345
					}
					pullCastCall52BFA0, pushCastCall52C000 = dispatch, dispatch
					before := bytes.Clone(a.raw)
					if got := pullPushNativeCall(route, cast, v, a.objects[0], a.objects[1], a.objects[2], a.argument, v); got != -12345 || calls != 1 || !bytes.Equal(before, a.raw) {
						t.Fatal("public return/calls/foreign owner changed")
					}
				})
			}
		}
	}
}

type pullPushNativeTestServer struct {
	Server
	world *server.Server
}

func (s *pullPushNativeTestServer) S() *server.Server { return s.world }
func TestPullPushNativeDefaultBindings(t *testing.T) {
	oldServer, oldMap, oldPull, oldPush := GetServer, mapPushUnitsAroundCall52E040, pullCastCall52BFA0, pushCastCall52C000
	oldData := datapath.Data()
	t.Cleanup(func() {
		GetServer, mapPushUnitsAroundCall52E040, pullCastCall52BFA0, pushCastCall52C000 = oldServer, oldMap, oldPull, oldPush
		datapath.SetData(oldData)
	})
	private := t.TempDir()
	if err := os.WriteFile(filepath.Join(private, "gamedata.yml"), []byte("PullPowerCoeff: 2.5\nPushPowerCoeff: 2.5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	datapath.SetData(private)
	world := new(server.Server)
	if err := world.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	GetServer = func() Server { return &pullPushNativeTestServer{world: world} }
	pullCastCall52BFA0, pushCastCall52C000 = castPullNative52BFA0, castPushNative52C000
	rows := pullPushNativeRows(t)
	for cast := 0; cast < 2; cast++ {
		for route := 0; route < 3; route++ {
			for _, placement := range []int{0, 64, 512} {
				for _, ownerSlot := range []int{-1, 1} {
					t.Run(fmt.Sprintf("cast%d/route%d/place%d/owner%d", cast, route, placement, ownerSlot), func(t *testing.T) {
						a := pullPushNativeArenaNew(t, placement)
						a.initialize(rows[0].Input)
						for _, o := range a.objects {
							off := a.offset(unsafe.Pointer(o))
							clear(a.raw[off : off+int(unsafe.Sizeof(*o))])
						}
						a.objects[0].PosVec = types.Ptf(123.5, -456.25)
						a.expected = bytes.Clone(a.raw)
						calls := 0
						mapPushUnitsAroundCall52E040 = func(ref *types.Pointf, outer, inner, force float32, source *server.Object, callback, callbackArg int) {
							a.check(t)
							calls++
							want := uint32(0x40f00000)
							if cast == 0 {
								want |= 0x80000000
							}
							if ref != &a.objects[0].PosVec || *ref != types.Ptf(123.5, -456.25) || outer != 600 || inner != 10 || math.Float32bits(force) != want || source != nil || callback != 0 || callbackArg != 0 {
								t.Fatal("actual balance/native position/radial binding differs")
							}
							m := pullPushNativeMutation{Action: "object", Slot: 1, Field: 1, Value: int64(math.Float32bits(-999))}
							a.patch(a.raw, m)
							a.patch(a.expected, m)
						}
						id := int32(spell.SPELL_PULL)
						if cast == 1 {
							id = int32(spell.SPELL_PUSH)
						}
						if got := pullPushNativeCall(route, cast, id, a.objects[2], a.object(ownerSlot), a.objects[0], a.argument, 3); got != 1 || calls != 1 {
							t.Fatal("actual default composition return/calls", got, calls)
						}
						a.check(t)
					})
				}
			}
		}
	}
}
