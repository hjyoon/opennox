//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/server"
)

type rddPinnedAllocation4E0B30 struct {
	Role string
	Base unsafe.Pointer
	Size uintptr
}

// Fourteen canonical input roles, one identity-only opaque attribution carrier,
// and one independent clock allocation. The carrier preserves the distinct
// original0x10203040 alias group; it does not model that address's memory domain.
// Object fields are typed Go pointers. UpdateData is deliberately a pinned
// pointer-free composite buffer: its player/monster scalar views can change
// after a class mutation without storing fake pointers in a Go struct.
type rddPinnedGraph4E0B30 struct {
	Objects     map[string]*server.Object
	Wands       map[string]*server.WandUseData
	Buffers     map[string][]byte
	Health      *server.HealthData
	Opaque      *server.Object
	Clock       *uint32
	Roles       map[string]unsafe.Pointer
	Canonical   map[uint32]unsafe.Pointer
	Allocations []rddPinnedAllocation4E0B30
	Definitions []ddiField4E0B30
	Offsets     [27]uintptr
	Case        ddiCase4E0B30
	Input       rddInput4E0B30
	Pin         runtime.Pinner
}

func rddPinnedRequireEngine4E0B30(t *testing.T) {
	t.Helper()
	if rawParentABI4E0B30() != 2 || rawDefaultDamageNativeEngine4E0B30 == nil ||
		reflect.ValueOf(rawDefaultDamageNativeEngine4E0B30).Pointer() != reflect.ValueOf(rawDefaultDamageNative4E0B30).Pointer() {
		t.Fatal("actual registered raw engine and public typed C ABI2 required")
	}
	if rawDefaultDamageNativeObserverStack4E0B30 != nil || damagePredicateRawObserve4E1400 != nil ||
		rawDefaultDamageCurrentContext4E0B30 != nil || rawDefaultDamageContextStack4E0B30 != nil {
		t.Fatal("fresh nil ordinary observers required")
	}
}

