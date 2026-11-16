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
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/fumble_cast_original_52c060_v1.json.gz
var fumbleCastNativeOriginal []byte

type fumbleCastNativeMutation struct {
	Kind, Action    string
	At, Slot, Field int
	Value           int64
}
type fumbleCastNativeInput struct {
	Name                       string
	ID                         int32
	Second, Owner, Caster, Arg int
	Cache                      uint32
	Objects                    [8][10]int64
	Arguments                  [2][3]int64
	LookupResult               uint32   `json:"lookup_result"`
	ForceDropResults           []uint32 `json:"force_drop_results"`
	DropAllResults             []uint32 `json:"drop_all_results"`
	ForceResults               []uint32 `json:"force_results"`
	ClearResults               []uint32 `json:"clear_results"`
	AudioIDResult              uint32   `json:"audio_id_result"`
	AudioResults               []uint32 `json:"audio_results"`
	Mutations                  []fumbleCastNativeMutation
}
type fumbleCastNativeEvent struct {
	Kind          string
	At            int
	Args          json.RawMessage
	Result        uint32
	Before, After string
}
type fumbleCastNativeCase struct {
	Input         fumbleCastNativeInput
	Before, After [87]uint32
	Events        []fumbleCastNativeEvent
	CacheReads    []uint32 `json:"cache_reads"`
	CacheWrites   []uint32 `json:"cache_writes"`
	Return        *uint32  `json:"return_u32"`
	Fault         map[string]any
}
type fumbleCastNativeArena struct {
	raw, expected []byte
	objects       [8]*server.Object
	arguments     [2]*server.SpellAcceptArg
	cache         uint32
}

var fumbleCastNativeOffsets = [10]uintptr{
	unsafe.Offsetof(server.Object{}.TypeInd), unsafe.Offsetof(server.Object{}.ObjClass),
	unsafe.Offsetof(server.Object{}.ObjSubClass), unsafe.Offsetof(server.Object{}.ObjFlags),
	unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4,
	unsafe.Offsetof(server.Object{}.InvFirstItem), unsafe.Offsetof(server.Object{}.InvNextItem),
	unsafe.Offsetof(server.Object{}.Field129), unsafe.Offsetof(server.Object{}.Field128),
}

