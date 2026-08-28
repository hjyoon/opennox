package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

type objectExtendedDataTestObject4F40A0 struct {
	name         string
	idPresent    bool
	inventory    *objectExtendedDataTestObject4F40A0
	field129     *objectExtendedDataTestObject4F40A0
	teamID       uint8
	typeInd      uint16
	flags        uint32
	field5       uint32
	field189     int
	scriptPickup int32
}

type objectExtendedDataTestType4F40A0 struct {
	name   string
	flags  uint32
	field9 uint32
}

type objectExtendedDataTestWorld4F40A0 struct {
	object      *objectExtendedDataTestObject4F40A0
	typ         *objectExtendedDataTestType4F40A0
	modeResult  int32
	hostResult  int32
	textLengths map[int]uintptr
	events      []string
	faultAt     int
	after       map[string]func()
}

func objectExtendedDataTestObjectName4F40A0(object *objectExtendedDataTestObject4F40A0) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func objectExtendedDataTestTypeName4F40A0(typ *objectExtendedDataTestType4F40A0) string {
	if typ == nil {
		return "nil"
	}
	return typ.name
}

func (w *objectExtendedDataTestWorld4F40A0) event(value string) {
	w.events = append(w.events, value)
	if callback := w.after[value]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *objectExtendedDataTestWorld4F40A0) deps() objectExtendedDataDeps4F40A0[
	*objectExtendedDataTestObject4F40A0,
	*objectExtendedDataTestType4F40A0,
	int,
] {
	return objectExtendedDataDeps4F40A0[
		*objectExtendedDataTestObject4F40A0,
		*objectExtendedDataTestType4F40A0,
		int,
	]{
		loadIDPointerPresent: func(object *objectExtendedDataTestObject4F40A0) bool {
			value := object.idPresent
			valueInt := 0
			if value {
				valueInt = 1
			}
			w.event(fmt.Sprintf("id:%s=%d", objectExtendedDataTestObjectName4F40A0(object), valueInt))
			return value
		},
		loadInventoryHead: func(object *objectExtendedDataTestObject4F40A0) *objectExtendedDataTestObject4F40A0 {
			value := object.inventory
			w.event(fmt.Sprintf("inventory:%s=%s", objectExtendedDataTestObjectName4F40A0(object), objectExtendedDataTestObjectName4F40A0(value)))
			return value
		},
		loadField129: func(object *objectExtendedDataTestObject4F40A0) *objectExtendedDataTestObject4F40A0 {
			value := object.field129
			w.event(fmt.Sprintf("field129:%s=%s", objectExtendedDataTestObjectName4F40A0(object), objectExtendedDataTestObjectName4F40A0(value)))
			return value
		},
		loadTeamID: func(object *objectExtendedDataTestObject4F40A0) uint8 {
			value := object.teamID
			w.event(fmt.Sprintf("team:%s=%02x", objectExtendedDataTestObjectName4F40A0(object), value))
			return value
		},
		loadTypeInd: func(object *objectExtendedDataTestObject4F40A0) uint16 {
			value := object.typeInd
			w.event(fmt.Sprintf("type-ind:%s=%04x", objectExtendedDataTestObjectName4F40A0(object), value))
			return value
		},
		lookupType: func(typeInd uint16) *objectExtendedDataTestType4F40A0 {
			value := w.typ
			w.event(fmt.Sprintf("lookup:%04x=%s", typeInd, objectExtendedDataTestTypeName4F40A0(value)))
			return value
		},
		loadTypeFlags: func(typ *objectExtendedDataTestType4F40A0) uint32 {
			if typ == nil {
				panic("nil type flags")
			}
			value := typ.flags
			w.event(fmt.Sprintf("type-flags:%s=%08x", objectExtendedDataTestTypeName4F40A0(typ), value))
			return value
		},
		loadObjectFlags: func(object *objectExtendedDataTestObject4F40A0) uint32 {
			value := object.flags
			w.event(fmt.Sprintf("object-flags:%s=%08x", objectExtendedDataTestObjectName4F40A0(object), value))
			return value
		},
		loadTypeField9: func(typ *objectExtendedDataTestType4F40A0) uint32 {
			if typ == nil {
				panic("nil type field9")
			}
			value := typ.field9
			w.event(fmt.Sprintf("type-field9:%s=%08x", objectExtendedDataTestTypeName4F40A0(typ), value))
			return value
		},
		loadObjectField5: func(object *objectExtendedDataTestObject4F40A0) uint32 {
			value := object.field5
			w.event(fmt.Sprintf("object-field5:%s=%08x", objectExtendedDataTestObjectName4F40A0(object), value))
			return value
		},
		gameFlags: func(mask uint32) int32 {
			var value int32
			switch mask {
			case objectExtendedDataModeMask4F40A0:
				value = w.modeResult
			case objectExtendedDataHostMask4F40A0:
				value = w.hostResult
			default:
				panic(fmt.Sprintf("unexpected mask %#x", mask))
			}
			w.event(fmt.Sprintf("game-flags:%08x=%08x", mask, uint32(value)))
			return value
		},
		loadField189: func(object *objectExtendedDataTestObject4F40A0) int {
			value := object.field189
			w.event(fmt.Sprintf("field189:%s=%d", objectExtendedDataTestObjectName4F40A0(object), value))
			return value
		},
		stringLength: func(text int) uintptr {
			value := w.textLengths[text]
			w.event(fmt.Sprintf("strlen:%d=%d", text, value))
			return value
		},
		loadScriptPickupFunc: func(object *objectExtendedDataTestObject4F40A0) int32 {
			value := object.scriptPickup
			w.event(fmt.Sprintf("script-pickup:%s=%08x", objectExtendedDataTestObjectName4F40A0(object), uint32(value)))
			return value
		},
	}
}

