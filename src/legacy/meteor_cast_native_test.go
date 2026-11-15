package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/meteor_cast_original_52d8a0_52d9d0_v1.json.gz
var meteorCastOriginal []byte

type meteorCastEvent struct {
	Kind   string   `json:"kind"`
	At     int      `json:"at"`
	Args   []uint32 `json:"args"`
	Result uint32   `json:"result"`
	Before string   `json:"before"`
	After  string   `json:"after"`
}

type meteorCastCase struct {
	Input struct {
		Name           string        `json:"name"`
		Root           int           `json:"root"`
		ID             int32         `json:"id"`
		Second         int           `json:"second"`
		Owner          int           `json:"owner"`
		Caster         int           `json:"caster"`
		Arg            int           `json:"arg"`
		Level          int32         `json:"level"`
		Lookup         uint32        `json:"lookup"`
		Flag           uint32        `json:"flag"`
		Trace          uint32        `json:"trace"`
		Allocate       int           `json:"allocate"`
		Sound          uint32        `json:"sound"`
		Objects        [6][11]uint32 `json:"objects"`
		Arguments      [2][3]uint32  `json:"arguments"`
		Damage         [3]uint32     `json:"damage"`
		UpdatePlayer   [2]uint32     `json:"update_player"`
		PlayerIndex    [2]uint32     `json:"player_index"`
		Cache          [2]uint32     `json:"cache"`
		BalanceValues  [32]uint32    `json:"balance_values"`
		BalanceCount   uint32        `json:"balance_count"`
		BalancePresent uint32        `json:"balance_present"`
		SpeedValue     uint32        `json:"speed_value"`
		SpeedPresent   uint32        `json:"speed_present"`
		Mutations      []struct {
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

type meteorCastArena struct {
	raw, expected []byte
	objects       [6]*server.Object
	arguments     [2]*server.SpellAcceptArg
	damages       [3]*server.MeteorUpdateData
	updates       [2]*server.PlayerUpdateData
	players       [2]*server.Player
	layout        [37]uintptr
	cache         [2]uint32
	table         [32]uint32
	control       [4]uint32
}

func meteorCastArenaNew(t *testing.T, placement int) *meteorCastArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("native meteor owners must be above4GiB")
	}
	a := &meteorCastArena{raw: raw, layout: meteorCastLayoutC()}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x400+placement]))
	}
	for i := range a.arguments {
		a.arguments[i] = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x4000+i*0x100+placement]))
		a.updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x6000+i*0x400+placement]))
		a.players[i] = (*server.Player)(unsafe.Pointer(&raw[0x8000+i*0x2000+placement]))
	}
	for i := range a.damages {
		a.damages[i] = (*server.MeteorUpdateData)(unsafe.Pointer(&raw[0x5000+i*0x40+placement]))
	}
	return a
}

func (a *meteorCastArena) object(i int) *server.Object {
	if i < 0 || uint32(i) == math.MaxUint32 {
		return nil
	}
	return a.objects[i]
}

func (a *meteorCastArena) owner(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if o == p {
			return uint32(i)
		}
	}
	panic("unknown native meteor object")
}

func (a *meteorCastArena) data(i uint32) unsafe.Pointer {
	if i == math.MaxUint32 {
		return nil
	}
	if i < 3 {
		return unsafe.Pointer(a.damages[i])
	}
	return unsafe.Pointer(a.updates[i-3])
}

func (a *meteorCastArena) dataOwner(p unsafe.Pointer) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i := uint32(0); i < 5; i++ {
		if p == a.data(i) {
			return i
		}
	}
	panic("unknown native meteor update record")
}

func (a *meteorCastArena) player(i uint32) *server.Player {
	if i == math.MaxUint32 {
		return nil
	}
	return a.players[i]
}

func (a *meteorCastArena) playerOwner(p *server.Player) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, player := range a.players {
		if p == player {
			return uint32(i)
		}
	}
	panic("unknown native meteor player")
}

