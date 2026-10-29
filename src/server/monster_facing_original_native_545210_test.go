package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func facingOriginalAllocation545210[T comparable](t *testing.T) *T {
	t.Helper()
	var zero T
	p, free := alloc.New(zero)
	t.Cleanup(free)
	if uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatal("native comparison allocation must exceed4GiB")
	}
	return p
}

// All1348 numerical/action contexts compare complete observable original
// results under startup PC53/chop. Four additional original cancel-thunk
// return controls stay in the fixture: their non-numerical pointer forwarding
// is checked by the frozen actual C/native dispatcher probes. PopAction's body
// is a declared service here. Internal normalized-vector scratch stays in the
// original fixture and is not an observable output of these native functions.
func TestMonsterFacingOriginalNative545210(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native C-owned comparison above4GiB")
	}
	path := os.Getenv("NOX_MONSTER_FACING_ORACLE_FIXTURE")
	if path == "" {
		path = "testdata/monster_facing_native_original_545210.json"
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
	if fixture.Control != 0xe3f || len(fixture.Cases) != 1352 {
		t.Fatal("sealed startup original corpus")
	}
	u := func(c map[string]any, key string) uint32 { return uint32(c[key].(float64)) }
	pt := func(value any) types.Pointf {
		p := value.([]any)
		return types.Ptf(math.Float32frombits(uint32(p[0].(float64))), math.Float32frombits(uint32(p[1].(float64))))
	}
	pointBits := func(p types.Pointf) []uint32 { return []uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	canonical := func(value any) any {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var copy map[string]any
		if err := json.Unmarshal(encoded, &copy); err != nil {
			t.Fatal(err)
		}
		return copy
	}
	compared, thunks := 0, 0
	for _, row := range fixture.Cases {
		c := row.Input
		if c["function"] == "pop" {
			thunks++
			continue
		}
		compared++
		t.Run(c["name"].(string), func(t *testing.T) {
			type guardedObject struct {
				Before [32]byte
				Object Object
				After  [32]byte
			}
			type guardedUpdate struct {
				Before [32]byte
				Update MonsterUpdateData
				After  [32]byte
			}
			f := facingOriginalAllocation545210[guardedObject](t)
			g := facingOriginalAllocation545210[guardedUpdate](t)
			target := facingOriginalAllocation545210[Object](t)
			for i := range f.Before {
				f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
				g.Before[i], g.After[i] = byte(0x73+i), byte(0xc5-i)
			}
			unit := &f.Object
			*unit = Object{ObjClass: object.ClassMonster, PosVec: pt(c["pos"]), Direction1: Dir16(u(c, "direction1")), Direction2: Dir16(u(c, "direction2")), UpdateData: unsafe.Pointer(&g.Update)}
			target.PosVec = pt(c["target_pos"])
			function := c["function"].(string)
			actions := map[string]ai.ActionType{"location": ai.ACTION_FACE_LOCATION, "point-helper": ai.ACTION_FACE_LOCATION, "object": ai.ACTION_FACE_OBJECT, "angle": ai.ACTION_FACE_ANGLE, "set-angle": ai.ACTION_SET_ANGLE, "facing-helper": ai.ACTION_FACE_ANGLE}
			index := int(u(c, "head_index"))
			g.Update.AIStackInd = int8(index)
			head := &g.Update.AIStack[index]
			head.Action = uint32(actions[function])
			switch function {
			case "object":
				if !c["target_nil"].(bool) {
					head.Args[0] = uintptr(unsafe.Pointer(target))
				}
				head.Args[1] = 0x12345678
			case "angle", "set-angle", "facing-helper":
				head.Args[0], head.Args[1] = uintptr(u(c, "angle")), 0x12345678
			default:
				head.SetArgs(pt(c["target_pos"]))
			}
			beforeObject := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))...)
			beforeUpdate := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(g)), int(unsafe.Sizeof(*g)))...)
			beforeTarget := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(target)), int(unsafe.Sizeof(*target)))...)
			state := func() map[string]any {
				i := int(g.Update.AIStackInd)
				h := &g.Update.AIStack[i]
				arg0 := int64(h.ArgU32(0))
				if function == "object" {
					p := h.ArgObj(0)
					switch p {
					case nil:
						arg0 = -1
					case target:
						arg0 = 1
					default:
						t.Fatal("target outside sealed native identity set")
					}
				}
				return map[string]any{"pos": pointBits(unit.PosVec), "target_pos": pointBits(target.PosVec),
					"direction1": uint16(unit.Direction1), "direction2": uint16(unit.Direction2), "head_index": i, "head_action": h.Action, "head_arg0": arg0, "head_arg1": h.ArgU32(1)}
			}
			events := make([]map[string]any, 0)
			pop := func() int {
				events = append(events, map[string]any{"kind": "pop", "args": []int{0}, "state": state()})
				return int(u(c, "pop_return"))
			}
			var result int
			switch function {
			case "location":
				result = monsterActionFaceLocation545210(unit, pop)
			case "point-helper":
				result = monsterActionFacePoint545240(unit, pt(c["target_pos"]), pop)
			case "object":
				result = monsterActionFaceObject545300(unit, pop)
			case "angle":
				result = monsterActionFaceAngle545340(unit, pop)
			case "set-angle":
				result = monsterActionSetAngle5453E0(unit, pop)
			case "facing-helper":
				if monsterFacingDirection534120(unit, pt(c["vector"])) {
					result = 1
				}
			default:
				t.Fatal("unknown sealed original function")
			}
			actual := map[string]any{"return_value": uint32(result), "events": events, "state": state()}
			if !reflect.DeepEqual(canonical(actual), canonical(row.Result)) {
				a, _ := json.Marshal(actual)
				w, _ := json.Marshal(row.Result)
				t.Fatalf("native=%s\noriginal=%s", a, w)
			}
			allowed := make([]bool, len(beforeObject))
			for _, field := range []struct{ off, n uintptr }{{unsafe.Offsetof(unit.Direction1), unsafe.Sizeof(unit.Direction1)}, {unsafe.Offsetof(unit.Direction2), unsafe.Sizeof(unit.Direction2)}} {
				for i := field.off; i < field.off+field.n; i++ {
					allowed[i] = true
				}
			}
			for i, v := range unsafe.Slice((*byte)(unsafe.Pointer(unit)), len(beforeObject)) {
				if v != beforeObject[i] && !allowed[i] {
					t.Fatalf("unrelated object byte changed at%d", i)
				}
			}
			if !reflect.DeepEqual(unsafe.Slice((*byte)(unsafe.Pointer(g)), len(beforeUpdate)), beforeUpdate) || !reflect.DeepEqual(unsafe.Slice((*byte)(unsafe.Pointer(target)), len(beforeTarget)), beforeTarget) {
				t.Fatal("native update/target canary changed")
			}
			for i := range f.Before {
				if f.Before[i] != byte(0x51+i) || f.After[i] != byte(0xe7-i) {
					t.Fatal("native object canary")
				}
			}
		})
	}
	if compared != 1348 || thunks != 4 {
		t.Fatalf("original context partition=%d/%d", compared, thunks)
	}
}