func newObjectExtendedDataTestWorld4F40A0() *objectExtendedDataTestWorld4F40A0 {
	return &objectExtendedDataTestWorld4F40A0{
		object: &objectExtendedDataTestObject4F40A0{
			name:         "object",
			typeInd:      0xabcd,
			flags:        0x01020304,
			field5:       0xa5,
			scriptPickup: -1,
		},
		typ: &objectExtendedDataTestType4F40A0{
			name:   "type",
			flags:  0x01020304,
			field9: 0xa5,
		},
		modeResult:  1,
		textLengths: make(map[int]uintptr),
		after:       make(map[string]func()),
	}
}

func verifyObjectExtendedDataFaultPrefixes4F40A0(
	t *testing.T,
	want []string,
	build func() *objectExtendedDataTestWorld4F40A0,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			objectExtendedDataAdmission4F40A0(w.object, w.deps())
		})
	}
}

func TestObjectExtendedDataAdmission4F40A0NullDoesNotReadDependencies(t *testing.T) {
	deps := objectExtendedDataDeps4F40A0[
		*objectExtendedDataTestObject4F40A0,
		*objectExtendedDataTestType4F40A0,
		int,
	]{}
	if got := objectExtendedDataAdmission4F40A0((*objectExtendedDataTestObject4F40A0)(nil), deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestObjectExtendedDataAdmission4F40A0EarlyRejectOrder(t *testing.T) {
	item := &objectExtendedDataTestObject4F40A0{name: "item"}
	tests := []struct {
		name   string
		mutate func(*objectExtendedDataTestWorld4F40A0)
		want   []string
	}{
		{
			name:   "non-null ID pointer regardless of contents",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.idPresent = true },
			want:   []string{"id:object=1"},
		},
		{
			name:   "inventory",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.inventory = item },
			want:   []string{"id:object=0", "inventory:object=item"},
		},
		{
			name:   "field129",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.field129 = item },
			want:   []string{"id:object=0", "inventory:object=nil", "field129:object=item"},
		},
		{
			name:   "team ID",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.teamID = 0x80 },
			want:   []string{"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=80"},
		},
		{
			name:   "masked flags mismatch",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.flags ^= 0x2 },
			want: []string{
				"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=00",
				"type-ind:object=abcd", "lookup:abcd=type", "type-flags:type=01020304", "object-flags:object=01020306",
			},
		},
		{
			name:   "masked field5 mismatch",
			mutate: func(w *objectExtendedDataTestWorld4F40A0) { w.object.field5 ^= 0x2 },
			want: []string{
				"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=00",
				"type-ind:object=abcd", "lookup:abcd=type", "type-flags:type=01020304", "object-flags:object=01020304",
				"type-field9:type=000000a5", "object-field5:object=000000a7",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newObjectExtendedDataTestWorld4F40A0()
			test.mutate(w)
			if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != -1 {
				t.Fatalf("result = %d, want -1", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %v, want %v", w.events, test.want)
			}
		})
	}
}