func rddPinnedNewGraph4E0B30(t *testing.T, fixture *ddiFixture4E0B30, chosen ddiCase4E0B30, input rddInput4E0B30) *rddPinnedGraph4E0B30 {
	t.Helper()
	g := &rddPinnedGraph4E0B30{
		Objects: map[string]*server.Object{}, Wands: map[string]*server.WandUseData{}, Buffers: map[string][]byte{},
		Roles: map[string]unsafe.Pointer{}, Canonical: map[uint32]unsafe.Pointer{0: nil},
		Definitions: fixture.Fields, Offsets: rawDefaultDamageNativeOffsets4E0B30(), Case: chosen, Input: input,
		Health: new(server.HealthData),
	}
	clockAllocation := new([8]uint32) // separate backing allocation; avoid tiny-object coallocation.
	clockAllocation[1] = 0x84ea0484
	g.Clock = &clockAllocation[0]
	var pins []any
	add := func(role string, pointer unsafe.Pointer, size uintptr, owner any) {
		if pointer == nil || uintptr(pointer) <= math.MaxUint32 {
			t.Fatal("every pinned allocation must exercise high native address bits", role, pointer)
		}
		g.Roles[role] = pointer
		g.Allocations = append(g.Allocations, rddPinnedAllocation4E0B30{role, pointer, size})
		pins = append(pins, owner)
	}
	for _, role := range ddiRoles4E0B30[:6] {
		unit := new(server.Object)
		unit.Field38 = 0x6da5a56d // scalar-only canary; never fill pointer slots.
		g.Objects[role] = unit
		add(role, unsafe.Pointer(unit), unsafe.Sizeof(*unit), unit)
	}
	updateSize := max(int(unsafe.Sizeof(server.MonsterUpdateData{})), int(unsafe.Sizeof(server.PlayerUpdateData{})))
	for _, role := range []string{"target_update", "source_update", "alt_update", "soundset", "sound_resource"} {
		size := updateSize
		if role == "soundset" || role == "sound_resource" {
			size = 64
		}
		buffer := make([]byte, size) // pointer-free, initially zero; no random pointer-shaped bytes.
		g.Buffers[role] = buffer
		add(role, unsafe.Pointer(&buffer[0]), uintptr(len(buffer)), &buffer[0])
	}
	for _, role := range []string{"weapon_use", "alt_use"} {
		wand := new(server.WandUseData)
		g.Wands[role] = wand
		add(role, unsafe.Pointer(wand), unsafe.Sizeof(*wand), wand)
	}
	add("health", unsafe.Pointer(g.Health), unsafe.Sizeof(*g.Health), g.Health)
	add("clock", unsafe.Pointer(g.Clock), unsafe.Sizeof(*clockAllocation), clockAllocation)
	for role, original := range input.Roles {
		pointer, ok := g.Roles[role]
		if !ok || original == 0 || g.Canonical[original] != nil {
			t.Fatal("selected full aligned graph requires distinct known role identities", role)
		}
		shape := chosen.Input.Domain.Shapes[role]
		if shape.Unreadable || shape.Unaligned != 0 || shape.Readable != nil {
			t.Fatal("this Go-owned integration scope requires full aligned role allocations", role)
		}
		g.Canonical[original] = pointer
	}
	for original, kind := range map[uint32]uint32{0x23001000: 2, 0x23002000: 3} {
		pointer, bits := ddiCallback4E0B30(kind)
		if pointer == nil || uintptr(pointer) != bits || bits <= math.MaxUint32 {
			t.Fatal("actual foreign C function token identity")
		}
		g.Canonical[original] = pointer
	}
	// Input-only audit: every selected input has the same opaque token in
	// Obj130 of its six Object roles. It is not one of those roles, a callback,
	// or a dereferenced address. Bind it to one additional typed Go allocation.
	// Never turn its numeric bits into a Go pointer or alias it to a known role.
	const opaqueToken = uint32(0x10203040)
	const opaqueRole = "opaque-attribution-10203040"
	if _, exists := g.Canonical[opaqueToken]; exists {
		t.Fatal("opaque identity must remain distinct from all existing canonical tokens")
	}
	g.Opaque = new(server.Object)
	g.Opaque.Field38 = 0x5ec01020 // scalar-only carrier canary.
	for _, pointer := range g.Canonical {
		if pointer == unsafe.Pointer(g.Opaque) {
			t.Fatal("opaque carrier must not alias any canonical role or foreign function")
		}
	}
	g.Objects[opaqueRole] = g.Opaque
	add(opaqueRole, unsafe.Pointer(g.Opaque), unsafe.Sizeof(*g.Opaque), g.Opaque)
	g.Canonical[opaqueToken] = unsafe.Pointer(g.Opaque)
	opaqueAliases := 0
	for _, field := range chosen.Fields {
		data, err := hex.DecodeString(field.Hex)
		if err != nil || len(data) != fixture.Fields[field.Field].Width {
			t.Fatal("full independent initial field bytes", field)
		}
		if field.Field == 9 && binary.LittleEndian.Uint32(data) == opaqueToken {
			opaqueAliases++
		}
		g.put(t, field.Role, field.Field, data)
	}
	if opaqueAliases != 6 {
		t.Fatal("selected original Obj130 opaque alias group must contain all six object roles")
	}
	clockFound := false
	for _, init := range input.ImageInitializers {
		if init.Address == 0x84ea04 {
			data, err := hex.DecodeString(init.Hex)
			if err != nil || len(data) != 4 {
				t.Fatal("independent DWORD clock input")
			}
			*g.Clock = binary.LittleEndian.Uint32(data)
			clockFound = true
		}
	}
	if !clockFound || len(g.Allocations) != 16 || len(pins) != 16 {
		t.Fatal("all14roles plus opaque identity carrier and actual independent clock required")
	}
	// Pinner is not recursive. Pin every object, each backing array, each wand,
	// HealthData, opaque attribution carrier and clock separately, including
	// initially unreachable alt_use. Every nonnil Go destination is pinned.
	for _, value := range pins {
		g.Pin.Pin(value)
	}
	t.Cleanup(func() { g.Pin.Unpin(); runtime.KeepAlive(g) })
	return g
}

