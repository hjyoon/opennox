package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type monsterDeathRawContract549D80 interface {
	MonsterDeathExplosionNative549D80(uint32, server.MonsterDeathExplosionRuntime549D80) int32
}

type monsterDeathInput549D80 struct {
	Name          string
	Entry         uint32
	Unit          int
	Positions     [2][2]uint32
	FlagResult    uint32 `json:"flag_result"`
	ReadPositions bool   `json:"read_positions"`
	Script        map[string]struct {
		Positions      *[2][2]uint32
		AfterPositions *[2][2]uint32 `json:"after_positions"`
		FlagResult     *uint32       `json:"flag_result"`
		Reenter        *struct {
			Entry uint32
			Unit  int
		}
	}
}

type monsterDeathOutput549D80 struct {
	Name       string
	Result     *int32
	Fault      map[string]any
	Events     []map[string]any
	Positions  [2][2]uint32
	HeapSHA256 string `json:"original_heap_sha256"`
}

type monsterDeathWorldReadFault549D80 struct{}

func monsterDeathFixtureBytes549D80(t *testing.T, name, compressedSHA, decodedSHA string) []byte {
	t.Helper()
	data, err := os.ReadFile("../server/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != compressedSHA {
		t.Fatal("frozen compressed original fixture hash", name)
	}
	stream, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(stream)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != decodedSHA {
		t.Fatal("frozen decoded original fixture hash", name)
	}
	return decoded
}

func monsterDeathKind549D80(t *testing.T, entry uint32) server.MonsterDeadCallbackKind549D80 {
	t.Helper()
	switch entry {
	case 0x549d80:
		return server.MonsterDeadCallbackEmberDemon549D80
	case 0x549e00:
		return server.MonsterDeadCallbackDemon549E00
	case 0x549e70:
		return server.MonsterDeadCallbackImp549E70
	case 0x54a250:
		return server.MonsterDeadCallbackSpider54A250
	default:
		t.Fatal("unknown original death entry", entry)
		return 0
	}
}

func TestMonsterDeathExplosionOriginal549D80(t *testing.T) {
	TestMonsterDeathExplosionNativeLayout549D80(t)
	var inputs []monsterDeathInput549D80
	if err := json.Unmarshal(monsterDeathFixtureBytes549D80(t, "monster_death_explosion_independent_inputs_549d80.json.gz", "acec7841c6a433cfaa36972977c9b471049c1f5780e4605e5cf53521bb39caeb", "4d6206ffcec9fafe20371a51db93687fb60b7563ce99a0b552ed8edc34b4f452"), &inputs); err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 305 {
		t.Fatal("complete independent original input count", len(inputs))
	}
	for _, profile := range []string{"pc53_tozero", "pc53_nearest"} {
		t.Run(profile, func(t *testing.T) {
			var outputs []monsterDeathOutput549D80
			if err := json.Unmarshal(monsterDeathFixtureBytes549D80(t, "monster_death_explosion_native_original_549d80_"+profile+".json.gz", "951ae5cda4a689cd7b717c1479e6a201a6add9380edcbe5104df62008a90f54f", "3f3255d2dc3c6c276b735532ed704a67012709d1cdba8cf21fb3341cee21288c"), &outputs); err != nil {
				t.Fatal(err)
			}
			if len(outputs) != len(inputs) {
				t.Fatal("complete original output count", len(outputs))
			}
			for _, mode := range []string{"Go", "C", "ordinary"} {
				t.Run(mode, func(t *testing.T) {
					for i, in := range inputs {
						want := outputs[i]
						if want.Name != in.Name {
							t.Fatal("original input/output identities", i)
						}
						// Ordinary admission is deliberately separate from the raw
						// nil/unmapped pointer contract, including no-read services.
						if mode == "ordinary" && in.Unit < 0 {
							continue
						}
						t.Run(in.Name, func(t *testing.T) {
							got := monsterDeathRunOriginal549D80(t, in, mode)
							a, err := json.Marshal(got)
							if err != nil {
								t.Fatal(err)
							}
							b, err := json.Marshal(want)
							if err != nil {
								t.Fatal(err)
							}
							var normalizedA, normalizedB any
							if err := json.Unmarshal(a, &normalizedA); err != nil {
								t.Fatal(err)
							}
							if err := json.Unmarshal(b, &normalizedB); err != nil {
								t.Fatal(err)
							}
							if !reflect.DeepEqual(normalizedA, normalizedB) {
								t.Fatalf("complete original output differs\ngot %s\nwant%s", a, b)
							}
						})
					}
				})
			}
		})
	}
}

