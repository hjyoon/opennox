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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/cleansing_flame_original_52d5c0_v1.json.gz
var cleansingFlameOriginal []byte

//go:embed testdata/cleansing_flame_direction_table_52d5c0_v1.bin
var cleansingFlameDirectionTable []byte

type cleansingFlameEvent struct {
	Kind   string   `json:"kind"`
	At     int      `json:"at"`
	Args   []uint32 `json:"args"`
	Result uint32   `json:"result"`
	Before string   `json:"before"`
	After  string   `json:"after"`
}

type cleansingFlameCase struct {
	Input struct {
		Name       string         `json:"name"`
		ID         int32          `json:"id"`
		Second     int            `json:"second"`
		Recipient  int            `json:"recipient"`
		Caster     int            `json:"caster"`
		Arg        int            `json:"arg"`
		Level      int32          `json:"level"`
		Cache      [10]uint32     `json:"cache"`
		Lookup     [10]uint32     `json:"lookup"`
		Objects    [54][16]uint32 `json:"objects"`
		FPS        uint32         `json:"fps"`
		Frame      uint32         `json:"frame"`
		Flag       uint32         `json:"flag"`
		Power      [48]int32      `json:"power"`
		Directions [48]uint32     `json:"directions"`
		Duration   [48]uint32     `json:"duration"`
		Allocate   [48]int        `json:"allocate"`
		Trace      [48]int        `json:"trace"`
		Sound      uint32         `json:"sound"`
		Mutations  []struct {
			Kind   string `json:"kind"`
			At     int    `json:"at"`
			Action string `json:"action"`
			Value  uint32 `json:"value"`
			Slot   int    `json:"slot"`
			Object int    `json:"object"`
			Field  int    `json:"field"`
		} `json:"mutations"`
	} `json:"input"`
	Before [876]uint32           `json:"before"`
	Events []cleansingFlameEvent `json:"events"`
	After  [876]uint32           `json:"after"`
	Return uint32                `json:"return_u32"`
}

type cleansingFlameArena struct {
	raw, expected []byte
	objects       [54]*server.Object
	layout        [34]uintptr
	callback      unsafe.Pointer
	cache         [10]uint32
	fps, frame    uint32
}

func cleansingFlameArenaNew(t *testing.T, placement int) *cleansingFlameArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("foreign native objects must be above4GiB")
	}
	a := &cleansingFlameArena{raw: raw, layout: cleansingFlameLayoutC(), callback: Get_nox_xxx_updateFlameCleanse_53D510()}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x400+i*0x400+placement]))
	}
	return a
}

func (a *cleansingFlameArena) object(i int) *server.Object {
	if i < 0 || uint32(i) == math.MaxUint32 {
		return nil
	}
	return a.objects[i]
}

func (a *cleansingFlameArena) owner(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, obj := range a.objects {
		if p == obj {
			return uint32(i)
		}
	}
	panic("unknown CleansingFlame owner")
}

func (a *cleansingFlameArena) words() (out [876]uint32) {
	copy(out[:10], a.cache[:])
	out[10], out[11] = a.fps, a.frame
	for i, o := range a.objects {
		callback := uint32(0)
		if o.Update != nil {
			if o.Update != a.callback {
				panic("changed/truncated native CleansingFlame callback")
			}
			callback = 1
		}
		v := [16]uint32{uint32(o.TypeInd), uint32(o.ObjClass), math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.VelVec.X), math.Float32bits(o.VelVec.Y), math.Float32bits(o.Float28), uint32(o.Direction1), uint32(o.Direction2), o.Field34, math.Float32bits(o.Pos39.X), math.Float32bits(o.Pos39.Y), math.Float32bits(o.Shape.Circle.R), a.owner(o.Field128), a.owner(o.Field129), callback}
		copy(out[12+16*i:], v[:])
	}
	return
}

func (a *cleansingFlameArena) hash() string {
	words := a.words()
	var raw [3504]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(raw[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw[:]))
}

