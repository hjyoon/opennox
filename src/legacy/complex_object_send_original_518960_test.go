package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type senderNetworkInput518960 struct {
	Player      int32  `json:"player"`
	Report      uint32 `json:"report"`
	HP          bool   `json:"hp_present"`
	CurrentHP   uint16 `json:"hp"`
	PreviousHP  uint16 `json:"previous_hp"`
	Frame       uint32 `json:"frame"`
	LastFrame   uint32 `json:"last_frame"`
	Class       uint32 `json:"object_class"`
	Subclass    uint32 `json:"subclass"`
	Primary     uint32 `json:"primary_code"`
	Secondary   uint32 `json:"secondary_code"`
	X           uint32 `json:"x"`
	Y           uint32 `json:"y"`
	Direction   int16  `json:"direction"`
	Sprite      uint8  `json:"sprite_frame"`
	OtherSprite uint8  `json:"other_sprite_frame"`
	NPC         uint8  `json:"NPC_random"`
	OtherNPC    uint8  `json:"other_NPC_random"`
	RNG         uint32 `json:"rng"`
	Frames      int32  `json:"player_frames"`
	Duration    int32  `json:"player_duration"`
	Return      uint32 `json:"queue_return"`
}

type senderRow518960 struct {
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Output struct {
		Return    uint32           `json:"return_word"`
		Returns   []uint32         `json:"returns"`
		Events    []map[string]any `json:"events"`
		Mutations []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
			Call  int             `json:"call"`
		} `json:"mutations"`
		Previous uint16    `json:"previous_hp_after"`
		RNG      uint32    `json:"RNG_index_after"`
		Caches   [4]uint32 `json:"caches_after"`
		Guards   bool      `json:"full_heap_stack_global_neighbors_nonvolatile_and_FP_control_guards_preserved"`
	} `json:"output"`
}

type senderFixture518960 struct {
	t            *testing.T
	b            *animationFixture533790
	health       *animationGuard533790[server.HealthData]
	frames       *animationGuard533790[[53][2]int32]
	runtime      server.ComplexObjectSendNativeRuntime518960
	row          *senderRow518960
	selector     animationRow533790
	network      senderNetworkInput518960
	phaseB       bool
	call         int
	inMimic      bool
	inWord       string
	events       []map[string]any
	mutations    []animationMutation533790
	lookupCounts map[string]int
	before       [][]byte
	mimicBefore  []byte
}

func (f *senderFixture518960) event(e map[string]any) {
	if f.phaseB {
		e["call"] = f.call
	}
	f.events = append(f.events, e)
}

func senderCanonical518960(v any) any {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		panic(err)
	}
	return out
}

