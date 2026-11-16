package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/inversion_original_52beb0_v1.json.gz
var inversionNativeOriginal []byte

type inversionNativeMutation struct {
	Kind            string
	At, Slot, Field int
	Value           uint32
}
type inversionNativeInput struct {
	Name                  string
	ID                    int32
	Second, Owner, Caster int
	RangeBits             uint64 `json:"range_bits"`
	Objects               [6][2]uint32
	Missiles              []int
	EnumerationResult     uint32 `json:"enumeration_result"`
	ChangeResult          uint32 `json:"change_result"`
	AudioIDResult         uint32 `json:"audio_id_result"`
	AudioResult           uint32 `json:"audio_result"`
	Mutations             []inversionNativeMutation
}
type inversionNativeEvent struct {
	Kind          string
	At            int
	Args          json.RawMessage
	Result        uint64
	Before, After string
}
type inversionNativeCase struct {
	Input         inversionNativeInput
	Before, After [12]uint32
	Events        []inversionNativeEvent
	Spills        []uint32 `json:"binary32_spills"`
	Return        *uint32  `json:"return_u32"`
	Fault         map[string]any
}
type inversionNativeArena struct {
	raw, expected []byte
	objects       [6]*server.Object
	argument      *server.SpellAcceptArg
}

func inversionNativeArenaNew(t *testing.T, placement int) *inversionNativeArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign owner must exceed4GiB")
	}
	a := &inversionNativeArena{raw: raw}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
	}
	a.argument = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x8000+placement]))
	return a
}
func (a *inversionNativeArena) object(slot int) *server.Object {
	if slot == -1 {
		return nil
	}
	return a.objects[slot]
}
func (a *inversionNativeArena) role(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if o == p {
			return uint32(i)
		}
	}
	panic("unowned Inversion object")
}
func (a *inversionNativeArena) offset(p unsafe.Pointer) int {
	n := int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0])))
	if n < 0 || n >= len(a.raw) {
		panic("outside foreign owner")
	}
	return n
}
func (a *inversionNativeArena) put(b []byte, m inversionNativeMutation) {
	off := a.offset(unsafe.Pointer(a.objects[m.Slot])) + int(unsafe.Offsetof(server.Object{}.PosVec)) + 4*m.Field
	binary.LittleEndian.PutUint32(b[off:], m.Value)
}
func (a *inversionNativeArena) initialize(in inversionNativeInput) {
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	for i, row := range in.Objects {
		for j, v := range row {
			a.put(a.raw, inversionNativeMutation{Slot: i, Field: j, Value: v})
		}
	}
	a.expected = bytes.Clone(a.raw)
}
func (a *inversionNativeArena) words() (out [12]uint32) {
	for i, o := range a.objects {
		out[2*i], out[2*i+1] = math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)
	}
	return
}
func (a *inversionNativeArena) state() string {
	words := a.words()
	var b [48]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(b[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b[:]))
}
func (a *inversionNativeArena) check(t *testing.T) {
	t.Helper()
	if !bytes.Equal(a.raw, a.expected) {
		t.Fatal("unexpected write in entire64KiB owner")
	}
}
func (a *inversionNativeArena) mutate(in inversionNativeInput, kind string, at int) {
	for _, m := range in.Mutations {
		if m.Kind == kind && m.At == at {
			a.put(a.raw, m)
			a.put(a.expected, m)
		}
	}
}
func inversionNativeRows(t *testing.T) []inversionNativeCase {
	t.Helper()
	var rows []inversionNativeCase
	npcVoiceOriginalDecode(t, inversionNativeOriginal, "09ef7b359c4b988988f2a4ceb472375e432d0f1ab89f1f3f1fea2e50eac38ead", "9027a202f77f366bc2e61716708d5f2a509b131f5bd685d0741a75e77f912eb2", 79, &rows)
	return rows
}
func TestInversionNativeLayout(t *testing.T) {
	var arg server.SpellAcceptArg
	want := [8]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(server.Object{}), unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4, unsafe.Sizeof(arg), unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos), unsafe.Offsetof(arg.Pos) + 4}
	got := inversionNativeLayoutC()
	if got[1] < got[3]+4 || got[1] > 0x1000 || want[1] > 0x1000 {
		t.Fatal("native record extent", got, want)
	}
	for i := range want {
		if i != 1 && got[i] != want[i] {
			t.Fatal("C/Go native object/Go-only acceptance layout", got, want)
		}
	}
}
func TestInversionNativeOriginal(t *testing.T) {
	rows := inversionNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := inversionNativeArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					a.initialize(c.Input)
					if a.words() != c.Before {
						t.Fatal("canonical native input differs")
					}
					cursor := 0
					counts := make(map[string]int)
					observe := func(kind string, args []any, result uint64) {
						a.check(t)
						at := counts[kind]
						before := a.state()
						a.mutate(c.Input, kind, at)
						data, err := json.Marshal(args)
						if err != nil {
							t.Fatal(err)
						}
						e := inversionNativeEvent{Kind: kind, At: at, Args: data, Result: result, Before: before, After: a.state()}
						if cursor >= len(c.Events) {
							t.Fatalf("unexpected service%d:%+v", cursor, e)
						}
						if !reflect.DeepEqual(e, c.Events[cursor]) {
							t.Fatalf("service%d differs:%+v;want:%+v", cursor, e, c.Events[cursor])
						}
						cursor++
						counts[kind]++
					}
					originRole := func(pos types.Pointf) uint32 {
						found := uint32(math.MaxUint32)
						for i, o := range a.objects {
							if math.Float32bits(pos.X) == math.Float32bits(o.PosVec.X) && math.Float32bits(pos.Y) == math.Float32bits(o.PosVec.Y) {
								if found != math.MaxUint32 {
									t.Fatal("ambiguous live origin")
								}
								found = uint32(i)
							}
						}
						if found == math.MaxUint32 {
							t.Fatal("origin is not an owned live position")
						}
						return found
					}
					old := inversionCastCall52BEB0
					defer func() { inversionCastCall52BEB0 = old }()
					inversionCastCall52BEB0 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						if int32(id) != c.Input.ID || second != a.object(c.Input.Second) || owner != a.object(c.Input.Owner) || caster != a.object(c.Input.Caster) || arg != nil || level != 0 {
							t.Fatal("original C4/full Go selector identity differs")
						}
						return int(server.InversionCastNative52BEB0(int32(id), owner, caster, server.InversionNativeRuntime52BEB0{
							Balance: func(name string) float64 {
								observe("balance", []any{name}, c.Input.RangeBits)
								return math.Float64frombits(c.Input.RangeBits)
							},
							EachMissile: func(origin types.Pointf, radius float32, visit func(*server.Object) bool) {
								observe("enumerate", []any{originRole(origin), math.Float32bits(origin.X), math.Float32bits(origin.Y), math.Float32bits(radius), "change-owner-52be40", a.role(owner)}, uint64(c.Input.EnumerationResult))
								for _, slot := range c.Input.Missiles {
									if !visit(a.object(slot)) {
										t.Fatal("native callback stopped declared enumeration")
									}
								}
							},
							ChangeOwner: func(missile, owner *server.Object) {
								observe("change-owner", []any{a.role(missile), a.role(owner)}, uint64(c.Input.ChangeResult))
							},
							CastSound: func(id int32) sound.ID {
								observe("audio-id", []any{uint32(id), uint32(0)}, uint64(c.Input.AudioIDResult))
								return sound.ID(c.Input.AudioIDResult)
							},
							Audio: func(id sound.ID, unit *server.Object, kind int, code uint32) {
								observe("audio", []any{uint32(id), a.role(unit), uint32(kind), code}, uint64(c.Input.AudioResult))
							},
						}))
					}
					var got int32
					var caught any
					func() {
						defer func() { caught = recover() }()
						got = inversionNativeCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), nil, 0)
					}()
					if c.Fault != nil {
						err, ok := caught.(runtime.Error)
						if !ok || !strings.Contains(err.Error(), "invalid memory address") || c.Return != nil {
							t.Fatal("original nil-caster fault prefix differs", caught)
						}
					} else if caught != nil || c.Return == nil || uint32(got) != *c.Return {
						t.Fatal("original return/fault differs", got, caught)
					}
					if cursor != len(c.Events) || a.words() != c.After {
						t.Fatal("complete original service/state projection differs")
					}
					a.check(t)
				})
			}
		}
	}
}
func TestInversionNativePublicPointers(t *testing.T) {
	a := inversionNativeArenaNew(t, 64)
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	old := inversionCastCall52BEB0
	t.Cleanup(func() { inversionCastCall52BEB0 = old })
	for route := 0; route < 3; route++ {
		for _, v := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/id%d", route, v), func(t *testing.T) {
				calls := 0
				inversionCastCall52BEB0 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
					calls++
					wantLevel := 0
					var wantArg *server.SpellAcceptArg
					if route == 2 {
						wantLevel = int(v)
						wantArg = a.argument
					}
					if int32(id) != v || second != a.objects[1] || owner != a.objects[2] || caster != a.objects[0] || arg != wantArg || level != wantLevel {
						t.Fatal("C4/full Go6 pointer or signed scalar forwarding differs")
					}
					return -12345
				}
				before := bytes.Clone(a.raw)
				if got := inversionNativeCall(route, v, a.objects[1], a.objects[2], a.objects[0], a.argument, v); got != -12345 || calls != 1 || !bytes.Equal(before, a.raw) {
					t.Fatal("native return/calls/owner differs")
				}
			})
		}
	}
}
func TestInversionNativeDefaultBindings(t *testing.T) {
	oldServer, oldCast, oldOwner := GetServer, inversionCastCall52BEB0, Nox_xxx_changeOwner_52BE40
	t.Cleanup(func() { GetServer, inversionCastCall52BEB0, Nox_xxx_changeOwner_52BE40 = oldServer, oldCast, oldOwner })
	inversionCastCall52BEB0 = castInversionNative52BEB0
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			for _, kind := range []string{"owned-missile", "nil-owner"} {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, kind), func(t *testing.T) {
					a := inversionNativeArenaNew(t, placement)
					for i := range a.raw {
						a.raw[i] = byte(i*31 + 19)
					}
					for _, o := range a.objects {
						off := a.offset(unsafe.Pointer(o))
						clear(a.raw[off : off+int(unsafe.Sizeof(*o))])
					}
					caster, missile, owner := a.objects[0], a.objects[3], a.objects[2]
					if kind == "nil-owner" {
						owner = nil
					}
					caster.PosVec = types.Ptf(123.5, 456.25)
					*missile = server.Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, PosVec: caster.PosVec, NewPos: caster.PosVec}
					world := server.NewHigh(nil, nil, strman.New())
					t.Cleanup(world.Close)
					world.Map.Init()
					world.Map.AddObjectToIndex(missile)
					a.expected = bytes.Clone(a.raw)
					calls := 0
					GetServer = func() Server { return &monsterMainLegacyServer547210{srv: world} }
					Nox_xxx_changeOwner_52BE40 = func(gotMissile, gotOwner *server.Object) {
						calls++
						if gotMissile != missile || gotOwner != owner {
							t.Fatal("actual native map/owner binding differs")
						}
					}
					if got := inversionNativeCall(route, int32(spell.SPELL_INVERSION), a.objects[1], owner, caster, nil, -7); got != 1 || calls != 1 {
						t.Fatal("actual zero-radius map result differs", got, calls)
					}
					a.check(t)
				})
			}
		}
	}
}
