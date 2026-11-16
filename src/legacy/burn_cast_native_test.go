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

//go:embed testdata/burn_cast_original_52c3e0_v1.json.gz
var burnCastNativeOriginal []byte

type burnCastNativeMutation struct {
	Kind, Action    string
	At, Slot, Field int
	Value           int64
}
type burnCastNativeInput struct {
	Name                       string
	ID                         int32
	Second, Owner, Caster, Arg int
	Caches                     [2]uint32
	Objects                    [5][5]int64
	Arguments                  [2][3]int64
	Updates                    [2]int64
	Players                    [2]uint32
	LookupGlyph                uint32 `json:"lookup_glyph"`
	LookupFlame                uint32 `json:"lookup_flame"`
	TraceResult                uint32 `json:"trace_result"`
	NewSlot                    int    `json:"new_slot"`
	DurationBits               uint64 `json:"duration_bits"`
	AudioIDResult              uint32 `json:"audio_id_result"`
	ServiceResult              uint32 `json:"service_result"`
	Mutations                  []burnCastNativeMutation
}
type burnCastNativeEvent struct {
	Kind          string
	At            int
	Args          json.RawMessage
	Result        uint64
	Before, After string
}
type burnCastNativeCase struct {
	Input         burnCastNativeInput
	Before, After [37]uint32
	Events        []burnCastNativeEvent
	CacheReads    [][2]uint32 `json:"cache_reads"`
	CacheWrites   [][2]uint32 `json:"cache_writes"`
	Return        *uint32     `json:"return_u32"`
	Fault         map[string]any
}
type burnCastNativeArena struct {
	raw, expected []byte
	objects       [5]*server.Object
	arguments     [2]*server.SpellAcceptArg
	updates       [2]*server.PlayerUpdateData
	players       [2]*server.Player
	caches        [2]uint32
}

var burnCastNativeOffsets = [5]uintptr{
	unsafe.Offsetof(server.Object{}.TypeInd), unsafe.Offsetof(server.Object{}.ObjClass),
	unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4,
	unsafe.Offsetof(server.Object{}.UpdateData),
}

