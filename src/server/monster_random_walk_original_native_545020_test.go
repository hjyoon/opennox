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

type randomWalkOriginalState545020 struct {
	Update      int       `json:"update"`
	Class       uint32    `json:"cls"`
	Subclass    uint32    `json:"subclass"`
	Flags       uint32    `json:"flags"`
	Position    [2]uint32 `json:"position"`
	Directions  [2]uint16 `json:"directions"`
	Force       [2]uint32 `json:"force"`
	Speed       uint32    `json:"speed"`
	Statuses    [2]uint32 `json:"statuses"`
	Definitions [2]int    `json:"definitions"`
	Multipliers [3]uint32 `json:"multipliers"`
}

type randomWalkOriginalEvent545020 struct {
	Kind  string                        `json:"kind"`
	Args  []int64                       `json:"args"`
	State randomWalkOriginalState545020 `json:"state"`
}

// The root's existing Go API returns handled bool. Its raw original EAX and
// public byte remain in the reference; they are not replaced with that bool.
// The standalone direction helper's real public int is compared explicitly.
func TestMonsterRandomWalkOriginalNative545020(t *testing.T) {
	path := "testdata/monster_random_walk_native_original_545020.json"
	if p := os.Getenv("NOX_MONSTER_RANDOM_WALK_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA     string `json:"original_sha256"`
		Control uint16 `json:"control_word"`
		Rows    []struct {
			Input struct {
				Name       string    `json:"name"`
				Entry      string    `json:"entry"`
				Direction  uint16    `json:"direction"`
				Subclass   uint32    `json:"subclass"`
				Position   [2]uint32 `json:"position"`
				Speed      uint32    `json:"speed"`
				Multiplier uint32    `json:"run_multiplier"`
				Status     uint32    `json:"status"`
				Random     int       `json:"random"`
				Tile       int       `json:"tile"`
				Mutation   struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
					Value  int    `json:"value"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events []randomWalkOriginalEvent545020 `json:"events"`
				State  randomWalkOriginalState545020   `json:"state"`
			} `json:"result"`
			EAX uint32 `json:"raw_original_eax"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || (fixture.Control != 0xe3f && fixture.Control != 0x23f) || len(fixture.Rows) != 2651 {
		t.Fatal("original RandomWalk identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			type guarded struct {
				Before      [32]byte
				Unit        Object
				Updates     [2]MonsterUpdateData
				Definitions [3]MonsterDef
				After       [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, object.SubClass(c.Subclass), 0x12000004
			unit.PosVec = types.Ptf(math.Float32frombits(c.Position[0]), math.Float32frombits(c.Position[1]))
			unit.ForceVec = types.Ptf(math.Float32frombits(0x3f400001), math.Float32frombits(0xbf000001))
			unit.SpeedCur, unit.Direction1, unit.Direction2 = math.Float32frombits(c.Speed), Dir16(c.Direction), 0xbabe
			unit.UpdateData = unsafe.Pointer(&p.Updates[0])
			p.Updates[0].StatusFlags = object.MonsterStatus(c.Status)
			p.Updates[1].StatusFlags = object.MonStatusRunning
			p.Updates[0].MonsterDef, p.Updates[1].MonsterDef = &p.Definitions[0], &p.Definitions[2]
			for i, bits := range []uint32{c.Multiplier, 0x40500000, 0x3f400000} {
				p.Definitions[i].RunMultiplier96 = math.Float32frombits(bits)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, ptr := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(&p.Updates[0]), unsafe.Pointer(&p.Updates[1]), unsafe.Pointer(&p.Definitions[0])} {
					if uintptr(ptr) <= 0xffffffff {
						t.Fatal("native unit/update/definition links must exceed4GiB")
					}
				}
			}
			expected := *p
			state := func() randomWalkOriginalState545020 {
				v := randomWalkOriginalState545020{Update: -1, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Flags: uint32(unit.ObjFlags), Position: [2]uint32{math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)}, Directions: [2]uint16{uint16(unit.Direction1), uint16(unit.Direction2)}, Force: [2]uint32{math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y)}, Speed: math.Float32bits(unit.SpeedCur)}
				for i := range p.Updates {
					if unit.UpdateData == unsafe.Pointer(&p.Updates[i]) {
						v.Update = i
					}
					v.Statuses[i] = uint32(p.Updates[i].StatusFlags)
					v.Definitions[i] = -1
					for j := range p.Definitions {
						if p.Updates[i].MonsterDef == &p.Definitions[j] {
							v.Definitions[i] = j
						}
					}
				}
				for i := range p.Definitions {
					v.Multipliers[i] = math.Float32bits(p.Definitions[i].RunMultiplier96)
				}
				return v
			}
			mutate := func(stage string) {
				m := c.Mutation
				if m.Stage != stage {
					return
				}
				a, n := m.Action, m.Value
				if a == "direction" || a == "all" {
					unit.Direction1 = 129
					if n != 0 {
						unit.Direction1 = 0xffff
					}
					unit.Direction2 = 0x1234
				}
				if a == "subclass" || a == "all" {
					unit.ObjSubClass = 0x80
					if n != 0 {
						unit.ObjSubClass = 0x480
					}
				}
				if a == "class" || a == "all" {
					unit.ObjClass = object.ClassMonster
					if n != 0 {
						unit.ObjClass = 0x400000
					}
				}
				if a == "update" || a == "all" {
					unit.UpdateData = unsafe.Pointer(&p.Updates[n])
				}
				if a == "entry-status" || a == "all" {
					p.Updates[0].StatusFlags = 0
					if n != 0 {
						p.Updates[0].StatusFlags = object.MonStatusRunning
					}
				}
				if a == "alternate-status" || a == "all" {
					p.Updates[1].StatusFlags = object.MonStatusRunning
					if n != 0 {
						p.Updates[1].StatusFlags = 0
					}
				}
				if a == "entry-definition" || a == "all" {
					p.Updates[0].MonsterDef = &p.Definitions[1]
				}
				if a == "alternate-definition" || a == "all" {
					p.Updates[1].MonsterDef = &p.Definitions[n]
				}
				if a == "run-multiplier" || a == "all" {
					p.Definitions[0].RunMultiplier96 = math.Float32frombits(0x3fe00000)
					if n != 0 {
						p.Definitions[0].RunMultiplier96 = math.Float32frombits(0x3f000001)
					}
				}
				if a == "speed" || a == "all" {
					unit.SpeedCur = math.Float32frombits(0x40499999)
					if n != 0 {
						unit.SpeedCur = math.Float32frombits(0xbf900000)
					}
				}
				if a == "position" || a == "all" {
					unit.PosVec = types.Ptf(-37.5, 19.25)
				}
			}
			events := make([]randomWalkOriginalEvent545020, 0)
			emit := func(kind string, args []int64) {
				events = append(events, randomWalkOriginalEvent545020{Kind: kind, Args: args, State: state()})
				mutate(kind)
			}
			hooks := monsterActionRandomWalkHooks545020{
				random: func(min, max int) int {
					if min != -20 || max != 20 {
						t.Fatal("original random bounds")
					}
					emit("random", []int64{-20, 20, 435})
					return c.Random
				},
				tileAt: func(pos types.Pointf) int {
					emit("tile", []int64{int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y))})
					return c.Tile
				},
				moveAudio: func(got *Object) {
					if got != unit {
						t.Fatal("native move-audio unit identity")
					}
					emit("move_audio", []int64{0})
				},
			}
			// The same frozen comparator compiles before the optional explicit
			// nearest-profile hook exists. Production/default is startup chop.
			if f := reflect.ValueOf(&hooks).Elem().FieldByName("Nearest"); f.IsValid() {
				f.SetBool(fixture.Control == 0x23f)
			}
			if c.Entry == "root" {
				if !monsterActionRandomWalkMove545020(unit, hooks) {
					t.Fatal("valid original root context not handled")
				}
			} else if c.Entry == "direction" {
				if got := uint32(monsterRandomWalkDirection545090(unit, hooks)); got != row.EAX {
					t.Errorf("original helper int=%d want=%d", got, row.EAX)
				}
			} else {
				t.Fatal("unknown original entry")
			}
			if !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				got, _ := json.Marshal(struct {
					Events []randomWalkOriginalEvent545020
					State  randomWalkOriginalState545020
				}{events, state()})
				want, _ := json.Marshal(row.Result)
				t.Errorf("original ordered services/raw state differ\ngot=%s\nwant=%s", got, want)
			}
			v := row.Result.State
			expected.Unit.UpdateData = unsafe.Pointer(&p.Updates[v.Update])
			expected.Unit.ObjClass = object.Class(v.Class)
			expected.Unit.ObjSubClass = object.SubClass(v.Subclass)
			expected.Unit.ObjFlags = object.Flags(v.Flags)
			expected.Unit.PosVec = types.Ptf(math.Float32frombits(v.Position[0]), math.Float32frombits(v.Position[1]))
			expected.Unit.ForceVec = types.Ptf(math.Float32frombits(v.Force[0]), math.Float32frombits(v.Force[1]))
			expected.Unit.SpeedCur = math.Float32frombits(v.Speed)
			expected.Unit.Direction1, expected.Unit.Direction2 = Dir16(v.Directions[0]), Dir16(v.Directions[1])
			for i := range p.Updates {
				expected.Updates[i].StatusFlags = object.MonsterStatus(v.Statuses[i])
				expected.Updates[i].MonsterDef = &p.Definitions[v.Definitions[i]]
			}
			for i := range p.Definitions {
				expected.Definitions[i].RunMultiplier96 = math.Float32frombits(v.Multipliers[i])
			}
			raw := func(v unsafe.Pointer) []byte { return unsafe.Slice((*byte)(v), int(unsafe.Sizeof(*p))) }
			if !bytes.Equal(raw(unsafe.Pointer(p)), raw(unsafe.Pointer(&expected))) {
				t.Error("whole native unit/two-update/three-definition records and canaries differ")
			}
		})
	}
}