func TestObjectExtendedDataAdmission4F40A0LiveModeOrderAndFaultPrefixes(t *testing.T) {
	build := func() *objectExtendedDataTestWorld4F40A0 {
		w := newObjectExtendedDataTestWorld4F40A0()
		w.object.flags = 0
		w.object.field5 = 0
		w.typ.field9 = 0
		w.modeResult = math.MinInt32
		w.textLengths[1] = 0
		w.after["type-flags:type=01020304"] = func() { w.object.flags = 0x01020304 }
		w.after["object-flags:object=01020304"] = func() { w.typ.field9 = 0xa5 }
		w.after["type-field9:type=000000a5"] = func() { w.object.field5 = 0xa5 }
		w.after["game-flags:00600000=80000000"] = func() { w.object.field189 = 1 }
		w.after["field189:object=1"] = func() { w.object.field189 = 2 }
		return w
	}
	want := []string{
		"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=00",
		"type-ind:object=abcd", "lookup:abcd=type", "type-flags:type=01020304", "object-flags:object=01020304",
		"type-field9:type=000000a5", "object-field5:object=000000a5",
		"game-flags:00600000=80000000", "field189:object=1", "strlen:1=0",
	}
	w := build()
	if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyObjectExtendedDataFaultPrefixes4F40A0(t, want, build)
}

func TestObjectExtendedDataAdmission4F40A0ModeTextGate(t *testing.T) {
	tests := []struct {
		name       string
		text       int
		length     uintptr
		wantResult int8
		wantTail   []string
	}{
		{name: "null pointer", text: 0, wantResult: 0, wantTail: []string{"game-flags:00600000=00000001", "field189:object=0"}},
		{name: "non-null empty string", text: 1, length: 0, wantResult: 0, wantTail: []string{"game-flags:00600000=00000001", "field189:object=1", "strlen:1=0"}},
		{name: "non-empty string", text: 2, length: 5, wantResult: -1, wantTail: []string{"game-flags:00600000=00000001", "field189:object=2", "strlen:2=5"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newObjectExtendedDataTestWorld4F40A0()
			w.object.field189 = test.text
			w.textLengths[test.text] = test.length
			if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if got := w.events[len(w.events)-len(test.wantTail):]; !reflect.DeepEqual(got, test.wantTail) {
				t.Fatalf("event tail = %v, want %v", got, test.wantTail)
			}
			for _, event := range w.events {
				if strings.HasPrefix(event, "game-flags:00000001") {
					t.Fatalf("mode path read host flag: %v", w.events)
				}
			}
		})
	}
}

