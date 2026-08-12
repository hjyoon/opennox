package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectNeedSyncCMatchesGo(t *testing.T) {
	got := &server.Object{Field37: 0x13579BDF, Field38: 0x2468ACE0}
	want := *got
	want.NeedSync()

	objectNeedSyncC(got)
	if got.Field38 != want.Field38 {
		t.Fatalf("Field38: C = %#08x, Go = %#08x", got.Field38, want.Field38)
	}
	if got.Field37 != want.Field37 {
		t.Fatalf("C overwrote Field37: got %#08x, want %#08x", got.Field37, want.Field37)
	}
}

func TestObjectStatusMaskCMatchesGo(t *testing.T) {
	const (
		val1 = uint32(0x020000)
		val2 = uint32(0x000002)
	)
	for _, set := range []bool{false, true} {
		t.Run(map[bool]string{false: "clear", true: "set"}[set], func(t *testing.T) {
			got := &server.Object{Field37: 0x80000015, Field38: 0xA5A5A5A5}
			for i := range got.Field140 {
				got.Field140[i] = 0x5A000000 | val1 | val2 | uint32(i<<4)
			}
			want := *got
			want.Sub_4E4500(val1, val2, set)

			objectStatusMaskC(got, val1, val2, set)
			if got.Field140 != want.Field140 {
				for i := range got.Field140 {
					if got.Field140[i] != want.Field140[i] {
						t.Errorf("Field140[%d]: C = %#08x, Go = %#08x", i, got.Field140[i], want.Field140[i])
					}
				}
			}
			if got.Field37 != want.Field37 || got.Field38 != want.Field38 {
				t.Fatalf("C overwrote adjacent state: Field37=%#08x Field38=%#08x", got.Field37, got.Field38)
			}
		})
	}
}

func TestObjectSetOnOffCMatchesGo(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disable", true: "enable"}[enabled], func(t *testing.T) {
			got := &server.Object{
				ObjClass: object.ClassClientPersist,
				ObjFlags: object.FlagActive | object.FlagEnabled,
				Field37:  0x13579BDF,
				Field38:  0x2468ACE0,
			}
			if enabled {
				got.ObjFlags &^= object.FlagEnabled
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | uint32(i<<4) | uint32(i&7)
			}
			want := *got
			want.SetOnOff(enabled)

			resultOffset := objectSetOnOffC(got, enabled)
			wantOffset := unsafe.Offsetof(server.Object{}.Field140) + unsafe.Sizeof(server.Object{}.Field140)
			if resultOffset != wantOffset {
				t.Errorf("return offset: C = %d, want %d", resultOffset, wantOffset)
			}
			if got.ObjFlags != want.ObjFlags || got.Field38 != want.Field38 {
				t.Errorf("flags/sync: C = (%#08x, %#08x), Go = (%#08x, %#08x)", got.ObjFlags, got.Field38, want.ObjFlags, want.Field38)
			}
			if got.Field140 != want.Field140 {
				t.Errorf("Field140 differs: C = %#v, Go = %#v", got.Field140, want.Field140)
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass {
				t.Fatal("C overwrote state outside the original function contract")
			}
		})
	}
}

func TestObjectSetBuffFlagsCWrapperMatchesGo(t *testing.T) {
	got := &server.Object{
		ObjClass: object.ClassClientPersist,
		Buffs:    0x55555555,
		Field37:  0x13579BDF,
		Field38:  0x2468ACE0,
	}
	for i := range got.Field140 {
		got.Field140[i] = 0xA5000000 | uint32(i<<4) | uint32(i&7)
	}
	want := *got
	want.SetBuffFlags(0x89ABCDEF, nil)

	resultOffset := objectSetBuffFlagsC(got, 0x89ABCDEF)
	wantOffset := unsafe.Offsetof(server.Object{}.Field140) + unsafe.Sizeof(server.Object{}.Field140)
	if resultOffset != wantOffset {
		t.Errorf("return offset: C wrapper = %d, want %d", resultOffset, wantOffset)
	}
	if got.Buffs != want.Buffs || got.Field38 != want.Field38 || got.Field140 != want.Field140 {
		t.Error("C wrapper did not preserve the Go implementation result")
	}
}

