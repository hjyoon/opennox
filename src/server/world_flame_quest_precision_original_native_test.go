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
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type worldFlameQuestOriginalCase struct {
	Name           string       `json:"name"`
	Player         bool         `json:"player"`
	Raw            int32        `json:"raw"`
	Scale          uint32       `json:"quest_scale_bits32"`
	FlameClass     object.Class `json:"flame_class"`
	FlameType      uint16       `json:"flame_type"`
	SourceProfile  string       `json:"source_profile"`
	SourceClass    object.Class `json:"source_class"`
	Quest          bool         `json:"quest"`
	God            bool         `json:"god"`
	Carry          uint32       `json:"carry_bits32"`
	Absorption     uint32       `json:"absorption_bits32"`
	ExpectedMarker uint32       `json:"expected_marker"`
	ExpectedKind   uint32       `json:"expected_marker_type"`
	ExpectedCarry  uint32       `json:"expected_carry_bits32"`
	ExpectedDamage []uint32     `json:"expected_damage_bits32"`
}

func observeWorldFlameQuestOriginalNative(t *testing.T, c worldFlameQuestOriginalCase, full bool) map[string]any {
	t.Helper()
	var frees []func()
	defer func() {
		for i := len(frees) - 1; i >= 0; i-- {
			frees[i]()
		}
	}()
	target, f := alloc.New(Object{})
	frees = append(frees, f)
	player, f := alloc.New(Player{})
	frees = append(frees, f)
	pu, f := alloc.New(PlayerUpdateData{})
	frees = append(frees, f)
	mu, f := alloc.New(MonsterUpdateData{})
	frees = append(frees, f)
	hp, f := alloc.New(HealthData{})
	frees = append(frees, f)
	source, f := alloc.New(Object{})
	frees = append(frees, f)
	sourcePlayer, f := alloc.New(Player{})
	frees = append(frees, f)
	sourcePU, f := alloc.New(PlayerUpdateData{})
	frees = append(frees, f)
	sourceMU, f := alloc.New(MonsterUpdateData{})
	frees = append(frees, f)
	flame, f := alloc.New(Object{})
	frees = append(frees, f)
	*player = Player{}
	*pu = PlayerUpdateData{Player: player, State: PlayerState13, Field57: c.Absorption, Field21: c.Carry, Field76: 99, Field75: 77}
	*mu = MonsterUpdateData{Field518: c.Absorption, Field1: c.Carry, Field547: 99, Field546: 77}
	*hp = HealthData{Cur: 20, Max: 20, Field2: 20}
	*target = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(mu), HealthData: hp}
	marker, kind, carry := &mu.Field547, &mu.Field546, &mu.Field1
	if c.Player {
		target.ObjClass = object.ClassPlayer
		target.ObjSubClass = 0
		target.UpdateData = unsafe.Pointer(pu)
		marker, kind, carry = &pu.Field76, &pu.Field75, &pu.Field21
	}
	*sourcePlayer = Player{}
	*sourcePU = PlayerUpdateData{Player: sourcePlayer}
	*sourceMU = MonsterUpdateData{}
	*source = Object{TypeInd: 0x6701, ObjClass: c.SourceClass}
	if c.SourceClass.Has(object.ClassPlayer) {
		source.UpdateData = unsafe.Pointer(sourcePU)
	}
	if c.SourceClass.Has(object.ClassMonster) {
		source.UpdateData = unsafe.Pointer(sourceMU)
		source.ObjSubClass = 0x11012
	}
	*flame = Object{TypeInd: c.FlameType, ObjClass: c.FlameClass, PosVec: types.Ptf(100, 0), PrevPos: types.Ptf(100, 0)}
	attacker := source
	switch c.SourceProfile {
	case "nil":
		attacker = nil
	case "self":
		attacker = flame
	}
	frames := [][]byte{
		unsafe.Slice((*byte)(unsafe.Pointer(target)), int(unsafe.Sizeof(*target))),
		unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player))),
		unsafe.Slice((*byte)(unsafe.Pointer(pu)), int(unsafe.Sizeof(*pu))),
		unsafe.Slice((*byte)(unsafe.Pointer(mu)), int(unsafe.Sizeof(*mu))),
		unsafe.Slice((*byte)(unsafe.Pointer(hp)), int(unsafe.Sizeof(*hp))),
		unsafe.Slice((*byte)(unsafe.Pointer(source)), int(unsafe.Sizeof(*source))),
		unsafe.Slice((*byte)(unsafe.Pointer(sourcePlayer)), int(unsafe.Sizeof(*sourcePlayer))),
		unsafe.Slice((*byte)(unsafe.Pointer(sourcePU)), int(unsafe.Sizeof(*sourcePU))),
		unsafe.Slice((*byte)(unsafe.Pointer(sourceMU)), int(unsafe.Sizeof(*sourceMU))),
		unsafe.Slice((*byte)(unsafe.Pointer(flame)), int(unsafe.Sizeof(*flame))),
	}
	before := make([][]byte, len(frames))
	for i, b := range frames {
		if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(unsafe.Pointer(&b[0])) <= 0xffffffff {
			t.Fatal("expected C-owned world flame input frames above4GiB")
		}
		before[i] = append([]byte(nil), b...)
	}
	damage := []uint32{}
	unsupported := []string{}
	r := PlayerDamageRuntime4E17B0{
		GodMode: func() bool { return c.God }, QuestMode: func() bool { return c.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(c.Scale) },
		BlockSourceExcluded: func(v *Object) bool {
			if v != flame {
				t.Fatal("changed original flame exclusion identity")
			}
			return false
		},
		BlockSourceOnlyExcluded: func(v *Object) bool { t.Fatal("world flame used nil-weapon exclusion"); return false },
		BlockDirection: func(v *Object, pos types.Pointf) bool {
			if v != target || pos != flame.PrevPos {
				t.Fatal("changed original direction input")
			}
			return false
		},
		DefaultDamage: func(v, s, w *Object, n int32, typ object.DamageType) bool {
			if v != target || s != attacker || w != flame || typ != object.DamageFlame {
				t.Fatal("changed original DefaultDamage identity/type")
			}
			damage = append(damage, uint32(n))
			return true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			unsupported = append(unsupported, reason)
		},
	}
	if full {
		r.CoopMode = func() bool { return false }
	}
	handled, result := PlayerDamageNative4E17B0(target, attacker, flame, c.Raw, object.DamageFlame, r)
	// Only the two cached hit-marker fields are outputs in the original frame.
	index := 3
	off, ko := unsafe.Offsetof(mu.Field547), unsafe.Offsetof(mu.Field546)
	if c.Player {
		index = 2
		off, ko = unsafe.Offsetof(pu.Field76), unsafe.Offsetof(pu.Field75)
	}
	copy(before[index][off:off+4], frames[index][off:off+4])
	copy(before[index][ko:ko+4], frames[index][ko:ko+4])
	preserved := true
	for i, b := range frames {
		preserved = preserved && bytes.Equal(before[i], b)
	}
	match := handled && result && len(unsupported) == 0 && reflect.DeepEqual(damage, c.ExpectedDamage) && *marker == c.ExpectedMarker && *kind == c.ExpectedKind && *carry == c.ExpectedCarry && preserved
	return map[string]any{"name": c.Name, "full_runtime": full, "handled": handled, "result": result, "damage": damage, "marker": *marker, "marker_type": *kind, "carry_bits32": *carry, "all_native_nonoutput_frames_preserved": preserved, "unsupported": unsupported, "matches_original": match}
}

