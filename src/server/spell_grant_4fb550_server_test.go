package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

func TestSpellGrant4FB550NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantUpdateTrade := uintptr(280)
	wantPlayerSize := uintptr(4828)
	wantPlayerIndex := uintptr(2064)
	wantSpellLevels := uintptr(3696)
	wantProtection := uintptr(4636)
	wantNotifyField := uintptr(4792)
	wantTradeSize := uintptr(64)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectUpdate = 872
		wantUpdateSize = 656
		wantUpdatePlayer = 336
		wantUpdateTrade = 344
		wantPlayerSize = 6160
		wantPlayerIndex = 2068
		wantSpellLevels = 4992
		wantProtection = 5940
		wantNotifyField = 6096
		wantTradeSize = 104
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"PlayerUpdateData.Trade70", unsafe.Offsetof(PlayerUpdateData{}.Trade70), wantUpdateTrade},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerUnit", unsafe.Offsetof(Player{}.PlayerUnit), 2056},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"Player.Prot4636", unsafe.Offsetof(Player{}.Prot4636), wantProtection},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantNotifyField},
		{"TradeSession size", unsafe.Sizeof(TradeSession{}), wantTradeSize},
		{"spell level width", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{"spell level count", uintptr(len(Player{}.SpellLvl)), 137},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellGrantToPlayerNative4FB550PreservesNativePointersAndArguments(t *testing.T) {
	const spellID = int32(10)
	override := int32(-0x1234567)
	trade := new(TradeSession)
	player := &Player{Prot4636: 0x89abcdef, Field4792: 0x10203040}
	update := &PlayerUpdateData{Player: player, Trade70: trade}
	unit := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}

	var (
		protectedToken  uint32
		protectedSpell  int32
		protectedLevel  int32
		audioUnit       *Object
		rewardRecipient *Object
		rewardSource    *Object
		shopTrade       *TradeSession
		reportUnit      *Object
		reportArgs      [3]int32
		soloChecks      int
	)
	deps := spellGrantNativeDeps4FB550{
		gameFlagsCheck: func(mask uint32) int32 {
			switch mask {
			case spellGrantCoopQuestFlag4FB550, spellGrantQuestFlag4FB550:
				return 0
			case spellGrantSoloFlag4FB550:
				soloChecks++
				if soloChecks == 2 {
					return 1
				}
				return 0
			default:
				t.Fatalf("unexpected game flag %#x", mask)
				return 0
			}
		},
		loadString: func(string, string, int) string {
			t.Fatal("valid spell loaded an error string")
			return ""
		},
		sendLineMessage: func(*Object, string) {
			t.Fatal("valid spell sent an error line")
		},
		awardProtection: func(token uint32, gotSpell, level int32) {
			protectedToken = token
			protectedSpell = gotSpell
			protectedLevel = level
		},
		spellHasFlags: func(gotSpell, flags int32) int32 {
			if gotSpell != spellID {
				t.Fatalf("spell flag ID = %d, want %d", gotSpell, spellID)
			}
			switch flags {
			case spellGrantFamilySourceA4FB550, spellGrantFamilySourceB4FB550, spellGrantFamilySourceC4FB550:
				return 0
			default:
				t.Fatalf("unexpected spell flags %#x", flags)
				return 0
			}
		},
		spellIsValid: func(int32) int32 {
			t.Fatal("non-family spell checked a related spell")
			return 0
		},
		audio: func(id uint32, gotUnit *Object, kind int32, code uint32) {
			if id != spellGrantSound4FB550 || kind != spellGrantSoundKind4FB550 || code != 0 {
				t.Fatalf("audio args = %d/%d/%d", id, kind, code)
			}
			audioUnit = gotUnit
		},
		rewardNotify: func(recipient *Object, kind int32, source *Object, gotSpell int32) {
			if kind != spellGrantRewardKind4FB550 || gotSpell != spellID {
				t.Fatalf("reward args = %d/%d", kind, gotSpell)
			}
			rewardRecipient = recipient
			rewardSource = source
		},
		checkPlayerState: func(got *Object) int32 {
			if got != unit {
				t.Fatalf("state unit = %p, want %p", got, unit)
			}
			return 1
		},
		firstPlayer: func() *Player {
			t.Fatal("state-gated grant iterated players")
			return nil
		},
		nextPlayer: func(*Player) *Player {
			t.Fatal("state-gated grant advanced players")
			return nil
		},
		shopExit: func(got *TradeSession) {
			shopTrade = got
		},
		reportSpellAward: func(gotUnit *Object, gotSpell, notify, shop int32) {
			reportUnit = gotUnit
			reportArgs = [3]int32{gotSpell, notify, shop}
		},
	}

	var pin runtime.Pinner
	pin.Pin(trade)
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()

	if got := spellGrantToPlayerNative4FB550(unit, spellID, 1, 1, override, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := player.SpellLvl[spellID]; got != uint32(override) {
		t.Fatalf("spell level = %#x, want override bits %#x", got, uint32(override))
	}
	if protectedToken != player.Prot4636 || protectedSpell != spellID || protectedLevel != override {
		t.Fatalf("protection = %#x/%d/%d", protectedToken, protectedSpell, protectedLevel)
	}
	if audioUnit != unit || rewardRecipient != unit || rewardSource != unit {
		t.Fatalf("native object links = audio %p, reward %p/%p; want %p", audioUnit, rewardRecipient, rewardSource, unit)
	}
	if shopTrade != trade {
		t.Fatalf("shop trade = %p, want %p", shopTrade, trade)
	}
	if reportUnit != unit || reportArgs != [3]int32{spellID, 1, 1} {
		t.Fatalf("report = %p/%v, want %p/[10 1 1]", reportUnit, reportArgs, unit)
	}
	if soloChecks != 2 {
		t.Fatalf("solo checks = %d, want 2", soloChecks)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"trade":  uintptr(unsafe.Pointer(trade)),
			"player": uintptr(unsafe.Pointer(player)),
			"update": uintptr(unsafe.Pointer(update)),
			"unit":   uintptr(unsafe.Pointer(unit)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(trade)
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestSpellGrantRewardNotifyNative4FAD50BuildsExactPacket(t *testing.T) {
	s := new(Server)
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Player: player}
	recipient := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}
	source := &Object{NetCode: 0x1234abcd}

	var (
		gotRecipient int
		gotPacket    []byte
		gotRelated   *Object
		gotRemove    int
		gotSequence  int
	)
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		gotRecipient = index
		gotPacket = append([]byte(nil), packet...)
		gotRelated = related
		gotRemove = remove
		gotSequence = sequence
		return 77
	}

	spellGrantRewardNotifyNative4FAD50(s, recipient, 2, source, 0x123)
	want := []byte{byte(netmsg.MSG_INFORM), 32, 0x23, 0xcd, 0xab}
	if gotRecipient != 0xfe || string(gotPacket) != string(want) || gotRelated != nil || gotRemove != 1 || gotSequence != 0 {
		t.Fatalf("packet = recipient %d bytes %v related %p remove %d sequence %d", gotRecipient, gotPacket, gotRelated, gotRemove, gotSequence)
	}

	gotPacket = nil
	spellGrantRewardNotifyNative4FAD50(s, recipient, 3, source, 10)
	if gotPacket != nil {
		t.Fatalf("invalid reward kind sent %v", gotPacket)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(recipient)
	runtime.KeepAlive(source)
}

func TestSpellGrantReportNative4D7F90BuildsExactPacket(t *testing.T) {
	s := new(Server)
	player := &Player{PlayerInd: 0x81}
	player.SpellLvl[10] = 0x102
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}

	var (
		gotRecipient int
		gotPacket    []byte
		gotRelated   *Object
		gotRemove    int
		gotSequence  int
	)
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		gotRecipient = index
		gotPacket = append([]byte(nil), packet...)
		gotRelated = related
		gotRemove = remove
		gotSequence = sequence
		return 88
	}

	spellGrantReportNative4D7F90(s, unit, 10, 0x123, -1)
	want := []byte{byte(netmsg.MSG_REPORT_SPELL_AWARD), 10, 2, 0xa3}
	if gotRecipient != 0x81 || string(gotPacket) != string(want) || gotRelated != nil || gotRemove != 1 || gotSequence != 1 {
		t.Fatalf("packet = recipient %d bytes %v related %p remove %d sequence %d", gotRecipient, gotPacket, gotRelated, gotRemove, gotSequence)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