func TestObjectSetBuffFlagsCWrapperPlayerPath(t *testing.T) {
	udMem, freeUD := alloc.Malloc(unsafe.Sizeof(server.PlayerUpdateData{}))
	defer freeUD()
	plMem, freePlayer := alloc.Malloc(unsafe.Sizeof(server.Player{}))
	defer freePlayer()
	ud := (*server.PlayerUpdateData)(udMem)
	pl := (*server.Player)(plMem)
	pl.ProtUnitBuffs = 0 // The legacy protection reset treats zero as an inactive handle.
	ud.Player = pl

	got := &server.Object{
		ObjClass:   object.ClassPlayer,
		Buffs:      0x55555555,
		Field37:    0x13579BDF,
		Field38:    0x2468ACE0,
		UpdateData: udMem,
	}
	for i := range got.Field140 {
		got.Field140[i] = 0xA5000000 | uint32(i<<4) | uint32(i&7)
	}
	want := *got
	want.SetBuffFlags(0x89ABCDEF, func(gotPlayer *server.Player, flags uint32) {
		if gotPlayer != pl || flags != 0x89ABCDEF {
			t.Fatal("Go reference resolved the wrong native Player")
		}
	})

	resultOffset := objectSetBuffFlagsC(got, 0x89ABCDEF)
	wantOffset := unsafe.Offsetof(server.Object{}.Field140) + unsafe.Sizeof(server.Object{}.Field140)
	if resultOffset != wantOffset {
		t.Errorf("return offset: C wrapper = %d, want %d", resultOffset, wantOffset)
	}
	if got.Buffs != want.Buffs || got.Field38 != want.Field38 || got.Field140 != want.Field140 {
		t.Error("C wrapper player path differs from the Go implementation")
	}
}

func TestObjectSetModifierAttrsCWrapperMatchesGo(t *testing.T) {
	mod0, freeMod0 := alloc.New(server.ModifierEff{})
	defer freeMod0()
	mod1, freeMod1 := alloc.New(server.ModifierEff{})
	defer freeMod1()
	full := server.ModifierInitData{
		Modifiers: [4]*server.ModifierEff{mod0, nil, mod1, nil},
		Field16:   0x89ABCDEF,
	}
	empty := server.ModifierInitData{Field16: 0x13579BDF}
	tests := []struct {
		name     string
		class    object.Class
		subClass object.SubClass
		typeInd  uint16
		teamBase uint32
		attrs    server.ModifierInitData
	}{
		{name: "empty scans only four modifier pointers", class: object.ClassWeapon, teamBase: 11, attrs: empty},
		{name: "ineligible returns TeamBase id", class: object.ClassMissile, typeInd: 7, teamBase: 12, attrs: full},
		{name: "allowed special object", class: object.ClassWeapon | object.ClassClientPersist, teamBase: 13, attrs: full},
		{name: "forced wand accepts empty attributes", class: object.ClassWand | object.ClassPlayer, subClass: 0x00010000, teamBase: 14, attrs: empty},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attrs, freeAttrs := alloc.New(server.ModifierInitData{})
			defer freeAttrs()
			*attrs = tc.attrs
			dst, freeDst := alloc.New(server.ModifierInitData{})
			defer freeDst()
			*dst = server.ModifierInitData{
				Modifiers: [4]*server.ModifierEff{nil, mod1, nil, mod0},
				Field16:   0x2468ACE0,
			}
			wantDst := *dst
			got := &server.Object{
				TypeInd:     tc.typeInd,
				ObjClass:    tc.class,
				ObjSubClass: tc.subClass,
				ObjFlags:    object.FlagActive,
				Field37:     0x80000015,
				Field38:     0x2468ACE0,
				InitData:    unsafe.Pointer(dst),
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | 0x02000000 | 0x200 | uint32(i<<12)
			}
			want := *got
			want.InitData = unsafe.Pointer(&wantDst)
			applied := want.SetModifierAttrs(&tc.attrs, tc.teamBase)

			result := objectSetModifierAttrsC(got, attrs, tc.teamBase)
			forced := tc.class.Has(object.ClassWand) && uint32(tc.subClass)&0x047F0000 != 0
			var wantResult uintptr
			switch {
			case !forced && !tc.attrs.HasModifiers():
				wantResult = uintptr(unsafe.Pointer(attrs)) + unsafe.Offsetof(server.ModifierInitData{}.Field16)
			case !applied:
				wantResult = uintptr(tc.teamBase)
			default:
				wantResult = uintptr(unsafe.Pointer(got)) + unsafe.Offsetof(server.Object{}.Field140) + unsafe.Sizeof(server.Object{}.Field140)
			}
			if result != wantResult {
				t.Errorf("return: C = %#x, want %#x", result, wantResult)
			}
			if *dst != wantDst {
				t.Errorf("attributes: C = %#v, Go = %#v", *dst, wantDst)
			}
			if got.Field38 != want.Field38 || got.Field140 != want.Field140 {
				t.Error("C wrapper object state differs from Go implementation")
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass || got.ObjSubClass != want.ObjSubClass || got.ObjFlags != want.ObjFlags || got.TypeInd != want.TypeInd || got.InitData != unsafe.Pointer(dst) {
				t.Fatal("C wrapper overwrote state outside the original function contract")
			}
		})
	}
}

