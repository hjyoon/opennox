package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestGameBallCarrierStateNative4EB9B0UsesNativeOwnerChainAndRecord(t *testing.T) {
	update := &GameBallUpdateData4EA800{
		Carrier:            &Object{},
		TeamID:             0xaabbccdd,
		Ticks:              0x0123456789abcdef,
		CarrierFrame:       7,
		PossessionDuration: 91,
		ResetVelocity:      12.5,
		Reserved:           0x76543210,
	}
	ball := &Object{UpdateData: unsafe.Pointer(update)}
	player := &Object{ObjClass: object.ClassPlayer, TeamVal: ObjectTeam{ID: 0xab}}
	child := &Object{ObjOwner: player}
	frameLoads := 0
	got := gameBallCarrierStateNative4EB9B0(ball, child, func() uint32 {
		frameLoads++
		return 0x89abcdef
	})
	if got != player {
		t.Fatalf("result = %p, want player %p", got, player)
	}
	if update.Carrier != player || update.TeamID != 0xab || update.CarrierFrame != 0x89abcdef {
		t.Fatalf("carrier state = %+v", *update)
	}
	if frameLoads != 1 {
		t.Fatalf("frame loads = %d, want 1", frameLoads)
	}
	if update.Ticks != 0x0123456789abcdef || update.PossessionDuration != 91 ||
		update.ResetVelocity != 12.5 || update.Reserved != 0x76543210 {
		t.Fatalf("unrelated GameBall fields changed: %+v", *update)
	}
}

func TestGameBallCarrierStateNative4EB9B0ReturnsNonPlayerWhileClearingPrefix(t *testing.T) {
	oldCarrier := &Object{}
	update := &GameBallUpdateData4EA800{
		Carrier:      oldCarrier,
		TeamID:       9,
		CarrierFrame: 0x11223344,
	}
	ball := &Object{UpdateData: unsafe.Pointer(update)}
	terminal := &Object{ObjClass: object.ClassMonster}
	got := gameBallCarrierStateNative4EB9B0(ball, terminal, func() uint32 {
		t.Fatal("failure path loaded frame")
		return 0
	})
	if got != terminal {
		t.Fatalf("result = %p, want non-Player terminal %p", got, terminal)
	}
	if update.Carrier != nil || update.TeamID != 0 || update.CarrierFrame != 0x11223344 {
		t.Fatalf("cleared state = %+v", *update)
	}
	update.Carrier = oldCarrier
	update.TeamID = 17
	got = gameBallCarrierStateNative4EB9B0(ball, nil, func() uint32 {
		t.Fatal("nil-target path loaded frame")
		return 0
	})
	if got != nil {
		t.Fatalf("nil-target result = %p, want nil", got)
	}
	if update.Carrier != nil || update.TeamID != 0 || update.CarrierFrame != 0x11223344 {
		t.Fatalf("nil-target state = %+v", *update)
	}
}

func TestGameBallCarrierStateNative4EB9B0PreservesNativeFaults(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	for _, tc := range []struct {
		name   string
		ball   *Object
		target *Object
	}{
		{name: "nil ball", target: player},
		{name: "nil update data", ball: &Object{}, target: player},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("native nil dereference did not fault")
				}
			}()
			gameBallCarrierStateNative4EB9B0(tc.ball, tc.target, func() uint32 {
				t.Fatal("fault path loaded frame")
				return 0
			})
		})
	}
}

func TestGameBallCarrierStateServer4EB9B0UsesServerFrame(t *testing.T) {
	update := &GameBallUpdateData4EA800{}
	ball := &Object{UpdateData: unsafe.Pointer(update)}
	player := &Object{ObjClass: object.ClassPlayer, TeamVal: ObjectTeam{ID: 6}}
	s := new(Server)
	s.SetFrame(0xfedcba98)
	if got := s.GameBallCarrierState4EB9B0(ball, player); got != player {
		t.Fatalf("result = %p, want player %p", got, player)
	}
	if update.Carrier != player || update.TeamID != 6 || update.CarrierFrame != 0xfedcba98 {
		t.Fatalf("server carrier state = %+v", *update)
	}
}