func (g *rddPinnedGraph4E0B30) pointer(t *testing.T, word uint32) unsafe.Pointer {
	t.Helper()
	pointer, ok := g.Canonical[word]
	if !ok {
		t.Fatal("unknown pointer token: never synthesize a Go heap pointer", word)
	}
	return pointer
}

func (g *rddPinnedGraph4E0B30) put(t *testing.T, role string, id int, data []byte) {
	t.Helper()
	definition := g.Definitions[id]
	if len(data) != definition.Width {
		t.Fatal("exact independent field storage width")
	}
	base, ok := g.Roles[role]
	if !ok {
		t.Fatal("unknown Go role", role)
	}
	if definition.pointer() {
		unit := g.Objects[role]
		if unit == nil {
			t.Fatal("pointer fields only belong to the typed Object view")
		}
		pointer := g.pointer(t, binary.LittleEndian.Uint32(data))
		isObject := pointer == nil
		for _, object := range g.Objects {
			isObject = isObject || pointer == unsafe.Pointer(object)
		}
		switch id {
		case 8:
			if !isObject || pointer == unsafe.Pointer(g.Opaque) {
				t.Fatal("owner destination must be a separately pinned typed Object")
			}
			unit.ObjOwner = (*server.Object)(pointer)
		case 9:
			if !isObject {
				t.Fatal("attribution destination must be a separately pinned typed Object")
			}
			unit.Obj130 = (*server.Object)(pointer)
		case 14:
			if pointer != nil && pointer != unsafe.Pointer(g.Health) {
				t.Fatal("actual pinned HealthData destination")
			}
			unit.HealthData = (*server.HealthData)(pointer)
		case 15, 16:
			if pointer != nil && pointer != g.Canonical[0x23001000] && pointer != g.Canonical[0x23002000] {
				t.Fatal("selected callback field needs a declared foreign function endpoint")
			}
			if id == 15 {
				unit.Damage = pointer
			} else {
				unit.DamageSound = pointer
			}
		case 17:
			if pointer != nil && pointer != g.Roles["weapon_use"] && pointer != g.Roles["alt_use"] {
				t.Fatal("use destination must be an individually pinned typed WandUseData")
			}
			unit.UseData.Ptr = pointer
		case 18:
			if pointer != nil && pointer != g.Roles["target_update"] && pointer != g.Roles["source_update"] && pointer != g.Roles["alt_update"] {
				t.Fatal("update destination must be an individually pinned pointer-free composite buffer")
			}
			unit.UpdateData = pointer
		default:
			t.Fatal("undeclared typed pointer field", id)
		}
		return
	}
	address := unsafe.Add(base, g.Offsets[id])
	switch len(data) {
	case 1:
		*(*uint8)(address) = data[0]
	case 2:
		*(*uint16)(address) = binary.LittleEndian.Uint16(data)
	case 4:
		*(*uint32)(address) = binary.LittleEndian.Uint32(data)
	default:
		t.Fatal("undeclared pointer-free scalar width")
	}
	runtime.KeepAlive(g)
}

func (g *rddPinnedGraph4E0B30) snapshot() map[string][]byte {
	out := map[string][]byte{}
	for _, allocation := range g.Allocations {
		out[allocation.Role] = bytes.Clone(unsafe.Slice((*byte)(allocation.Base), allocation.Size))
	}
	runtime.KeepAlive(g)
	return out
}

