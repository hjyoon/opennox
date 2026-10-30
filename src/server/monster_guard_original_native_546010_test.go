package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type guardOriginalInput546010 struct {
	Name, Entry               string
	Aggression, Sight         uint32
	Positions                 [4][2]uint32
	Previous                  [2]uint32
	Guard                     [3]uint32
	Classes, Flags            [4]uint32
	Direction                 uint16
	Net, Buffs                uint32
	Enemy                     int
	Seen                      []int
	Sound                     [4]uint32
	LastMove                  uint32 `json:"last_move"`
	FPS, Frame                uint32
	Random                    int32
	Notice, Look, Interesting uint32
	Mimics                    [4]uint32
	Plant, Heal, Accept       uint32
	Alias                     int
	Mutation                  struct {
		Stage, Action string
		Value         int
	}
}

type guardOriginalState546010 struct {
	Update            int
	Classes, Flags    [4]uint32
	Positions         [4][2]uint32
	Previous          [2]uint32
	Direction         uint16
	Net, Buffs        uint32
	Indices           [2]uint8
	Aggression, Sight [2]uint32
	Sound             [2][4]uint32
	LastMove          [2]uint32 `json:"last_move"`
	Enemies           [2]int
	SeenCount         [2]uint8 `json:"seen_count"`
	Seen              [2][16]int
	Items             [4][5]uint32
	FPS, Frame        uint32
}

type guardOriginalEvent546010 struct {
	Kind  string
	Args  []int64
	State guardOriginalState546010
}

type guardOriginalWorld546010 struct {
	objects        [4]*Object
	updates        [2]*MonsterUpdateData
	items          [4]*AIStackItem
	fps, frame     uint32
	events         []guardOriginalEvent546010
	pushes, mimics int
	checks         []func()
}

func (w *guardOriginalWorld546010) ref(p *Object) int {
	if p == nil {
		return -1
	}
	for i, v := range w.objects {
		if p == v {
			return i
		}
	}
	panic("unknown Guard original object identity")
}

func (w *guardOriginalWorld546010) expand(v uint32) uintptr {
	if v >= 0xf0000000 && v < 0xf0000004 {
		return uintptr(w.objects[v-0xf0000000].CObj())
	}
	return uintptr(v)
}

func (w *guardOriginalWorld546010) canonical(v uintptr) uint32 {
	for i, p := range w.objects {
		if v == uintptr(p.CObj()) {
			return 0xf0000000 + uint32(i)
		}
	}
	return uint32(v)
}

func (w *guardOriginalWorld546010) state() guardOriginalState546010 {
	s := guardOriginalState546010{Update: -1, Previous: [2]uint32{math.Float32bits(w.objects[0].PrevPos.X), math.Float32bits(w.objects[0].PrevPos.Y)}, Direction: uint16(w.objects[0].Direction1), Net: uint32(w.objects[0].NetCode), Buffs: w.objects[0].Buffs, FPS: w.fps, Frame: w.frame}
	for i, p := range w.objects {
		s.Classes[i], s.Flags[i] = uint32(p.ObjClass), uint32(p.ObjFlags)
		s.Positions[i] = [2]uint32{math.Float32bits(p.PosVec.X), math.Float32bits(p.PosVec.Y)}
	}
	for i, v := range w.updates {
		if w.objects[0].UpdateData == unsafe.Pointer(v) {
			s.Update = i
		}
		s.Indices[i] = uint8(v.AIStackInd)
		s.Aggression[i], s.Sight[i] = math.Float32bits(v.Aggression), math.Float32bits(v.SightRange)
		s.Sound[i] = [4]uint32{v.Field97, math.Float32bits(v.Field99X), math.Float32bits(v.Field99Y), v.Field101}
		s.LastMove[i], s.Enemies[i], s.SeenCount[i] = v.Field137, w.ref(v.CurrentEnemy), v.Field282_1
		for j, p := range v.SeenEnemies {
			s.Seen[i][j] = w.ref(p)
		}
	}
	for i, v := range w.items {
		s.Items[i][0] = v.Action
		for j, arg := range v.Args {
			s.Items[i][j+1] = w.canonical(arg)
		}
	}
	if s.Update < 0 {
		panic("unknown Guard update identity")
	}
	return s
}

