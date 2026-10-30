package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type retreatOriginalInput545440 struct {
	Name, Entry           string
	Cur, Max              uint16
	Resume, Status, Buffs uint32
	Enemy                 int
	Related               uint32
	Food                  int
	Quest                 bool
	Accept, FPS, Frame    uint32
	Random                int32
	PopReturn             uint32 `json:"pop_return"`
	Positions             [4][2]uint32
	Mutation              struct {
		Stage, Action string
		Value         int
	}
}

type retreatOriginalState545440 struct {
	Update, Health    int
	Cls, Buffs        uint32
	Positions         [4][2]uint32
	Healths           [2][2]uint16
	Statuses, Resume  [2]uint32
	Enemies           [2]int
	Items             [2][5]uint32
	FPS, Frame, Flags uint32
}

type retreatOriginalEvent545440 struct {
	Kind  string
	Args  []int64
	State retreatOriginalState545440
}

type retreatOriginalWrite545440 struct {
	Size  int
	Value uint32
}
type retreatOriginalGuard545440[T comparable] struct {
	Head  [32]byte
	Value T
	Tail  [32]byte
}

func retreatOriginalAlloc545440[T comparable](t *testing.T) (*T, func()) {
	t.Helper()
	g, free := alloc.New(retreatOriginalGuard545440[T]{})
	t.Cleanup(free)
	clear(unsafe.Slice((*byte)(unsafe.Pointer(&g.Value)), int(unsafe.Sizeof(g.Value))))
	for i := range g.Head {
		g.Head[i], g.Tail[i] = 0xa5, 0x5a
	}
	return &g.Value, func() {
		t.Helper()
		for i := range g.Head {
			if g.Head[i] != 0xa5 || g.Tail[i] != 0x5a {
				t.Fatal("native allocation canary changed")
			}
		}
	}
}

type retreatOriginalWorld545440 struct {
	objects           [4]*Object
	updates           [2]*MonsterUpdateData
	healths           [2]*HealthData
	items             [2]*AIStackItem
	fps, frame, flags uint32
	events            []retreatOriginalEvent545440
	pushes            int
	checks            []func()
}

func (w *retreatOriginalWorld545440) ref(p *Object) int {
	if p == nil {
		return -1
	}
	for i, v := range w.objects {
		if v == p {
			return i
		}
	}
	panic("unknown original object identity")
}

func (w *retreatOriginalWorld545440) canonical(v uintptr) uint32 {
	for i, p := range w.objects {
		if v == uintptr(p.CObj()) {
			return 0xf0000000 + uint32(i)
		}
	}
	return uint32(v)
}

func (w *retreatOriginalWorld545440) expand(v uint32) uintptr {
	if v >= 0xf0000000 && v < 0xf0000004 {
		return uintptr(w.objects[v-0xf0000000].CObj())
	}
	return uintptr(v)
}

func (w *retreatOriginalWorld545440) state() retreatOriginalState545440 {
	s := retreatOriginalState545440{Update: -1, Health: -1, Cls: uint32(w.objects[0].ObjClass), Buffs: w.objects[0].Buffs, FPS: w.fps, Frame: w.frame, Flags: w.flags}
	for i, p := range w.objects {
		s.Positions[i] = [2]uint32{math.Float32bits(p.PosVec.X), math.Float32bits(p.PosVec.Y)}
	}
	for i := range w.updates {
		if w.objects[0].UpdateData == unsafe.Pointer(w.updates[i]) {
			s.Update = i
		}
		if w.objects[0].HealthData == w.healths[i] {
			s.Health = i
		}
		s.Healths[i] = [2]uint16{w.healths[i].Cur, w.healths[i].Max}
		s.Statuses[i], s.Resume[i], s.Enemies[i] = uint32(w.updates[i].StatusFlags), math.Float32bits(w.updates[i].ResumeLevel), w.ref(w.updates[i].CurrentEnemy)
		s.Items[i][0] = w.items[i].Action
		for j, v := range w.items[i].Args {
			s.Items[i][j+1] = w.canonical(v)
		}
	}
	if s.Update < 0 || s.Health < 0 {
		panic("unknown native update/health identity")
	}
	return s
}

