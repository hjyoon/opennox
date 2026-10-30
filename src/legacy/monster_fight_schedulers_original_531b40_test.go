package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

type schedObject531B40 = server.Object
type schedUpdate531B40 = server.MonsterUpdateData
type schedDef531B40 = server.MonsterDef
type schedServer531B40 = server.Server
type schedObjectType531B40 = server.ObjectType
type schedSlotNative531B40 = server.AIStackItem
type schedShapeKind531B40 = server.ShapeKind

func schedNewServer531B40() *schedServer531B40 { return server.New(nil, nil, strman.New()) }
func schedDefaultReference531B40() string {
	return "../server/testdata/monster_fight_schedulers_original_531b40.json.gz"
}

type schedInput531B40 map[string]json.RawMessage

func (v schedInput531B40) number(k string) uint32 {
	var n uint32
	if raw := v[k]; raw != nil {
		if err := json.Unmarshal(raw, &n); err != nil {
			panic(err)
		}
	}
	return n
}
func (v schedInput531B40) head() int8 {
	var n int8
	if err := json.Unmarshal(v["head"], &n); err != nil {
		panic(err)
	}
	return n
}
func (v schedInput531B40) text(k string) string {
	var s string
	if raw := v[k]; raw != nil {
		if err := json.Unmarshal(raw, &s); err != nil {
			panic(err)
		}
	}
	return s
}
func (v schedInput531B40) vector(k string) []uint32 {
	var p []uint32
	if raw := v[k]; raw != nil {
		if err := json.Unmarshal(raw, &p); err != nil {
			panic(err)
		}
	}
	return p
}

type schedWord531B40 struct {
	Kind   string `json:"kind"`
	Value  uint32 `json:"value"`
	Update int    `json:"update"`
	Index  int    `json:"index"`
}
type schedSlot531B40 struct {
	Action  uint32            `json:"action"`
	Args    []schedWord531B40 `json:"args"`
	Started uint32            `json:"started"`
}
type schedUpdateState531B40 struct {
	Head  int8              `json:"head"`
	Reset []uint32          `json:"reset_fields"`
	Slots []schedSlot531B40 `json:"slots"`
}
type schedPush531B40 struct {
	Action   uint32          `json:"action"`
	Update   int             `json:"unit_update"`
	Returned schedWord531B40 `json:"return_value"`
}
type schedRelated531B40 struct {
	Update   int    `json:"unit_update_after"`
	Returned uint32 `json:"return_u32"`
	Mutation string `json:"mutation"`
}
type schedCancel531B40 struct {
	Cached int    `json:"cached_update"`
	Live   int    `json:"live_update_after"`
	Effect string `json:"effect"`
}
type schedRow531B40 struct {
	Name   string           `json:"name"`
	Entry  uint32           `json:"entry"`
	Input  schedInput531B40 `json:"input"`
	Output struct {
		Returned    schedWord531B40          `json:"return_value"`
		Live        int                      `json:"unit_update"`
		Updates     []schedUpdateState531B40 `json:"updates"`
		Changed     uint32                   `json:"stack_changed"`
		OldRange    uint32                   `json:"old_range_after"`
		Pushes      []schedPush531B40        `json:"pushes"`
		Related     []schedRelated531B40     `json:"related"`
		Cancels     []schedCancel531B40      `json:"cancels"`
		TargetAfter []uint32                 `json:"target_pos_after"`
		RangesAfter []uint32                 `json:"definition_ranges_after"`
	} `json:"output"`
}
type schedObjectGuard531B40 struct {
	Before [64]byte
	Value  schedObject531B40
	After  [64]byte
}
type schedUpdateGuard531B40 struct {
	Before [64]byte
	Value  schedUpdate531B40
	After  [64]byte
}
type schedDefGuard531B40 struct {
	Before [64]byte
	Value  schedDef531B40
	After  [64]byte
}
type schedRegion531B40 struct {
	Pointer unsafe.Pointer
	Size    uintptr
}

func schedSnapshot531B40(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}

