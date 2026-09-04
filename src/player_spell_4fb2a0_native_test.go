package opennox

import (
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerSpellNative4FB2A0Layouts(t *testing.T) {
	wantObjectUpdate := uintptr(748)
	wantLeaf := uintptr(184)
	wantPlayer := uintptr(276)
	wantCursorObj := uintptr(288)
	wantPlayerInd := uintptr(2064)
	wantCursor := uintptr(2284)
	wantTarget := uintptr(3640)
	wantSpellLevels := uintptr(3696)
	wantArgSize := uintptr(12)
	wantArgPos := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectUpdate = 872
		wantLeaf = 232
		wantPlayer = 336
		wantCursorObj = 360
		wantPlayerInd = 2068
		wantCursor = 2288
		wantTarget = 4928
		wantSpellLevels = 4992
		wantArgSize = 16
		wantArgPos = 8
	}

	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(server.Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.State", unsafe.Offsetof(server.PlayerUpdateData{}.State), 88},
		{"PlayerUpdateData.SpellPhonemeLeaf", unsafe.Offsetof(server.PlayerUpdateData{}.SpellPhonemeLeaf), wantLeaf},
		{"PlayerUpdateData.Player", unsafe.Offsetof(server.PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.CursorObj", unsafe.Offsetof(server.PlayerUpdateData{}.CursorObj), wantCursorObj},
		{"Player.PlayerInd", unsafe.Offsetof(server.Player{}.PlayerInd), wantPlayerInd},
		{"Player.CursorVec", unsafe.Offsetof(server.Player{}.CursorVec), wantCursor},
		{"Player.Obj3640", unsafe.Offsetof(server.Player{}.Obj3640), wantTarget},
		{"Player.SpellLvl", unsafe.Offsetof(server.Player{}.SpellLvl), wantSpellLevels},
		{"SpellAcceptArg size", unsafe.Sizeof(server.SpellAcceptArg{}), wantArgSize},
		{"SpellAcceptArg.Obj", unsafe.Offsetof(server.SpellAcceptArg{}.Obj), 0},
		{"SpellAcceptArg.Pos", unsafe.Offsetof(server.SpellAcceptArg{}.Pos), wantArgPos},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerSpellNative4FB2A0PreservesPointersAndLiveReloads(t *testing.T) {
	unit := &server.Object{}
	target2 := &server.Object{}
	target3 := &server.Object{}
	root := &server.PhonemeLeaf{}
	leaves := []*server.PhonemeLeaf{{Ind: 10}, {Ind: 11}, {Ind: 12}, {Ind: 13}, {Ind: 14}}
	players := []*server.Player{
		{PlayerInd: 1},
		{PlayerInd: 2, Obj3640: target2},
		{PlayerInd: 3, Obj3640: target3},
		{PlayerInd: 4, CursorVec: image.Pt(-123, 456)},
		{PlayerInd: 5},
		{PlayerInd: 6},
	}
	players[0].SpellLvl[10] = 1
	if unsafe.Sizeof(uintptr(0)) == 8 {
		players[3].CursorVec = image.Pt(int(int64(1)<<32|123), int(-(int64(1)<<32)-456))
	}
	update := &server.PlayerUpdateData{
		State:            server.PlayerState2,
		SpellPhonemeLeaf: leaves[0],
		Player:           players[0],
	}
	unit.UpdateData = unsafe.Pointer(update)

	var gameFlagCalls int
	var informedIndex ntype.PlayerInd
	var informedID int
	var reportIndex ntype.PlayerInd
	var reportID spell.ID
	var reportStatus byte
	var castArgAddress uintptr
	deps := playerSpellNativeDeps4FB2A0{
		phonemeRoot: root,
		hasGameFlag: func(flag uint32) bool {
			gameFlagCalls++
			if flag != playerSpellQuestFlag4FB2A0 {
				t.Fatalf("game flag = %#x, want %#x", flag, playerSpellQuestFlag4FB2A0)
			}
			return true
		},
		hasSpellFlags: func(id spell.ID, flags things.SpellFlags) bool {
			switch flags {
			case things.SpellOffensive:
				if id != 12 {
					t.Fatalf("offensive spell = %d, want 12", id)
				}
				update.Player = players[2]
				return true
			case things.SpellFlagUnk21:
				if id != 13 {
					t.Fatalf("report-filter spell = %d, want 13", id)
				}
				update.SpellPhonemeLeaf = leaves[4]
				update.Player = players[5]
				return false
			default:
				t.Fatalf("unexpected spell flags %#x", flags)
				return false
			}
		},
		isEnemy: func(gotUnit, target *server.Object) bool {
			if gotUnit != unit || target != target3 {
				t.Fatalf("enemy args = %p/%p, want %p/%p", gotUnit, target, unit, target3)
			}
			update.Player = players[3]
			return true
		},
		precheck: func(gotUnit *server.Object, id spell.ID) int32 {
			if gotUnit != unit || id != 10 {
				t.Fatalf("precheck = %p/%d, want %p/10", gotUnit, id, unit)
			}
			update.SpellPhonemeLeaf = leaves[1]
			return 0
		},
		checkCantCast: func(gotUnit *server.Object, id spell.ID, bypass int32) int32 {
			if gotUnit != unit || id != 11 || bypass != 0 {
				t.Fatalf("cast check = %p/%d/%d, want %p/11/0", gotUnit, id, bypass, unit)
			}
			update.SpellPhonemeLeaf = leaves[2]
			return 0
		},
		informText: func(index ntype.PlayerInd, code byte, value int) {
			if code != 1 {
				t.Fatalf("inform code = %d, want success 1", code)
			}
			informedIndex = index
			informedID = value
		},
		audioEvent: func(sound.ID, *server.Object, int, uint32) {
			t.Fatal("successful spell emitted failure audio")
		},
		chargeMana: func(gotUnit *server.Object, id spell.ID, amount int32) int32 {
			if gotUnit != unit || id != 12 || amount != 1 {
				t.Fatalf("mana charge = %p/%d/%d, want %p/12/1", gotUnit, id, amount, unit)
			}
			update.Player = players[1]
			return 37
		},
		castSpell: func(id spell.ID, gotUnit *server.Object, arg *server.SpellAcceptArg) bool {
			castArgAddress = uintptr(unsafe.Pointer(arg))
			if id != 12 || gotUnit != unit || arg.Obj != target2 {
				t.Fatalf("cast = %d/%p/%p, want 12/%p/%p", id, gotUnit, arg.Obj, unit, target2)
			}
			wantX, wantY := float32(-123), float32(456)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				wantX, wantY = 123, -456
			}
			if arg.Pos.X != wantX || arg.Pos.Y != wantY {
				t.Fatalf("cast position = %v, want (%g,%g)", arg.Pos, wantX, wantY)
			}
			update.Player = players[4]
			update.SpellPhonemeLeaf = leaves[3]
			return true
		},
		refundMana: func(*server.Object, int32) {
			t.Fatal("successful spell refunded mana")
		},
		setState: func(gotUnit *server.Object, state server.PlayerState) {
			if gotUnit != unit || state != server.PlayerState13 {
				t.Fatalf("state = %p/%d, want %p/13", gotUnit, state, unit)
			}
			update.State = state
		},
		unknownMessage: func() string {
			t.Fatal("known spell requested unknown message")
			return ""
		},
		lineMessage: func(*server.Object, string) {
			t.Fatal("known spell emitted unknown line")
		},
		reportSpell: func(index ntype.PlayerInd, id spell.ID, status byte) {
			reportIndex, reportID, reportStatus = index, id, status
		},
	}

	playerSpellNative4FB2A0(unit, deps)
	if gameFlagCalls != 2 {
		t.Fatalf("game flag calls = %d, want 2", gameFlagCalls)
	}
	if informedIndex != 5 || informedID != 13 {
		t.Fatalf("success inform = %d/%d, want 5/13", informedIndex, informedID)
	}
	if update.State != server.PlayerState13 {
		t.Fatalf("state = %d, want 13", update.State)
	}
	if reportIndex != 6 || reportID != 14 || reportStatus != 15 {
		t.Fatalf("report = %d/%d/%d, want 6/14/15", reportIndex, reportID, reportStatus)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := map[string]uintptr{
			"unit": uintptr(unsafe.Pointer(unit)), "target-2": uintptr(unsafe.Pointer(target2)),
			"target-3": uintptr(unsafe.Pointer(target3)), "update": uintptr(unsafe.Pointer(update)),
			"root": uintptr(unsafe.Pointer(root)),
		}
		for i, leaf := range leaves {
			pointers["leaf-"+string(rune('0'+i))] = uintptr(unsafe.Pointer(leaf))
		}
		for i, player := range players {
			pointers["player-"+string(rune('0'+i))] = uintptr(unsafe.Pointer(player))
		}
		for name, pointer := range pointers {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}
	// SpellAcceptArg is deliberately allocated on the C heap by the native
	// adapter. A 64-bit allocator may place it below 4 GiB; the callback's exact
	// arg.Obj == target2 check above proves that its pointer field stayed native
	// width without depending on allocator placement.
	if castArgAddress == 0 {
		t.Fatal("cast spell received a nil argument")
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(target2)
	runtime.KeepAlive(target3)
	runtime.KeepAlive(root)
	runtime.KeepAlive(leaves)
	runtime.KeepAlive(players)
	runtime.KeepAlive(update)
}