func burnCastNativeArenaNew(t *testing.T, placement int) *burnCastNativeArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign owner must exceed4GiB")
	}
	a := &burnCastNativeArena{raw: raw}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
	}
	for i := range a.arguments {
		a.arguments[i] = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x8000+i*0x100+placement]))
		a.updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x9000+i*0x1000+placement]))
		a.players[i] = (*server.Player)(unsafe.Pointer(&raw[0xb000+i*0x1000+placement]))
	}
	return a
}
func (a *burnCastNativeArena) object(slot int) *server.Object {
	if slot == -1 {
		return nil
	}
	return a.objects[slot]
}
func (a *burnCastNativeArena) update(slot int) *server.PlayerUpdateData {
	if slot == -1 {
		return nil
	}
	return a.updates[slot]
}
func (a *burnCastNativeArena) player(slot int) *server.Player {
	if slot == -1 {
		return nil
	}
	return a.players[slot]
}
func (a *burnCastNativeArena) role(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if o == p {
			return uint32(i)
		}
	}
	panic("unowned Burn object")
}
func (a *burnCastNativeArena) updateRole(p unsafe.Pointer) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.updates {
		if unsafe.Pointer(o) == p {
			return uint32(i)
		}
	}
	panic("unowned Burn update")
}
func (a *burnCastNativeArena) playerRole(p *server.Player) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.players {
		if o == p {
			return uint32(i)
		}
	}
	panic("unowned Burn player")
}
func (a *burnCastNativeArena) offset(p unsafe.Pointer) int {
	n := int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0])))
	if n < 0 || n >= len(a.raw) {
		panic("pointer outside foreign owner")
	}
	return n
}
func (a *burnCastNativeArena) patch(b []byte, m burnCastNativeMutation) {
	size := 4
	value := uint64(m.Value)
	var offset int
	switch m.Action {
	case "object":
		offset = a.offset(unsafe.Pointer(a.objects[m.Slot])) + int(burnCastNativeOffsets[m.Field])
		if m.Field == 0 {
			size = 2
		} else if m.Field == 4 {
			size = int(unsafe.Sizeof(uintptr(0)))
			value = uint64(uintptr(unsafe.Pointer(a.update(int(m.Value)))))
		}
	case "argument":
		offset = a.offset(unsafe.Pointer(a.arguments[m.Slot]))
		if m.Field == 0 {
			size = int(unsafe.Sizeof(uintptr(0)))
			value = uint64(uintptr(unsafe.Pointer(a.object(int(m.Value)))))
		} else {
			offset += int(unsafe.Offsetof(server.SpellAcceptArg{}.Pos)) + 4*(m.Field-1)
		}
	case "update":
		offset = a.offset(unsafe.Pointer(a.updates[m.Slot])) + int(unsafe.Offsetof(server.PlayerUpdateData{}.Player))
		size = int(unsafe.Sizeof(uintptr(0)))
		value = uint64(uintptr(unsafe.Pointer(a.player(int(m.Value)))))
	case "player":
		offset = a.offset(unsafe.Pointer(a.players[m.Slot])) + int(unsafe.Offsetof(server.Player{}.PlayerInd))
		size = 1
	default:
		panic("unknown Burn field patch")
	}
	poisonCastNativePut(b, offset, size, value)
}
func (a *burnCastNativeArena) initialize(in burnCastNativeInput) {
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	a.caches = in.Caches
	for i, row := range in.Objects {
		for j, v := range row {
			a.patch(a.raw, burnCastNativeMutation{Action: "object", Slot: i, Field: j, Value: v})
		}
	}
	for i, row := range in.Arguments {
		for j, v := range row {
			a.patch(a.raw, burnCastNativeMutation{Action: "argument", Slot: i, Field: j, Value: v})
		}
	}
	for i, v := range in.Updates {
		a.patch(a.raw, burnCastNativeMutation{Action: "update", Slot: i, Value: v})
	}
	for i, v := range in.Players {
		a.patch(a.raw, burnCastNativeMutation{Action: "player", Slot: i, Value: int64(v)})
	}
	a.expected = bytes.Clone(a.raw)
}
func (a *burnCastNativeArena) words() (out [37]uint32) {
	copy(out[:2], a.caches[:])
	at := 2
	for _, o := range a.objects {
		values := [5]uint32{uint32(o.TypeInd), uint32(o.ObjClass), math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), a.updateRole(o.UpdateData)}
		copy(out[at:], values[:])
		at += 5
	}
	for _, arg := range a.arguments {
		out[at], out[at+1], out[at+2] = a.role(arg.Obj), math.Float32bits(arg.Pos.X), math.Float32bits(arg.Pos.Y)
		at += 3
	}
	for _, ud := range a.updates {
		out[at] = a.playerRole(ud.Player)
		at++
	}
	for _, pl := range a.players {
		out[at] = uint32(pl.PlayerInd)
		at++
	}
	return
}
func (a *burnCastNativeArena) state() string {
	words := a.words()
	var b [148]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(b[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b[:]))
}
func (a *burnCastNativeArena) check(t *testing.T) {
	t.Helper()
	if !bytes.Equal(a.raw, a.expected) {
		t.Fatal("unexpected write in whole64KiB foreign owner")
	}
}
func (a *burnCastNativeArena) mutate(in burnCastNativeInput, kind string, at int) {
	for _, m := range in.Mutations {
		if m.Kind == kind && m.At == at {
			if m.Action == "cache" {
				a.caches[m.Slot] = uint32(m.Value)
			} else {
				a.patch(a.raw, m)
				a.patch(a.expected, m)
			}
		}
	}
}
func burnCastNativeRows(t *testing.T) []burnCastNativeCase {
	t.Helper()
	var rows []burnCastNativeCase
	npcVoiceOriginalDecode(t, burnCastNativeOriginal, "6a929ab3dfc0fef0168b0ec4a34edce3daa695074d0f7b22c045817cca36c4c7", "2719a30f035d50d2c9e0ddb5f2cdb569b86289502b31900be151e42dc94c1b1e", 145, &rows)
	// JSON whitespace is immaterial; retain every original argument value.
	for i := range rows {
		for j := range rows[i].Events {
			var b bytes.Buffer
			if err := json.Compact(&b, rows[i].Events[j].Args); err != nil {
				t.Fatal(err)
			}
			rows[i].Events[j].Args = bytes.Clone(b.Bytes())
		}
	}
	return rows
}
func TestBurnNativeLayout(t *testing.T) {
	var arg server.SpellAcceptArg
	want := [19]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(server.Object{}), unsafe.Sizeof(arg),
		unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos), unsafe.Offsetof(arg.Pos) + 4,
		unsafe.Offsetof(server.Object{}.TypeInd), unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4,
		unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Sizeof(server.Object{}.UpdateData),
		unsafe.Sizeof(server.PlayerUpdateData{}), unsafe.Offsetof(server.PlayerUpdateData{}.Player), unsafe.Sizeof(server.PlayerUpdateData{}.Player),
		unsafe.Sizeof(server.Player{}), unsafe.Offsetof(server.Player{}.PlayerInd), unsafe.Sizeof(server.Player{}.PlayerInd), unsafe.Sizeof(server.Object{}.TypeInd)}
	got := burnCastNativeLayoutC()
	for i := range want {
		if i != 1 && i != 12 && i != 15 && want[i] != got[i] {
			t.Fatal("independent C/Go field layout", got, want)
		}
	}
	if got[1] < got[10]+got[11] || got[12] < got[13]+got[14] || got[15] < got[16]+got[17] {
		t.Fatal("C record extent", got)
	}
	if want[1] > 0x1000 || want[13]+want[14] > 0x1000 || want[16]+want[17] > 0x1000 {
		t.Fatal("read fields exceed owned placement", want)
	}
}
func TestBurnNativeOriginal(t *testing.T) {
	rows := burnCastNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := burnCastNativeArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					a.initialize(c.Input)
					if a.words() != c.Before {
						t.Fatal("canonical native input differs")
					}
					cursor := 0
					counts := make(map[string]int)
					var cacheReads, cacheWrites [][2]uint32
					observe := func(kind string, args []any, result uint64) {
						a.check(t)
						at := counts[kind]
						before := a.state()
						a.mutate(c.Input, kind, at)
						data, err := json.Marshal(args)
						if err != nil {
							t.Fatal(err)
						}
						e := burnCastNativeEvent{Kind: kind, At: at, Args: data, Result: result, Before: before, After: a.state()}
						if cursor >= len(c.Events) {
							t.Fatalf("unexpected service%d:%+v", cursor, e)
						}
						if !reflect.DeepEqual(e, c.Events[cursor]) {
							t.Fatalf("service%d differs:%+v;want:%+v", cursor, e, c.Events[cursor])
						}
						cursor++
						counts[kind]++
					}
					load := func(slot int) uint32 {
						a.check(t)
						cacheReads = append(cacheReads, [2]uint32{uint32(slot), a.caches[slot]})
						return a.caches[slot]
					}
					store := func(slot int, value uint32) {
						cacheWrites = append(cacheWrites, [2]uint32{uint32(slot), value})
						a.caches[slot] = value
					}
					old := burnCastCall52C3E0
					defer func() { burnCastCall52C3E0 = old }()
					burnCastCall52C3E0 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						var wantedArg *server.SpellAcceptArg
						if c.Input.Arg != -1 {
							wantedArg = a.arguments[c.Input.Arg]
						}
						if int32(id) != c.Input.ID || second != a.object(c.Input.Second) || owner != a.object(c.Input.Owner) || caster != a.object(c.Input.Caster) || arg != wantedArg || level != 0 {
							t.Fatal("native five-argument C/full Go selector identity differs")
						}
						return int(server.BurnCastNative52C3E0(int32(id), caster, arg, server.BurnNativeRuntime52C3E0{
							LoadGlyph: func() uint32 { return load(0) }, StoreGlyph: func(v uint32) { store(0, v) },
							LoadFlame: func() uint32 { return load(1) }, StoreFlame: func(v uint32) { store(1, v) },
							LookupType: func(name string) uint32 {
								kind := "lookup-glyph"
								value := c.Input.LookupGlyph
								if name == "MediumFlame" {
									kind = "lookup-flame"
									value = c.Input.LookupFlame
								} else if name != "Glyph" {
									t.Fatal(name)
								}
								observe(kind, []any{name}, uint64(value))
								return value
							},
							TraceRay: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
								observe("trace", []any{math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y), uint32(flags)}, uint64(c.Input.TraceResult))
								return c.Input.TraceResult&255 != 0
							},
							NewObject: func(kind uint32) *server.Object {
								value := a.object(c.Input.NewSlot)
								observe("new", []any{kind}, uint64(a.role(value)))
								return value
							},
							CreateAt: func(obj, owner *server.Object, pos types.Pointf) {
								observe("create", []any{a.role(obj), a.role(owner), math.Float32bits(pos.X), math.Float32bits(pos.Y), uint32(0)}, uint64(c.Input.ServiceResult))
							},
							Balance: func(name string) float64 {
								observe("balance", []any{name}, c.Input.DurationBits)
								return math.Float64frombits(c.Input.DurationBits)
							},
							SetDecay: func(obj *server.Object, delay uint32) {
								observe("decay", []any{a.role(obj), delay}, uint64(c.Input.ServiceResult))
							},
							Spark: func(pos types.Pointf, count byte) {
								flame := a.object(c.Input.NewSlot)
								if flame == nil || math.Float32bits(pos.X) != math.Float32bits(flame.PosVec.X) || math.Float32bits(pos.Y) != math.Float32bits(flame.PosVec.Y) {
									t.Fatal("spark is not the live allocated flame position")
								}
								observe("spark", []any{a.role(flame), math.Float32bits(pos.X), math.Float32bits(pos.Y), uint32(count)}, uint64(c.Input.ServiceResult))
							},
							CastSound: func(id int32) sound.ID {
								observe("audio-id", []any{uint32(id), uint32(0)}, uint64(c.Input.AudioIDResult))
								return sound.ID(c.Input.AudioIDResult)
							},
							Audio: func(id sound.ID, pos types.Pointf, kind int, code uint32) {
								observe("audio", []any{uint32(id), uint32(c.Input.Arg), math.Float32bits(pos.X), math.Float32bits(pos.Y), uint32(kind), code}, uint64(c.Input.ServiceResult))
							},
							Inform: func(index uint8, code byte, value int32) {
								observe("inform", []any{uint32(index), uint32(code), uint32(value)}, uint64(c.Input.ServiceResult))
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
						got = burnCastNativeCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), arg, 0)
					}()
					if c.Fault != nil {
						err, ok := caught.(runtime.Error)
						if !ok || !strings.Contains(err.Error(), "invalid memory address") || c.Return != nil {
							t.Fatalf("original nil-player-chain prefix differs:%v", caught)
						}
					} else if caught != nil || c.Return == nil || uint32(got) != *c.Return {
						t.Fatalf("original return/fault changed:%d/%v", got, caught)
					}
					if cursor != len(c.Events) || a.words() != c.After || !slices.Equal(cacheReads, c.CacheReads) || !slices.Equal(cacheWrites, c.CacheWrites) {
						t.Fatal("complete original service/state/cache observations differ")
					}
					a.check(t)
				})
			}
		}
	}
}
func TestBurnNativePublicPointers(t *testing.T) {
	a := burnCastNativeArenaNew(t, 64)
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	old := burnCastCall52C3E0
	t.Cleanup(func() { burnCastCall52C3E0 = old })
	for route := 0; route < 3; route++ {
		for _, v := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/id%d", route, v), func(t *testing.T) {
				calls := 0
				burnCastCall52C3E0 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
					calls++
					wantLevel := 0
					if route == 2 {
						wantLevel = int(v)
					}
					if int32(id) != v || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.arguments[1] || level != wantLevel {
						t.Fatal("native five-argument C/full signed Go selector forwarding differs")
					}
					return -12345
				}
				before := bytes.Clone(a.raw)
				if got := burnCastNativeCall(route, v, a.objects[0], a.objects[1], a.objects[2], a.arguments[1], v); got != -12345 || calls != 1 || !bytes.Equal(before, a.raw) {
					t.Fatal("native public return/calls/owner changed")
				}
			})
		}
	}
}
func TestBurnNativeDefaultBindings(t *testing.T) {
	oldServer, oldCast := GetServer, burnCastCall52C3E0
	glyph, flame := Get_dword_5d4594_2487712_ptr(), memmap.PtrUint32(0x5d4594, 2487732)
	oldGlyph, oldFlame := *glyph, *flame
	t.Cleanup(func() { GetServer, burnCastCall52C3E0 = oldServer, oldCast; *glyph, *flame = oldGlyph, oldFlame })
	burnCastCall52C3E0 = castBurnNative52C3E0
	rows := burnCastNativeRows(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			for _, kind := range []string{"nil-argument", "nil-caster", "both-nil"} {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, kind), func(t *testing.T) {
					a := burnCastNativeArenaNew(t, placement)
					a.initialize(rows[0].Input)
					caster, arg := a.objects[0], a.arguments[0]
					if kind != "nil-caster" {
						arg = nil
					}
					if kind != "nil-argument" {
						caster = nil
					}
					*glyph, *flame = 11, 23
					GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: new(server.Server)} }
					if got := burnCastNativeCall(route, int32(spell.SPELL_BURN), a.objects[1], a.objects[2], caster, arg, -7); got != 0 || *glyph != 11 || *flame != 23 {
						t.Fatal("actual default cached-type/nil guard binding differs", got, *glyph, *flame)
					}
					a.check(t)
				})
			}
		}
	}
}