// C supplies every field offset/width. Bytes outside those fields, including
// unused pointers, private handles and the entire foreign arena, stay canaries.
func (a *cleansingFlameArena) patch(raw []byte, words [876]uint32) {
	for i, o := range a.objects {
		start := int(uintptr(unsafe.Pointer(o)) - uintptr(unsafe.Pointer(&a.raw[0])))
		for field := 0; field < 16; field++ {
			value := uint64(words[12+16*i+field])
			if field == 13 || field == 14 {
				value = uint64(uintptr(unsafe.Pointer(a.object(int(uint32(value))))))
			} else if field == 15 && value != 0 {
				value = uint64(uintptr(a.callback))
			}
			offset, width := start+int(a.layout[field]), a.layout[16+field]
			for b := uintptr(0); b < width; b++ {
				raw[offset+int(b)] = byte(value >> (8 * b))
			}
		}
	}
}

func (a *cleansingFlameArena) check(t *testing.T) {
	t.Helper()
	a.patch(a.expected, a.words())
	if !bytes.Equal(a.raw, a.expected) {
		for i := range a.raw {
			if a.raw[i] != a.expected[i] {
				t.Fatalf("undeclared foreign arena write at%x: got%02x want%02x", i, a.raw[i], a.expected[i])
			}
		}
	}
}

func TestCleansingFlameNativeCLayout52D5C0(t *testing.T) {
	var o server.Object
	got := cleansingFlameLayoutC()
	want := [16]uintptr{unsafe.Offsetof(o.TypeInd), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.Y), unsafe.Offsetof(o.VelVec) + unsafe.Offsetof(o.VelVec.X), unsafe.Offsetof(o.VelVec) + unsafe.Offsetof(o.VelVec.Y), unsafe.Offsetof(o.Float28), unsafe.Offsetof(o.Direction1), unsafe.Offsetof(o.Direction2), unsafe.Offsetof(o.Field34), unsafe.Offsetof(o.Pos39) + unsafe.Offsetof(o.Pos39.X), unsafe.Offsetof(o.Pos39) + unsafe.Offsetof(o.Pos39.Y), unsafe.Offsetof(o.Shape) + unsafe.Offsetof(o.Shape.Circle) + unsafe.Offsetof(o.Shape.Circle.R), unsafe.Offsetof(o.Field128), unsafe.Offsetof(o.Field129), unsafe.Offsetof(o.Update)}
	if !reflect.DeepEqual(got[:16], want[:]) || got[33] != unsafe.Sizeof(uintptr(0)) || got[32] > unsafe.Sizeof(o) || unsafe.Sizeof(o) > 1024 {
		t.Fatalf("native shared prefix: C%v Go%v, bytes C%d Go%d", got, want, got[32], unsafe.Sizeof(o))
	}
	for i, size := range got[16:32] {
		wantSize := uintptr(4)
		if i == 0 || i == 7 || i == 8 {
			wantSize = 2
		} else if i >= 13 {
			wantSize = unsafe.Sizeof(uintptr(0))
		}
		if size != wantSize {
			t.Fatalf("field%d C width%d want%d", i, size, wantSize)
		}
	}
}