func (w *guardOriginalWorld546010) event(kind string, args ...int64) {
	if args == nil {
		args = []int64{}
	}
	w.events = append(w.events, guardOriginalEvent546010{Kind: kind, Args: args, State: w.state()})
}

func guardOriginalPoint546010(v [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1]))
}

func (w *guardOriginalWorld546010) mutate(c guardOriginalInput546010, stage string) {
	m := c.Mutation
	if m.Stage != stage {
		return
	}
	a, v := m.Action, m.Value
	is := func(name string) bool { return a == name || a == "all" }
	if is("update") {
		w.objects[0].UpdateData = unsafe.Pointer(w.updates[v])
	}
	if is("enemy") {
		w.updates[0].CurrentEnemy, w.updates[1].CurrentEnemy = w.objects[1+v], w.objects[2-v]
	}
	if a == "nil-enemy" {
		w.updates[0].CurrentEnemy = nil
	}
	if is("aggression") {
		w.updates[0].Aggression, w.updates[1].Aggression = math.Float32frombits(0x7fc12345), 0
		if v != 0 {
			w.updates[0].Aggression, w.updates[1].Aggression = 1, 0.5
		}
	}
	if is("sight") {
		w.updates[0].SightRange = math.Float32frombits(0x7fc54321)
		if v != 0 {
			w.updates[0].SightRange = math.Float32frombits(0x42c80001)
		}
		w.updates[1].SightRange = math.Float32frombits(0x3f000001)
	}
	if is("positions") {
		for i, p := range w.objects {
			bits := [2]uint32{0x80000000, 0x7fc54321}
			if v != 0 {
				bits = [2]uint32{0x42c80001 + uint32(i), 0xc1a00001 - uint32(i)}
			}
			p.PosVec = guardOriginalPoint546010(bits)
		}
	}
	if is("previous") {
		bits := [2]uint32{0x7fc12345, 0x41a00000}
		if v != 0 {
			bits = [2]uint32{0x3f000001, 0x80000000}
		}
		w.objects[0].PrevPos = guardOriginalPoint546010(bits)
	}
	if is("direction") {
		w.objects[0].Direction1 = 129
		if v != 0 {
			w.objects[0].Direction1 = 65535
		}
	}
	if is("net") {
		w.objects[0].NetCode = 0
		if v != 0 {
			w.objects[0].NetCode = 0xffffffff
		}
	}
	if is("buffs") {
		w.objects[0].Buffs = 0
		if v != 0 {
			w.objects[0].Buffs = 0x20000000
		}
	}
	if is("classes") {
		for i, p := range w.objects {
			p.ObjClass = object.ClassMonster
			if v != 0 && i == 2 {
				p.ObjClass = object.ClassPlayer
			}
		}
	}
	if is("flags") {
		for i, p := range w.objects {
			p.ObjFlags = object.Flags(4)
			if i == v {
				p.ObjFlags = object.Flags(0x1000000)
			}
		}
	}
	if is("frame") {
		w.frame = 17
		if v != 0 {
			w.frame = 0xfffffff0
		}
	}
	if is("fps") {
		w.fps = 0xffffffff
		if v != 0 {
			w.fps = 0x80000001
		}
	}
	if is("last-move") {
		w.updates[0].Field137 = 1
		if v != 0 {
			w.updates[0].Field137 = 0xfffffff0
		}
		w.updates[1].Field137 = 17
	}
	if is("sound") {
		for i, p := range w.updates {
			p.Field97, p.Field101 = 0, 3
			if v != 0 {
				p.Field97, p.Field101 = 7, 0xfffffff0
			}
			p.Field99X, p.Field99Y = math.Float32frombits(0x41a00001+uint32(i)), math.Float32frombits(0xc1200001-uint32(i))
		}
	}
	if is("seen") {
		for _, p := range w.updates {
			p.Field282_1 = 0
			if v != 0 {
				p.Field282_1 = 3
			}
			for j := range p.SeenEnemies {
				p.SeenEnemies[j] = w.objects[3-j%3]
			}
		}
	}
	if is("head-args") {
		for i, item := range w.items[2:] {
			item.Args = [4]uintptr{uintptr(0x42c80001 + i), uintptr(0xc1a00001 - i), 64, 0x6b000004}
		}
	}
	if is("item-args") {
		for i, item := range w.items[:2] {
			for j := range item.Args {
				item.Args[j] = uintptr(0x6b000000 + i*256 + j + 1)
			}
		}
	}
}

