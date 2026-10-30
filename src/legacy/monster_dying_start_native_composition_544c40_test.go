package legacy

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type dyingStartCState544C40 struct {
	Update   int        `json:"update"`
	Class    uint32     `json:"cls"`
	Subclass uint32     `json:"subclass"`
	Flags    uint32     `json:"flags"`
	Position [2]float32 `json:"position"`
	Updates  [2]struct {
		Definition  int    `json:"definition"`
		Sound       int    `json:"sound"`
		ScriptFlags uint32 `json:"script_flags"`
		ScriptFunc  int32  `json:"script_func"`
	} `json:"updates"`
	Callbacks [3]int        `json:"callbacks"`
	Sounds    [2][16]uint32 `json:"sounds"`
}

type dyingStartCRow544C40 struct {
	Input struct {
		Name           string `json:"name"`
		Entry          string `json:"entry"`
		Update         int    `json:"update"`
		Class          uint32 `json:"cls"`
		NilUnit        bool   `json:"nil_unit"`
		Sound          int    `json:"sound"`
		Callback       int    `json:"callback"`
		CallbackReturn uint32 `json:"callback_return"`
		Mutation       struct {
			Stage  string `json:"stage"`
			Action string `json:"action"`
			Value  int    `json:"value"`
		} `json:"mutation"`
	} `json:"input"`
	Result struct {
		Events []struct {
			Kind string   `json:"kind"`
			Args []uint32 `json:"args"`
		} `json:"events"`
		State dyingStartCState544C40 `json:"state"`
	} `json:"result"`
	RawEAX       uint32 `json:"raw_original_eax"`
	GetterReturn int    `json:"getter_return_sound"`
}