func (w *retreatOriginalWorld545440) event(kind string, args ...int64) {
	if args == nil {
		args = []int64{}
	}
	w.events = append(w.events, retreatOriginalEvent545440{Kind: kind, Args: args, State: w.state()})
}

func (w *retreatOriginalWorld545440) mutate(c retreatOriginalInput545440, stage string) {
	m := c.Mutation
	if m.Stage != stage {
		return
	}
	a, v := m.Action, m.Value
	is := func(name string) bool { return a == name || a == "all" }
	if is("enemy") {
		w.updates[0].CurrentEnemy = w.objects[1+v]
		w.updates[1].CurrentEnemy = w.objects[2-v]
	}
	if a == "nil-enemy" {
		w.updates[0].CurrentEnemy = nil
	}
	if is("update") {
		w.objects[0].UpdateData = unsafe.Pointer(w.updates[v])
	}
	if is("health-pointer") {
		w.objects[0].HealthData = w.healths[v]
	}
	if is("health") {
		for _, h := range w.healths {
			if v != 0 {
				h.Cur, h.Max = 100, 100
			} else {
				h.Cur, h.Max = 1, 0
			}
		}
	}
	if is("status") {
		w.updates[0].StatusFlags, w.updates[1].StatusFlags = 0, object.MonsterStatus(0x20)
		if v != 0 {
			w.updates[0].StatusFlags, w.updates[1].StatusFlags = object.MonsterStatus(0x20), 0
		}
	}
	if is("resume") {
		w.updates[0].ResumeLevel = math.Float32frombits(0x7fc12345)
		if v != 0 {
			w.updates[0].ResumeLevel = 0.75
		}
		w.updates[1].ResumeLevel = 0.25
	}
	if is("buffs") {
		w.objects[0].Buffs = 0
		if v != 0 {
			w.objects[0].Buffs = 0x20000000
		}
	}
	if is("class") {
		w.objects[0].ObjClass = object.ClassMonster
		if v != 0 {
			w.objects[0].ObjClass = object.Class(0x400000)
		}
	}
	if is("positions") {
		for i, p := range w.objects {
			bits := uint32(0x7fc54321)
			if i&1 != 0 {
				bits = 0x80000000
			}
			p.PosVec = types.Ptf(math.Float32frombits(bits), math.Float32frombits(0x42c80001+uint32(i)))
		}
	}
	if is("fps") {
		w.fps = 0xffffffff
		if v != 0 {
			w.fps = 0x40000001
		}
	}
	if is("frame") {
		w.frame = 17
		if v != 0 {
			w.frame = 0xfffffffd
		}
	}
	if is("flags") {
		w.flags = 4096
		if v != 0 {
			w.flags = 0
		}
	}
	if is("item-args") {
		for i, item := range w.items {
			for j := range item.Args {
				item.Args[j] = uintptr(0x6b000000 + i*256 + j + 1)
			}
		}
	}
}