func guardOriginalWorldNew546010(t *testing.T, c guardOriginalInput546010) *guardOriginalWorld546010 {
	t.Helper()
	w := &guardOriginalWorld546010{fps: c.FPS, frame: c.Frame, events: []guardOriginalEvent546010{}}
	for i := range w.objects {
		var check func()
		w.objects[i], check = retreatOriginalAlloc545440[Object](t)
		w.checks = append(w.checks, check)
		p := w.objects[i]
		p.ObjClass, p.ObjFlags = object.Class(c.Classes[i]), object.Flags(c.Flags[i])
		p.PosVec = guardOriginalPoint546010(c.Positions[i])
		p.ForceVec = types.Ptf(0.75, -0.5)
		p.Direction2 = 0xbabe
	}
	w.objects[0].NetCode, w.objects[0].Buffs = c.Net, c.Buffs
	w.objects[0].PrevPos = guardOriginalPoint546010(c.Previous)
	w.objects[0].Direction1 = Dir16(c.Direction)
	for i := range w.updates {
		var check func()
		w.updates[i], check = retreatOriginalAlloc545440[MonsterUpdateData](t)
		w.checks = append(w.checks, check)
		w.items[i], check = retreatOriginalAlloc545440[AIStackItem](t)
		w.checks = append(w.checks, check)
		w.items[i+2] = &w.updates[i].AIStack[0]
		p := w.updates[i]
		p.Aggression, p.SightRange, p.Field137 = 1, 100, 17
		p.Field97, p.Field99X, p.Field99Y, p.Field101 = 7, 20, -10, 0
		p.Field282_1 = 2
		p.CurrentEnemy = w.objects[2]
		for j := range p.SeenEnemies {
			p.SeenEnemies[j] = w.objects[1+j%3]
		}
	}
	w.objects[0].UpdateData = unsafe.Pointer(w.updates[0])
	u := w.updates[0]
	u.Aggression, u.SightRange, u.Field137 = math.Float32frombits(c.Aggression), math.Float32frombits(c.Sight), c.LastMove
	u.Field97, u.Field99X, u.Field99Y, u.Field101 = c.Sound[0], math.Float32frombits(c.Sound[1]), math.Float32frombits(c.Sound[2]), c.Sound[3]
	u.CurrentEnemy = nil
	if c.Enemy >= 0 {
		u.CurrentEnemy = w.objects[c.Enemy]
	}
	u.Field282_1 = uint8(len(c.Seen))
	clear(u.SeenEnemies[:])
	for j, idx := range c.Seen {
		u.SeenEnemies[j] = w.objects[idx]
	}
	for i, p := range w.items {
		p.Action = 0xa5a50000 + uint32(i)
		p.Field5 = 0x13572468
		if i >= 2 {
			p.Action = uint32(ai.ACTION_GUARD)
		}
		for j := range p.Args {
			low := uint32(0x5a000000 + i*256 + j + 1)
			if i >= 2 && j < 3 {
				low = [3]uint32{0x41a00000, 0xc1200000, 64}[j]
				if i == 2 {
					low = c.Guard[j]
				}
			}
			p.Args[j] = uintptr(low)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				high := uint64(0xface0000)
				p.Args[j] |= uintptr(high << 32)
			}
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{w.objects[0].CObj(), unsafe.Pointer(w.updates[0]), unsafe.Pointer(w.items[0])} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatal("Guard comparator requires actual native pointers above4GiB")
			}
		}
	}
	return w
}

