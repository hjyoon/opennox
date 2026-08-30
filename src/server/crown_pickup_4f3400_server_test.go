package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

func defaultCrownPickupNativeDeps4F3400() crownPickupNativeDeps4F3400 {
	return crownPickupNativeDeps4F3400{
		defaultPickup: func(*Object, *Object, int32, int32) uint32 { return 0 },
		loadFrame:     func() uint32 { return 0 },
		setOwner:      func(*Object, *Object) {},
		applyEnchant:  func(*Object, EnchantID, uint32, uint32) {},
		playAudio:     func(sound.ID, *Object, int32, uint32) {},
		informPickup:  func(uint8, uint32, uint32) {},
		unmarkMinimap: func(*Object, uint32) {},
	}
}

func TestCrownPickup4F3400NativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantNetCode := uintptr(36)
	wantTeam := uintptr(48)
	wantObjectUpdate := uintptr(748)
	wantPlayerFrame := uintptr(264)
	wantCrownUpdateSize := uintptr(12)
	wantCrownPending := uintptr(4)
	if ptrSize == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantNetCode = 40
		wantTeam = 52
		wantObjectUpdate = 872
		wantPlayerFrame = 324
		wantCrownUpdateSize = 24
		wantCrownPending = 8
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantTeam},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Field66", unsafe.Offsetof(PlayerUpdateData{}.Field66), wantPlayerFrame},
		{"CrownUpdateData size", unsafe.Sizeof(CrownUpdateData{}), wantCrownUpdateSize},
		{"CrownUpdateData.Field0", unsafe.Offsetof(CrownUpdateData{}.Field0), 0},
		{"CrownUpdateData.PickupTarget", unsafe.Offsetof(CrownUpdateData{}.PickupTarget), wantCrownPending},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestCrownPickupNative4F3400BindsFieldsAndEffects(t *testing.T) {
	oldPending := &Object{}
	crownUpdate := &CrownUpdateData{PickupTarget: oldPending}
	crown := &Object{UpdateData: unsafe.Pointer(crownUpdate)}
	playerUpdate := &PlayerUpdateData{}
	who := &Object{
		ObjClass:   object.ClassPlayer | object.Class(0x80000000),
		NetCode:    0xfedcba98,
		TeamVal:    ObjectTeam{ID: 0x7f},
		UpdateData: unsafe.Pointer(playerUpdate),
	}
	events := make([]string, 0, 8)
	deps := defaultCrownPickupNativeDeps4F3400()
	deps.defaultPickup = func(gotWho, gotCrown *Object, flag1, flag2 int32) uint32 {
		if gotWho != who || gotCrown != crown || flag1 != -3 || flag2 != 9 {
			t.Fatalf("default args = (%p,%p,%d,%d)", gotWho, gotCrown, flag1, flag2)
		}
		events = append(events, "default")
		return 0x80000001
	}
	deps.loadFrame = func() uint32 {
		events = append(events, "frame")
		return 0x89abcdef
	}
	deps.setOwner = func(owner, item *Object) {
		if owner != who || item != crown || playerUpdate.Field66 != 0x89abcdef {
			t.Fatalf("owner args/frame = (%p,%p,%#x)", owner, item, playerUpdate.Field66)
		}
		events = append(events, "owner")
	}
	deps.applyEnchant = func(obj *Object, enchant EnchantID, duration, power uint32) {
		if obj != who || enchant != ENCHANT_CROWN || duration != 0 || power != 5 {
			t.Fatalf("enchant args = (%p,%d,%d,%d)", obj, enchant, duration, power)
		}
		events = append(events, "enchant")
	}
	deps.playAudio = func(id sound.ID, obj *Object, kind int32, code uint32) {
		if id != sound.SoundCrownChange || obj != who || kind != 0 || code != 0 {
			t.Fatalf("audio args = (%d,%p,%d,%d)", id, obj, kind, code)
		}
		events = append(events, "audio")
	}
	deps.informPickup = func(code uint8, netCode, teamID uint32) {
		if code != 10 || netCode != 0xfedcba98 || teamID != 0x7f {
			t.Fatalf("inform args = (%d,%#x,%#x)", code, netCode, teamID)
		}
		events = append(events, "inform")
	}
	deps.unmarkMinimap = func(obj *Object, flags uint32) {
		if obj != crown || flags != 1 || crownUpdate.PickupTarget != oldPending {
			t.Fatalf("unmark args/pending = (%p,%d,%p)", obj, flags, crownUpdate.PickupTarget)
		}
		events = append(events, "unmark")
	}

	got := crownPickupNative4F3400(who, crown, -3, 9, deps)
	if got != 0x80000001 {
		t.Fatalf("result = %#x, want 0x80000001", got)
	}
	if crownUpdate.PickupTarget != nil {
		t.Fatal("pending target was not cleared last")
	}
	if !reflect.DeepEqual(events, []string{
		"default", "frame", "owner", "enchant", "audio", "inform", "unmark",
	}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestCrownPickupNative4F3400ClearsCachedUpdateAfterFailure(t *testing.T) {
	oldUpdate := &CrownUpdateData{PickupTarget: &Object{}}
	newPending := &Object{}
	newUpdate := &CrownUpdateData{PickupTarget: newPending}
	crown := &Object{UpdateData: unsafe.Pointer(oldUpdate)}
	who := &Object{ObjClass: object.ClassPlayer}
	deps := defaultCrownPickupNativeDeps4F3400()
	deps.defaultPickup = func(*Object, *Object, int32, int32) uint32 {
		crown.UpdateData = unsafe.Pointer(newUpdate)
		return 0
	}

	if got := crownPickupNative4F3400(who, crown, 1, 1, deps); got != 0 {
		t.Fatalf("result = %#x, want 0", got)
	}
	if oldUpdate.PickupTarget != nil {
		t.Fatal("cached update was not cleared")
	}
	if newUpdate.PickupTarget != newPending {
		t.Fatal("replacement update was cleared")
	}
}

func TestCrownPickupInformPacket4F3400(t *testing.T) {
	packet := crownPickupInformPacket4F3400(10, 0xf1234567, 0x000000ab)
	if packet[0] != byte(netmsg.MSG_INFORM) || packet[1] != 10 {
		t.Fatalf("packet header = %#v", packet[:2])
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0xf1234567 {
		t.Fatalf("packet net code = %#x", got)
	}
	if got := binary.LittleEndian.Uint32(packet[6:10]); got != 0xab {
		t.Fatalf("packet team ID = %#x", got)
	}
}

func TestCrownPickup4F3400ServerDefaultKeepsBothFlags(t *testing.T) {
	s := &Server{}
	oldUpdate := &CrownUpdateData{PickupTarget: &Object{}}
	crown := &Object{UpdateData: unsafe.Pointer(oldUpdate)}
	who := &Object{ObjClass: object.ClassPlayer}
	s.Objs.DefaultPickup = func(gotWho, gotCrown *Object, flag1, flag2 int) bool {
		if gotWho != who || gotCrown != crown || flag1 != -17 || flag2 != 23 {
			t.Fatalf("default args = (%p,%p,%d,%d)", gotWho, gotCrown, flag1, flag2)
		}
		return false
	}
	runtime := CrownPickupRuntime4F3400{
		ApplyEnchant: func(*Object, EnchantID, uint32, uint32) {
			t.Fatal("failed pickup applied enchant")
		},
	}

	if got := s.CrownPickup4F3400(who, crown, -17, 23, runtime); got != 0 {
		t.Fatalf("result = %#x, want 0", got)
	}
	if oldUpdate.PickupTarget != nil {
		t.Fatal("server binding did not clear pending target")
	}
}
