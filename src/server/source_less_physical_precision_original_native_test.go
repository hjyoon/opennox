package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type sourceLessPhysicalOriginalInput struct {
	Name           string            `json:"name"`
	Player         bool              `json:"player"`
	Type           object.DamageType `json:"damage_type"`
	Raw            int32             `json:"raw"`
	Armor          uint32            `json:"absorption_bits32"`
	Carry          uint32            `json:"carry_bits32"`
	Quest          bool              `json:"quest"`
	Scale          uint32            `json:"quest_scale_bits32"`
	God            bool              `json:"god"`
	ExpectedCarry  uint32            `json:"expected_carry_bits32"`
	ExpectedDamage []uint32          `json:"expected_damage_bits32"`
}

func sourceLessPhysicalOriginalFrame(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), int(n))
}

func observeSourceLessPhysicalOriginalNative(c sourceLessPhysicalOriginalInput) map[string]any {
	v, freeV := alloc.New(Object{})
	defer freeV()
	hp, freeHP := alloc.New(HealthData{})
	defer freeHP()
	*hp = HealthData{Cur: 200, Max: 200, Field2: 200}
	*v = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, HealthData: hp}
	var carry, marker, kind *uint32
	var updateBefore []byte
	var update []byte
	var offsets []uintptr
	if c.Player {
		u, freeU := alloc.New(PlayerUpdateData{})
		defer freeU()
		info, freeInfo := alloc.New(Player{})
		defer freeInfo()
		*info = Player{}
		*u = PlayerUpdateData{Player: info, State: PlayerState13, Field57: c.Armor, Field21: c.Carry, Field76: 99, Field75: 77}
		v.ObjClass, v.ObjSubClass, v.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(u)
		carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
		update = sourceLessPhysicalOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
		offsets = []uintptr{unsafe.Offsetof(u.Field21), unsafe.Offsetof(u.Field76), unsafe.Offsetof(u.Field75)}
	} else {
		u, freeU := alloc.New(MonsterUpdateData{})
		defer freeU()
		*u = MonsterUpdateData{Field518: c.Armor, Field1: c.Carry, Field547: 99, Field546: 77}
		v.UpdateData = unsafe.Pointer(u)
		carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
		update = sourceLessPhysicalOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
		offsets = []uintptr{unsafe.Offsetof(u.Field1), unsafe.Offsetof(u.Field547), unsafe.Offsetof(u.Field546)}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(v)) <= 0xffffffff || uintptr(v.UpdateData) <= 0xffffffff) {
		panic("expected C-owned high native pointers")
	}
	updateBefore = append([]byte(nil), update...)
	objectBefore := append([]byte(nil), sourceLessPhysicalOriginalFrame(unsafe.Pointer(v), unsafe.Sizeof(*v))...)
	healthBefore := *hp
	forward := []uint32{}
	unsupported := []string{}
	r := PlayerDamageRuntime4E17B0{
		CoopMode: func() bool { return false }, GodMode: func() bool { return c.God }, QuestMode: func() bool { return c.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(c.Scale) },
		DefaultDamage: func(got, source, weapon *Object, amount int32, typ object.DamageType) bool {
			if got != v || source != nil || weapon != nil || typ != c.Type {
				panic("changed native identities")
			}
			forward = append(forward, uint32(amount))
			return true
		},
		Unsupported: func(s string, _, _, _ *Object, _ int32, _ object.DamageType) {
			unsupported = append(unsupported, s)
		},
	}
	handled, result := PlayerDamageNative4E17B0(v, nil, nil, c.Raw, c.Type, r)
	for _, off := range offsets {
		copy(updateBefore[off:off+4], update[off:off+4])
	}
	preserved := bytes.Equal(objectBefore, sourceLessPhysicalOriginalFrame(unsafe.Pointer(v), unsafe.Sizeof(*v))) && bytes.Equal(updateBefore, update) && healthBefore == *hp
	match := handled && result && len(unsupported) == 0 && reflect.DeepEqual(forward, c.ExpectedDamage) && *carry == c.ExpectedCarry && *marker == 2 && *kind == uint32(c.Type) && preserved
	return map[string]any{"name": c.Name, "player": c.Player, "damage_type": c.Type, "handled": handled, "result": result, "forward_bits32": forward, "carry_bits32": *carry, "marker": *marker, "marker_type": *kind, "unsupported": unsupported, "all_native_nonoutput_frame_bytes_preserved": preserved, "matches_original_startup": match}
}

func TestSourceLessPhysicalPrecisionOriginal4E17B0Native(t *testing.T) {
	b, err := os.ReadFile("testdata/source_less_physical_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []sourceLessPhysicalOriginalInput `json:"cases"`
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 6622 {
		t.Fatal("incomplete independently executed original physical contexts")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := observeSourceLessPhysicalOriginalNative(c)
			if !got["matches_original_startup"].(bool) {
				t.Fatalf("actual native source-less physical mismatch=%v original damage=%v carry=%08x", got, c.ExpectedDamage, c.ExpectedCarry)
			}
		})
	}
}

// Historical expectations are captured independently from original GAME.EXE.
// The old test input loops, callback effects, and assertions remain in place.
type sourceLessPhysicalHistoricalInput struct {
	sourceLessPhysicalOriginalInput
	SourceFixture string `json:"source_fixture"`
	ExpectedWear  uint32 `json:"expected_wear_bits32"`
}

var sourceLessPhysicalHistoricalRows4E17B0 = sync.OnceValues(func() ([]sourceLessPhysicalHistoricalInput, error) {
	b, err := os.ReadFile("testdata/source_less_physical_historical_original.json")
	if err != nil {
		return nil, err
	}
	var f struct {
		Cases []sourceLessPhysicalHistoricalInput `json:"cases"`
	}
	err = json.Unmarshal(b, &f)
	return f.Cases, err
})

func sourceLessPhysicalHistoricalCase4E17B0(t *testing.T, family string, player bool, typ object.DamageType, raw int32, armor, carry float32, quest bool, scale float32) sourceLessPhysicalHistoricalInput {
	t.Helper()
	rows, err := sourceLessPhysicalHistoricalRows4E17B0()
	if err != nil || len(rows) != 2163 {
		t.Fatalf("incomplete original historical physical expectations: %v", err)
	}
	for _, c := range rows {
		if c.SourceFixture == family && c.Player == player && c.Type == typ && c.Raw == raw && c.Armor == math.Float32bits(armor) && c.Carry == math.Float32bits(carry) && c.Quest == quest && c.Scale == math.Float32bits(scale) && len(c.ExpectedDamage) == 1 {
			return c
		}
	}
	t.Fatalf("missing original historical physical input: %s %t %d %d %g %g %t %g", family, player, typ, raw, armor, carry, quest, scale)
	return sourceLessPhysicalHistoricalInput{}
}

func sourceLessPhysicalHistoricalNamed4E17B0(t *testing.T, name string) sourceLessPhysicalHistoricalInput {
	t.Helper()
	rows, err := sourceLessPhysicalHistoricalRows4E17B0()
	if err != nil || len(rows) != 2163 {
		t.Fatalf("incomplete original historical physical expectations: %v", err)
	}
	for _, c := range rows {
		if c.Name == name && len(c.ExpectedDamage) == 1 {
			return c
		}
	}
	t.Fatalf("missing original historical physical case: %s", name)
	return sourceLessPhysicalHistoricalInput{}
}