func senderNewFixture518960(t *testing.T) *senderFixture518960 {
	handles.Init()
	t.Cleanup(handles.Release)
	f := &senderFixture518960{t: t, b: animationNewFixture533790(t)}
	f.health = animationAllocate533790[server.HealthData](t, f.b)
	f.frames = animationAllocate533790[[53][2]int32](t, f.b)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantListHeadsC(0, 0)
	setImportantAllocClassC(nil)
	setImportantRecipientMaskC(0)
	owner := GetServer
	GetServer = func() Server {
		if f.inWord != "" {
			f.lookup(f.inWord)
		} else if f.inMimic {
			f.event(map[string]any{"entry": "0x4e3aa0"})
			f.lookup("Mimic")
		}
		return owner()
	}
	t.Cleanup(func() { GetServer = owner })
	oldTypes := zombieEntryTypes534A10
	zombieEntryTypes534A10 = func() server.TypePredicateRuntime534A10 {
		return server.TypePredicateRuntime534A10{
			Read:  func(i int) uint32 { return f.b.cache.Value[i] },
			Write: func(i int, v uint32) { f.b.cache.Value[i] = v },
			Lookup: func(name string) uint32 {
				f.event(map[string]any{"entry": "0x4e3aa0"})
				f.inWord = name
				defer func() { f.inWord = "" }()
				return complexObjectWordIDC518960(name)
			},
		}
	}
	t.Cleanup(func() { zombieEntryTypes534A10 = oldTypes })
	predicate := func(index int) func(*server.Object) uint32 {
		return func(unit *server.Object) uint32 {
			if unit != &f.b.unit.Value {
				f.t.Fatal("wrong predicate native object")
			}
			e := map[string]any{"entry": []string{"0x534840", "0x534a10", "0x534a40"}[index]}
			if f.phaseB {
				live := "entry"
				if unit.UpdateData == unsafe.Pointer(&f.b.updates[1].Value) {
					live = "other"
				}
				e["caches_before"], e["unit_type"], e["unit_class"], e["entry_status"], e["live_update"] = f.b.caches(), unit.TypeInd, uint32(unit.ObjClass), uint32(f.b.updates[0].Value.StatusFlags), live
			}
			f.event(e)
			if index == 0 {
				f.inMimic = true
				defer func() { f.inMimic = false }()
			}
			return animationPredicatePublicC533790(index, unit)
		}
	}
	f.b.runtime = server.MonsterAnimationNativeRuntime533790{Mimic: predicate(0), Plant: predicate(1), Zombie: predicate(2)}
	oldFrames := playerAnimFrames4F9F90
	playerAnimFrames4F9F90 = func(i int) (int, int) {
		if i < 0 || i >= 53 {
			f.t.Fatal("frame index outside original table")
		}
		v := f.frames.Value[i]
		return int(v[0]), int(v[1])
	}
	t.Cleanup(func() { playerAnimFrames4F9F90 = oldFrames })
	f.runtime = server.ComplexObjectSendNativeRuntime518960{
		Frame: func() uint32 { return f.b.srv.Frame() },
		UnitCode: func(u *server.Object) uint32 {
			f.event(map[string]any{"entry": "0x578ac0"})
			return complexObjectUnitCodeC518960(u)
		},
		FloatToInt: func(v float32) int32 {
			f.event(map[string]any{"entry": "0x419a70", "argument": math.Float32bits(v)})
			return complexObjectFloatToIntC518960(v)
		},
		Animation: func(u *server.Object) int32 {
			f.event(map[string]any{"entry": "0x533790"})
			return int32(ccall.CallIntPtr(animationSelectorPublicC533790(), u.CObj()))
		},
		Random: func(min, max int32) int32 {
			f.event(map[string]any{"entry": "0x415fa0", "minimum": uint32(min), "maximum": uint32(max), "index_before": f.b.srv.Rand.Logic.Index()})
			return complexObjectRandomC518960(min, max)
		},
		Frames: func(i int32) (int32, int32) {
			f.event(map[string]any{"entry": "0x4f9f90", "argument": uint32(i)})
			return NPCFramesActualC533D00(i)
		},
		Direction: func(i int32, out *server.IndexedDirectionVector509E20) {
			f.event(map[string]any{"entry": "0x509e20", "argument": uint32(i)})
			complexObjectDirectionC518960(i, out)
		},
		ReportHealth: func(player int32, code uint16, delta int16) {
			f.event(map[string]any{"entry": "0x4d8760"})
			first, _ := importantNativeListC()
			if first != nil {
				f.t.Fatal("health publication queue was not empty")
			}
			complexObjectReportHealthC518960(player, code, delta)
			recipient, data := complexObjectTakeImportantPacketC518960()
			if data != nil {
				f.event(map[string]any{"service": "0x4e5420", "player": uint32(recipient), "packet": hex.EncodeToString(data)})
			}
			first, _ = importantNativeListC()
			if first != nil {
				f.t.Fatal("health publication queue leaked")
			}
		},
		Enqueue: func(player int32, packet [11]byte) int32 {
			f.event(map[string]any{"service": "0x40ef40", "player": uint32(player), "packet": hex.EncodeToString(packet[:])})
			return int32(f.network.Return)
		},
	}
	oldFactory := complexObjectSendRuntimeNative518960
	complexObjectSendRuntimeNative518960 = func() server.ComplexObjectSendNativeRuntime518960 { return f.runtime }
	t.Cleanup(func() { complexObjectSendRuntimeNative518960 = oldFactory })
	return f
}