func retreatOriginalWorldNew545440(t *testing.T, c retreatOriginalInput545440) *retreatOriginalWorld545440 {
	t.Helper()
	w := &retreatOriginalWorld545440{fps: c.FPS, frame: c.Frame, events: []retreatOriginalEvent545440{}}
	if c.Quest {
		w.flags = 4096
	}
	for i := range w.objects {
		var check func()
		w.objects[i], check = retreatOriginalAlloc545440[Object](t)
		w.checks = append(w.checks, check)
		w.objects[i].PosVec = types.Ptf(math.Float32frombits(c.Positions[i][0]), math.Float32frombits(c.Positions[i][1]))
		w.objects[i].ForceVec = types.Ptf(math.Float32frombits(0x3f400001), math.Float32frombits(0xbf000001))
		w.objects[i].Direction1, w.objects[i].Direction2, w.objects[i].SpeedCur = 0xabcd, 0xbabe, 2.5
	}
	w.objects[0].ObjClass, w.objects[0].Buffs = object.ClassMonster, c.Buffs
	for i := range w.updates {
		var check func()
		w.updates[i], check = retreatOriginalAlloc545440[MonsterUpdateData](t)
		w.checks = append(w.checks, check)
		w.healths[i], check = retreatOriginalAlloc545440[HealthData](t)
		w.checks = append(w.checks, check)
		w.items[i], check = retreatOriginalAlloc545440[AIStackItem](t)
		w.checks = append(w.checks, check)
		w.updates[i].AIStack[0].Action = uint32(ai.ACTION_RETREAT)
		w.updates[i].StatusFlags, w.updates[i].ResumeLevel, w.updates[i].CurrentEnemy = object.MonsterStatus(0x20), 0.75, w.objects[2]
		w.healths[i].Cur, w.healths[i].Max = 75, 100
		w.items[i].Action = 0xa5a50000 + uint32(i)
		for j := range w.items[i].Args {
			value := uintptr(0x5a000000 + i*256 + j + 1)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				high := uint64(0xface0000)
				value |= uintptr(high << 32)
			}
			w.items[i].Args[j] = value
		}
	}
	w.objects[0].UpdateData, w.objects[0].HealthData = unsafe.Pointer(w.updates[0]), w.healths[0]
	w.updates[0].StatusFlags, w.updates[0].ResumeLevel = object.MonsterStatus(c.Status), math.Float32frombits(c.Resume)
	w.updates[0].CurrentEnemy = nil
	if c.Enemy >= 0 {
		w.updates[0].CurrentEnemy = w.objects[c.Enemy]
	}
	w.healths[0].Cur, w.healths[0].Max = c.Cur, c.Max
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{w.objects[0].CObj(), unsafe.Pointer(w.updates[0]), unsafe.Pointer(w.healths[0]), unsafe.Pointer(w.items[0])} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatal("original comparator requires native pointers above4GiB")
			}
		}
	}
	return w
}

