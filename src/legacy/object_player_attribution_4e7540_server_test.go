package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestRecordPlayerAttributionNative4E7540WritesNativePlayerState(t *testing.T) {
	frameCalls := 0
	frameHook := func() uint32 {
		frameCalls++
		return 0xfedcba98
	}

	sourcePlayer := &server.Player{PlayerInd: 0xfe}
	targetPlayer := &server.Player{}
	sourceUpdate := &server.PlayerUpdateData{Player: sourcePlayer}
	targetUpdate := &server.PlayerUpdateData{Player: targetPlayer}
	source := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate)}
	target := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(targetUpdate)}

	recordPlayerAttributionNative4E7540(source, target, frameHook)
	pending, playerIndex, frame := targetPlayer.LastAggressorState()
	if pending != 1 || playerIndex != 0xfe || frame != 0xfedcba98 {
		t.Fatalf("attribution = (%d, %#x, %#x), want (1, 0xfe, 0xfedcba98)", pending, playerIndex, frame)
	}
	if frameCalls != 1 {
		t.Fatalf("frame calls = %d, want 1", frameCalls)
	}
}

func TestRecordPlayerAttributionNative4E7540GuardsBeforeUpdateData(t *testing.T) {
	frameHook := func() uint32 {
		t.Fatal("guarded attribution read the frame")
		return 0
	}

	player := &server.Object{ObjClass: object.ClassPlayer}
	nonPlayer := &server.Object{ObjClass: object.ClassMonster}
	for _, tc := range []struct {
		name           string
		source, target *server.Object
	}{
		{name: "nil source", target: player},
		{name: "nil target", source: player},
		{name: "non-player source", source: nonPlayer, target: player},
		{name: "non-player target", source: player, target: nonPlayer},
		{name: "same object", source: player, target: player},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recordPlayerAttributionNative4E7540(tc.source, tc.target, frameHook)
		})
	}
}