func (f *senderFixture518960) lookup(name string) {
	f.t.Helper()
	index := -1
	for i, v := range animationNames533790 {
		if v == name {
			index = i
		}
	}
	if index < 0 {
		f.t.Fatal("unknown actual dictionary name", name)
	}
	f.lookupCounts[name]++
	e := map[string]any{"service": "dictionary4E3830", "name": name, "occurrence": f.lookupCounts[name], "caches_before": f.b.caches()}
	mutation := f.selector.Input.Mutation
	if mutation != nil && mutation.Name == name && (mutation.Occurrence == 0 && f.lookupCounts[name] == 1 || mutation.Occurrence == f.lookupCounts[name]) {
		for _, effect := range mutation.Effects {
			f.mutate(effect.Field, effect.Value)
			f.mutations = append(f.mutations, animationMutation533790{Field: effect.Field, Value: effect.Value, Call: f.call})
		}
	}
	actual := uint16(f.b.srv.Types.IndByID(name))
	found := actual != 0
	if found != f.selector.Input.Found[index] || found && actual != f.selector.Input.Lookup[index] {
		f.t.Fatal("real registry does not match original dictionary input", name, actual)
	}
	e["found"], e["index"], e["caches_after_declared_mutations"] = found, any(nil), f.b.caches()
	if found {
		e["index"] = actual
	}
	f.event(e)
}

func (f *senderFixture518960) mutate(field string, raw json.RawMessage) {
	f.t.Helper()
	u := &f.b.unit.Value
	entry, other := &f.b.updates[0].Value, &f.b.updates[1].Value
	var number uint32
	var flag bool
	var text string
	switch field {
	case "live_update":
		if err := json.Unmarshal(raw, &text); err != nil {
			f.t.Fatal(err)
		}
		switch text {
		case "entry":
			u.UpdateData = unsafe.Pointer(entry)
		case "other":
			u.UpdateData = unsafe.Pointer(other)
		default:
			f.t.Fatal("unknown live update")
		}
		return
	case "entry_enemy", "other_enemy":
		if err := json.Unmarshal(raw, &flag); err != nil {
			f.t.Fatal(err)
		}
		d := entry
		if field == "other_enemy" {
			d = other
		}
		d.CurrentEnemy = nil
		if flag {
			d.CurrentEnemy = &f.b.target.Value
		}
		return
	}
	if err := json.Unmarshal(raw, &number); err != nil {
		f.t.Fatal(err)
	}
	switch field {
	case "unit_type":
		u.TypeInd = uint16(number)
	case "unit_class":
		u.ObjClass = object.Class(number)
	case "unit_subclass":
		u.ObjSubClass = object.SubClass(number)
	case "unit_direction":
		u.Direction1 = server.Dir16(number)
	case "unit_x":
		u.PosVec.X = math.Float32frombits(number)
	case "unit_y":
		u.PosVec.Y = math.Float32frombits(number)
	case "primary_code":
		u.NetCode = number
	case "secondary_code":
		u.Extent = number
	case "entry_status":
		entry.StatusFlags = object.MonsterStatus(number)
	case "other_status":
		other.StatusFlags = object.MonsterStatus(number)
	case "entry_action":
		item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&entry.AIStack[0]), int(f.selector.Input.Head)*int(unsafe.Sizeof(entry.AIStack[0]))))
		item.Action = number
	case "entry_sprite_frame":
		entry.Field120_1 = uint8(number)
	case "other_sprite_frame":
		other.Field120_1 = uint8(number)
	case "entry_NPC_random":
		entry.Field523_2 = uint8(number)
	case "other_NPC_random":
		other.Field523_2 = uint8(number)
	case "cache_0", "cache_1", "cache_2", "cache_3":
		f.b.setCache(int(field[len(field)-1]-'0'), number)
	default:
		f.t.Fatal("unknown original declared mutation", field)
	}
}

