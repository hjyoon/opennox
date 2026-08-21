package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestGoldInit4F04B0NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantInitData := uintptr(692)
	wantPlayerSize := uintptr(4828)
	wantPlayerIndex := uintptr(2064)
	wantPlayerActive := uintptr(2092)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
		wantInitData = 760
		wantPlayerSize = 6160
		wantPlayerIndex = 2068
		wantPlayerActive = 2096
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerUnit", unsafe.Offsetof(Player{}.PlayerUnit), 2056},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.Active", unsafe.Offsetof(Player{}.Active), wantPlayerActive},
		{"GoldInitData size", unsafe.Sizeof(GoldInitData{}), 4},
		{"GoldInitData.Amount", unsafe.Offsetof(GoldInitData{}.Amount), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestGoldInit4F04B0NativeCachesInitDataAndUsesExactFields(t *testing.T) {
	type guardedData struct {
		data  GoldInitData
		guard uint32
	}
	entry := &guardedData{guard: 0xa5a5a5a5}
	replacement := &guardedData{data: GoldInitData{Amount: 0x11111111}, guard: 0x5a5a5a5a}
	gold := &Object{Field6_2: 0x1234, Worth: 0x55667788, InitData: unsafe.Pointer(&entry.data)}
	firstUnit := &Object{Field6_2: 0xaaaa, Experience: 1000, Worth: 0xbbbbbbbb}
	first := &Player{PlayerUnit: firstUnit}
	second := &Player{}
	nextCalls := 0
	randomCalls := 0

	got := goldInitNative4F04B0(gold, goldInitNativeDeps4F04B0{
		firstPlayer: func() *Player { return first },
		nextPlayer: func(player *Player) *Player {
			nextCalls++
			if player == first {
				return second
			}
			if player != second {
				t.Fatalf("unexpected player %p", player)
			}
			return nil
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			randomCalls++
			switch randomCalls {
			case 1:
				if minimum != 5 || maximum != 10 || path != goldInitScaledRandomPath4F04B0 || line != 1017 {
					t.Fatalf("scaled RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				gold.InitData = unsafe.Pointer(&replacement.data)
				return 7
			case 2:
				if minimum != 15 || maximum != 30 || path != goldInitBaseRandomPath4F04B0 || line != 1018 {
					t.Fatalf("base RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				return 20
			default:
				t.Fatalf("unexpected RNG call %d", randomCalls)
				return 0
			}
		},
	})

	if got != 20 || entry.data.Amount != 37 {
		t.Fatalf("return/entry Amount = %d/%d, want 20/37", got, entry.data.Amount)
	}
	if nextCalls != 2 || randomCalls != 2 {
		t.Fatalf("next/RNG calls = %d/%d, want 2/2", nextCalls, randomCalls)
	}
	if entry.guard != 0xa5a5a5a5 || replacement.data.Amount != 0x11111111 || replacement.guard != 0x5a5a5a5a {
		t.Fatalf("records changed unexpectedly: entry=%+v replacement=%+v", *entry, *replacement)
	}
	if gold.InitData != unsafe.Pointer(&replacement.data) || gold.Field6_2 != 0x1234 || gold.Worth != 0x55667788 {
		t.Fatalf("gold object fields changed: InitData=%p Field6_2=%#x Worth=%#x", gold.InitData, gold.Field6_2, gold.Worth)
	}
	if firstUnit.Experience != 1000 || firstUnit.Field6_2 != 0xaaaa || firstUnit.Worth != 0xbbbbbbbb {
		t.Fatalf("player unit fields changed: %+v", firstUnit)
	}
}

func TestGoldInit4F04B0NativeUsesServerPlayersAndLogicRNG(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	firstUnit := &Object{Experience: 1200}
	thirdUnit := &Object{Experience: 1800}
	s.Players.list = []Player{
		{Active: 1, PlayerInd: 0, PlayerUnit: firstUnit},
		{Active: 1, PlayerInd: 1},
		{Active: 1, PlayerInd: 2, PlayerUnit: thirdUnit},
	}
	data := &GoldInitData{}
	gold := &Object{InitData: unsafe.Pointer(data)}

	average := goldInitAverage4F04B0(3000, 3)
	upper := goldInitTruncQwordLow4F04B0(goldInitScale4F04B0(average, goldInitUpperScaleBits4F04B0))
	lower := goldInitTruncQwordLow4F04B0(goldInitScale4F04B0(average, goldInitLowerScaleBits4F04B0))
	negative := goldInitTruncQwordLow4F04B0(goldInitScale4F04B0(average, goldInitNegativeScaleBits4F04B0))
	wantRandom := prand.New(0)
	scaled := int32(wantRandom.IntClamp(int(lower), int(upper)))
	base := int32(wantRandom.IntClamp(15, 30))
	wantAmount := uint32(scaled) - uint32(negative) + uint32(base)

	got := s.GoldInit4F04B0(gold)
	if got != base || data.Amount != wantAmount {
		t.Fatalf("return/Amount = %d/%d, want %d/%d", got, data.Amount, base, wantAmount)
	}
	if index := s.Rand.Logic.Index(); index != 2 {
		t.Fatalf("logic RNG index = %d, want 2", index)
	}
}

func TestGoldInit4F04B0NativeNonzeroAmountReturnsPointerLowDword(t *testing.T) {
	data := &GoldInitData{Amount: 1}
	unit := &Object{InitData: unsafe.Pointer(data)}
	want := int32(uint32(uintptr(unsafe.Pointer(unit))))
	got := goldInitNative4F04B0(unit, goldInitNativeDeps4F04B0{})
	if got != want {
		t.Fatalf("return = %#x, want pointer low dword %#x", uint32(got), uint32(want))
	}
	if data.Amount != 1 {
		t.Fatalf("Amount = %d, want 1", data.Amount)
	}
}

func TestGoldInit4F04B0NativeNilFaultBoundaries(t *testing.T) {
	t.Run("nil-unit", func(t *testing.T) {
		playerCalls := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil Object did not preserve the original InitData-load fault")
			}
			if playerCalls != 0 {
				t.Fatalf("player calls = %d, want 0", playerCalls)
			}
		}()
		goldInitNative4F04B0(nil, goldInitNativeDeps4F04B0{
			firstPlayer: func() *Player { playerCalls++; return nil },
		})
	})

	t.Run("nil-init-data", func(t *testing.T) {
		playerCalls := 0
		defer func() {
			if recover() == nil {
				t.Fatal("nil GoldInitData did not preserve the original Amount-load fault")
			}
			if playerCalls != 0 {
				t.Fatalf("player calls = %d, want 0", playerCalls)
			}
		}()
		goldInitNative4F04B0(&Object{}, goldInitNativeDeps4F04B0{
			firstPlayer: func() *Player { playerCalls++; return nil },
		})
	})
}