type schedFixture531B40 struct {
	server     *schedServer531B40
	units      [2]*schedObjectGuard531B40
	updates    [2]*schedUpdateGuard531B40
	defs       [2]*schedDefGuard531B40
	prototypes [2]schedObject531B40
	regions    []schedRegion531B40
	before     [][]byte
	row        *schedRow531B40
	related    []schedRelated531B40
	cancels    []schedCancel531B40
	pushes     []schedPush531B40
}

func schedNewFixture531B40(t *testing.T) *schedFixture531B40 {
	f := &schedFixture531B40{server: schedNewServer531B40()}
	t.Cleanup(f.server.Close)
	if !f.server.Objs.Init(2) {
		t.Fatal("native object owner allocation")
	}
	t.Cleanup(f.server.Objs.FreeObjects)
	f.server.SetFrame(1400)
	for i := 0; i < 2; i++ {
		f.prototypes[i] = *f.server.Objs.NewObject(&schedObjectType531B40{})
		u, fu := alloc.New(schedObjectGuard531B40{})
		t.Cleanup(fu)
		f.units[i] = u
		d, fd := alloc.New(schedUpdateGuard531B40{})
		t.Cleanup(fd)
		f.updates[i] = d
		m, fm := alloc.New(schedDefGuard531B40{})
		t.Cleanup(fm)
		f.defs[i] = m
		f.regions = append(f.regions, schedRegion531B40{unsafe.Pointer(u), unsafe.Sizeof(*u)}, schedRegion531B40{unsafe.Pointer(d), unsafe.Sizeof(*d)}, schedRegion531B40{unsafe.Pointer(m), unsafe.Sizeof(*m)})
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, r := range f.regions {
			if uintptr(r.Pointer) <= math.MaxUint32 {
				t.Fatal("native fixture must exceed4GiB")
			}
		}
	}
	return f
}
func (f *schedFixture531B40) live() int {
	for i, u := range f.updates {
		if f.units[0].Value.UpdateData == unsafe.Pointer(&u.Value) {
			return i
		}
	}
	panic("unknown live update")
}
func (f *schedFixture531B40) word(v schedWord531B40) uintptr {
	switch v.Kind {
	case "nil":
		return 0
	case "sentinel":
		if v.Value != 1 {
			panic("bad sentinel")
		}
		return 1
	case "u32":
		return uintptr(v.Value)
	case "target":
		return uintptr(unsafe.Pointer(&f.units[1].Value))
	case "slot":
		return uintptr(unsafe.Pointer(&f.updates[v.Update].Value.AIStack[v.Index]))
	default:
		panic("unknown original role " + v.Kind)
	}
}
func (f *schedFixture531B40) reset(row *schedRow531B40) {
	f.row = row
	f.related = nil
	f.cancels = nil
	f.pushes = nil
	c := row.Input
	for _, r := range f.regions {
		for i := range unsafe.Slice((*byte)(r.Pointer), int(r.Size)) {
			unsafe.Slice((*byte)(r.Pointer), int(r.Size))[i] = byte(i*31 + 19)
		}
	}
	for i := 0; i < 2; i++ {
		u := &f.units[i].Value
		*u = f.prototypes[i]
		u.TypeInd = 1
		u.ObjClass = object.Class(c.number("cls"))
		u.ObjSubClass = object.SubClass(c.number("subclass"))
		u.NetCode = 117
		key := "unit"
		if i == 1 {
			key = "target"
		}
		p := c.vector(key + "_pos")
		u.PosVec = types.Pointf{X: math.Float32frombits(p[0]), Y: math.Float32frombits(p[1])}
		u.Shape.Kind = schedShapeKind531B40(c.number(key + "_shape"))
		u.Shape.Circle.R = math.Float32frombits(c.number(key + "_radius"))
		u.Shape.Box.W = math.Float32frombits(c.number(key + "_width"))
		u.Shape.Box.H = math.Float32frombits(c.number(key + "_height"))
		d := &f.updates[i].Value
		*d = schedUpdate531B40{}
		m := &f.defs[i].Value
		*m = schedDef531B40{}
		d.MonsterDef = m
		d.AIStackInd = c.head()
		prefix := ""
		if i == 1 {
			prefix = "alt_"
		}
		d.StatusFlags = object.MonsterStatus(c.number(prefix + "status"))
		d.WeaponEquipFlags = c.number(prefix + "weapons")
		m.MeleeAttackRange112 = math.Float32frombits(c.number(prefix + "melee_bits"))
		m.MissileAttackRange212 = math.Float32frombits(c.number(prefix + "missile_bits"))
		m.MissileName148[0] = byte(c.number(prefix + "missile_first"))
		repeated := uint32(0x71717171) + uint32(i)*0x01010101
		d.Field2 = repeated
		d.Field67 = repeated
		d.Field74 = repeated
		d.Field91 = repeated
		d.Field120_1 = byte(repeated)
		d.Field120_2 = byte(repeated)
		d.Field120_3 = byte(repeated)
		d.Field124 = repeated
		d.Field137 = repeated
		for j := range d.AIStack {
			d.AIStack[j].Action = 1
			for k := range d.AIStack[j].Args {
				d.AIStack[j].Args[k] = uintptr(0x61000000 + uint32(i)*0x10000 + uint32(j)*0x100 + uint32(k)*0x10 + c.number("seed"))
			}
		}
		if d.AIStackInd >= 0 {
			d.AIStack[d.AIStackInd].Action = c.number("action")
			if d.AIStackInd > 0 {
				d.AIStack[0].Action = c.number("root_action")
			}
			key := "started"
			if i == 1 {
				key = "alt_started"
			}
			d.AIStack[d.AIStackInd].Field5 = c.number(key)
		}
	}
	f.units[0].Value.Buffs = c.number("buffs")
	f.units[0].Value.UpdateData = unsafe.Pointer(&f.updates[0].Value)
	f.server.AI.StackChanged = c.number("initial_changed") != 0
	f.before = nil
	for _, r := range f.regions {
		f.before = append(f.before, schedSnapshot531B40(r.Pointer, r.Size))
	}
}
func (f *schedFixture531B40) relatedCall(unit, target *schedObject531B40) bool {
	if unit != &f.units[0].Value || target != &f.units[1].Value {
		panic("related binding changed native arguments")
	}
	c := f.row.Input
	mutation := c.text("related_mutation")
	if mutation == "replace-update" || mutation == "replace-update-and-old-range" {
		unit.UpdateData = unsafe.Pointer(&f.updates[1].Value)
	}
	if mutation == "replace-update-and-old-range" {
		f.defs[0].Value.MissileAttackRange212 = math.Float32frombits(c.number("old_range_after_related"))
	}
	returned := c.number("related_return")
	if returned > 1 {
		panic("noncanonical declared spell result")
	}
	f.related = append(f.related, schedRelated531B40{f.live(), returned, mutation})
	return returned != 0
}