func TestMonsterGuardOriginalNative546010(t *testing.T) {
	file := os.Getenv("NOX_MONSTER_GUARD_ORIGINAL_FIXTURE")
	if file == "" {
		file = "testdata/monster_guard_native_original_546010.json.gz"
	}
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		ControlWord    uint16 `json:"control_word"`
		Rows           []struct {
			Input  guardOriginalInput546010
			Result struct {
				Events     []guardOriginalEvent546010
				State      guardOriginalState546010
				Fault      bool
				ItemWrites map[string][]retreatOriginalWrite545440 `json:"item_writes"`
			}
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(io.Discard, z); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 5713 || fixture.ControlWord != 0xe3f && fixture.ControlWord != 0x23f {
		t.Fatal("unsealed original Guard reference")
	}
	t.Logf("ORIGINAL GUARD PROFILE: fixture=%s control=%#x contexts=%d", file, fixture.ControlWord, len(fixture.Rows))
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			w := guardOriginalWorldNew546010(t, c)
			var beforeObjects [4]Object
			var beforeUpdates [2]MonsterUpdateData
			var beforeItems [2]AIStackItem
			for i, p := range w.objects {
				beforeObjects[i] = *p
			}
			for i, p := range w.updates {
				beforeUpdates[i] = *p
				beforeItems[i] = *w.items[i]
			}
			hooks := monsterActionGuardHooks546010{
				frame:            func() uint32 { w.event("frame"); return w.frame },
				tickRate:         func() uint32 { w.event("tick"); return w.fps },
				noticeThreat:     func() int { w.event("notice"); w.mutate(c, "notice"); return int(int32(c.Notice)) },
				lookAtDamager:    func() bool { w.event("look"); w.mutate(c, "look"); return c.Look != 0 },
				interestingSound: func() int { w.event("interesting"); w.mutate(c, "interesting"); return int(int32(c.Interesting)) },
				isPlant:          func() bool { w.event("plant"); w.mutate(c, "plant"); return c.Plant != 0 },
				isMimic: func() bool {
					w.mimics++
					w.event("mimic")
					w.mutate(c, fmt.Sprintf("mimic%d", w.mimics))
					return c.Mimics[w.mimics-1] != 0
				},
				healSomeone: func() int { w.event("heal"); w.mutate(c, "heal"); return int(int32(c.Heal)) },
				random: func(low, high int) int {
					w.event("random", int64(low), int64(high), 42)
					w.mutate(c, "random")
					return int(c.Random)
				},
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					w.pushes++
					index := (w.pushes - 1) % 2
					if c.Alias == 1 || c.Alias == 2 {
						index = c.Alias + 1
					} else if c.Alias == 3 {
						index = (w.pushes - 1) % 4
					}
					var ret *AIStackItem
					if c.Accept&(1<<uint(w.pushes-1)) != 0 {
						ret = w.items[index]
					}
					w.event("push", int64(action))
					w.mutate(c, fmt.Sprintf("push%d", w.pushes))
					if ret != nil {
						ret.Action = uint32(action)
						ret.SetArgs(args...)
					}
					return ret
				},
			}
			// Optional profile/observation fields may be added by the port. Their
			// absence in the old implementation cannot suppress fixture events.
			value := reflect.ValueOf(&hooks).Elem()
			if f := value.FieldByName("Nearest"); f.IsValid() {
				f.SetBool(fixture.ControlWord == 0x23f)
			}
			if f := value.FieldByName("ObserveFacing"); f.IsValid() {
				f.Set(reflect.ValueOf(func(v types.Pointf) { w.event("facing", int64(math.Float32bits(v.X)), int64(math.Float32bits(v.Y))) }))
			}
			fault := false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				if c.Entry == "nil-buff" {
					if (*Object)(nil).UnitBuffTest4FF350(29) != 0 {
						t.Fatal("nil original Buff result")
					}
				} else {
					monsterActionGuard546010(w.objects[0], hooks)
				}
			}()
			if fault != row.Result.Fault || !reflect.DeepEqual(w.events, row.Result.Events) || w.state() != row.Result.State {
				index := -1
				for i := 0; i < len(w.events) && i < len(row.Result.Events); i++ {
					if !reflect.DeepEqual(w.events[i], row.Result.Events[i]) {
						index = i
						break
					}
				}
				t.Fatalf("original Guard effects differ: fault=%t want=%t events=%d want=%d first-difference=%d final-state-match=%t\nactual kinds=%v\noriginal kinds=%v", fault, row.Result.Fault, len(w.events), len(row.Result.Events), index, w.state() == row.Result.State, guardOriginalKinds546010(w.events), guardOriginalKinds546010(row.Result.Events))
			}
			want := row.Result.State
			for i := range beforeObjects {
				beforeObjects[i].ObjClass, beforeObjects[i].ObjFlags = object.Class(want.Classes[i]), object.Flags(want.Flags[i])
				beforeObjects[i].PosVec = guardOriginalPoint546010(want.Positions[i])
			}
			beforeObjects[0].PrevPos = guardOriginalPoint546010(want.Previous)
			beforeObjects[0].Direction1 = Dir16(want.Direction)
			beforeObjects[0].NetCode, beforeObjects[0].Buffs = want.Net, want.Buffs
			beforeObjects[0].UpdateData = unsafe.Pointer(w.updates[want.Update])
			for i := range beforeUpdates {
				v := &beforeUpdates[i]
				v.AIStackInd = int8(want.Indices[i])
				v.Aggression, v.SightRange, v.Field137 = math.Float32frombits(want.Aggression[i]), math.Float32frombits(want.Sight[i]), want.LastMove[i]
				v.Field97, v.Field99X, v.Field99Y, v.Field101 = want.Sound[i][0], math.Float32frombits(want.Sound[i][1]), math.Float32frombits(want.Sound[i][2]), want.Sound[i][3]
				v.CurrentEnemy = nil
				if want.Enemies[i] >= 0 {
					v.CurrentEnemy = w.objects[want.Enemies[i]]
				}
				v.Field282_1 = want.SeenCount[i]
				for j, ref := range want.Seen[i] {
					v.SeenEnemies[j] = nil
					if ref >= 0 {
						v.SeenEnemies[j] = w.objects[ref]
					}
				}
			}
			for i := 0; i < 4; i++ {
				var item *AIStackItem
				if i < 2 {
					item = &beforeItems[i]
				} else {
					item = &beforeUpdates[i-2].AIStack[0]
				}
				for field := 0; field < 5; field++ {
					for _, write := range row.Result.ItemWrites[fmt.Sprintf("%d:%d", i, field)] {
						if write.Size != 4 {
							t.Fatal("original Guard DWORD width")
						}
						if field == 0 {
							item.Action = write.Value
						} else {
							item.Args[field-1] = w.expand(write.Value)
						}
					}
				}
			}
			compare := func(got, want unsafe.Pointer, size uintptr) {
				t.Helper()
				if !bytes.Equal(unsafe.Slice((*byte)(got), int(size)), unsafe.Slice((*byte)(want), int(size))) {
					t.Fatal("whole native Guard record differs from original write-derived result")
				}
			}
			for i, p := range w.objects {
				compare(p.CObj(), unsafe.Pointer(&beforeObjects[i]), unsafe.Sizeof(beforeObjects[i]))
			}
			for i, p := range w.updates {
				compare(unsafe.Pointer(p), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
				compare(unsafe.Pointer(w.items[i]), unsafe.Pointer(&beforeItems[i]), unsafe.Sizeof(beforeItems[i]))
			}
			for _, check := range w.checks {
				check()
			}
		})
	}
}

func guardOriginalKinds546010(events []guardOriginalEvent546010) []string {
	values := make([]string, len(events))
	for i, e := range events {
		values[i] = fmt.Sprintf("%s%v", e.Kind, e.Args)
	}
	return values
}