func TestObjectExtendedDataAdmission4F40A0HostGateAndFaultPrefixes(t *testing.T) {
	build := func() *objectExtendedDataTestWorld4F40A0 {
		w := newObjectExtendedDataTestWorld4F40A0()
		w.modeResult = 0
		w.hostResult = math.MinInt32
		w.object.scriptPickup = 123
		w.after["game-flags:00000001=80000000"] = func() { w.object.scriptPickup = -1 }
		return w
	}
	want := []string{
		"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=00",
		"type-ind:object=abcd", "lookup:abcd=type", "type-flags:type=01020304", "object-flags:object=01020304",
		"type-field9:type=000000a5", "object-field5:object=000000a5",
		"game-flags:00600000=00000000", "game-flags:00000001=80000000", "script-pickup:object=ffffffff",
	}
	w := build()
	if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyObjectExtendedDataFaultPrefixes4F40A0(t, want, build)

	t.Run("not host skips script", func(t *testing.T) {
		w := newObjectExtendedDataTestWorld4F40A0()
		w.modeResult = 0
		w.hostResult = 0
		w.object.scriptPickup = 7
		if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if got := w.events[len(w.events)-1]; got != "game-flags:00000001=00000000" {
			t.Fatalf("last event = %q, want host flag check", got)
		}
	})

	t.Run("host callback mismatch rejects", func(t *testing.T) {
		w := newObjectExtendedDataTestWorld4F40A0()
		w.modeResult = 0
		w.hostResult = 2
		w.object.scriptPickup = math.MinInt32
		if got := objectExtendedDataAdmission4F40A0(w.object, w.deps()); got != -1 {
			t.Fatalf("result = %d, want -1", got)
		}
	})
}

func TestObjectExtendedDataAdmission4F40A0MissingTypeFaultsAtTypeFlags(t *testing.T) {
	w := newObjectExtendedDataTestWorld4F40A0()
	w.typ = nil
	defer func() {
		if got := recover(); got != "nil type flags" {
			t.Fatalf("panic = %v, want nil type flags", got)
		}
		want := []string{
			"id:object=0", "inventory:object=nil", "field129:object=nil", "team:object=00",
			"type-ind:object=abcd", "lookup:abcd=nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	objectExtendedDataAdmission4F40A0(w.object, w.deps())
}

func TestObjectExtendedDataAdmission4F40A0NativeRejectsNonNullEmptyID(t *testing.T) {
	empty := byte(0)
	object := &Object{IDPtr: unsafe.Pointer(&empty)}
	if got := new(Server).Sub_4F40A0(object); got != -1 {
		t.Fatalf("result = %d, want -1 for a non-null empty ID pointer", got)
	}
}

func TestObjectExtendedDataAdmission4F40A0NativeLayout(t *testing.T) {
	type check struct {
		name string
		got  uintptr
		want uintptr
	}
	checks32 := []check{
		{"Object.size", unsafe.Sizeof(Object{}), 780},
		{"Object.IDPtr", unsafe.Offsetof(Object{}.IDPtr), 0},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 4},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16},
		{"Object.Field5", unsafe.Offsetof(Object{}.Field5), 20},
		{"Object.TeamVal.ID", unsafe.Offsetof(Object{}.TeamVal) + unsafe.Offsetof(ObjectTeam{}.ID), 52},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 504},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), 516},
		{"Object.Field189", unsafe.Offsetof(Object{}.Field189), 756},
		{"Object.ScriptPickup.Func", unsafe.Offsetof(Object{}.ScriptPickup) + unsafe.Offsetof(ScriptCallback{}.Func), 768},
	}
	checks64 := []check{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.IDPtr", unsafe.Offsetof(Object{}.IDPtr), 0},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 8},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 20},
		{"Object.Field5", unsafe.Offsetof(Object{}.Field5), 24},
		{"Object.TeamVal.ID", unsafe.Offsetof(Object{}.TeamVal) + unsafe.Offsetof(ObjectTeam{}.ID), 56},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 544},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), 568},
		{"Object.Field189", unsafe.Offsetof(Object{}.Field189), 888},
		{"Object.ScriptPickup.Func", unsafe.Offsetof(Object{}.ScriptPickup) + unsafe.Offsetof(ScriptCallback{}.Func), 908},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, test := range checks {
		if test.got != test.want {
			t.Errorf("%s on %s/%s = %d, want %d", test.name, runtime.GOOS, runtime.GOARCH, test.got, test.want)
		}
	}
}