func fumbleCastNativeArenaNew(t *testing.T, placement int) *fumbleCastNativeArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign owner must exceed4GiB")
	}
	a := &fumbleCastNativeArena{raw: raw}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
	}
	for i := range a.arguments {
		a.arguments[i] = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0xa000+i*0x100+placement]))
	}
	return a
}
func (a *fumbleCastNativeArena) object(slot int) *server.Object {
	if slot == -1 {
		return nil
	}
	return a.objects[slot]
}
func (a *fumbleCastNativeArena) role(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if p == o {
			return uint32(i)
		}
	}
	panic("unowned Fumble object")
}
func (a *fumbleCastNativeArena) offset(p unsafe.Pointer) int {
	n := int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0])))
	if n < 0 || n >= len(a.raw) {
		panic("pointer outside foreign owner")
	}
	return n
}
func (a *fumbleCastNativeArena) patch(b []byte, m fumbleCastNativeMutation) {
	if m.Action == "object" {
		size := 4
		value := uint64(m.Value)
		if m.Field == 0 {
			size = 2
		} else if m.Field >= 6 {
			size = int(unsafe.Sizeof(uintptr(0)))
			value = uint64(uintptr(unsafe.Pointer(a.object(int(m.Value)))))
		}
		poisonCastNativePut(b, a.offset(unsafe.Pointer(a.objects[m.Slot]))+int(fumbleCastNativeOffsets[m.Field]), size, value)
	} else {
		off := a.offset(unsafe.Pointer(a.arguments[m.Slot]))
		if m.Field == 0 {
			poisonCastNativePut(b, off, int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(a.object(int(m.Value))))))
		} else {
			poisonCastNativePut(b, off+int(unsafe.Offsetof(server.SpellAcceptArg{}.Pos))+4*(m.Field-1), 4, uint64(m.Value))
		}
	}
}
func (a *fumbleCastNativeArena) initialize(in fumbleCastNativeInput) {
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	a.cache = in.Cache
	for i, row := range in.Objects {
		for j, v := range row {
			a.patch(a.raw, fumbleCastNativeMutation{Action: "object", Slot: i, Field: j, Value: v})
		}
	}
	for i, row := range in.Arguments {
		for j, v := range row {
			a.patch(a.raw, fumbleCastNativeMutation{Action: "argument", Slot: i, Field: j, Value: v})
		}
	}
	a.expected = bytes.Clone(a.raw)
}
func (a *fumbleCastNativeArena) words() (out [87]uint32) {
	out[0] = a.cache
	at := 1
	for _, o := range a.objects {
		values := [10]uint32{uint32(o.TypeInd), uint32(o.ObjClass), uint32(o.ObjSubClass), uint32(o.ObjFlags),
			math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), a.role(o.InvFirstItem), a.role(o.InvNextItem), a.role(o.Field129), a.role(o.Field128)}
		copy(out[at:], values[:])
		at += 10
	}
	for _, arg := range a.arguments {
		off := a.offset(unsafe.Pointer(arg)) + int(unsafe.Offsetof(arg.Pos))
		out[at], out[at+1], out[at+2] = a.role(arg.Obj), binary.LittleEndian.Uint32(a.raw[off:]), binary.LittleEndian.Uint32(a.raw[off+4:])
		at += 3
	}
	return
}
func (a *fumbleCastNativeArena) state() string {
	words := a.words()
	var b [348]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(b[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b[:]))
}
func (a *fumbleCastNativeArena) check(t *testing.T) {
	t.Helper()
	if !bytes.Equal(a.raw, a.expected) {
		t.Fatal("unexpected write in entire64KiB foreign owner")
	}
}
func (a *fumbleCastNativeArena) mutate(in fumbleCastNativeInput, kind string, at int) {
	for _, m := range in.Mutations {
		if m.Kind == kind && m.At == at {
			if m.Action == "cache" {
				a.cache = uint32(m.Value)
			} else {
				a.patch(a.raw, m)
				a.patch(a.expected, m)
			}
		}
	}
}
func fumbleCastNativeRows(t *testing.T) []fumbleCastNativeCase {
	t.Helper()
	var rows []fumbleCastNativeCase
	npcVoiceOriginalDecode(t, fumbleCastNativeOriginal, "941806e05d365e79b1b4868eab084b5a261787f0053e32cd505f68d016ba4648", "bbd8a1ef0f312cdbc6d2f1208fe006eb4094ad0cb542155b7f04b2d2f23cafe9", 146, &rows)
	return rows
}
func TestFumbleNativeLayout(t *testing.T) {
	var arg server.SpellAcceptArg
	want := [16]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(server.Object{}), unsafe.Sizeof(arg),
		unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos), unsafe.Offsetof(arg.Pos) + 4}
	copy(want[6:], fumbleCastNativeOffsets[:])
	got := fumbleCastNativeLayoutC()
	// C's object extent and Go's runtime extent both fit the owned record;
	// all fields read by the original must have identical independent offsets.
	if got[1] < got[14]+unsafe.Sizeof(uintptr(0)) || got[1] > 0x1000 || want[1] > 0x1000 {
		t.Fatal("native record extent", got[1], want[1])
	}
	for i := range want {
		if i != 1 && got[i] != want[i] {
			t.Fatal("native C/Go object/acceptance layout", got, want)
		}
	}
}