func testWorldFlameQuestOriginalNative(t *testing.T, full bool) {
	b, err := os.ReadFile("testdata/world_flame_quest_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []worldFlameQuestOriginalCase `json:"cases"`
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 1096 {
		t.Fatal("incomplete independently executed original worldFLAME contexts")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := observeWorldFlameQuestOriginalNative(t, c, full)
			if !got["matches_original"].(bool) {
				t.Fatalf("actual native world flame Quest mismatch=%v original damage=%v", got, c.ExpectedDamage)
			}
		})
	}
}
func TestPlayerDamageWorldFlameQuestPrecisionOriginal4E17B0Native(t *testing.T) {
	testWorldFlameQuestOriginalNative(t, true)
}
func TestPlayerDamageWorldFlamePartialQuestPrecisionOriginal4E17B0Native(t *testing.T) {
	testWorldFlameQuestOriginalNative(t, false)
}

// These expected Quest amounts are complete original-executable outputs for
// exact historical raw/scale inputs. No native rounding computes the oracle.
var worldFlameQuestHistoricalFixture struct {
	once  sync.Once
	cases []worldFlameQuestOriginalCase
	err   error
}

func worldFlameQuestHistoricalExpected(t *testing.T, player bool, raw int32, scale float32) int32 {
	t.Helper()
	worldFlameQuestHistoricalFixture.once.Do(func() {
		b, err := os.ReadFile("testdata/world_flame_quest_historical_original.json")
		if err != nil {
			worldFlameQuestHistoricalFixture.err = err
			return
		}
		var f struct {
			Cases []worldFlameQuestOriginalCase `json:"cases"`
		}
		worldFlameQuestHistoricalFixture.err = json.Unmarshal(b, &f)
		worldFlameQuestHistoricalFixture.cases = f.Cases
	})
	if err := worldFlameQuestHistoricalFixture.err; err != nil {
		t.Fatal(err)
	}
	for _, c := range worldFlameQuestHistoricalFixture.cases {
		if c.Player == player && c.Raw == raw && c.Scale == math.Float32bits(scale) && c.Quest == (scale >= 0) {
			if len(c.ExpectedDamage) != 1 {
				t.Fatal("missing exact original Quest output")
			}
			return int32(c.ExpectedDamage[0])
		}
	}
	t.Fatalf("missing original world flame Quest input raw=%d scale=%08x", raw, math.Float32bits(scale))
	return 0
}