func (g *rddPinnedGraph4E0B30) compare(t *testing.T, expected map[string][]byte) {
	t.Helper()
	for _, allocation := range g.Allocations {
		if !bytes.Equal(unsafe.Slice((*byte)(allocation.Base), allocation.Size), expected[allocation.Role]) {
			t.Fatal("complete pinned native allocation differs", allocation.Role)
		}
		if uintptr(g.Roles[allocation.Role]) != uintptr(allocation.Base) {
			t.Fatal("pinned identity changed", allocation.Role)
		}
	}
	runtime.KeepAlive(g)
}

// This projects only selected PE fields into independent expected byte slices.
// Pointer bits are written ONLY to these byte slices, never to a Go pointer
// field. All other native bytes retain their actual before canary. It is not
// an assertion of PE stack/caller/full-image or opaque PE-byte equivalence.
func (g *rddPinnedGraph4E0B30) expectedAfter(t *testing.T, bundle *rddBundle4E0B30, row *rddRow4E0B30) map[string][]byte {
	t.Helper()
	expected := g.snapshot()
	for _, field := range g.Case.Fields {
		definition := g.Definitions[field.Field]
		address := uint64(g.Input.Roles[field.Role]) + uint64(definition.Offset)
		var data []byte
		for _, ref := range row.MemoryAfter {
			start := mustRDDNumber4E0B30(ref.Address)
			if address >= start && address+uint64(definition.Width) <= start+uint64(ref.Size) {
				memory, err := bundle.snapshot(ref.Snapshot)
				if err != nil {
					t.Fatal(err)
				}
				data = memory[address-start : address-start+uint64(definition.Width)]
				break
			}
		}
		if len(data) != definition.Width {
			t.Fatal("original complete AFTER field unavailable", field.Role, field.Field)
		}
		offset := g.Offsets[field.Field]
		if definition.pointer() {
			pointer := g.pointer(t, binary.LittleEndian.Uint32(data))
			binary.LittleEndian.PutUint64(expected[field.Role][offset:offset+8], uint64(uintptr(pointer)))
		} else {
			copy(expected[field.Role][offset:offset+uintptr(len(data))], data)
		}
	}
	return expected
}

func (g *rddPinnedGraph4E0B30) apply(t *testing.T, effects []rddEffect4E0B30) {
	t.Helper()
	for _, effect := range effects {
		if effect.Address == nil || effect.Arg != nil || effect.Role != "" {
			t.Fatal("selected stimulus must be an explicit independent address")
		}
		data, err := hex.DecodeString(effect.Hex)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, field := range g.Case.Fields {
			definition := g.Definitions[field.Field]
			if uint64(g.Input.Roles[field.Role])+uint64(definition.Offset) == uint64(*effect.Address) && len(data) == definition.Width {
				g.put(t, field.Role, field.Field, data)
				found = true
				break
			}
		}
		if !found {
			t.Fatal("stimulus is outside the selected typed Go graph fields")
		}
	}
}

type rddPinnedDefaultWorld4E0B30 struct {
	Server
	Value *server.Server
}

func (s *rddPinnedDefaultWorld4E0B30) S() *server.Server { return s.Value }