func monsterDeathRunOriginal549D80(t *testing.T, in monsterDeathInput549D80, mode string) monsterDeathOutput549D80 {
	t.Helper()
	base, free := alloc.Malloc(0x10000)
	defer free()
	buffer := unsafe.Slice((*byte)(base), 0x10000)
	canonical, expected := make([]byte, len(buffer)), make([]byte, len(buffer))
	for i := range buffer {
		buffer[i] = byte(i*17 + (i>>7)*29)
	}
	copy(canonical, buffer)
	copy(expected, buffer)
	objects := [2]*server.Object{(*server.Object)(unsafe.Add(base, 0x1000)), (*server.Object)(unsafe.Add(base, 0x3000))}
	offsets := [2]int{0x1000, 0x3000}
	positionOffset := int(unsafe.Offsetof(server.Object{}.PosVec))
	if unsafe.Sizeof(server.Object{}) >= 0x1000 {
		t.Fatal("fixture must hold each complete native object and separate guards")
	}
	setPositions := func(positions [2][2]uint32) {
		for i, value := range positions {
			objects[i].PosVec = types.Ptf(math.Float32frombits(value[0]), math.Float32frombits(value[1]))
			for k, bits := range value {
				binary.LittleEndian.PutUint32(canonical[offsets[i]+56+4*k:], bits)
				binary.LittleEndian.PutUint32(expected[offsets[i]+positionOffset+4*k:], bits)
			}
		}
	}
	setPositions(in.Positions)
	positions := func() (values [2][2]uint32) {
		for i, obj := range objects {
			values[i] = [2]uint32{math.Float32bits(obj.PosVec.X), math.Float32bits(obj.PosVec.Y)}
		}
		return
	}
	bad := monsterDeathOriginalBadObject549D80(-2)
	at := func(role int) *server.Object {
		if role < 0 {
			return monsterDeathOriginalBadObject549D80(role)
		}
		return objects[role]
	}
	role := func(obj *server.Object) int {
		switch obj {
		case nil:
			return -1
		case bad:
			return -2
		case objects[0]:
			return 0
		case objects[1]:
			return 1
		default:
			t.Fatalf("unknown full native object identity %p", obj)
			return -3
		}
	}
	for _, obj := range objects {
		if got := monsterDeathOriginalControls549D80(obj); got != [3]uintptr{uintptr(unsafe.Pointer(obj)), uintptr(unsafe.Pointer(&obj.PosVec.X)), uintptr(unsafe.Pointer(&obj.PosVec.Y))} {
			t.Fatal("independent C native object/point identities", got)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			t.Fatal("complete native original fixture must be above4GiB")
		}
	}
	got := monsterDeathOutput549D80{Name: in.Name, Events: make([]map[string]any, 0)}
	flagResult, index := in.FlagResult, 0
	var call func(uint32, *server.Object) int32
	world := func(kind string, fields map[string]any, owner *server.Object, hasPoint bool, supplied *types.Pointf) uint32 {
		number := index
		index++
		event := map[string]any{"kind": kind, "index": number, "positions": positions()}
		for k, value := range fields {
			event[k] = value
		}
		if hasPoint && in.ReadPositions {
			if identity := role(owner); identity < 0 {
				got.Events = append(got.Events, event)
				got.Fault = map[string]any{"kind": "declared-world-point-read", "service": kind, "index": number, "point_unit": identity}
				panic(monsterDeathWorldReadFault549D80{})
			}
			value := owner.PosVec
			if supplied != nil {
				value = *supplied
			}
			event["point"] = [2]uint32{math.Float32bits(value.X), math.Float32bits(value.Y)}
		}
		got.Events = append(got.Events, event)
		result := uint32(0xa5a50000) + uint32(number)
		if kind == "flag" {
			result = flagResult
		}
		step := in.Script[strconv.Itoa(number)]
		if step.Positions != nil {
			setPositions(*step.Positions)
		}
		if step.FlagResult != nil {
			flagResult = *step.FlagResult
		}
		if nested := step.Reenter; nested != nil {
			got.Events = append(got.Events, map[string]any{"kind": "nested", "index": number, "entry": nested.Entry, "unit": nested.Unit, "positions": positions()})
			if result := call(nested.Entry, at(nested.Unit)); result != 1 {
				t.Fatal("nested original entry return", result)
			}
			got.Events = append(got.Events, map[string]any{"kind": "nested-return", "index": number, "result": 1, "positions": positions()})
		}
		if step.AfterPositions != nil {
			setPositions(*step.AfterPositions)
		}
		got.Events = append(got.Events, map[string]any{"kind": kind + "-return", "index": number, "result": result, "positions": positions()})
		return result
	}
	push := func(p *server.Object, outer, inner, force float32, source *server.Object, callback, arg uintptr, supplied *types.Pointf) uint32 {
		return world("push", map[string]any{"unit": role(source), "point_unit": role(p), "outer": math.Float32bits(outer), "inner": math.Float32bits(inner), "force": math.Float32bits(force), "callback": callback, "callback_arg": arg}, p, true, supplied)
	}
	damage := func(p *server.Object, outer, inner float32, amount, typ uint32, source, excluded *server.Object, supplied *types.Pointf) uint32 {
		return world("damage", map[string]any{"unit": role(source), "point_unit": role(p), "outer": math.Float32bits(outer), "inner": math.Float32bits(inner), "damage": amount, "damage_type": typ, "exclude": role(excluded)}, p, true, supplied)
	}
	spark := func(p *server.Object, code uint32, supplied *types.Pointf) uint32 {
		return world("spark", map[string]any{"point_unit": role(p), "code": code}, p, true, supplied)
	}
	point := func(code uint32, p *server.Object, supplied *types.Pointf) uint32 {
		return world("point", map[string]any{"point_unit": role(p), "code": code}, p, true, supplied)
	}
	runtime := server.MonsterDeathExplosionRuntime549D80{
		GameFlag: func(mask uint32) uint32 { return world("flag", map[string]any{"mask": mask}, nil, false, nil) },
		Push: func(p *server.Object, a, b, f float32, source *server.Object, callback, arg uintptr) uint32 {
			return push(p, a, b, f, source, callback, arg, nil)
		},
		Damage: func(p *server.Object, a, b float32, d, typ uint32, source, excluded *server.Object) uint32 {
			return damage(p, a, b, d, typ, source, excluded, nil)
		},
		Spark: func(p *server.Object, code uint32) uint32 { return spark(p, code, nil) },
		Point: func(code uint32, p *server.Object) uint32 { return point(code, p, nil) },
		Audio: func(id uint32, unit *server.Object, kind, arg uint32) uint32 {
			return world("audio", map[string]any{"sound": id, "unit": role(unit), "kind_arg": kind, "arg": arg}, nil, false, nil)
		},
		Delete: func(unit *server.Object) uint32 {
			return world("delete", map[string]any{"unit": role(unit)}, nil, false, nil)
		},
	}
	oldOverride := monsterDeathOriginalRuntimeOverride549D80
	monsterDeathOriginalRuntimeOverride549D80 = &runtime
	defer func() { monsterDeathOriginalRuntimeOverride549D80 = oldOverride }()
	call = func(entry uint32, unit *server.Object) int32 {
		switch mode {
		case "C":
			if monsterDeathOriginalABI549D80(entry) != 64 {
				t.Fatal("raw C native ABI is not implemented", entry)
			}
			return monsterDeathOriginalC549D80(entry, unit)
		case "Go":
			impl, ok := any(unit).(monsterDeathRawContract549D80)
			if !ok {
				t.Fatal("raw Go original contract is not implemented")
			}
			return impl.MonsterDeathExplosionNative549D80(entry, runtime)
		case "ordinary":
			ordinary := server.MonsterDeadCallbackRuntime549D80{
				CoopMode: func() bool { return runtime.GameFlag(2048) != 0 },
				PushUnits: func(value types.Pointf, a, b, f float32, source *server.Object) {
					push(source, a, b, f, source, 0, 0, &value)
				},
				DamageUnits: func(value types.Pointf, a, b float32, d int, typ object.DamageType, source *server.Object) {
					damage(source, a, b, uint32(d), uint32(typ), source, nil, &value)
				},
				SparkExplosion: func(value types.Pointf, code byte) { spark(unit, uint32(code), &value) },
				PointFX:        func(code netmsg.Op, value types.Pointf) { point(uint32(code), unit, &value) },
				Audio:          func(id sound.ID, source *server.Object) { runtime.Audio(uint32(id), source, 0, 0) },
				DelayedDelete:  func(source *server.Object) { runtime.Delete(source) },
			}
			if !server.MonsterDeadCallbackNative549D80(unit, monsterDeathKind549D80(t, entry), ordinary) {
				t.Fatal("ordinary valid original fixture was not admitted")
			}
			return 1
		default:
			t.Fatal("unknown original consumer mode", mode)
			return 0
		}
	}
	func() {
		defer func() {
			if value := recover(); value != nil {
				if _, ok := value.(monsterDeathWorldReadFault549D80); !ok || got.Fault == nil {
					t.Fatalf("unexpected native panic before declared world fault: %T %v", value, value)
				}
			}
		}()
		result := call(in.Entry, at(in.Unit))
		got.Result = &result
	}()
	if !bytes.Equal(buffer, expected) {
		for i, value := range buffer {
			if value != expected[i] {
				t.Fatalf("complete native64KiB allocation/guards changed outside declared writes at %x: %02x/%02x", i, value, expected[i])
			}
		}
	}
	got.Positions = positions()
	hash := sha256.Sum256(canonical)
	got.HeapSHA256 = hex.EncodeToString(hash[:])
	return got
}
