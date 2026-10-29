package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func manaBombOriginalAllocation530F90[T comparable](t *testing.T) *T {
	t.Helper()
	var zero T
	p, free := alloc.New(zero)
	t.Cleanup(free)
	if uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatal("native comparison allocation must exceed4GiB")
	}
	return p
}

// The fixture contains complete original root execution. Only the discarded
// destructor EAX is excluded. The native charge sidecar is compared with the
// original PE32 pointer slot; unrelated native record bytes stay intact.
func TestSpellManaBombOriginalNative530F90(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native C-owned comparison above4GiB")
	}
	path := os.Getenv("NOX_MANA_BOMB_ORACLE_FIXTURE")
	if path == "" {
		path = "testdata/spell_mana_bomb_native_original_530f90.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Control uint32 `json:"control_word"`
		Cases   []struct {
			Input  map[string]any `json:"input"`
			Result map[string]any `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Control != 0xe3f || len(fixture.Cases) != 1969 {
		t.Fatal("sealed startup original corpus")
	}
	u := func(c map[string]any, key string) uint32 { return uint32(c[key].(float64)) }
	pt := func(value any) types.Pointf {
		p := value.([]any)
		return types.Ptf(math.Float32frombits(uint32(p[0].(float64))), math.Float32frombits(uint32(p[1].(float64))))
	}
	pointBits := func(p types.Pointf) []uint32 { return []uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	canonical := func(value any, destructor bool) any {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var copy map[string]any
		if err := json.Unmarshal(encoded, &copy); err != nil {
			t.Fatal(err)
		}
		if destructor {
			copy["return_value"] = float64(0)
		}
		return copy
	}
	for _, row := range fixture.Cases {
		c := row.Input
		t.Run(c["name"].(string), func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("native comparison panicked: %v", p)
				}
			}()
			objects := make([]*Object, 4)
			for i := range objects {
				objects[i] = manaBombOriginalAllocation530F90[Object](t)
				*objects[i] = Object{ObjClass: object.ClassPlayer, Mass: 17.5 + float32(i),
					VelVec:   types.Ptf(float32(1+i), float32(2+i)),
					ForceVec: types.Ptf(float32(3+i), float32(4+i)), Pos24: types.Ptf(float32(5+i), float32(6+i))}
			}
			index := func(p *Object) int {
				if p == nil {
					return -1
				}
				for i, o := range objects {
					if p == o {
						return i
					}
				}
				t.Fatal("object outside sealed identity set")
				return -2
			}
			updates := []*PlayerUpdateData{manaBombOriginalAllocation530F90[PlayerUpdateData](t), manaBombOriginalAllocation530F90[PlayerUpdateData](t)}
			for i := range updates {
				objects[i].UpdateData = unsafe.Pointer(updates[i])
				updates[i].ManaCur = 37
			}
			updates[0].ManaCur = uint16(u(c, "mana"))
			objects[0].ObjClass, objects[0].PosVec = object.Class(u(c, "class")), pt(c["caster_pos"])
			objects[1].PosVec = types.Ptf(211.25, -41.5)
			type frame struct {
				Before [32]byte
				Record DurSpell
				After  [32]byte
			}
			f := manaBombOriginalAllocation530F90[frame](t)
			for i := range f.Before {
				f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
			}
			r := &f.Record
			*r = DurSpell{ID: 0xabcd, Level: u(c, "level"), Caster16: objects[0], Flag20: u(c, "mode"),
				Pos: pt(c["pos"]), Frame68: u(c, "expiry"), Field72: int32(u(c, "power")),
				Field76: 0x3fdccccc, Field80: u(c, "saved_mass"), Field84: u(c, "exploded")}
			if c["caster_nil"].(bool) {
				r.Caster16 = nil
			}
			before := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(r)), int(unsafe.Sizeof(*r)))...)
			var charge *Object
			if c["charge"].(bool) {
				charge = objects[2]
			}
			cache, frameNumber, fps, walls := u(c, "cache"), u(c, "frame"), u(c, "fps"), u(c, "walls")
			state := func() map[string]any {
				actors := make([]any, 2)
				for i, o := range objects[:2] {
					actors[i] = map[string]any{"class_bits": uint32(o.ObjClass), "mana": updates[i].ManaCur,
						"mass": math.Float32bits(o.Mass), "pos": pointBits(o.PosVec), "velocity": pointBits(o.VelVec),
						"force": pointBits(o.ForceVec), "pos24": pointBits(o.Pos24)}
				}
				return map[string]any{"level": r.Level, "caster": index(r.Caster16), "mode": r.Flag20, "pos": pointBits(r.Pos),
					"expiry": r.Frame68, "power": uint32(r.Field72), "charge": index(charge), "saved_mass": r.Field80, "exploded": r.Field84,
					"cache": cache, "frame": frameNumber, "fps": fps, "walls": walls, "actors": actors, "charge_pos": pointBits(objects[3].PosVec)}
			}
			events := make([]map[string]any, 0)
			calls := make(map[string]uint32)
			changed := false
			mutation := c["mutation"].(map[string]any)
			event := func(kind string, values []any) {
				events = append(events, map[string]any{"kind": kind, "args": values, "state": state()})
				n := calls[kind]
				calls[kind]++
				if changed || mutation["stage"].(string) != kind || uint32(mutation["at"].(float64)) != n {
					return
				}
				changed = true
				switch mutation["action"].(string) {
				case "caster-b":
					r.Caster16 = objects[1]
				case "caster-nil":
					r.Caster16 = nil
				case "mode-one":
					r.Flag20 = 1
				case "mode-zero":
					r.Flag20 = 0
				case "record-position":
					r.Pos = types.Ptf(-73.25, 97.125)
				case "caster-position":
					objects[0].PosVec = types.Ptf(8193.125, -4097.25)
				case "power":
					r.Field72 = 0x76543210
				case "level":
					r.Level = 2
				case "fps":
					fps = 120
				case "frame":
					frameNumber = 0xfffffffd
				case "expiry":
					r.Frame68 = frameNumber + 1
				case "saved-mass":
					r.Field80 = math.Float32bits(9.75)
				case "cache":
					cache = 137
				case "walls":
					walls = 0
				default:
					t.Fatal("unknown sealed mutation")
				}
			}
			balance := func(key string, level uint32) float32 {
				event("lookup", []any{key})
				for _, missing := range c["missing"].([]any) {
					if missing.(string) == key {
						return 0
					}
				}
				if level >= 5 {
					return 0
				}
				values := c["balance"].(map[string]any)[key].([]any)
				return math.Float32frombits(uint32(values[level].(float64)))
			}
			runtime := SpellManaBombRuntime530F90{
				Frame: func() uint32 { return frameNumber }, TickRate: func() uint32 { return fps },
				Balance: func(key string) float32 { return balance(key, 0) }, BalanceLevel: balance,
				LoadTypeID: func() uint32 { return cache }, StoreTypeID: func(v uint32) { cache = v },
				NameID: func(name string) uint32 { event("name", []any{name}); return u(c, "resolved_type") },
				AllocateCharge: func(id uint32) *Object {
					event("allocate", []any{id})
					if c["allocate"].(bool) {
						return objects[3]
					}
					return nil
				},
				NewObject: func(name string) *Object {
					event("new-by-name", []any{name})
					if c["allocate"].(bool) {
						return objects[3]
					}
					return nil
				},
				CreateAt: func(p, owner *Object, pos types.Pointf) {
					event("create", []any{index(p), index(owner), pointBits(pos), 0})
					p.PosVec = pos
				},
				LoadCharge: func(*DurSpell) *Object { return charge }, StoreCharge: func(_ *DurSpell, p *Object) { charge = p },
				DelayedDelete: func(p *Object) { event("delete", []any{index(p)}) },
				ApplyBuff: func(p *Object, buff EnchantID, duration int16, power int8) {
					event("buff", []any{index(p), uint32(buff), uint32(uint16(duration)), uint32(uint8(power))})
				},
				BuffOff:     func(p *Object, buff EnchantID) { event("off", []any{index(p), uint32(buff)}) },
				BeginDamage: func() { event("walls", []any{1}); walls = 1 },
				DamageAround: func(pos types.Pointf, outer, inner float32, damage int, p *Object) {
					event("damage", []any{pointBits(pos), math.Float32bits(outer), math.Float32bits(inner), uint32(damage), 15, index(p), 0})
				},
				Earthquake: func(pos types.Pointf, magnitude int) { event("quake", []any{pointBits(pos), uint32(magnitude)}) },
				PointFX:    func(effect netmsg.Op, pos types.Pointf) { event("fx", []any{uint32(effect), pointBits(pos)}) },
				Audio:      func(p *Object) { event("audio", []any{81, index(p), 0, 0}) },
				ManaSub:    func(p *Object, amount int32) { event("mana", []any{index(p), uint32(amount)}) },
			}
			function := c["function"].(string)
			var result int32
			switch function {
			case "create":
				result = SpellManaBombCreate530F90(r, runtime)
			case "update":
				result = SpellManaBombUpdate5310C0(r, runtime)
			case "destroy":
				SpellManaBombDestroy531290(r, runtime)
			default:
				t.Fatal("unknown sealed entry")
			}
			actual := map[string]any{"return_value": uint32(result), "events": events, "state": state()}
			if !reflect.DeepEqual(canonical(actual, function == "destroy"), canonical(row.Result, function == "destroy")) {
				a, _ := json.Marshal(actual)
				w, _ := json.Marshal(row.Result)
				t.Fatalf("native=%s\noriginal=%s", a, w)
			}
			wantOpaque := uintptr(0x3fdccccc)
			if function == "create" {
				wantOpaque = 0
			}
			if r.Field76 != wantOpaque {
				t.Fatal("charge pointer escaped into opaque Field76")
			}
			allowed := make([]bool, len(before))
			for _, field := range []struct{ off, n uintptr }{
				{unsafe.Offsetof(r.Level), unsafe.Sizeof(r.Level)}, {unsafe.Offsetof(r.Caster16), unsafe.Sizeof(r.Caster16)},
				{unsafe.Offsetof(r.Flag20), unsafe.Sizeof(r.Flag20)}, {unsafe.Offsetof(r.Pos), unsafe.Sizeof(r.Pos)},
				{unsafe.Offsetof(r.Frame68), unsafe.Sizeof(r.Frame68)}, {unsafe.Offsetof(r.Field72), unsafe.Sizeof(r.Field72)},
				{unsafe.Offsetof(r.Field76), unsafe.Sizeof(r.Field76)}, {unsafe.Offsetof(r.Field80), unsafe.Sizeof(r.Field80)},
				{unsafe.Offsetof(r.Field84), unsafe.Sizeof(r.Field84)},
			} {
				for i := field.off; i < field.off+field.n; i++ {
					allowed[i] = true
				}
			}
			for i, v := range unsafe.Slice((*byte)(unsafe.Pointer(r)), len(before)) {
				if v != before[i] && !allowed[i] {
					t.Fatalf("unrelated duration byte changed at%d", i)
				}
			}
			for i := range f.Before {
				if f.Before[i] != byte(0x51+i) || f.After[i] != byte(0xe7-i) {
					t.Fatal("duration canary")
				}
			}
		})
	}
}