func TestObjectGetMassCMatchesGAMEEXEContract(t *testing.T) {
	tests := []struct {
		name string
		bits uint32
	}{
		{name: "positive zero", bits: 0x00000000},
		{name: "negative zero", bits: 0x80000000},
		{name: "one", bits: 0x3F800000},
		{name: "negative finite", bits: 0xC0200000},
		{name: "smallest subnormal", bits: 0x00000001},
		{name: "largest finite", bits: 0x7F7FFFFF},
		{name: "positive infinity", bits: 0x7F800000},
		{name: "negative infinity", bits: 0xFF800000},
		{name: "quiet NaN", bits: 0x7FC12345},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := &server.Object{
				Field29:    0x13579BDF,
				Mass:       math.Float32frombits(tc.bits),
				Direction1: server.Dir16(0x2468),
				Direction2: server.Dir16(0xACE0),
			}
			beforeMass := math.Float32bits(got.Mass)
			want := float64(got.Mass)
			result := objectMassC(got)

			if math.IsNaN(want) {
				if !math.IsNaN(result) {
					t.Fatalf("mass: C = %v, want NaN", result)
				}
			} else if math.Float64bits(result) != math.Float64bits(want) {
				t.Fatalf("mass bits: C = %#016x, Go = %#016x", math.Float64bits(result), math.Float64bits(want))
			}
			if math.Float32bits(got.Mass) != beforeMass || got.Field29 != 0x13579BDF ||
				got.Direction1 != server.Dir16(0x2468) || got.Direction2 != server.Dir16(0xACE0) {
				t.Fatal("mass getter modified the object or an adjacent field")
			}
		})
	}
}

func TestObjectRaiseCMatchesGo(t *testing.T) {
	tests := []struct {
		name string
		from float32
		to   float32
	}{
		{name: "changed", from: 2.5, to: 7.25},
		{name: "equal", from: 2.5, to: 2.5},
		{name: "negative zero", from: float32(math.Copysign(0, -1)), to: 0},
		{name: "source NaN is unordered", from: math.Float32frombits(0x7FC00001), to: 7.25},
		{name: "target NaN is unordered", from: 2.5, to: math.Float32frombits(0x7FC00002)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := &server.Object{
				ObjClass: object.ClassPlayer,
				ZVal:     tc.from,
				Field37:  0x13579BDF,
				Field38:  0x2468ACE0,
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | uint32(i<<4) | uint32(i&7)
			}
			want := *got
			want.Raise(tc.to)

			objectRaiseC(got, tc.to)
			if math.Float32bits(got.ZVal) != math.Float32bits(want.ZVal) {
				t.Errorf("ZVal: C = %#08x, Go = %#08x", math.Float32bits(got.ZVal), math.Float32bits(want.ZVal))
			}
			if got.Field38 != want.Field38 {
				t.Errorf("Field38: C = %#08x, Go = %#08x", got.Field38, want.Field38)
			}
			if got.Field140 != want.Field140 {
				t.Errorf("Field140 differs: C = %#v, Go = %#v", got.Field140, want.Field140)
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass {
				t.Fatal("C overwrote state outside the original function contract")
			}
		})
	}
}

func TestObjectMarkAnimFrameCMatchesGo(t *testing.T) {
	got := &server.Object{
		ObjClass: object.ClassImmobile,
		Field33:  0xAAAAAAAA,
		Field37:  0x13579BDF,
		Field38:  0x2468ACE0,
	}
	for i := range got.Field140 {
		got.Field140[i] = 0x5A000000 | uint32(i<<4) | uint32(i&7)
	}
	want := *got
	want.MarkAnimFrame(0xFEDCBA98)

	objectMarkAnimFrameC(got, 0xFEDCBA98)
	if got.Field33 != want.Field33 || got.Field38 != want.Field38 {
		t.Errorf("frame/sync: C = (%#08x, %#08x), Go = (%#08x, %#08x)", got.Field33, got.Field38, want.Field33, want.Field38)
	}
	if got.Field140 != want.Field140 {
		t.Errorf("Field140 differs: C = %#v, Go = %#v", got.Field140, want.Field140)
	}
	if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass {
		t.Fatal("C overwrote state outside the original function contract")
	}
}