func TestCleansingFlameNativeOriginal52D5C0(t *testing.T) {
	var rows []cleansingFlameCase
	npcVoiceOriginalDecode(t, cleansingFlameOriginal, "177dad1f9d98a8b466fbab21a4800248c7cc139466c00d1d9588eeecb329ac76", "07b6bb201f8947cd7b065c03f81c7561543fea812a0c856a7a14e7ca129fd413", 66, &rows)
	if fmt.Sprintf("%x", sha256.Sum256(cleansingFlameDirectionTable)) != "fbfdbd07ab3c6fbec2eb3768e69f1c32eb63c77b6e80e9957b9567c78d61bfcf" {
		t.Fatal("original direction table changed")
	}
	previous := cleansingFlameCastCall52D5C0
	t.Cleanup(func() { cleansingFlameCastCall52D5C0 = previous })
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := cleansingFlameArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					for i := range a.raw {
						a.raw[i] = byte(i*31 + 19)
					}
					a.cache, a.fps, a.frame = c.Input.Cache, c.Input.FPS, c.Input.Frame
					a.patch(a.raw, c.Before)
					a.expected = bytes.Clone(a.raw)
					if a.words() != c.Before {
						t.Fatal("native C field initialization differs from original")
					}
					cursor, attempt := 0, -1
					counts := make(map[string]int)
					emit := func(e cleansingFlameEvent) {
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
							a.object(int(args[0])).PosVec = types.Ptf(math.Float32frombits(args[2]), math.Float32frombits(args[3]))
						}
						for _, m := range c.Input.Mutations {
							if m.Kind != kind || m.At != counts[kind] {
								continue
							}
							switch m.Action {
							case "cache":
								a.cache[m.Slot] = m.Value
							case "fps":
								a.fps = m.Value
							case "frame":
								a.frame = m.Value
							case "object":
								state := a.words()
								state[12+16*m.Object+m.Field] = m.Value
								a.patch(a.raw, state)
							default:
								t.Fatal(m.Action)
							}
						}
						emit(cleansingFlameEvent{Kind: kind, Args: args, Result: result, Before: before, After: a.hash()})
						a.check(t)
						return result
					}
					const ignored = 0xaabbccdd
					rt := server.CleansingFlameNativeRuntime52D5C0{
						LoadCache: func(slot int32) uint32 {
							v := a.cache[slot]
							emit(cleansingFlameEvent{Kind: "cache-read", Args: []uint32{uint32(slot)}, Result: v})
							return v
						},
						StoreCache: func(slot int32, value uint32) {
							a.cache[slot] = value
							emit(cleansingFlameEvent{Kind: "cache-write", Args: []uint32{uint32(slot), value}})
						},
						LookupType: func(name string) uint32 {
							i := counts["lookup"]
							want := [10]string{"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse", "SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse"}
							if name != want[i] {
								t.Fatal(name)
							}
							return service("lookup", []uint32{uint32(i)}, c.Input.Lookup[i])
						},
						PriorityMsg: func(o *server.Object, id string, arg byte) {
							if id != "plyrspel.c:TooManySpells" {
								t.Fatal(id)
							}
							service("priority", []uint32{a.owner(o), uint32(arg)}, ignored)
						},
						HasGameFlag: func(flag uint32) bool { return service("flag", []uint32{flag}, c.Input.Flag) != 0 },
						Random: func(min, max int32) int32 {
							line, value := uint32(0x788), uint32(0)
							if min == 0 && max == 1 {
								attempt++
								line, value = 0x760, uint32(c.Input.Power[attempt])
							} else if min == 0 && max == 255 {
								line, value = 0x76b, c.Input.Directions[attempt]
							} else {
								value = c.Input.Duration[attempt]
							}
							return int32(service("random", []uint32{uint32(min), uint32(max), line}, value))
						},
						NewObject: func(kind uint32) *server.Object {
							owner := uint32(math.MaxUint32)
							if c.Input.Allocate[attempt] != 0 {
								owner = uint32(6 + attempt)
							}
							service("allocate", []uint32{kind}, owner)
							return a.object(int(owner))
						},
						DirectionVector: func(dir int16) types.Pointf {
							offset := int(dir) * 8
							x := binary.LittleEndian.Uint32(cleansingFlameDirectionTable[offset:])
							y := binary.LittleEndian.Uint32(cleansingFlameDirectionTable[offset+4:])
							emit(cleansingFlameEvent{Kind: "direction", Args: []uint32{uint32(int32(dir)), x, y}})
							return types.Ptf(math.Float32frombits(x), math.Float32frombits(y))
						},
						TraceRay: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
							return service("trace", []uint32{math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y), uint32(flags)}, uint32(c.Input.Trace[attempt])) != 0
						},
						CreateAt: func(o, owner *server.Object, pos types.Pointf) {
							service("create", []uint32{a.owner(o), a.owner(owner), math.Float32bits(pos.X), math.Float32bits(pos.Y), 0}, ignored)
						},
						DelayedDelete:  func(o *server.Object) { service("delete", []uint32{a.owner(o)}, ignored) },
						Fps:            func() uint32 { emit(cleansingFlameEvent{Kind: "fps", Result: a.fps}); return a.fps },
						Frame:          func() uint32 { emit(cleansingFlameEvent{Kind: "frame", Result: a.frame}); return a.frame },
						UpdateCallback: a.callback,
						AddUpdatable:   func(o *server.Object) { service("add", []uint32{a.owner(o)}, ignored) },
						PredictLinear:  func(o *server.Object) { service("predict", []uint32{a.owner(o)}, ignored) },
						CastSound: func(id int32) sound.ID {
							return sound.ID(int32(service("cast-sound", []uint32{uint32(id), 0}, c.Input.Sound)))
						},
						Audio: func(id sound.ID, o *server.Object, kind int, code uint32) {
							service("audio", []uint32{uint32(id), a.owner(o), uint32(kind), code}, ignored)
						},
					}
					cleansingFlameCastCall52D5C0 = func(id spell.ID, second, recipient, caster *server.Object, _ *server.SpellAcceptArg, level int) int {
						return int(server.CleansingFlameNative52D5C0(int32(id), second, recipient, caster, int32(level), rt))
					}
					got := cleansingFlameActualCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Recipient), a.object(c.Input.Caster), (*server.SpellAcceptArg)(unsafe.Pointer(a.object(c.Input.Arg))), c.Input.Level)
					if uint32(got) != c.Return || cursor != len(c.Events) || a.words() != c.After {
						t.Fatalf("return/events/final state got%d/%d want%d/%d", got, cursor, c.Return, len(c.Events))
					}
					a.check(t)
				})
			}
		}
	}
}