func TestMonsterRetreatOriginalNative545440(t *testing.T) {
	file := os.Getenv("NOX_MONSTER_RETREAT_ORIGINAL_FIXTURE")
	if file == "" {
		file = "testdata/monster_retreat_native_original_545440.json"
	}
	blob, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		ControlWord    uint16 `json:"control_word"`
		Rows           []struct {
			Input  retreatOriginalInput545440
			Result struct {
				Events     []retreatOriginalEvent545440
				State      retreatOriginalState545440
				Fault      bool
				ItemWrites map[string][]retreatOriginalWrite545440 `json:"item_writes"`
			}
		}
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 1465 || (fixture.ControlWord != 0xe3f && fixture.ControlWord != 0x23f) {
		t.Fatal("unsealed original reference")
	}
	t.Logf("ORIGINAL RETREAT PROFILE: fixture=%s control=%#x contexts=%d", file, fixture.ControlWord, len(fixture.Rows))
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			w := retreatOriginalWorldNew545440(t, c)
			var beforeObjects [4]Object
			var beforeUpdates [2]MonsterUpdateData
			var beforeHealths [2]HealthData
			var beforeItems [2]AIStackItem
			for i, p := range w.objects {
				beforeObjects[i] = *p
			}
			for i := range w.updates {
				beforeUpdates[i], beforeHealths[i], beforeItems[i] = *w.updates[i], *w.healths[i], *w.items[i]
			}
			hooks := monsterActionRetreatHooks545440{
				frame:    func() uint32 { w.event("frame"); return w.frame },
				tickRate: func() uint32 { w.event("tick"); return w.fps },
				random: func(low, high int) int {
					w.event("random", int64(low), int64(high), 149)
					w.mutate(c, "random")
					return int(c.Random)
				},
				castRelated: func(unit *Object) bool {
					if unit != w.objects[0] {
						t.Fatal("related lost native unit")
					}
					w.event("related")
					w.mutate(c, "related")
					return c.Related != 0
				},
				searchFood: func(unit *Object, radius float32) *Object {
					if unit != w.objects[0] {
						t.Fatal("food search lost native unit")
					}
					var ret *Object
					if c.Food >= 0 {
						ret = w.objects[c.Food]
					}
					w.event("search", int64(math.Float32bits(radius)))
					w.mutate(c, "search")
					return ret
				},
				quest: c.Quest,
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					if len(args) != 0 {
						t.Fatal("original push has no preloaded payload")
					}
					w.pushes++
					var ret *AIStackItem
					if c.Accept&(1<<uint(w.pushes-1)) != 0 {
						ret = w.items[(w.pushes-1)%2]
					}
					w.event("push", int64(action))
					w.mutate(c, fmt.Sprintf("push%d", w.pushes))
					if ret != nil {
						ret.Action = uint32(action)
					}
					return ret
				},
				pop: func() int { w.event("pop"); w.mutate(c, "pop"); return int(int32(c.PopReturn)) },
			}
			fault := false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				if c.Entry == "nil-buff" {
					if value := (*Object)(nil).UnitBuffTest4FF350(29); value != 0 {
						t.Fatalf("original nil Buff result0 got%d", value)
					}
				} else {
					monsterActionRetreat545440(w.objects[0], hooks)
				}
			}()
			if fault != row.Result.Fault || !reflect.DeepEqual(w.events, row.Result.Events) || w.state() != row.Result.State {
				t.Fatalf("original ordered effects differ: fault=%t want=%t\nevents=%+v\nwant=%+v\nstate=%+v\nwant=%+v", fault, row.Result.Fault, w.events, row.Result.Events, w.state(), row.Result.State)
			}
			want := row.Result.State
			for i := range beforeObjects {
				beforeObjects[i].PosVec = types.Ptf(math.Float32frombits(want.Positions[i][0]), math.Float32frombits(want.Positions[i][1]))
			}
			beforeObjects[0].ObjClass, beforeObjects[0].Buffs = object.Class(want.Cls), want.Buffs
			beforeObjects[0].UpdateData, beforeObjects[0].HealthData = unsafe.Pointer(w.updates[want.Update]), w.healths[want.Health]
			for i := range beforeUpdates {
				beforeUpdates[i].StatusFlags, beforeUpdates[i].ResumeLevel = object.MonsterStatus(want.Statuses[i]), math.Float32frombits(want.Resume[i])
				beforeUpdates[i].CurrentEnemy = nil
				if want.Enemies[i] >= 0 {
					beforeUpdates[i].CurrentEnemy = w.objects[want.Enemies[i]]
				}
				beforeHealths[i].Cur, beforeHealths[i].Max = want.Healths[i][0], want.Healths[i][1]
				for field := 0; field < 5; field++ {
					for _, write := range row.Result.ItemWrites[fmt.Sprintf("%d:%d", i, field)] {
						if field == 0 {
							if write.Size != 4 {
								t.Fatal("action width")
							}
							beforeItems[i].Action = write.Value
							continue
						}
						if write.Size == 4 {
							beforeItems[i].Args[field-1] = w.expand(write.Value)
						} else if write.Size == 1 {
							beforeItems[i].Args[field-1] = beforeItems[i].Args[field-1]&^uintptr(255) | uintptr(write.Value)
						} else {
							t.Fatal("original store width")
						}
					}
				}
			}
			compare := func(got, want unsafe.Pointer, size uintptr) {
				t.Helper()
				if !bytes.Equal(unsafe.Slice((*byte)(got), int(size)), unsafe.Slice((*byte)(want), int(size))) {
					t.Fatal("whole native record/canary differs from original write-derived result")
				}
			}
			for i, p := range w.objects {
				compare(p.CObj(), unsafe.Pointer(&beforeObjects[i]), unsafe.Sizeof(beforeObjects[i]))
			}
			for i := range w.updates {
				compare(unsafe.Pointer(w.updates[i]), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
				compare(unsafe.Pointer(w.healths[i]), unsafe.Pointer(&beforeHealths[i]), unsafe.Sizeof(beforeHealths[i]))
				compare(unsafe.Pointer(w.items[i]), unsafe.Pointer(&beforeItems[i]), unsafe.Sizeof(beforeItems[i]))
			}
			for _, check := range w.checks {
				check()
			}
		})
	}
}