func (a *meteorCastArena) words() (out [117]uint32) {
	for i, o := range a.objects {
		v := [11]uint32{uint32(o.TypeInd), uint32(o.ObjClass), o.Field5, math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.ZVal), math.Float32bits(o.Field27), a.owner(o.Field128), a.owner(o.Field129), a.dataOwner(o.UpdateData), a.owner(o.ObjOwner)}
		copy(out[11*i:], v[:])
	}
	for i, arg := range a.arguments {
		out[66+3*i], out[67+3*i], out[68+3*i] = a.owner(arg.Obj), math.Float32bits(arg.Pos.X), math.Float32bits(arg.Pos.Y)
	}
	for i, data := range a.damages {
		out[72+i] = uint32(data.Damage)
	}
	for i := range a.updates {
		out[75+i], out[77+i] = a.playerOwner(a.updates[i].Player), uint32(a.players[i].PlayerInd)
	}
	copy(out[79:81], a.cache[:])
	copy(out[81:113], a.table[:])
	copy(out[113:], a.control[:])
	return
}

func (a *meteorCastArena) hash() string {
	words := a.words()
	var raw [468]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(raw[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw[:]))
}

// Patch only C-observed field locations; all other bytes stay canaries.
func (a *meteorCastArena) patch(raw []byte, words [117]uint32) {
	offset := func(p unsafe.Pointer) int { return int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0]))) }
	put := func(at int, width uintptr, v uint64) {
		for b := uintptr(0); b < width; b++ {
			raw[at+int(b)] = byte(v >> (8 * b))
		}
	}
	for i, o := range a.objects {
		start := offset(unsafe.Pointer(o))
		for field := 0; field < 11; field++ {
			v := uint64(words[11*i+field])
			if field == 7 || field == 8 || field == 10 {
				v = uint64(uintptr(unsafe.Pointer(a.object(int(uint32(v))))))
			} else if field == 9 {
				v = uint64(uintptr(a.data(uint32(v))))
			}
			put(start+int(a.layout[field]), a.layout[11+field], v)
		}
	}
	for i, arg := range a.arguments {
		start := offset(unsafe.Pointer(arg))
		for field := 0; field < 3; field++ {
			v := uint64(words[66+3*i+field])
			if field == 0 {
				v = uint64(uintptr(unsafe.Pointer(a.object(int(uint32(v))))))
			}
			put(start+int(a.layout[22+field]), a.layout[25+field], v)
		}
		put(offset(unsafe.Pointer(a.updates[i]))+int(a.layout[29]), a.layout[34], uint64(uintptr(unsafe.Pointer(a.player(words[75+i])))))
		raw[offset(unsafe.Pointer(a.players[i]))+int(a.layout[30])] = byte(words[77+i])
	}
	for i, data := range a.damages {
		put(offset(unsafe.Pointer(data))+int(a.layout[35]), a.layout[36], uint64(words[72+i]))
	}
}

func (a *meteorCastArena) setState(words [117]uint32) {
	a.patch(a.raw, words)
	copy(a.cache[:], words[79:81])
	copy(a.table[:], words[81:113])
	copy(a.control[:], words[113:])
}

func (a *meteorCastArena) check(t *testing.T) {
	t.Helper()
	a.patch(a.expected, a.words())
	if !bytes.Equal(a.raw, a.expected) {
		for i := range a.raw {
			if a.raw[i] != a.expected[i] {
				t.Fatalf("undeclared64KiB meteor arena write at%x: got%02x want%02x", i, a.raw[i], a.expected[i])
			}
		}
	}
}