var schedCancelActions531B40 = []ai.ActionType{0, 1, 4, 7, 31, 39, 40, 60, 0xffffffff}

func (f *schedFixture531B40) cancelLookup(action ai.ActionType) func(*schedObject531B40) {
	if _, ok := f.row.Input["cancel_present"]; !ok {
		return nil
	}
	if f.row.Input.number("cancel_present") == 0 {
		return nil
	}
	for _, candidate := range schedCancelActions531B40 {
		if candidate == action {
			return f.cancelCall
		}
	}
	return nil
}
func (f *schedFixture531B40) cancelCall(unit *schedObject531B40) {
	if unit != &f.units[0].Value {
		panic("cancel binding changed native unit")
	}
	cached := f.live()
	c := f.row.Input
	effect := c.text("cancel_effect")
	switch effect {
	case "none":
	case "swap-update":
		unit.UpdateData = unsafe.Pointer(&f.updates[1-cached].Value)
	case "swap-update-and-old-range":
		unit.UpdateData = unsafe.Pointer(&f.updates[1-cached].Value)
		f.defs[0].Value.MissileAttackRange212 = math.Float32frombits(c.number("cancel_range"))
	case "cached-range":
		f.defs[cached].Value.MissileAttackRange212 = math.Float32frombits(c.number("cancel_range"))
	case "cap-cached-head":
		f.updates[cached].Value.AIStackInd = 22
	case "move-target":
		p := c.vector("cancel_target_pos")
		f.units[1].Value.PosVec = types.Pointf{X: math.Float32frombits(p[0]), Y: math.Float32frombits(p[1])}
	default:
		panic("unknown original cancel effect")
	}
	f.cancels = append(f.cancels, schedCancel531B40{cached, f.live(), effect})
}
func (f *schedFixture531B40) observedPush(action ai.ActionType, args ...any) *schedSlotNative531B40 {
	current := f.live()
	item := f.units[0].Value.MonsterPushAction(action, args...)
	result := schedWord531B40{Kind: "nil"}
	if item != nil {
		found := false
		for i, u := range f.updates {
			for j := range u.Value.AIStack {
				if item == &u.Value.AIStack[j] {
					result = schedWord531B40{Kind: "slot", Update: i, Index: j}
					found = true
				}
			}
		}
		if !found {
			panic("push returned outside native records")
		}
	}
	f.pushes = append(f.pushes, schedPush531B40{uint32(action), current, result})
	return item
}

