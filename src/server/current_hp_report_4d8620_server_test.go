package server

import (
	"reflect"
	"testing"
)

func TestCurrentHPReport4D8620NativeReloadsHealthAndSendsExactPacket(t *testing.T) {
	first := &HealthData{Cur: 2}
	second := &HealthData{Cur: 0xffff}
	obj := &Object{HealthData: first}
	var packet [4]byte
	var recipient int32
	var related *Object
	var remove int32
	got := currentHPReportNative4D8620(37, obj, currentHPReportNativeDeps4D8620{
		getUnitNetCode: func(got *Object) uint32 {
			if got != obj {
				t.Fatalf("net-code object = %p, want %p", got, obj)
			}
			obj.HealthData = second
			return 0x12345678
		},
		sendReliable: func(gotRecipient int32, gotPacket [4]byte, gotRelated *Object, gotRemove int32) int32 {
			recipient, packet, related, remove = gotRecipient, gotPacket, gotRelated, gotRemove
			return -1 << 31
		},
	})
	if got != -1<<31 {
		t.Fatalf("result = %d, want MinInt32", got)
	}
	if recipient != 37 || related != nil || remove != 1 {
		t.Fatalf("send metadata = (%d,%p,%d), want (37,nil,1)", recipient, related, remove)
	}
	if want := [4]byte{65, 0x78, 0x56, 0xff}; !reflect.DeepEqual(packet, want) {
		t.Fatalf("packet = %v, want %v", packet, want)
	}
}

func TestCurrentHPReport4D8620NativeNullInitialHealthReturnsZero(t *testing.T) {
	obj := &Object{}
	got := currentHPReportNative4D8620(1, obj, currentHPReportNativeDeps4D8620{
		getUnitNetCode: func(*Object) uint32 {
			t.Fatal("net code called for nil HealthData")
			return 0
		},
		sendReliable: func(int32, [4]byte, *Object, int32) int32 {
			t.Fatal("send called for nil HealthData")
			return 0
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestCurrentHPReport4D8620NativePreservesReloadedNullFault(t *testing.T) {
	obj := &Object{HealthData: &HealthData{Cur: 1}}
	defer func() {
		if recover() == nil {
			t.Fatal("expected reloaded nil HealthData panic")
		}
	}()
	currentHPReportNative4D8620(1, obj, currentHPReportNativeDeps4D8620{
		getUnitNetCode: func(*Object) uint32 {
			obj.HealthData = nil
			return 1
		},
		sendReliable: func(int32, [4]byte, *Object, int32) int32 {
			t.Fatal("send called after reloaded nil HealthData")
			return 0
		},
	})
}