func TestRawDefaultDamageDefaultPinnedGoGraph4E0B30(t *testing.T) {
	rddPinnedRequireEngine4E0B30(t)
	fixture, err := ddiLoadFixture4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	chosen := rddFixtureIndex4E0B30(fixture)["stimulus-child-use-pointer-cache-variant-1"]
	for _, name := range []string{"nil-target", "real-Buff23-and-clock5"} {
		t.Run("C/"+name, func(t *testing.T) {
			g := rddPinnedNewGraph4E0B30(t, fixture, chosen, rddInput4E0B30{ddiInput4E0B30: chosen.Input, ImageInitializers: []struct {
				Address uint32 `json:"address"`
				Hex     string `json:"hex"`
			}{{0x84ea04, "05000000"}}})
			s := server.NewHigh(nil, nil, strman.New())
			s.SetFrame(5) // actual fixture world; its real lazy frame read avoids audio.
			previous := GetServer
			GetServer = func() Server { return &rddPinnedDefaultWorld4E0B30{Value: s} }
			defer func() { GetServer = previous; s.Close() }()
			target := unsafe.Pointer(g.Objects["target"])
			g.Objects["target"].Buffs = uint32(1) << 23
			if name == "nil-target" {
				target = nil
			}
			before := g.snapshot()
			runtime.GC()
			result := rawParentPublic4E0B30(target, g.Roles["source"], g.Roles["weapon"], 27, 9)
			runtime.GC()
			if result != 1 || s.Frame() != 5 {
				t.Fatal("actual public C early result and unchanged real world clock", result)
			}
			rddPinnedRequireEngine4E0B30(t)
			g.compare(t, before)
			t.Logf("RAW_DEFAULT_DAMAGE_PINNED_DEFAULT_C case=%s pinned=16 opaque_carriers=1 observers=nil real_Buff_helper=%t real_clock=%t audio_body=false", name, name != "nil-target", name != "nil-target")
			runtime.KeepAlive(g)
		})
	}
}