func TestMeteorCastNativeLayout(t *testing.T) {
	var o server.Object
	var arg server.SpellAcceptArg
	var update server.PlayerUpdateData
	var player server.Player
	var damage server.MeteorUpdateData
	c := meteorCastLayoutC()
	want := [11]uintptr{unsafe.Offsetof(o.TypeInd), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.Field5), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.Y), unsafe.Offsetof(o.ZVal), unsafe.Offsetof(o.Field27), unsafe.Offsetof(o.Field128), unsafe.Offsetof(o.Field129), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.ObjOwner)}
	widths := [11]uintptr{unsafe.Sizeof(o.TypeInd), unsafe.Sizeof(o.ObjClass), unsafe.Sizeof(o.Field5), unsafe.Sizeof(o.PosVec.X), unsafe.Sizeof(o.PosVec.Y), unsafe.Sizeof(o.ZVal), unsafe.Sizeof(o.Field27), unsafe.Sizeof(o.Field128), unsafe.Sizeof(o.Field129), unsafe.Sizeof(o.UpdateData), unsafe.Sizeof(o.ObjOwner)}
	if !reflect.DeepEqual(c[:11], want[:]) || !reflect.DeepEqual(c[11:22], widths[:]) {
		t.Fatalf("native object C offsets/widths%v Go%v/%v", c, want, widths)
	}
	argFields := [7]uintptr{unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.X), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.Y), unsafe.Sizeof(arg.Obj), unsafe.Sizeof(arg.Pos.X), unsafe.Sizeof(arg.Pos.Y), unsafe.Sizeof(arg)}
	if !reflect.DeepEqual(c[22:29], argFields[:]) || c[29] != unsafe.Offsetof(update.Player) || c[30] != unsafe.Offsetof(player.PlayerInd) || c[34] != unsafe.Sizeof(uintptr(0)) || c[35] != unsafe.Offsetof(damage.Damage) || c[36] != unsafe.Sizeof(damage.Damage) {
		t.Fatalf("native argument/update/player/damage layout C%v GoArg%v", c, argFields)
	}
	if c[31] > unsafe.Sizeof(o) || unsafe.Sizeof(o) > 1024 || unsafe.Sizeof(update) > 1024 || unsafe.Sizeof(player) > 0x2000 {
		t.Fatalf("native arena spacing: object C%d Go%d, update C%d Go%d, player C%d Go%d", c[31], unsafe.Sizeof(o), c[32], unsafe.Sizeof(update), c[33], unsafe.Sizeof(player))
	}
	t.Logf("native record sizes: object C%d Go%d, update C%d Go%d, player C%d Go%d", c[31], unsafe.Sizeof(o), c[32], unsafe.Sizeof(update), c[33], unsafe.Sizeof(player))
}

func TestMeteorCastNativePublicPointers(t *testing.T) {
	previousShower, previousMeteor := meteorShowerCastCall52D8A0, meteorCastCall52D9D0
	t.Cleanup(func() { meteorShowerCastCall52D8A0, meteorCastCall52D9D0 = previousShower, previousMeteor })
	a := meteorCastArenaNew(t, 64)
	for route := 0; route < 3; route++ {
		for root := 0; root < 2; root++ {
			for _, value := range []int32{math.MinInt32, -1, 0, 52, 53, math.MaxInt32} {
				t.Run(fmt.Sprintf("route%d/root%d/value%d", route, root, value), func(t *testing.T) {
					calls := 0
					call := func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						calls++
						if int32(id) != value || int32(level) != value || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.arguments[1] {
							t.Fatal("C entry changed native owners, argument or signed DWORD values")
						}
						return -12345
					}
					meteorShowerCastCall52D8A0, meteorCastCall52D9D0 = call, call
					before := bytes.Clone(a.raw)
					if got := meteorCastActualCall(route, root, value, a.objects[0], a.objects[1], a.objects[2], a.arguments[1], value); got != -12345 || calls != 1 || !bytes.Equal(a.raw, before) {
						t.Fatalf("return/calls=%d/%d or foreign bytes changed", got, calls)
					}
				})
			}
		}
	}
}