type schedOutcome531B40 struct {
	Returned    uintptr
	ReturnKnown bool
	PushKnown   bool
}

func (f *schedFixture531B40) validate(t *testing.T, outcome schedOutcome531B40, controlOnly bool) {
	want := f.row.Output
	if len(want.Updates) != 2 {
		t.Fatal("original update count")
	}
	expected := make([][]byte, len(f.before))
	for i, b := range f.before {
		expected[i] = append([]byte(nil), b...)
	}
	for i := 0; i < 2; i++ {
		shadow := (*schedUpdateGuard531B40)(unsafe.Pointer(&expected[i*3+1][0]))
		d := &shadow.Value
		state := want.Updates[i]
		if len(state.Reset) != 9 || len(state.Slots) != 24 {
			t.Fatal("original native state dimensions")
		}
		d.AIStackInd = state.Head
		d.Field2 = state.Reset[0]
		d.Field67 = state.Reset[1]
		d.Field74 = state.Reset[2]
		d.Field91 = state.Reset[3]
		d.Field120_1 = byte(state.Reset[4])
		d.Field120_2 = byte(state.Reset[5])
		d.Field120_3 = byte(state.Reset[6])
		d.Field124 = state.Reset[7]
		d.Field137 = state.Reset[8]
		for j, slot := range state.Slots {
			if len(slot.Args) != 4 {
				t.Fatal("original argument count")
			}
			d.AIStack[j].Action = slot.Action
			d.AIStack[j].Field5 = slot.Started
			for k, w := range slot.Args {
				d.AIStack[j].Args[k] = f.word(w)
			}
		}
		definition := &(*schedDefGuard531B40)(unsafe.Pointer(&expected[i*3+2][0])).Value
		if i == 0 {
			definition.MissileAttackRange212 = math.Float32frombits(want.OldRange)
		}
		if len(want.RangesAfter) == 2 {
			definition.MissileAttackRange212 = math.Float32frombits(want.RangesAfter[i])
		}
	}
	(*schedObjectGuard531B40)(unsafe.Pointer(&expected[0][0])).Value.UpdateData = unsafe.Pointer(&f.updates[want.Live].Value)
	if len(want.TargetAfter) == 2 {
		(*schedObjectGuard531B40)(unsafe.Pointer(&expected[3][0])).Value.PosVec = types.Pointf{X: math.Float32frombits(want.TargetAfter[0]), Y: math.Float32frombits(want.TargetAfter[1])}
	}
	if controlOnly {
		for i, r := range f.regions {
			if !bytes.Equal(f.before[i], schedSnapshot531B40(r.Pointer, r.Size)) {
				t.Errorf("control region%d changed", i)
			}
		}
		return
	}
	if outcome.ReturnKnown && outcome.Returned != f.word(want.Returned) {
		t.Errorf("return got=%#x original=%#x role=%s", outcome.Returned, f.word(want.Returned), want.Returned.Kind)
	}
	if f.server.AI.StackChanged != (want.Changed != 0) {
		t.Errorf("stack-changed got=%t original=%d", f.server.AI.StackChanged, want.Changed)
	}
	if !reflect.DeepEqual(f.related, want.Related) && !(len(f.related) == 0 && len(want.Related) == 0) {
		t.Errorf("related trace got=%v original=%v", f.related, want.Related)
	}
	if !reflect.DeepEqual(f.cancels, want.Cancels) && !(len(f.cancels) == 0 && len(want.Cancels) == 0) {
		t.Errorf("cancel trace got=%v original=%v", f.cancels, want.Cancels)
	}
	if outcome.PushKnown && !reflect.DeepEqual(f.pushes, want.Pushes) && !(len(f.pushes) == 0 && len(want.Pushes) == 0) {
		t.Errorf("push trace got=%v original=%v", f.pushes, want.Pushes)
	}
	for i, r := range f.regions {
		actual := schedSnapshot531B40(r.Pointer, r.Size)
		if bytes.Equal(expected[i], actual) {
			continue
		}
		if i%3 == 1 {
			d := &f.updates[i/3].Value
			state := want.Updates[i/3]
			if d.AIStackInd != state.Head {
				t.Errorf("update%d head got=%d original=%d", i/3, d.AIStackInd, state.Head)
			}
			for j, w := range state.Slots {
				for k, v := range w.Args {
					if d.AIStack[j].Args[k] != f.word(v) {
						t.Errorf("update%d slot%d action%d arg%d got=%#x original=%#x", i/3, j, w.Action, k, d.AIStack[j].Args[k], f.word(v))
					}
				}
			}
		}
		t.Errorf("native region%d differs from complete original-derived output shadow or guard", i)
	}
}
func schedLoad531B40(t *testing.T) []schedRow531B40 {
	path := os.Getenv("NOX_FIGHT_SCHEDULER_ORIGINAL_REFERENCE")
	if path == "" {
		path = schedDefaultReference531B40()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "f71aaff515e525c97765bf6c93dbd3d3fcb57475dbcabffd41aab39d38ad93b7" {
		t.Fatal("unsealed original reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string           `json:"schema"`
		Original string           `json:"original_sha256"`
		Control  uint16           `json:"control_word"`
		Rows     []schedRow531B40 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-fight-schedulers-531b40-v2" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || len(ref.Rows) != 4138 {
		t.Fatal("original provenance/dimensions")
	}
	return ref.Rows
}

type schedLegacyServer531B40 struct {
	Server
	fixture *schedFixture531B40
}

func (s *schedLegacyServer531B40) S() *server.Server { return s.fixture.server }
func (s *schedLegacyServer531B40) MonsterFightSchedulerRuntime531B40() server.MonsterFightSchedulerRuntime531B40 {
	return server.MonsterFightSchedulerRuntime531B40{RelatedSpell: s.fixture.relatedCall, CancelAction: s.fixture.cancelLookup}
}
func schedPrepare531B40(f *schedFixture531B40) func() {
	old := GetServer
	GetServer = func() Server { return &schedLegacyServer531B40{fixture: f} }
	return func() { GetServer = old }
}
func schedExecute531B40(f *schedFixture531B40, entry uint32) schedOutcome531B40 {
	result := ccall.CallUPtrPtr2(fightSchedulerPublicCBefore531B40(entry), f.units[0].Value.CObj(), f.units[1].Value.CObj())
	return schedOutcome531B40{Returned: result, ReturnKnown: true}
}
func TestFightSchedulersOriginalActualC531B40(t *testing.T) {
	rows := schedLoad531B40(t)
	f := schedNewFixture531B40(t)
	debug := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(debug) })
	control := os.Getenv("NOX_FIGHT_SCHEDULER_CONTROL_ONLY") == "true"
	for i := range rows {
		row := &rows[i]
		t.Run(row.Name, func(t *testing.T) {
			f.reset(row)
			undo := schedPrepare531B40(f)
			defer undo()
			result := schedOutcome531B40{}
			defer func() {
				if value := recover(); value != nil {
					t.Errorf("native engine panic for original-defined context: %v", value)
				}
				f.validate(t, result, control)
			}()
			if !control {
				result = schedExecute531B40(f, row.Entry)
			}
		})
	}
	t.Logf("original scheduler contexts=%d control-only=%t native-return-ABI=checked", len(rows), control)
}
