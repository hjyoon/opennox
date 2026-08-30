package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func gameCaptureMagicPlayer4FDC10(weaponEquip uint32, owned *server.Object, team server.TeamID) *server.Object {
	pl := &server.Player{WeaponEquip: weaponEquip}
	update := &server.PlayerUpdateData{Player: pl}
	return &server.Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    server.ObjectTeam{ID: team},
		Field129:   owned,
		UpdateData: unsafe.Pointer(update),
	}
}

func TestGameCaptureMagicAllowed4FDC10Gates(t *testing.T) {
	const ballID, crownID = uint16(0x1234), uint16(0xf234)
	if gameCaptureMagicAllowed4FDC10(true, nil, 0, ballID, crownID) {
		t.Fatal("nil unit was accepted")
	}
	nonPlayer := &server.Object{ObjClass: object.ClassMonster}
	if !gameCaptureMagicAllowed4FDC10(true, nonPlayer, noxflags.GameModeCTF, ballID, crownID) {
		t.Fatal("non-player was rejected")
	}
	player := gameCaptureMagicPlayer4FDC10(1, &server.Object{TypeInd: ballID}, 1)
	if !gameCaptureMagicAllowed4FDC10(false, player, noxflags.GameModeCTF, ballID, crownID) {
		t.Fatal("spell without restriction flag was rejected")
	}
}

func TestGameCaptureMagicAllowed4FDC10CTF(t *testing.T) {
	const ballID, crownID = uint16(11), uint16(12)
	for _, tc := range []struct {
		name   string
		equip  uint32
		flags  noxflags.GameFlag
		wanted bool
	}{
		{name: "clear", equip: 0, flags: noxflags.GameModeCTF, wanted: true},
		{name: "carrier", equip: 1, flags: noxflags.GameModeCTF, wanted: false},
		{name: "other bits", equip: 0xfffffffe, flags: noxflags.GameModeCTF, wanted: true},
		{name: "ctf priority", equip: 1, flags: noxflags.GameModeCTF | noxflags.GameModeFlagBall | noxflags.GameModeKOTR, wanted: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := gameCaptureMagicPlayer4FDC10(tc.equip, nil, 0)
			if got := gameCaptureMagicAllowed4FDC10(true, player, tc.flags, ballID, crownID); got != tc.wanted {
				t.Fatalf("allowed = %v, want %v", got, tc.wanted)
			}
		})
	}
}

func TestGameCaptureMagicAllowed4FDC10FlagBall(t *testing.T) {
	const ballID, crownID = uint16(0xf234), uint16(12)
	ball := &server.Object{TypeInd: ballID}
	first := &server.Object{TypeInd: 7, Field128: ball}
	player := gameCaptureMagicPlayer4FDC10(0, first, 0)
	if gameCaptureMagicAllowed4FDC10(true, player, noxflags.GameModeFlagBall, ballID, crownID) {
		t.Fatal("GameBall carrier was accepted")
	}
	ball.TypeInd = ballID - 1
	if !gameCaptureMagicAllowed4FDC10(true, player, noxflags.GameModeFlagBall, ballID, crownID) {
		t.Fatal("owned list without GameBall was rejected")
	}
}

func TestGameCaptureMagicAllowed4FDC10KOTR(t *testing.T) {
	const ballID, crownID = uint16(11), uint16(0xf234)
	crown := &server.Object{TypeInd: crownID}
	first := &server.Object{TypeInd: 7, Field128: crown}

	player := gameCaptureMagicPlayer4FDC10(0, first, 3)
	if gameCaptureMagicAllowed4FDC10(true, player, noxflags.GameModeKOTR, ballID, crownID) {
		t.Fatal("teamed Crown carrier was accepted")
	}
	player.TeamVal.ID = 0
	if !gameCaptureMagicAllowed4FDC10(true, player, noxflags.GameModeKOTR, ballID, crownID) {
		t.Fatal("unteamed Crown owner was rejected")
	}
	crown.TypeInd = crownID - 1
	player.TeamVal.ID = 3
	if !gameCaptureMagicAllowed4FDC10(true, player, noxflags.GameModeKOTR, ballID, crownID) {
		t.Fatal("teamed owner without Crown was rejected")
	}
}

func TestGameCaptureMagicAllowed4FDC10NativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	var wantUpdateData, wantFirstOwned, wantNextOwned, wantPlayer uintptr
	switch ptrSize {
	case 4:
		wantUpdateData, wantFirstOwned, wantNextOwned, wantPlayer = 748, 516, 512, 276
	case 8:
		wantUpdateData, wantFirstOwned, wantNextOwned, wantPlayer = 872, 568, 560, 336
	default:
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Object.UpdateData", got: unsafe.Offsetof(server.Object{}.UpdateData), want: wantUpdateData},
		{name: "Object.Field129", got: unsafe.Offsetof(server.Object{}.Field129), want: wantFirstOwned},
		{name: "Object.Field128", got: unsafe.Offsetof(server.Object{}.Field128), want: wantNextOwned},
		{name: "PlayerUpdateData.Player", got: unsafe.Offsetof(server.PlayerUpdateData{}.Player), want: wantPlayer},
		{name: "Player.WeaponEquip", got: unsafe.Offsetof(server.Player{}.WeaponEquip), want: 4},
	} {
		if field.got != field.want {
			t.Errorf("%s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
}