func TestFumbleNativeOriginal(t *testing.T) {
	rows := fumbleCastNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := fumbleCastNativeArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					a.initialize(c.Input)
					if a.words() != c.Before {
						t.Fatal("canonical native input differs")
					}
					cursor := 0
					counts := make(map[string]int)
					var cacheReads, cacheWrites []uint32
					observe := func(kind string, args []any, result uint32) {
						a.check(t)
						at := counts[kind]
						before := a.state()
						a.mutate(c.Input, kind, at)
						data, err := json.Marshal(args)
						if err != nil {
							t.Fatal(err)
						}
						e := fumbleCastNativeEvent{Kind: kind, At: at, Args: data, Result: result, Before: before, After: a.state()}
						if cursor >= len(c.Events) {
							t.Fatalf("unexpected service%d:%+v", cursor, e)
						}
						if !reflect.DeepEqual(e, c.Events[cursor]) {
							t.Fatalf("service%d differs:%+v; want:%+v", cursor, e, c.Events[cursor])
						}
						cursor++
						counts[kind]++
					}
					result := func(values []uint32, kind string) uint32 {
						at := counts[kind]
						if at >= len(values) {
							at = len(values) - 1
						}
						return values[at]
					}
					originRole := func(pos types.Pointf) uint32 {
						role := uint32(math.MaxUint32)
						for i, o := range a.objects {
							if math.Float32bits(pos.X) == math.Float32bits(o.PosVec.X) && math.Float32bits(pos.Y) == math.Float32bits(o.PosVec.Y) {
								if role != math.MaxUint32 {
									t.Fatal("ambiguous owned origin value")
								}
								role = uint32(i)
							}
						}
						if role == math.MaxUint32 {
							t.Fatal("origin value is not any live owned position")
						}
						return role
					}
					previous := fumbleCastCall52C060
					defer func() { fumbleCastCall52C060 = previous }()
					fumbleCastCall52C060 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						var wantedArg *server.SpellAcceptArg
						if c.Input.Arg != -1 {
							wantedArg = a.arguments[c.Input.Arg]
						}
						if int32(id) != c.Input.ID || second != a.object(c.Input.Second) || owner != a.object(c.Input.Owner) || caster != a.object(c.Input.Caster) || arg != wantedArg || level != 0 {
							t.Fatal("native five-argument C/Go selector identity differs")
						}
						return int(server.FumbleCastNative52C060(int32(id), caster, arg, server.FumbleNativeRuntime52C060{
							ForceDrop: func(owner, item *server.Object) int32 {
								v := result(c.Input.ForceDropResults, "force-drop")
								observe("force-drop", []any{a.role(owner), a.role(item)}, v)
								return int32(v)
							},
							DropAll: func(owner *server.Object) {
								observe("drop-all", []any{a.role(owner)}, result(c.Input.DropAllResults, "drop-all"))
							},
							LoadTypeCache: func() uint32 { a.check(t); cacheReads = append(cacheReads, a.cache); return a.cache },
							LookupType: func(name string) uint32 {
								observe("lookup", []any{name}, c.Input.LookupResult)
								return c.Input.LookupResult
							},
							StoreTypeCache: func(value uint32) { cacheWrites = append(cacheWrites, value); a.cache = value },
							ApplyForce: func(target *server.Object, pos types.Pointf, force float64) {
								observe("force", []any{originRole(pos), math.Float32bits(pos.X), math.Float32bits(pos.Y), a.role(target), math.Float32bits(float32(force))}, result(c.Input.ForceResults, "force"))
							},
							ClearOwner: func(ball *server.Object) {
								observe("clear-owner", []any{a.role(ball)}, result(c.Input.ClearResults, "clear-owner"))
							},
							EffectSound: func(id int32) sound.ID {
								observe("audio-id", []any{uint32(id), uint32(1)}, c.Input.AudioIDResult)
								return sound.ID(c.Input.AudioIDResult)
							},
							Audio: func(id sound.ID, target *server.Object, kind int, code uint32) {
								observe("audio", []any{uint32(id), a.role(target), uint32(kind), code}, result(c.Input.AudioResults, "audio"))
							},
						}))
					}
					var arg *server.SpellAcceptArg
					if c.Input.Arg != -1 {
						arg = a.arguments[c.Input.Arg]
					}
					var got int32
					var caught any
					func() {
						defer func() { caught = recover() }()
						got = fumbleCastNativeCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), arg, 0)
					}()
					if c.Fault != nil {
						err, ok := caught.(runtime.Error)
						if !ok || !strings.Contains(err.Error(), "invalid memory address") || c.Return != nil {
							t.Fatalf("original nil-read fault changed:%v", caught)
						}
					} else if caught != nil || c.Return == nil || uint32(got) != *c.Return {
						t.Fatalf("original return/fault changed:%d/%v", got, caught)
					}
					if cursor != len(c.Events) || a.words() != c.After || !slices.Equal(cacheReads, c.CacheReads) || !slices.Equal(cacheWrites, c.CacheWrites) {
						t.Fatal("complete original services/state/cache observations differ")
					}
					a.check(t)
				})
			}
		}
	}
}

