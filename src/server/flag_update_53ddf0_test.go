package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestFlagUpdate53DDF0NativeLayout(t *testing.T) {
	wants := []struct {
		name string
		got  uintptr
		v32  uintptr
		v64  uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8, 12},
		{"Object.TeamVal.ID", unsafe.Offsetof(Object{}.TeamVal) + unsafe.Offsetof(ObjectTeam{}.ID), 52, 56},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748, 872},
		{"FlagUpdateData size", unsafe.Sizeof(FlagUpdateData4EA490{}), 12, 12},
		{"FlagUpdateData.Home", unsafe.Offsetof(FlagUpdateData4EA490{}.Home), 0, 0},
		{"FlagUpdateData.State", unsafe.Offsetof(FlagUpdateData4EA490{}.State), 8, 8},
	}
	for _, tc := range wants {
		want := tc.v64
		if unsafe.Sizeof(uintptr(0)) == 4 {
			want = tc.v32
		}
		if tc.got != want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, want)
		}
	}
}

func TestFlagUpdate53DDF0InactiveAndExactDeadline(t *testing.T) {
	s := &Server{}
	s.SetTickRate(30)
	update := &FlagUpdateData4EA490{}
	flag := &Object{UpdateData: unsafe.Pointer(update)}
	called := false
	runtime := FlagUpdateRuntime53DDF0{
		AudioEvent: func(uint32, *Object) { called = true },
		FlagStatus: func(uint8, uint8, uint8, uint16) int32 { called = true; return 0 },
		Move:       func(*Object, types.Pointf) { called = true },
		InformHome: func(uint32) int32 { called = true; return 0 },
	}
	if got := s.FlagUpdate53DDF0(flag, runtime); got != 0 {
		t.Fatalf("inactive result = %d, want 0", got)
	}

	update.State = 100
	s.SetFrame(100 + 30*30)
	if got := s.FlagUpdate53DDF0(flag, runtime); got != 90 {
		t.Fatalf("deadline result = %d, want 90", got)
	}
	if update.State != 100 || called {
		t.Fatalf("deadline mutated state: state=%d called=%v", update.State, called)
	}
}

func TestFlagUpdate53DDF0ReturnHomeOrderAndWidths(t *testing.T) {
	s := &Server{}
	s.SetTickRate(30)
	s.SetFrame(100 + 30*30 + 1)
	update := &FlagUpdateData4EA490{
		Home:  types.Pointf{X: 123.5, Y: -456.25},
		State: 100,
	}
	flag := &Object{
		ObjClass:   object.ClassFlag,
		TeamVal:    ObjectTeam{ID: TeamID(0xab)},
		UpdateData: unsafe.Pointer(update),
	}
	materialName := []byte("MaterialTeamOrange\x00")
	material := &ModifierEff{name0: &materialName[0]}
	initData := &ModifierInitData{}
	initData.Modifiers[1] = material
	flag.InitData = unsafe.Pointer(initData)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(flag)) <= math.MaxUint32 {
		t.Fatal("expected native flag pointer above 4 GiB")
	}

	var events []string
	runtime := FlagUpdateRuntime53DDF0{
		AudioEvent: func(id uint32, got *Object) {
			events = append(events, "audio")
			if id != 305 || got != flag || update.State != 100 {
				t.Fatalf("audio = (%d, %p), state=%d", id, got, update.State)
			}
		},
		FlagStatus: func(teamID, status, flagIndex uint8, carrier uint16) int32 {
			events = append(events, "status")
			if teamID != 0xab || status != 0 || flagIndex != 9 || carrier != 0 || update.State != 0 {
				t.Fatalf("status = (%#x, %d, %d, %d), state=%d", teamID, status, flagIndex, carrier, update.State)
			}
			return 77
		},
		Move: func(got *Object, destination types.Pointf) {
			events = append(events, "move")
			if got != flag || destination != update.Home || update.State != 0 {
				t.Fatalf("move = (%p, %+v), state=%d", got, destination, update.State)
			}
		},
		InformHome: func(flagIndex uint32) int32 {
			events = append(events, "inform")
			if flagIndex != 9 || update.State != 0 {
				t.Fatalf("inform = %d, state=%d", flagIndex, update.State)
			}
			return -123
		},
	}
	if got := s.FlagUpdate53DDF0(flag, runtime); got != -123 {
		t.Fatalf("result = %d, want -123", got)
	}
	if want := []string{"audio", "status", "move", "inform"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if update.State != 0 {
		t.Fatalf("state = %d, want 0", update.State)
	}
}

func TestFlagUpdate53DDF0UnsignedWrapAndResultWrap(t *testing.T) {
	s := &Server{}
	s.SetTickRate(30)
	s.SetFrame(900)
	update := &FlagUpdateData4EA490{State: math.MaxUint32 - 100}
	flag := &Object{UpdateData: unsafe.Pointer(update)}
	returned := false
	if got := s.FlagUpdate53DDF0(flag, FlagUpdateRuntime53DDF0{
		InformHome: func(uint32) int32 { returned = true; return 17 },
	}); got != 17 || !returned || update.State != 0 {
		t.Fatalf("wrapped timeout = result %d returned=%v state=%d", got, returned, update.State)
	}

	update.State = 7
	s.SetFrame(7)
	tickRate := uint32(0x80000001)
	s.SetTickRate(tickRate)
	if got, want := s.FlagUpdate53DDF0(flag, FlagUpdateRuntime53DDF0{}), int32(uint32(3)*tickRate); got != want {
		t.Fatalf("wrapped active result = %#x, want %#x", uint32(got), uint32(want))
	}
}

func TestFlagUpdate53DDF0RejectsMissingNativeState(t *testing.T) {
	s := &Server{}
	if got := s.FlagUpdate53DDF0(nil, FlagUpdateRuntime53DDF0{}); got != 0 {
		t.Fatalf("nil flag = %d, want 0", got)
	}
	if got := s.FlagUpdate53DDF0(&Object{}, FlagUpdateRuntime53DDF0{}); got != 0 {
		t.Fatalf("nil update = %d, want 0", got)
	}
}