func (f *senderFixture518960) prepare(row *senderRow518960, phaseB bool) {
	f.t.Helper()
	f.row = row
	f.phaseB = phaseB
	f.selector = animationRow533790{}
	f.network = senderNetworkInput518960{}
	if phaseB {
		var input struct {
			Selector animationInput533790     `json:"selector_input"`
			Network  senderNetworkInput518960 `json:"network_input"`
		}
		if err := json.Unmarshal(row.Input, &input); err != nil {
			f.t.Fatal(err)
		}
		f.selector.Input, f.network = input.Selector, input.Network
	} else {
		if err := json.Unmarshal(row.Input, &f.selector.Input); err != nil {
			f.t.Fatal(err)
		}
		if err := json.Unmarshal(row.Input, &f.network); err != nil {
			f.t.Fatal(err)
		}
		f.selector.Input.Class = f.network.Class
		f.selector.Input.Flags = 0x1000000
		f.selector.Input.Repeat = 1
		f.selector.Input.Caches = [4]uint32{11, 22, 33, 44}
		f.selector.Input.Lookup = [4]uint16{11, 22, 33, 44}
		f.selector.Input.Found = [4]bool{true, true, true, true}
	}
	f.b.t = f.t
	f.b.recording = false
	f.b.prepare(&f.selector)
	n := f.network
	u := &f.b.unit.Value
	u.ObjSubClass = object.SubClass(n.Subclass)
	u.NetCode = n.Primary
	u.Extent = n.Secondary
	u.Frame134 = n.LastFrame
	u.PosVec.X, u.PosVec.Y = math.Float32frombits(n.X), math.Float32frombits(n.Y)
	u.Direction1 = server.Dir16(n.Direction)
	f.health.Value = server.HealthData{Cur: n.CurrentHP, Max: 65535}
	u.HealthData = nil
	if n.HP {
		u.HealthData = &f.health.Value
	}
	for i, g := range f.b.updates {
		d := &g.Value
		d.Field120_1, d.Field523_2 = n.Sprite, n.NPC
		if i == 1 {
			d.Field120_1, d.Field523_2 = n.OtherSprite, n.OtherNPC
		}
		d.HealthGraph103[n.Player] = n.PreviousHP
		item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&d.AIStack[0]), int(f.selector.Input.Head)*int(unsafe.Sizeof(d.AIStack[0]))))
		item.Action = f.selector.Input.Action
	}
	for i := range f.frames.Value {
		f.frames.Value[i] = [2]int32{0x12345678, 0x23456789}
	}
	f.frames.Value[50] = [2]int32{n.Frames, n.Duration}
	f.b.srv.SetFrame(n.Frame)
	f.b.srv.Rand.Logic.Reset(int(n.RNG))
	f.events = []map[string]any{}
	f.mutations = []animationMutation533790{}
	f.lookupCounts = map[string]int{}
	f.call = 0
	f.before = make([][]byte, len(f.b.regions))
	for i, r := range f.b.regions {
		f.before[i] = animationSnapshot533790(r.Ptr, r.Size)
	}
	f.mimicBefore = animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)
}