// Audio.EventObj and stock die callbacks run through the real production
// runtime. ScriptCallback is an observed mutation boundary; cooperative mode
// and peers are absent. This does not claim VM execution, audible output or drops.
func TestMonsterDyingStartNativeComposition544C40(t *testing.T) {
	path := "../server/testdata/monster_dying_start_native_original_544c40.json"
	if p := os.Getenv("NOX_MONSTER_DYING_START_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []dyingStartCRow544C40 `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	selected := make([]dyingStartCRow544C40, 0, 65)
	for _, row := range fixture.Rows {
		c := row.Input
		if c.Entry == "getter" {
			selected = append(selected, row)
			continue
		}
		if c.Update != 0 || c.CallbackReturn != 1 {
			continue
		}
		if c.Mutation.Stage == "" && (c.Sound == -1 || c.Sound == 0) {
			selected = append(selected, row)
			continue
		}
		m := c.Mutation
		if c.Sound == 0 && c.Callback == 1 && m.Stage == "script" && ((m.Action == "update" && m.Value == 2) || (m.Action == "entry-definition" && m.Value == 2) || (m.Action == "entry-function" && m.Value == 0) || (m.Action == "class" && m.Value == 2)) {
			selected = append(selected, row)
		}
	}
	if len(selected) != 65 {
		t.Fatal("49getter +12stock root +4cached/live script compositions", len(selected))
	}
	const child = "NOX_MONSTER_DYING_START_COMPOSITION_CHILD"
	if os.Getenv(child) == "" {
		for _, row := range selected {
			mode := row.Input.Entry + ":" + row.Input.Name
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDyingStartNativeComposition544C40$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual public C dying/getter composition: %v", err)
				}
			})
		}
		return
	}
	mode := os.Getenv(child)
	var ref *dyingStartCRow544C40
	for i := range selected {
		if mode == selected[i].Input.Entry+":"+selected[i].Input.Name {
			ref = &selected[i]
		}
	}
	if ref == nil {
		t.Fatal("unknown frozen composition", mode)
	}
	c := ref.Input
	f := dyingStartNativeFixture544C40(t, c.Update, c.Sound, c.Callback)
	f.unit.ObjClass = object.Class(c.Class)
	t.Cleanup(func() { f.unit.ObjClass = object.ClassMonster })
	functions := dyingStartStockFunctions544C40()
	for i, fn := range functions {
		if i != 0 && fn == nil {
			t.Fatal("missing stock native callback", i)
		}
	}
	entry := f.updates[c.Update]
	other := f.updates[1-c.Update]
	mutate := func() {
		m := c.Mutation
		if m.Stage == "" {
			return
		}
		if m.Stage != "script" {
			t.Fatal("unreviewed C composition mutation")
		}
		switch m.Action {
		case "update":
			f.unit.UpdateData = unsafe.Pointer(other)
		case "entry-definition":
			entry.MonsterDef = &f.extra.Definitions[1]
		case "entry-function":
			entry.MonsterDef.DieFunc228 = functions[m.Value]
		case "class":
			f.unit.ObjClass = 0x400000
			f.unit.ObjSubClass = 0xfedcba98
		default:
			t.Fatal("unknown C script mutation")
		}
	}
	f.script.mutate = mutate
	beforeUnit, beforeUpdate, beforeExtra, beforeMissile := *f.unit, *f.updates[0], *f.extra, *f.missile
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), int(n)) }
	compare := func(unit *server.Object, update *server.MonsterUpdateData, extra *dyingStartNativeExtra544C40) {
		for _, v := range []struct {
			got, want unsafe.Pointer
			n         uintptr
		}{{f.unit.CObj(), unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(f.updates[0]), unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(f.extra), unsafe.Pointer(extra), unsafe.Sizeof(*extra)}, {f.missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)}} {
			if !bytes.Equal(raw(v.got, v.n), raw(v.want, v.n)) {
				t.Fatal("full native unit/two-updates/definitions/sounds/missile/canaries differ")
			}
		}
	}
	if c.Entry == "getter" {
		arg := f.unit.CObj()
		if c.NilUnit {
			arg = nil
		}
		want := uintptr(0)
		if ref.GetterReturn >= 0 {
			want = uintptr(unsafe.Pointer(&f.extra.Sounds[ref.GetterReturn][0]))
		}
		t.Logf("DYING START COMPOSITION CONTROL COMPLETED: mode=%s getter-sound=%d unit=%p", mode, ref.GetterReturn, f.unit)
		got := ccall.CallUPtrPtr(monsterSoundSetPublicC424300(), arg)
		if got != want || f.script.calls != 0 {
			t.Fatalf("original getter native return=%x want=%x script=%d", got, want, f.script.calls)
		}
		compare(&beforeUnit, &beforeUpdate, &beforeExtra)
		t.Logf("DYING START COMPOSITION NATIVE C PASS: mode=%s getter=true result-sound=%d full-records=true", mode, ref.GetterReturn)
		return
	}
	callback := 0
	for _, e := range ref.Result.Events {
		if e.Kind == "callback" {
			callback = int(e.Args[0])
		}
	}
	if ref.RawEAX != uint32(0) && ref.RawEAX != 1 {
		t.Fatal("C composition requires stock int0/1 result")
	}
	const seed = 511
	f.s.Rand.Logic.Reset(seed)
	wantIndex := f.s.Rand.Logic.Index()
	if callback != 0 {
		if got := monsterDieCallbackCProbe54A7D0(server.MonsterDieCallbackKind54A7D0(callback), f.unit); got != 1 {
			t.Fatal("real stock C callback must preserve original int1", got)
		}
		wantIndex = f.s.Rand.Logic.Index()
		compare(&beforeUnit, &beforeUpdate, &beforeExtra)
	}
	expectedUnit, expectedUpdate, expectedExtra := beforeUnit, beforeUpdate, beforeExtra
	want := ref.Result.State
	expectedUnit.ObjClass, expectedUnit.ObjSubClass, expectedUnit.ObjFlags = object.Class(want.Class), object.SubClass(want.Subclass), object.Flags(want.Flags)
	expectedUnit.PosVec = types.Ptf(want.Position[0], want.Position[1])
	expectedUnit.UpdateData = unsafe.Pointer(f.updates[want.Update])
	expectedUpdates := [2]*server.MonsterUpdateData{&expectedUpdate, &expectedExtra.Alternate}
	for n, v := range want.Updates {
		ud := expectedUpdates[n]
		ud.MonsterDef = &f.extra.Definitions[v.Definition]
		ud.SoundSet122 = nil
		if v.Sound >= 0 {
			ud.SoundSet122 = unsafe.Pointer(&f.extra.Sounds[v.Sound][0])
		}
		ud.ScriptDeath = server.ScriptCallback{Flags: v.ScriptFlags, Func: v.ScriptFunc}
	}
	for i, id := range want.Callbacks {
		expectedExtra.Definitions[i].DieFunc228 = functions[id]
	}
	expectedExtra.Sounds = want.Sounds
	check := func() {
		if f.script.calls != 1 || f.s.Rand.Logic.Index() != wantIndex {
			t.Fatal("original cached script/RNG identity", f.script.calls, f.s.Rand.Logic.Index(), wantIndex)
		}
		compare(&expectedUnit, &expectedUpdate, &expectedExtra)
	}
	a, ok := server.GetAIAction(ai.ACTION_DYING).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary dying dispatcher")
	}
	f.s.Rand.Logic.Reset(seed)
	a.Start(f.unit)
	check()
	t.Logf("DYING START COMPOSITION CONTROL COMPLETED: mode=%s original-int=%d callback=%d rng-index=%d unit=%p update=%p", mode, ref.RawEAX, callback, wantIndex, f.unit, entry)
	*f.unit, *f.updates[0], *f.extra = beforeUnit, beforeUpdate, beforeExtra
	f.script.calls = 0
	f.s.Rand.Logic.Reset(seed)
	got := ccall.CallIntPtr(monsterDyingStartPublicC544C40(), f.unit.CObj())
	if uint32(got) != ref.RawEAX {
		t.Fatalf("public original int=%d want=%d", got, ref.RawEAX)
	}
	check()
	t.Logf("DYING START COMPOSITION NATIVE C PASS: mode=%s getter=false original-int=%d callback=%d real-stock-C-result=true real-RNG=true full-records=true", mode, ref.RawEAX, callback)
}
