package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultScavengerReportNativeDeps4D8CD0() scavengerReportNativeDeps4D8CD0 {
	return scavengerReportNativeDeps4D8CD0{
		unitCode:   func(*Object) uint32 { return 0 },
		sendPacket: func(int32, [7]byte, *Object, int32) int32 { return 0 },
	}
}

func TestScavengerReportNative4D8CD0BindsFieldsAndPacket(t *testing.T) {
	player := &Player{Field2152: 0x12345678, Field2156: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	events := make([]string, 0, 2)
	deps := defaultScavengerReportNativeDeps4D8CD0()
	deps.unitCode = func(gotOwner *Object) uint32 {
		events = append(events, "unit-code")
		if gotOwner != owner {
			t.Fatalf("unit-code owner = %p, want %p", gotOwner, owner)
		}
		return 0xfeed4321
	}
	deps.sendPacket = func(recipient int32, packet [7]byte, related *Object, remove int32) int32 {
		events = append(events, "send")
		want := [7]byte{85, 0x21, 0x43, 0x78, 0x56, 0xef, 0xcd}
		if recipient != 255 || packet != want || related != nil || remove != 1 {
			t.Fatalf("send = %d/% x/%p/%d", recipient, packet, related, remove)
		}
		return -91
	}

	got := scavengerReportNative4D8CD0(owner, deps)
	if got.kind != scavengerReportSendResult4D8CD0 || got.send != -91 {
		t.Fatalf("result = %#v", got)
	}
	if !reflect.DeepEqual(events, []string{"unit-code", "send"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestScavengerReportNative4D8CD0CachesUpdateAcrossUnitCode(t *testing.T) {
	oldPlayer := &Player{Field2152: 0x1111, Field2156: 0x2222}
	oldUpdate := &PlayerUpdateData{Player: oldPlayer}
	newPlayer := &Player{Field2152: 0xaaaa, Field2156: 0xbbbb}
	newUpdate := &PlayerUpdateData{Player: newPlayer}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(oldUpdate)}
	deps := defaultScavengerReportNativeDeps4D8CD0()
	deps.unitCode = func(*Object) uint32 {
		owner.UpdateData = unsafe.Pointer(newUpdate)
		return 0
	}
	deps.sendPacket = func(_ int32, packet [7]byte, _ *Object, _ int32) int32 {
		want := [7]byte{85, 0, 0, 0x11, 0x11, 0x22, 0x22}
		if packet != want {
			t.Fatalf("packet = % x, want % x", packet, want)
		}
		return 1
	}

	got := scavengerReportNative4D8CD0(owner, deps)
	if got.kind != scavengerReportSendResult4D8CD0 || got.send != 1 {
		t.Fatalf("result = %#v", got)
	}
}

func TestScavengerReportNative4D8CD0NonPlayerReturnsOwnerAfterUpdateLoad(t *testing.T) {
	update := &PlayerUpdateData{}
	owner := &Object{ObjClass: object.Class(0x80000000), UpdateData: unsafe.Pointer(update)}
	deps := defaultScavengerReportNativeDeps4D8CD0()
	deps.unitCode = func(*Object) uint32 {
		t.Fatal("non-Player requested unit code")
		return 0
	}
	deps.sendPacket = func(int32, [7]byte, *Object, int32) int32 {
		t.Fatal("non-Player sent packet")
		return 0
	}

	got := scavengerReportNative4D8CD0(owner, deps)
	if got.kind != scavengerReportOwnerResult4D8CD0 || got.owner != owner {
		t.Fatalf("result = %#v, want owner %p", got, owner)
	}
}
