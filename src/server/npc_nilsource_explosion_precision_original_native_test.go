package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type nilSourceNPCExplosionOriginalCase struct {
	Name           string            `json:"name"`
	Player         bool              `json:"player"`
	Type           object.DamageType `json:"damage_type"`
	Raw            int32             `json:"raw"`
	Armor          uint32            `json:"absorption_bits32"`
	Carry          uint32            `json:"carry_bits32"`
	Quest          bool              `json:"quest"`
	Scale          uint32            `json:"quest_scale_bits32"`
	God            bool              `json:"god"`
	SourceClass    object.Class      `json:"source_class"`
	SourceType     uint16            `json:"source_type"`
	ExpectedWear   uint32            `json:"expected_wear_bits32"`
	ExpectedCarry  uint32            `json:"expected_carry_bits32"`
	ExpectedDamage []uint32          `json:"expected_damage_bits32"`
}

func nilSourceNPCExplosionOriginalFrame(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), int(n))
}

func observeNilSourceNPCExplosionOriginalNative(t *testing.T, c nilSourceNPCExplosionOriginalCase, partial bool) map[string]any {
	t.Helper()
	var frees []func()
	defer func() {
		for i := len(frees) - 1; i >= 0; i-- {
			frees[i]()
		}
	}()
	v, f := alloc.New(Object{})
	frees = append(frees, f)
	u, f := alloc.New(MonsterUpdateData{})
	frees = append(frees, f)
	info, f := alloc.New(Player{})
	frees = append(frees, f)
	*info = Player{}
	hp, f := alloc.New(HealthData{})
	frees = append(frees, f)
	*hp = HealthData{Cur: 200, Max: 200, Field2: 200}
	*u = MonsterUpdateData{Field518: c.Armor, Field1: c.Carry, Field547: 99, Field546: 77}
	*v = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(u), HealthData: hp}
	source, f := alloc.New(Object{})
	frees = append(frees, f)
	*source = Object{TypeInd: c.SourceType, ObjClass: c.SourceClass, PosVec: types.Pointf{X: 100}, PrevPos: types.Pointf{X: 100}}
	sourceInfo, f := alloc.New(Player{})
	frees = append(frees, f)
	*sourceInfo = Player{}
	var sourceUpdate []byte
	if c.SourceClass.Has(object.ClassMonster) {
		update, f := alloc.New(MonsterUpdateData{})
		frees = append(frees, f)
		*update = MonsterUpdateData{}
		source.UpdateData = unsafe.Pointer(update)
		sourceUpdate = nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(update), unsafe.Sizeof(*update))
	} else {
		update, f := alloc.New(PlayerUpdateData{})
		frees = append(frees, f)
		*update = PlayerUpdateData{Player: sourceInfo}
		if c.SourceClass.Has(object.ClassPlayer) {
			source.UpdateData = unsafe.Pointer(update)
		}
		sourceUpdate = nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(update), unsafe.Sizeof(*update))
	}
	frames := [][]byte{
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(v), unsafe.Sizeof(*v)),
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u)),
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(info), unsafe.Sizeof(*info)),
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(hp), unsafe.Sizeof(*hp)),
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(source), unsafe.Sizeof(*source)),
		sourceUpdate,
		nilSourceNPCExplosionOriginalFrame(unsafe.Pointer(sourceInfo), unsafe.Sizeof(*sourceInfo)),
	}
	before := make([][]byte, len(frames))
	for i, frame := range frames {
		if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(unsafe.Pointer(&frame[0])) <= 0xffffffff {
			t.Fatal("expected C-owned native frames above4GiB")
		}
		before[i] = append([]byte(nil), frame...)
	}
	forward := []uint32{}
	unsupported := []string{}
	exclusions, directions := 0, 0
	r := PlayerDamageRuntime4E17B0{
		CoopMode: func() bool { return false }, GodMode: func() bool { return c.God }, QuestMode: func() bool { return c.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(c.Scale) },
		BlockSourceOnlyExcluded: func(got *Object) bool {
			if got != source {
				t.Fatal("changed source exclusion identity")
			}
			exclusions++
			return false
		},
		BlockDirection: func(got *Object, point types.Pointf) bool {
			if got != v || point != source.PrevPos {
				t.Fatal("changed direction service arguments")
			}
			directions++
			return false // Declared original excluded integer service result.
		},
		DefaultDamage: func(got, src, weapon *Object, amount int32, typ object.DamageType) bool {
			if got != v || src != nil || weapon != nil || typ != c.Type {
				t.Fatal("changed default damage identities or type")
			}
			forward = append(forward, uint32(amount))
			return true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			unsupported = append(unsupported, reason)
		},
	}
	if partial {
		r.CoopMode = nil // Existing partial runtime enters the dedicated tail.
	}
	handled, result := PlayerDamageNative4E17B0(v, nil, nil, c.Raw, c.Type, r)
	for _, off := range []uintptr{unsafe.Offsetof(u.Field1), unsafe.Offsetof(u.Field547), unsafe.Offsetof(u.Field546)} {
		copy(before[1][off:off+4], frames[1][off:off+4])
	}
	preserved := true
	for i, frame := range frames {
		preserved = preserved && bytes.Equal(before[i], frame)
	}
	match := handled && result && len(unsupported) == 0 && reflect.DeepEqual(forward, c.ExpectedDamage) && u.Field1 == c.ExpectedCarry && u.Field547 == 2 && u.Field546 == uint32(c.Type) && exclusions == 0 && directions == 0 && preserved
	return map[string]any{"name": c.Name, "handled": handled, "result": result, "forward_bits32": forward, "carry_bits32": u.Field1, "unsupported": unsupported, "exclusions": exclusions, "directions": directions, "all_native_nonoutput_frames_preserved": preserved, "matches_original": match}
}

func testNPCNilSourceExplosionPrecisionOriginal(t *testing.T, partial bool) {
	t.Helper()
	b, err := os.ReadFile("testdata/npc_nilsource_explosion_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []nilSourceNPCExplosionOriginalCase `json:"cases"`
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 473 {
		t.Fatal("incomplete independently executed original nil-source explosion contexts")
	}
	for _, c := range fixture.Cases {
		if c.Player {
			t.Fatal("NPC audit requires original NPC target inputs")
		}
		t.Run(c.Name, func(t *testing.T) {
			got := observeNilSourceNPCExplosionOriginalNative(t, c, partial)
			if !got["matches_original"].(bool) {
				t.Fatalf("actual native NPC weaponless explosion mismatch=%v original damage=%v carry=%08x", got, c.ExpectedDamage, c.ExpectedCarry)
			}
		})
	}
}

func TestPlayerDamageNPCNilSourceExplosionPrecisionOriginal4E17B0Native(t *testing.T) {
	testNPCNilSourceExplosionPrecisionOriginal(t, false)
}

func TestPlayerDamageNPCNilSourceExplosionPartialPrecisionOriginal4E17B0Native(t *testing.T) {
	testNPCNilSourceExplosionPrecisionOriginal(t, true)
}