func (f *senderFixture518960) patch(field string, raw json.RawMessage) {
	f.t.Helper()
	u := &f.b.unit.Value
	entry, other := &f.b.updates[0].Value, &f.b.updates[1].Value
	var n uint32
	var b bool
	var text string
	switch field {
	case "live_update":
		if err := json.Unmarshal(raw, &text); err != nil {
			f.t.Fatal(err)
		}
		ptr := unsafe.Pointer(entry)
		if text == "other" {
			ptr = unsafe.Pointer(other)
		} else if text != "entry" {
			f.t.Fatal(text)
		}
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.UpdateData, ptr)
		return
	case "entry_enemy", "other_enemy":
		if err := json.Unmarshal(raw, &b); err != nil {
			f.t.Fatal(err)
		}
		d := entry
		if field == "other_enemy" {
			d = other
		}
		var v *server.Object
		if b {
			v = &f.b.target.Value
		}
		animationPatchValue533790(f.t, f.before, f.b.regions, &d.CurrentEnemy, v)
		return
	}
	if err := json.Unmarshal(raw, &n); err != nil {
		f.t.Fatal(err)
	}
	switch field {
	case "unit_type":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.TypeInd, uint16(n))
	case "unit_class":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.ObjClass, object.Class(n))
	case "unit_subclass":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.ObjSubClass, object.SubClass(n))
	case "unit_direction":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.Direction1, server.Dir16(n))
	case "unit_x":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.PosVec.X, math.Float32frombits(n))
	case "unit_y":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.PosVec.Y, math.Float32frombits(n))
	case "primary_code":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.NetCode, n)
	case "secondary_code":
		animationPatchValue533790(f.t, f.before, f.b.regions, &u.Extent, n)
	case "entry_status":
		animationPatchValue533790(f.t, f.before, f.b.regions, &entry.StatusFlags, object.MonsterStatus(n))
	case "other_status":
		animationPatchValue533790(f.t, f.before, f.b.regions, &other.StatusFlags, object.MonsterStatus(n))
	case "entry_action":
		item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&entry.AIStack[0]), int(f.selector.Input.Head)*int(unsafe.Sizeof(entry.AIStack[0]))))
		animationPatchValue533790(f.t, f.before, f.b.regions, &item.Action, n)
	case "entry_sprite_frame":
		animationPatchValue533790(f.t, f.before, f.b.regions, &entry.Field120_1, uint8(n))
	case "other_sprite_frame":
		animationPatchValue533790(f.t, f.before, f.b.regions, &other.Field120_1, uint8(n))
	case "entry_NPC_random":
		animationPatchValue533790(f.t, f.before, f.b.regions, &entry.Field523_2, uint8(n))
	case "other_NPC_random":
		animationPatchValue533790(f.t, f.before, f.b.regions, &other.Field523_2, uint8(n))
	case "cache_0", "cache_1", "cache_2", "cache_3":
	default:
		f.t.Fatal("unknown original mutation patch", field)
	}
}

func (f *senderFixture518960) check(control bool) {
	f.t.Helper()
	o := f.row.Output
	if !o.Guards {
		f.t.Fatal("original guard contract missing")
	}
	if !control {
		if !reflect.DeepEqual(senderCanonical518960(f.events), senderCanonical518960(o.Events)) {
			f.t.Errorf("actual helper/dictionary/publication event trace differs from original: got=%v want=%v", f.events, o.Events)
		}
		expectedMutations := []animationMutation533790{}
		for _, m := range o.Mutations {
			expectedMutations = append(expectedMutations, animationMutation533790{Field: m.Field, Value: m.Value, Call: m.Call})
			f.patch(m.Field, m.Value)
		}
		if !reflect.DeepEqual(f.mutations, expectedMutations) {
			f.t.Error("declared mutation order or occurrence differs")
		}
		if !f.phaseB {
			animationPatchValue533790(f.t, f.before, f.b.regions, &f.b.updates[0].Value.HealthGraph103[f.network.Player], o.Previous)
		}
		expectedCaches := f.selector.Input.Caches
		if f.phaseB {
			expectedCaches = o.Caches
		}
		if f.b.caches() != expectedCaches {
			f.t.Errorf("cache=%v original=%v", f.b.caches(), expectedCaches)
		}
		for i := range f.b.cache.Value {
			animationPatchValue533790(f.t, f.before, f.b.regions, &f.b.cache.Value[i], expectedCaches[i+1])
		}
		*(*uint32)(unsafe.Pointer(&f.mimicBefore[64])) = expectedCaches[0]
		if uint32(f.b.srv.Rand.Logic.Index()) != o.RNG {
			f.t.Errorf("RNG=%d original=%d", f.b.srv.Rand.Logic.Index(), o.RNG)
		}
	}
	for i, r := range f.b.regions {
		if !bytes.Equal(f.before[i], unsafe.Slice((*byte)(r.Ptr), int(r.Size))) {
			f.t.Errorf("native guarded unit/two-MUD/target/cache/health/frame region%d differs beyond original outputs", i)
		}
	}
	if !bytes.Equal(f.mimicBefore, animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)) {
		f.t.Error("Mimic DWORD or neighbor guards differ")
	}
	if f.b.srv.Frame() != f.network.Frame {
		f.t.Error("frame changed")
	}
	first, last := importantNativeListC()
	if first != nil || last != nil {
		f.t.Error("native health publication queue not drained")
	}
}

