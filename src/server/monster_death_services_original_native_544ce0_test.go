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

type deathServicesState544CE0 struct {
	Owner         int          `json:"owner"`
	OwnerClasses  [2]uint32    `json:"owner_classes"`
	OwnerUpdates  [2]int       `json:"owner_updates"`
	Players       [2]int       `json:"players"`
	PlayerIndices [2]byte      `json:"player_indices"`
	Subclass      uint32       `json:"subclass"`
	Flags         uint32       `json:"flags"`
	Position      [2]uint32    `json:"position"`
	Directions    [3][2]uint16 `json:"directions"`
	Cache         uint32       `json:"cache"`
}

type deathServicesEvent544CE0 struct {
	Kind      string                   `json:"kind"`
	Args      []int64                  `json:"args"`
	Iteration int                      `json:"iteration"`
	State     deathServicesState544CE0 `json:"state"`
}

// Both original roots have a void public ABI. Their retained raw EAX values
// are deliberately outside the existing handled-bool component contract.
func TestMonsterDeathServicesOriginalNative544CE0(t *testing.T) {
	path := "testdata/monster_death_services_native_original_544ce0.json"
	if p := os.Getenv("NOX_MONSTER_DEATH_SERVICES_ORIGINAL_FIXTURE"); p != "" {
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
				Name         string    `json:"name"`
				Entry        string    `json:"entry"`
				Owner        int       `json:"owner"`
				OwnerClass   uint32    `json:"owner_class"`
				PlayerIndex  byte      `json:"player_index"`
				Subclass     uint32    `json:"subclass"`
				Position     [2]uint32 `json:"position"`
				Direction    uint16    `json:"direction"`
				Cache        uint32    `json:"cache"`
				LookupValues []uint32  `json:"lookup_values"`
				Allocations  []int     `json:"allocations"`
				Repeats      int       `json:"repeats"`
				Mutation     struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
					Value  int    `json:"value"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events    []deathServicesEvent544CE0 `json:"events"`
				State     deathServicesState544CE0   `json:"state"`
				Snapshots []deathServicesState544CE0 `json:"snapshots"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Control != 0xe3f || len(fixture.Rows) != 467 {
		t.Fatal("original death-service identity/profile/count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			type guarded struct {
				Before  [32]byte
				Unit    Object
				Owners  [2]Object
				Updates [2]PlayerUpdateData
				Players [2]Player
				Souls   [2]Object
				Cache   uint32
				After   [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			objects := [3]*Object{&p.Souls[0], &p.Souls[1], unit}
			unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, object.SubClass(c.Subclass), 0x12000004
			unit.PosVec = types.Ptf(math.Float32frombits(c.Position[0]), math.Float32frombits(c.Position[1]))
			unit.Direction1, unit.Direction2 = Dir16(c.Direction), 0xbabe
			if c.Owner >= 0 {
				unit.ObjOwner = &p.Owners[c.Owner]
			}
			for i := range p.Owners {
				p.Owners[i].ObjClass = object.ClassPlayer
				p.Owners[i].UpdateData = unsafe.Pointer(&p.Updates[i])
				p.Updates[i].Player = &p.Players[i]
				p.Players[i].PlayerInd = 9
				p.Souls[i].Direction1, p.Souls[i].Direction2 = Dir16(0xa500+i), Dir16(0x5a00+i)
			}
			p.Owners[0].ObjClass, p.Players[0].PlayerInd, p.Cache = object.Class(c.OwnerClass), c.PlayerIndex, c.Cache
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, ptr := range []unsafe.Pointer{unit.CObj(), p.Owners[0].CObj(), p.Owners[1].CObj(), unsafe.Pointer(&p.Updates[0]), unsafe.Pointer(&p.Players[0]), p.Souls[0].CObj()} {
					if uintptr(ptr) <= 0xffffffff {
						t.Fatal("native object/owner/update/player/soul links must exceed4GiB")
					}
				}
			}
			expected := *p
			state := func() deathServicesState544CE0 {
				v := deathServicesState544CE0{Owner: -1, Subclass: uint32(unit.ObjSubClass), Flags: uint32(unit.ObjFlags), Position: [2]uint32{math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)}, Cache: p.Cache}
				for i := range p.Owners {
					if unit.ObjOwner == &p.Owners[i] {
						v.Owner = i
					}
					v.OwnerClasses[i] = uint32(p.Owners[i].ObjClass)
					v.OwnerUpdates[i], v.Players[i] = -1, -1
					for j := range p.Updates {
						if p.Owners[i].UpdateData == unsafe.Pointer(&p.Updates[j]) {
							v.OwnerUpdates[i] = j
						}
						if p.Updates[i].Player == &p.Players[j] {
							v.Players[i] = j
						}
					}
					v.PlayerIndices[i] = p.Players[i].PlayerInd
				}
				for i, obj := range []*Object{unit, &p.Souls[0], &p.Souls[1]} {
					v.Directions[i] = [2]uint16{uint16(obj.Direction1), uint16(obj.Direction2)}
				}
				return v
			}
			mutate := func(stage string, obj int) {
				m := c.Mutation
				if m.Stage != stage {
					return
				}
				a, n := m.Action, m.Value
				if c.Entry == "burn" {
					if a == "owner" || a == "all" {
						unit.ObjOwner = &p.Owners[n]
					}
					if a == "owner-update" || a == "all" {
						p.Owners[0].UpdateData = unsafe.Pointer(&p.Updates[n])
					}
					if a == "player" || a == "all" {
						p.Updates[0].Player = &p.Players[n]
					}
					if a == "index" || a == "all" {
						p.Players[0].PlayerInd = 1
						if n != 0 {
							p.Players[0].PlayerInd = 255
						}
					}
					if a == "subclass" || a == "all" {
						unit.ObjSubClass = 0x12340080
						if n != 0 {
							unit.ObjSubClass = 0xfedcba98
						}
					}
					if a == "class" || a == "all" {
						p.Owners[0].ObjClass = 0
						if n != 0 {
							p.Owners[0].ObjClass = 0x400000
						}
					}
				} else {
					if a == "cache" || a == "all" {
						p.Cache = 0
						if n != 0 {
							p.Cache = 18
						}
					}
					if a == "unit-direction" || a == "all" {
						unit.Direction1 = 0x1234
						if n != 0 {
							unit.Direction1 = 0xffff
						}
					}
					if a == "soul-direction" || a == "all" {
						if obj < 0 {
							obj = 0
						}
						objects[obj].Direction1, objects[obj].Direction2 = 0xa55a, 0x5aa5
					}
				}
				if a == "position" || a == "all" {
					unit.PosVec = types.Ptf(-37.5, 19.25)
				}
			}
			events := make([]deathServicesEvent544CE0, 0)
			iteration, lookups, allocations := 0, 0, 0
			emit := func(kind string, args []int64, obj int) {
				events = append(events, deathServicesEvent544CE0{Kind: kind, Args: args, Iteration: iteration, State: state()})
				mutate(kind, obj)
			}
			burn := ZombieBurnDeleteRuntime544CE0{
				NetFxShield: func(index int, got *Object) {
					if got != unit {
						t.Fatal("shield native identity")
					}
					emit("shield", []int64{int64(index), 0}, -1)
				},
				UnmarkMinimap: func(index int, got *Object, flags uint32) {
					if got != unit || flags != 1 {
						t.Fatal("minimap native identity/flags")
					}
					emit("unmark", []int64{int64(index), 0, int64(flags)}, -1)
				},
				SoloMonsterKillReward: func(got *Object) {
					if got != unit {
						t.Fatal("reward native identity")
					}
					emit("reward", []int64{0}, -1)
				},
				MakeScorch: func(pos types.Pointf, size int) {
					emit("scorch", []int64{int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y)), int64(size)}, -1)
				},
				DelayedDelete: func(got *Object) {
					if got != unit {
						t.Fatal("delete native identity")
					}
					emit("delete", []int64{0}, -1)
				},
			}
			soul := MonsterReleasedSoulRuntime544E60{
				NewObjectByTypeID: func(id string) *Object {
					if id != "ReleasedSoul" {
						t.Fatal("legacy name allocation")
					}
					// This marks the genuine old high-level boundary. It must not
					// impersonate the original name-index/cache/allocator sequence.
					emit("by-name", []int64{}, -1)
					n := c.Allocations[allocations]
					allocations++
					if n < 0 {
						return nil
					}
					return objects[n]
				},
				CreateObjectAt: func(got, owner *Object, pos types.Pointf) {
					n := -1
					for i, o := range objects {
						if o == got {
							n = i
						}
					}
					if n < 0 || owner != nil {
						t.Fatal("create native object/owner identity")
					}
					// Original caller pushes an extra zero word; 4DAA50 ignores it.
					emit("create", []int64{int64(n), 0, int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y)), 0}, n)
				},
			}
			// Reflection keeps this exact test source runnable against the old
			// API before engine edits and the added indexed hooks after the port.
			fields := reflect.ValueOf(&soul).Elem()
			for name, value := range map[string]any{
				"TypeIndex": &p.Cache,
				"TypeIndexByID": func(id string) uint32 {
					if id != "ReleasedSoul" {
						t.Fatal("original name-index lookup")
					}
					ret := c.LookupValues[lookups]
					lookups++
					emit("lookup", []int64{}, -1)
					return ret
				},
				"NewObjectByTypeIndex": func(index uint32) *Object {
					n := c.Allocations[allocations]
					allocations++
					emit("new", []int64{int64(index)}, n)
					if n < 0 {
						return nil
					}
					return objects[n]
				},
			} {
				if f := fields.FieldByName(name); f.IsValid() {
					f.Set(reflect.ValueOf(value))
				}
			}
			snapshots := make([]deathServicesState544CE0, 0, c.Repeats)
			for iteration = 0; iteration < c.Repeats; iteration++ {
				var handled bool
				switch c.Entry {
				case "burn":
					handled = new(Server).ZombieBurnDelete544CE0(unit, burn)
				case "soul":
					handled = MonsterCreateReleasedSoul544E60(unit, soul)
				default:
					t.Fatal("unknown original entry")
				}
				if !handled {
					t.Fatal("valid original context not handled")
				}
				snapshots = append(snapshots, state())
			}
			if !reflect.DeepEqual(events, row.Result.Events) || !reflect.DeepEqual(snapshots, row.Result.Snapshots) || state() != row.Result.State {
				got, _ := json.Marshal(struct {
					Events    []deathServicesEvent544CE0
					Snapshots []deathServicesState544CE0
					State     deathServicesState544CE0
				}{events, snapshots, state()})
				want, _ := json.Marshal(row.Result)
				t.Errorf("original ordered services/states differ\ngot=%s\nwant=%s", got, want)
			}
			v := row.Result.State
			expected.Unit.ObjOwner = nil
			if v.Owner >= 0 {
				expected.Unit.ObjOwner = &p.Owners[v.Owner]
			}
			expected.Unit.ObjSubClass, expected.Unit.ObjFlags = object.SubClass(v.Subclass), object.Flags(v.Flags)
			expected.Unit.PosVec = types.Ptf(math.Float32frombits(v.Position[0]), math.Float32frombits(v.Position[1]))
			for i := range p.Owners {
				expected.Owners[i].ObjClass = object.Class(v.OwnerClasses[i])
				expected.Owners[i].UpdateData = unsafe.Pointer(&p.Updates[v.OwnerUpdates[i]])
				expected.Updates[i].Player = &p.Players[v.Players[i]]
				expected.Players[i].PlayerInd = v.PlayerIndices[i]
			}
			for i, obj := range []*Object{&expected.Unit, &expected.Souls[0], &expected.Souls[1]} {
				obj.Direction1, obj.Direction2 = Dir16(v.Directions[i][0]), Dir16(v.Directions[i][1])
			}
			expected.Cache = v.Cache
			raw := func(ptr unsafe.Pointer) []byte { return unsafe.Slice((*byte)(ptr), int(unsafe.Sizeof(*p))) }
			if !bytes.Equal(raw(unsafe.Pointer(p)), raw(unsafe.Pointer(&expected))) {
				t.Error("whole native object/owner/update/player/soul/cache records and canaries differ")
			}
		})
	}
}