func TestCleansingFlameNativePublicPointers52D5C0(t *testing.T) {
	a := cleansingFlameArenaNew(t, 64)
	previous := cleansingFlameCastCall52D5C0
	t.Cleanup(func() { cleansingFlameCastCall52D5C0 = previous })
	for route := 0; route < 3; route++ {
		for _, value := range []int32{math.MinInt32, -1, 0, 10, 11, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/value%d", route, value), func(t *testing.T) {
				calls := 0
				arg := (*server.SpellAcceptArg)(unsafe.Pointer(a.objects[5]))
				cleansingFlameCastCall52D5C0 = func(id spell.ID, second, recipient, caster *server.Object, gotArg *server.SpellAcceptArg, level int) int {
					calls++
					if int32(id) != value || int32(level) != value || second != a.objects[0] || recipient != a.objects[1] || caster != a.objects[2] || gotArg != arg {
						t.Fatal("C entry changed native arguments or signed DWORD values")
					}
					return -12345
				}
				if got := cleansingFlameActualCall(route, value, a.objects[0], a.objects[1], a.objects[2], arg, value); got != -12345 || calls != 1 {
					t.Fatalf("return/calls=%d/%d", got, calls)
				}
			})
		}
	}
}

// Exercise the actual default binding, native owned links and stock priority
// helper. An original owned-type match returns before caster/argument reads.
func TestCleansingFlameNativeDefaultOwned52D5C0(t *testing.T) {
	cache := unsafe.Slice(memmap.PtrUint32(0x5D4594, 2487760), 10)
	beforeCache := append([]uint32(nil), cache...)
	previousServer, previousCast := GetServer, cleansingFlameCastCall52D5C0
	t.Cleanup(func() {
		copy(cache, beforeCache)
		GetServer, cleansingFlameCastCall52D5C0 = previousServer, previousCast
	})
	GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: new(server.Server)} }
	cleansingFlameCastCall52D5C0 = castCleansingFlameNative52D5C0
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := cleansingFlameArenaNew(t, placement)
			for slot := 0; slot < 10; slot++ {
				t.Run(fmt.Sprintf("route%d/place%d/cache%d", route, placement, slot), func(t *testing.T) {
					for i := range a.raw {
						a.raw[i] = byte(i*31 + 19)
					}
					var state [876]uint32
					for i := 0; i < 54; i++ {
						state[12+16*i+13], state[12+16*i+14] = math.MaxUint32, math.MaxUint32
					}
					state[12+14] = 3
					state[12+16*3], state[12+16*3+13] = 77, 4
					state[12+16*4], state[12+16*4+13] = 78, 5
					state[12+16*5] = uint32(100 + slot)
					a.patch(a.raw, state)
					before := bytes.Clone(a.raw)
					for i := range cache {
						cache[i] = uint32(100 + i)
					}
					got := cleansingFlameActualCall(route, 10, a.objects[0], nil, nil, (*server.SpellAcceptArg)(unsafe.Pointer(a.objects[2])), math.MinInt32)
					if got != 0 || !bytes.Equal(a.raw, before) {
						t.Fatalf("default owned match return%d or modified native arena", got)
					}
					for i, value := range cache {
						if value != uint32(100+i) {
							t.Fatal("default owned match changed cache")
						}
					}
				})
			}
		}
	}
}