func senderRunOriginal518960(t *testing.T, kind string, invoke func(*senderFixture518960) int32) {
	root := os.Getenv("NOX_COMPLEX_OBJECT_SEND_REFERENCE_DIR")
	if root == "" {
		root = "../server/testdata/complex_object_send_518960"
	}
	profiles := []struct {
		File, Seal, Phase string
		Contexts, Calls   int
		B                 bool
	}{
		{"phaseA-startup-v2.json.gz", "f80a75388bee334824454bfc763f2819dea1573e822e6de5ddf1b639daf7f174", "A", 14110, 14110, false},
		{"phaseB-startup-v1.json.gz", "21c92133f0175135dafae5cfc11e48a81f81df9f9e5669446fdeed9d157a0655", "B", 15004, 15392, true},
	}
	f := senderNewFixture518960(t)
	control := os.Getenv("NOX_COMPLEX_OBJECT_SEND_CONTROL_ONLY") == "true"
	calls := 0
	for _, profile := range profiles {
		raw, err := os.ReadFile(root + "/" + profile.File)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		if hex.EncodeToString(sum[:]) != profile.Seal {
			t.Fatal("original sender reference seal")
		}
		z, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		var ref struct {
			Schema   string            `json:"schema"`
			Original string            `json:"original_sha256"`
			Control  uint16            `json:"control_word"`
			Rows     []senderRow518960 `json:"rows"`
		}
		if err = json.NewDecoder(z).Decode(&ref); err != nil {
			z.Close()
			t.Fatal(err)
		}
		z.Close()
		schema := "original-complex-object-send-phaseA-v1"
		if profile.B {
			schema = "original-complex-object-send-phaseB-v1"
		}
		if ref.Schema != schema || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || len(ref.Rows) != profile.Contexts {
			t.Fatal("original sender provenance/dimensions")
		}
		phaseCalls := 0
		for _, row := range ref.Rows {
			t.Run(profile.Phase+"/"+row.Name, func(t *testing.T) {
				f.t = t
				f.prepare(&row, profile.B)
				if !control {
					expected := row.Output.Returns
					if !profile.B {
						expected = []uint32{row.Output.Return}
					}
					if len(expected) != f.selector.Input.Repeat {
						t.Fatal("original sender call dimensions")
					}
					for call, want := range expected {
						f.call = call
						f.event(map[string]any{"entry": "0x518960"})
						got := uint32(invoke(f))
						calls++
						phaseCalls++
						if got != want {
							t.Errorf("raw sender return call%d=%#08x original=%#08x", call, got, want)
						}
					}
				}
				f.check(control)
			})
		}
		if !control && phaseCalls != profile.Calls {
			t.Fatal("actual phase sender call count", profile.Phase, phaseCalls)
		}
	}
	if !control && calls != 29502 {
		t.Fatal("actual sender call count", calls)
	}
	t.Logf("original %s complex-object sender phaseA=14110/14110 phaseB=15004/15392 calls=%d control-only=%t actual-C-scalar-helpers-real-registry-RNG-and-native-guards=true declared-dictionary-input-and-network-publication-boundaries=true", kind, calls, control)
}

func TestComplexObjectSendOriginalActualC518960(t *testing.T) {
	senderRunOriginal518960(t, "actual C", func(f *senderFixture518960) int32 {
		return complexObjectSendPublicC518960(f.network.Player, &f.b.unit.Value, int32(f.network.Report))
	})
}
func TestComplexObjectSendOriginalGo518960(t *testing.T) {
	senderRunOriginal518960(t, "Go", func(f *senderFixture518960) int32 {
		return server.ComplexObjectSendNative518960(f.network.Player, &f.b.unit.Value, int32(f.network.Report), f.runtime)
	})
}