// These compositions call the real default bindings and ordinary server
// helpers. No cast result, trace result or owned-list outcome is supplied.
func TestMeteorCastNativeDefaultBindings(t *testing.T) {
	showerCache := memmap.PtrUint32(0x5D4594, 2487800)
	meteorCache := Get_dword_5d4594_2487804_ptr()
	retiredCache := memmap.PtrUint32(0x5D4594, 2487804)
	if meteorCache == retiredCache {
		t.Fatal("Meteor cache aliases retired PE32 storage")
	}
	oldShower, oldMeteor, oldRetired := *showerCache, *meteorCache, *retiredCache
	oldServer, oldFlags := GetServer, noxflags.GetGame()
	oldShowerCast, oldMeteorCast := meteorShowerCastCall52D8A0, meteorCastCall52D9D0
	t.Cleanup(func() {
		*showerCache, *meteorCache, *retiredCache = oldShower, oldMeteor, oldRetired
		GetServer = oldServer
		meteorShowerCastCall52D8A0, meteorCastCall52D9D0 = oldShowerCast, oldMeteorCast
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	GetServer = func() Server { return &meteorCastLegacyServer52D9D0{srv: new(server.Server)} }
	meteorShowerCastCall52D8A0, meteorCastCall52D9D0 = castMeteorShowerNative52D8A0, castMeteorNative52D9D0
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := meteorCastArenaNew(t, placement)
			initial := func() {
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
			for match := 0; match < 3; match++ {
				t.Run(fmt.Sprintf("route%d/place%d/owned%d", route, placement, match), func(t *testing.T) {
					initial()
					*showerCache, *meteorCache, *retiredCache = 0xfedc0123, 321, 0xabcdef01
					a.objects[0].Field129 = a.objects[3]
					for i := 3; i < 6; i++ {
						a.objects[i].TypeInd = uint16(91 + i)
						if i < 5 {
							a.objects[i].Field128 = a.objects[i+1]
						}
					}
					a.objects[3+match].TypeInd = 321
					before := bytes.Clone(a.raw)
					// An owned match precedes all caster/argument reads.
					if got := meteorCastActualCall(route, 1, 52, nil, a.objects[0], nil, nil, math.MinInt32); got != 0 || !bytes.Equal(a.raw, before) || *meteorCache != 321 || *showerCache != 0xfedc0123 || *retiredCache != 0xabcdef01 {
						t.Fatalf("default owned match result=%d or state changed", got)
					}
				})
			}
			for root := 0; root < 2; root++ {
				for _, flag := range []noxflags.GameFlag{0, 2048} {
					for _, kind := range []uint32{0x1234, 0xfedc1234} {
						t.Run(fmt.Sprintf("route%d/place%d/blocked%d/flag%d/type%x", route, placement, root, flag, kind), func(t *testing.T) {
							initial()
							noxflags.ResetGame()
							noxflags.SetGame(flag)
							*showerCache, *meteorCache, *retiredCache = kind, kind, 0xabcdef01
							a.objects[2].PosVec = types.Ptf(-100, -200)
							a.arguments[0].Pos = types.Ptf(30, 40)
							before := bytes.Clone(a.raw)
							// Ordinary MapTraceRay rejects an out-of-grid origin.
							got := meteorCastActualCall(route, root, 52, a.objects[5], a.objects[0], a.objects[2], a.arguments[0], math.MaxInt32)
							if got != 0 || !bytes.Equal(a.raw, before) || *meteorCache != kind || *showerCache != kind || *retiredCache != 0xabcdef01 || noxflags.GetGame() != flag {
								t.Fatalf("default blocked cast result=%d or state changed", got)
							}
						})
					}
				}
			}
		}
	}
}

func TestMeteorCastNativeOriginalMatrix(t *testing.T) {
	var rows []meteorCastCase
	npcVoiceOriginalDecode(t, meteorCastOriginal, "523a45b59c9c640d4e1234fff89faf7dc32257058813244d76c753ed3812e754", "d53f5f8245f02e32bc97aaa2110f5172e1204561502842392e8ac828a6f883fd", 146, &rows)
	previousShower, previousMeteor := meteorShowerCastCall52D8A0, meteorCastCall52D9D0
	t.Cleanup(func() { meteorShowerCastCall52D8A0, meteorCastCall52D9D0 = previousShower, previousMeteor })
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
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
						if e.Args == nil {
							e.Args = []uint32{}
						}
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
						} else if kind == "raise" {
							a.object(int(args[0])).ZVal = math.Float32frombits(args[1])
						}
						for _, m := range c.Input.Mutations {
							if m.Kind != kind || m.At != counts[kind] {
								continue
							}
							state := a.words()
							index, value := 0, m.Value
							switch m.Action {
							case "object":
								index = 11*m.Object + m.Field
								if m.Field == 0 {
									value &= 65535
								}
							case "arg":
								index = 66 + 3*m.Arg + m.Field
							case "damage":
								index = 72 + m.Slot
							case "update-player":
								index = 75 + m.Slot
							case "player-index":
								index, value = 77+m.Slot, value&255
							case "cache":
								index = 79 + m.Slot
							case "balance-value":
								index = 81 + m.Slot
							case "balance-count":
								index, value = 113, value&255
							case "balance-present":
								index = 114
							case "speed-value":
								index = 115
							case "speed-present":
								index = 116
							default:
								t.Fatal(m.Action)
							}
							state[index] = value
							a.setState(state)
						}
						if kind == "balance" {
							index := int32(args[0])
							result = 0 // Original 0x58307c fallback word.
							if a.control[1] != 0 && index < int32(a.control[0]) {
								// The original PE32 getter multiplies the signed index
								// modulo DWORD before reading the owned table record.
								relative := int32(uint32(index)*4) / 4
								result = a.table[8+relative]
							}
						} else if kind == "speed" {
							result = 0
							if a.control[3] != 0 {
								result = a.control[2]
							}
						}
						emit(meteorCastEvent{Kind: kind, Args: args, Result: result, Before: before, After: a.hash()})
						a.check(t)
						return result
					}
					const ignored = uint32(0xaabbccdd)
					load := func() uint32 {
						v := a.cache[c.Input.Root]
						emit(meteorCastEvent{Kind: "cache-read", Args: []uint32{uint32(c.Input.Root)}, Result: v})
						return v
					}
					store := func(v uint32) {
						a.cache[c.Input.Root] = v
						emit(meteorCastEvent{Kind: "cache-write", Args: []uint32{uint32(c.Input.Root), v}})
					}
					lookup := func(name string) uint32 {
						want := "MeteorShower"
						if c.Input.Root == 1 {
							want = "Meteor"
						}
						if name != want {
							t.Fatal(name)
						}
						return service("lookup", []uint32{uint32(c.Input.Root)}, c.Input.Lookup)
					}
					flag := func(v uint32) bool { return service("flag", []uint32{v}, c.Input.Flag) != 0 }
					trace := func(from, to types.Pointf, flags server.MapTraceFlags) bool {
						return service("trace", []uint32{math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y), uint32(flags)}, c.Input.Trace) != 0
					}
					allocate := func(kind uint32) *server.Object {
						service("allocate", []uint32{kind}, uint32(c.Input.Allocate))
						return a.object(c.Input.Allocate)
					}
					balance := func(key string, index int32) float64 {
						if key != "MeteorDamage" {
							t.Fatal(key)
						}
						return float64(math.Float32frombits(service("balance", []uint32{uint32(index)}, 0)))
					}
					create := func(o, owner *server.Object, pos types.Pointf) {
						service("create", []uint32{a.owner(o), a.owner(owner), math.Float32bits(pos.X), math.Float32bits(pos.Y), 0}, ignored)
					}
					castSound := func(id int32) sound.ID {
						return sound.ID(int32(service("cast-sound", []uint32{uint32(id), 0}, c.Input.Sound)))
					}
					inform := func(index uint8, code byte, value int32) {
						service("inform", []uint32{uint32(index), uint32(code), uint32(value)}, ignored)
					}
					shower := server.MeteorShowerNativeRuntime52D8A0{LoadCache: load, StoreCache: store, LookupType: lookup, HasGameFlag: flag, TraceRay: trace, NewObject: allocate, Balance: balance, CreateAt: create, CastSound: castSound, Inform: inform,
						Audio: func(id sound.ID, owner *server.Object, kind int, code uint32) {
							service("object-audio", []uint32{uint32(id), a.owner(owner), uint32(kind), code}, ignored)
						},
					}
					meteor := server.MeteorNativeRuntime52D9D0{LoadCache: load, StoreCache: store, LookupType: lookup, HasGameFlag: flag, TraceRay: trace, NewObject: allocate, Balance: balance, CreateAt: create, CastSound: castSound, Inform: inform,
						PriorityMsg: func(owner *server.Object, id string, arg byte) {
							if id != "ExecSpel.c:TooManyMeteors" {
								t.Fatal(id)
							}
							service("priority", []uint32{a.owner(owner), uint32(arg)}, ignored)
						},
						Raise: func(o *server.Object, height float32) {
							service("raise", []uint32{a.owner(o), math.Float32bits(height)}, ignored)
						},
						Speed: func(key string) float64 {
							if key != "MeteorSpeed" {
								t.Fatal(key)
							}
							return float64(math.Float32frombits(service("speed", nil, 0)))
						},
						Audio: func(id sound.ID, pos types.Pointf, kind int, code uint32) {
							service("position-audio", []uint32{uint32(id), math.Float32bits(pos.X), math.Float32bits(pos.Y), uint32(kind), code}, ignored)
						},
					}
					meteorShowerCastCall52D8A0 = func(id spell.ID, _ *server.Object, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						return int(server.MeteorShowerNative52D8A0(int32(id), owner, caster, arg, int32(level), shower))
					}
					meteorCastCall52D9D0 = func(id spell.ID, _ *server.Object, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						return int(server.MeteorNative52D9D0(int32(id), owner, caster, arg, int32(level), meteor))
					}
					got := meteorCastActualCall(route, c.Input.Root, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), a.arguments[c.Input.Arg], c.Input.Level)
					if uint32(got) != c.Return || cursor != len(c.Events) || a.words() != c.After {
						t.Fatalf("return/events/final state got%d/%d want%d/%d", got, cursor, c.Return, len(c.Events))
					}
					a.check(t)
				})
			}
		}
	}
}