func TestFumbleNativePublicPointers(t *testing.T) {
	a := fumbleCastNativeArenaNew(t, 64)
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	old := fumbleCastCall52C060
	t.Cleanup(func() { fumbleCastCall52C060 = old })
	for route := 0; route < 3; route++ {
		for _, v := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/id%d", route, v), func(t *testing.T) {
				calls := 0
				fumbleCastCall52C060 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
					calls++
					wantLevel := 0
					if route == 2 {
						wantLevel = int(v)
					}
					if int32(id) != v || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.arguments[1] || level != wantLevel {
						t.Fatal("native five-argument C/full Go selector forwarding differs")
					}
					return -12345
				}
				before := bytes.Clone(a.raw)
				if got := fumbleCastNativeCall(route, v, a.objects[0], a.objects[1], a.objects[2], a.arguments[1], v); got != -12345 || calls != 1 || !bytes.Equal(before, a.raw) {
					t.Fatal("native public return/calls/owner changed")
				}
			})
		}
	}
}

type fumbleCastNativeTestServer struct {
	Server
	world *server.Server
	t     *testing.T
}

func (s *fumbleCastNativeTestServer) S() *server.Server { return s.world }
func (s *fumbleCastNativeTestServer) ApplyForce(*server.Object, types.Pointf, float64) {
	s.t.Fatal("nil/protected/empty-inventory default fixture reached force")
}
func TestFumbleNativeDefaultBindings(t *testing.T) {
	oldServer, oldCast := GetServer, fumbleCastCall52C060
	cell := memmap.PtrUint32(0x5d4594, 2487728)
	oldCache := *cell
	t.Cleanup(func() { GetServer, fumbleCastCall52C060 = oldServer, oldCast; *cell = oldCache })
	fumbleCastCall52C060 = castFumbleNative52C060
	rows := fumbleCastNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			for _, kind := range []string{"nil-target", "protected-monster", "empty-player-cache42", "empty-player-wide-cache"} {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, kind), func(t *testing.T) {
					a := fumbleCastNativeArenaNew(t, placement)
					a.initialize(rows[0].Input)
					for _, o := range a.objects {
						off := a.offset(unsafe.Pointer(o))
						clear(a.raw[off : off+int(unsafe.Sizeof(*o))])
					}
					a.objects[0].ObjClass = 4
					*cell = 42
					want := int32(1)
					if kind == "nil-target" {
						a.arguments[0].Obj = nil
						want = 0
					}
					if kind == "protected-monster" {
						a.objects[0].ObjClass = 2
						a.objects[0].ObjSubClass = 0x2000
					}
					if kind == "empty-player-wide-cache" {
						*cell = 0x1002a
					}
					cacheBefore := *cell
					a.expected = bytes.Clone(a.raw)
					world := new(server.Server)
					GetServer = func() Server { return &fumbleCastNativeTestServer{world: world, t: t} }
					got := fumbleCastNativeCall(route, int32(spell.SPELL_FUMBLE), a.objects[2], a.objects[3], a.objects[1], a.arguments[0], -7)
					if got != want || *cell != cacheBefore {
						t.Fatal("actual nil/protected/empty inventory/cache binding differs", got, *cell)
					}
					a.check(t)
				})
			}
		}
	}
}
