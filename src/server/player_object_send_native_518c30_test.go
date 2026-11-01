package server

import (
	"testing"
	"unsafe"
)

func playerSendOrderFixture518C30() (*Object, *Object, *PlayerUpdateData, *PlayerUpdateData, PlayerObjectSendRuntime518C30) {
	rp, up := &Player{PlayerInd: 3}, &Player{PlayerInd: 11}
	ru, uu := &PlayerUpdateData{Player: rp}, &PlayerUpdateData{Player: up}
	r, u := &Object{UpdateData: unsafe.Pointer(ru)}, &Object{UpdateData: unsafe.Pointer(uu)}
	runtime := PlayerObjectSendRuntime518C30{
		Frame:        func() uint32 { return 100 },
		HealthDelta:  func(int32, uint16, int16) {},
		Self:         func(*Object) {},
		UnitCode:     func(*Object) uint32 { return 0x1234 },
		FloatToInt:   func(value float32) int32 { return int32(value) },
		Action:       func(*Object) uint32 { return 0 },
		Random:       func(int32, int32) int32 { return 7 },
		Frames:       func(int32) (int32, int32) { return 3, 0 },
		Direction:    func(int32) IndexedDirectionVector509E20 { return IndexedDirectionVector509E20{X: 1} },
		ServerPacket: func(int32, [12]byte) int32 { return 1 },
		ClientPacket: func(int32, int32, [12]byte) int32 { return 1 },
	}
	return r, u, ru, uu, runtime
}

func TestPlayerSendHealthFrameBeforeDamageTimestamp518C30(t *testing.T) {
	r, u, _, uu, runtime := playerSendOrderFixture518C30()
	u.HealthData = &HealthData{Cur: 2}
	uu.HealthSamples[3] = 5
	reads := 0
	runtime.Frame = func() uint32 {
		reads++
		if reads == 1 {
			// Original518C71 reads the frame before518C77 reads Frame134.
			u.Frame134 = 100
		}
		return 100
	}
	runtime.HealthDelta = func(int32, uint16, int16) { t.Fatal("read stale damage timestamp before frame") }
	if got := PlayerObjectSendNative518C30(r, u, 1, 1, runtime); got != 1 || reads != 2 || uu.HealthSamples[3] != 5 {
		t.Fatalf("return=%d frameReads=%d sample=%d", got, reads, uu.HealthSamples[3])
	}
}

func TestPlayerSendPeerDestinationBeforeFrame518C30(t *testing.T) {
	r, u, ru, uu, runtime := playerSendOrderFixture518C30()
	initial := ru.Player
	other := &Player{PlayerInd: 29}
	runtime.Frame = func() uint32 {
		// Original518CC9/518CCF cache both slot and Player before518CD5.
		uu.Player.PlayerInd = 23
		ru.Player = other
		return 100
	}
	runtime.ServerPacket = func(index int32, _ [12]byte) int32 {
		if index != 29 {
			t.Fatalf("publication did not reload recipient slot: %d", index)
		}
		return 1
	}
	if PlayerObjectSendNative518C30(r, u, 1, 0, runtime) != 1 || initial.PeerUpdateFrames[11] != 100 || initial.PeerUpdateFrames[23] != 0 || other.PeerUpdateFrames[11] != 0 {
		t.Fatal("peer destination was reloaded after the frame accessor")
	}
}

func TestPlayerSendAnimationFrameBeforeExpiryReload518C30(t *testing.T) {
	r, u, _, uu, runtime := playerSendOrderFixture518C30()
	uu.Field41 = 80
	reads := 0
	runtime.Frame = func() uint32 {
		reads++
		if reads == 2 {
			// Original518DFC reads the frame before518E02 reloads expiry.
			uu.Field41 = 99
		}
		return 100
	}
	runtime.ServerPacket = func(_ int32, packet [12]byte) int32 {
		if packet[10] != 1 || packet[11] != 52 {
			t.Fatalf("expiry was cached before the frame read: %x", packet)
		}
		return 1
	}
	if PlayerObjectSendNative518C30(r, u, 1, 0, runtime) != 1 || reads != 2 || uu.Field41 != 99 {
		t.Fatal("animation frame/expiry reload ordering")
	}
}

func TestPlayerSendNativePeerLayout518C30(t *testing.T) {
	var p Player
	peer := unsafe.Offsetof(p.PeerUpdateFrames)
	end := peer + unsafe.Sizeof(p.PeerUpdateFrames)
	align := unsafe.Alignof(p.Field4580)
	// Field4580 retains the pointer alignment gap following the old 43 DWORDs.
	// The pre-port ARM64 product has BeastScrollLvl=5540, Field4580=5880.
	next := (end + align - 1) &^ (align - 1)
	if peer-unsafe.Offsetof(p.BeastScrollLvl) != 4452-4244 || unsafe.Sizeof(p.PeerUpdateFrames) != 128 || next != unsafe.Offsetof(p.Field4580) {
		t.Fatalf("peer=%d size=%d align=%d next=%d Field4580=%d", peer, unsafe.Sizeof(p.PeerUpdateFrames), align, next, unsafe.Offsetof(p.Field4580))
	}
	t.Logf("Player size=%d BeastScrollLvl=%d peer=%d Field4580=%d", unsafe.Sizeof(p), unsafe.Offsetof(p.BeastScrollLvl), peer, unsafe.Offsetof(p.Field4580))
}