func TestRawDefaultDamagePinnedGoGraphGC4E0B30(t *testing.T) {
	rddPinnedRequireEngine4E0B30(t)
	fixture, err := ddiLoadFixture4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := rddOpenBundle4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	defer bundle.close()
	names := []string{"stimulus-early-class-variant-0", "stimulus-child-use-pointer-cache-variant-1", "stimulus-sound-callback-cache-variant-0"}
	chosenRows := map[string]*rddRow4E0B30{}
	if err = bundle.rows(bundle.Index.Runs["baseline1034"], func(raw json.RawMessage) error {
		var identity struct{ Name, Mode string }
		if e := json.Unmarshal(raw, &identity); e != nil {
			return e
		}
		for _, name := range names {
			if identity.Name == name && identity.Mode == "startup-53-chop" {
				var row rddRow4E0B30
				if e := json.Unmarshal(raw, &row); e != nil {
					return e
				}
				if chosenRows[name] != nil {
					return fmt.Errorf("duplicate frozen Goheap source row")
				}
				chosenRows[name] = &row
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, public := range []bool{false, true} {
		backend := "Go"
		if public {
			backend = "C"
		}
		for _, name := range names {
			t.Run(backend+"/"+name, func(t *testing.T) {
				row := chosenRows[name]
				if row == nil || row.Status != "returned" || row.Control != 0x0e3f || len(row.UsedReads) != 1 || row.UsedReads[0] != 0 {
					t.Fatal("selected independent stimulus needs an actually completed original witness")
				}
				var input rddInput4E0B30
				if e := json.Unmarshal(row.Input, &input); e != nil {
					t.Fatal(e)
				}
				chosen := rddFixtureIndex4E0B30(fixture)[name]
				if chosen.InputSHA != row.InputSHA || len(input.AfterReads) != 1 || len(input.AfterServices) != 0 {
					t.Fatal("exact independent input/schedule scope")
				}
				g := rddPinnedNewGraph4E0B30(t, fixture, chosen, input)
				expected := g.expectedAfter(t, bundle, row)
				stimulus := input.AfterReads[0]
				critical := uint32(mustRDDNumber4E0B30(stimulus.PC))
				if stimulus.Occurrence != 1 {
					t.Fatal("single input after-read occurrence required")
				}
				gc, mutations, criticalReads, followupReads, callbackGC := 0, 0, 0, 0, 0
				forceGC := func() { runtime.GC(); gc++ }
				mutate := func() { g.apply(t, stimulus.Effects); mutations++; forceGC() }
				callbackPending := false
				var callbackSite uint32
				var callbackPointer unsafe.Pointer
				occurrences := map[uint32]int{}
				specs := map[uint32]rddCallSpec4E0B30{}
				for site, spec := range bundle.Calls {
					specs[uint32(mustRDDNumber4E0B30(site))] = spec
				}
				observer := &rawDefaultDamageNativeObserver4E0B30{ControlWord: uint16(row.Control)}
				observer.Frame = func(site uint32, width int) uint32 {
					if width == 1 {
						return uint32(uint8(*g.Clock))
					}
					if width != 4 {
						t.Fatal("clock width")
					}
					return *g.Clock
				}
				observer.Operand = func(value rawDefaultDamageNativeOperand4E0B30) {
					if value.Base == unsafe.Pointer(g.Opaque) {
						t.Fatal("opaque identity carrier is never a raw-engine memory domain")
					}
					if value.Write {
						return
					}
					if value.Site == critical {
						if !value.Complete {
							forceGC()
							return
						}
						criticalReads++
						if name == names[0] && value.Value.Word != 2 {
							t.Fatal("independent cached class BYTE")
						}
						if name == names[2] && value.Value.Pointer != g.pointer(t, 0x23002000) {
							t.Fatal("actual full-width cached callback")
						}
						mutate()
					}
					if name == names[0] && value.Site == 0x4e0b86 {
						if !value.Complete {
							forceGC()
						} else {
							followupReads++
							if value.Value.Pointer != g.Roles["target_update"] || g.Objects["target"].ObjClass != 4 {
								t.Fatal("class mutation must retain cached initial monster branch and actual Go update pointer")
							}
						}
					}
				}
				observer.BeforeCallback = func(site uint32, pointer unsafe.Pointer) {
					if callbackPending {
						t.Fatalf("duplicate callback attempt actual_site=%#x actual_pointer=%p pending_site=%#x pending_pointer=%p", site, pointer, callbackSite, callbackPointer)
					}
					callbackPending = true
					callbackSite, callbackPointer = site, pointer
					// This same frozen input also takes the earlier Shock callback
					// at4e0d05. Validate every pending callback above/below, but the
					// sound-only GC witness starts after the critical4e1039 read.
					if name == names[2] && (site == 0x4e1049 || site == 0x4e104f) {
						expectedSite := uint32(0x4e1049)
						if input.Args[2] == 0 {
							expectedSite = 0x4e104f
						}
						declared := g.pointer(t, 0x23002000)
						field := g.Objects["target"].DamageSound
						if site != expectedSite || critical != 0x4e1039 || criticalReads != 1 || mutations != 1 || pointer != declared || field != nil {
							t.Fatalf("cached sound callback must survive declared field-to-nil mutation actual_site=%#x expected_site=%#x actual_pointer=%p declared_pointer=%p field=%p critical=%#x reads=%d mutations=%d", site, expectedSite, pointer, declared, field, critical, criticalReads, mutations)
						}
						forceGC()
						callbackGC++
					}
				}
				observer.Service = func(call rawDefaultDamageNativeService4E0B30) (rawDefaultDamageNativeReply4E0B30, bool) {
					if call.Indirect {
						if call.Callback == nil || call.Callback != g.pointer(t, 0x23001000) && call.Callback != g.pointer(t, 0x23002000) {
							return rawDefaultDamageNativeReply4E0B30{}, false
						}
						if !callbackPending || callbackSite != call.Site || callbackPointer != call.Callback {
							t.Fatal("actual cached callback-before-entry identity")
						}
						callbackPending = false
					}
					spec, ok := specs[call.Site]
					if !ok || len(call.Arguments) != spec.Argc {
						t.Fatal("actual service ABI")
					}
					for _, arg := range call.Arguments {
						if arg.Kind == "object" && arg.Pointer == unsafe.Pointer(g.Opaque) {
							t.Fatal("opaque identity carrier must not enter an original service as an object")
						}
						if arg.Kind == "object" && arg.Pointer != nil && uintptr(arg.Pointer) <= math.MaxUint32 {
							t.Fatal("service object pointer lost high bits")
						}
					}
					schedule, ok := input.Services[fmt.Sprintf("%08x", call.Site)]
					if !ok {
						t.Fatal("missing explicit independent service input")
					}
					index := occurrences[call.Site]
					occurrences[call.Site]++
					if index >= len(schedule.Responses) {
						if !schedule.Repeat || len(schedule.Responses) == 0 {
							t.Fatal("exhausted explicit independent service responses")
						}
						index = len(schedule.Responses) - 1
					}
					var response rddResponse4E0B30
					if e := json.Unmarshal(schedule.Responses[index], &response); e != nil {
						t.Fatal(e)
					}
					if response.EAX == nil || len(response.Effects) != 0 || response.ECX != nil || response.EDX != nil {
						t.Fatal("selected Goheap rows require simple explicit response packets without service mutations")
					}
					reply := rawDefaultDamageNativeReply4E0B30{Word: *response.EAX}
					if spec.Name == "owner_chain" || spec.Name == "monster_sound_set" {
						reply.Pointer = g.pointer(t, *response.EAX)
					}
					if response.X87 != nil {
						payload, e := hex.DecodeString(response.X87.Hex)
						if e != nil || len(payload) != map[string]int{"f32": 4, "f64": 8, "ext80": 10}[response.X87.Format] {
							t.Fatal("exact input float packet")
						}
						reply.FloatFormat = response.X87.Format
						copy(reply.FloatPayload[:], payload)
					}
					return reply, true
				}
				oldChild := damagePredicateRawObserve4E1400
				damagePredicateRawObserve4E1400 = func(complete bool, base unsafe.Pointer, kind int, value damagePredicateRawValue4E1400) {
					if base == unsafe.Pointer(g.Opaque) {
						t.Fatal("opaque identity carrier must not be dereferenced by a child")
					}
					if name != names[1] {
						return
					}
					if kind == 4 {
						if !complete {
							forceGC()
						} else {
							criticalReads++
							if base != g.Roles["weapon"] || value.Pointer != g.Roles["weapon_use"] {
								t.Fatal("actual Go-owned UseData pointer")
							}
							mutate()
						}
					}
					if kind == 5 {
						if !complete {
							forceGC()
						} else {
							followupReads++
							if base != g.Roles["weapon_use"] || g.Objects["weapon"].UseData.Ptr != g.Roles["alt_use"] || value.Word != g.Wands["weapon_use"].Flags&0xff {
								t.Fatal("cached original Go use object must remain live after replacement and GC")
							}
						}
					}
				}
				pop := rawDefaultDamagePushNativeObserver4E0B30(observer)
				defer func() { pop(); damagePredicateRawObserve4E1400 = oldChild }()
				args := [3]unsafe.Pointer{g.pointer(t, input.Args[0]), g.pointer(t, input.Args[1]), g.pointer(t, input.Args[2])}
				var result int32
				if public {
					result = rawParentPublic4E0B30(args[0], args[1], args[2], int32(input.Args[3]), int32(input.Args[4]))
				} else {
					result = rawDefaultDamageNative4E0B30(args[0], args[1], args[2], int32(input.Args[3]), int32(input.Args[4]))
				}
				want, ok := row.RegistersAfter["eax"]
				if !ok || uint32(result) != want || gc != 3 || mutations != 1 || criticalReads != 1 || callbackPending {
					t.Fatal("actual original return, forced-GC and input-stimulus completion", result, gc, mutations, criticalReads)
				}
				if name == names[2] {
					if callbackGC != 1 {
						t.Fatal("one GC in actual cached indirect attempt")
					}
				} else if followupReads != 1 {
					t.Fatal("cached pointer followup read missing")
				}
				g.compare(t, expected)
				t.Logf("RAW_DEFAULT_DAMAGE_PINNED_GO_GC backend=%s input=%s pins=16 opaque_carriers=1 GC=%d stimuli=%d public_C=%t callback_body_executed=false selected_projection_and_all_native_canaries=true", backend, name, gc, mutations, public)
				runtime.KeepAlive(g)
			})
		}
	}
}
