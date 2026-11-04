package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func shopWandRepairFixture(t *testing.T, health uint16, charge, maximum uint8, progress uint32) (*Server, *Object, *Object, *TradeSession) {
	t.Helper()
	idata := nativeTradeTestValue(t, ShopkeeperInitData{BuyMultiplier: 2})
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	attrs := nativeTradeTestValue(t, ModifierInitData{})
	data := nativeTradeTestValue(t, WandUseData{Charge: charge, MaxCharge: maximum, Progress: progress})
	item := nativeTradeTestValue(t, Object{
		ObjClass: object.ClassWand, ObjSubClass: 0x80000, NetCode: 0x4321, Worth: 100,
		HealthData: nativeTradeTestValue(t, HealthData{Cur: health, Max: 100}),
		InitData:   unsafe.Pointer(attrs), UseData: UseDataPtr{Ptr: unsafe.Pointer(data)},
	})
	player := nativeTradeTestValue(t, Player{GoldVal: 200, ProtPlayerGold: 0x89abcdef})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	unit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: item})
	item.InvHolder = unit
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(unit, merchant)
	update.Trade70 = session
	t.Cleanup(func() {
		if !s.ReleaseTradeSessionNative510000(session) {
			t.Error("native shop session was not released")
		}
	})
	return s, unit, item, session
}

func TestShopRepairWandQuote5108D0(t *testing.T) {
	for _, tc := range []struct {
		name    string
		class   object.Class
		sub     uint32
		charge  uint8
		maximum uint8
		noUse   bool
		noHP    bool
		quote   bool
	}{
		{name: "fireball partial", class: object.ClassWand, sub: 0x80000, charge: 3, maximum: 4, quote: true},
		{name: "fireball empty", class: object.ClassWand, sub: 0x80000, maximum: 4, quote: true},
		{name: "other rechargeable staff", class: object.ClassWand, sub: 0x4000000, charge: 3, maximum: 4, quote: true},
		{name: "full staff", class: object.ClassWand, sub: 0x80000, charge: 4, maximum: 4},
		{name: "unlimited staff", class: object.ClassWand, sub: 0x80000},
		{name: "not a wand", class: object.ClassWeapon, sub: 0x80000, charge: 3, maximum: 4},
		{name: "not rechargeable subclass", class: object.ClassWand, sub: 0x8000000, charge: 3, maximum: 4},
		{name: "missing use data", class: object.ClassWand, sub: 0x80000, charge: 3, maximum: 4, noUse: true},
		{name: "charge only without durability", class: object.ClassWand, sub: 0x80000, charge: 3, maximum: 4, noHP: true, quote: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, item, session := shopWandRepairFixture(t, 100, tc.charge, tc.maximum, 75)
			item.ObjClass, item.ObjSubClass = tc.class, object.SubClass(tc.sub)
			if tc.noUse {
				item.UseData.Ptr = nil
			}
			if tc.noHP {
				item.HealthData = nil
			}
			quotes, rejects := 0, 0
			got := s.QuoteShopRepairNative5108D0(unit, session, uint16(item.NetCode), ShopRepairRuntime5108D0{
				RepairCoefficient: 0.5,
				SendQuote: func(p *Player, got *Object, cost uint32) {
					quotes++
					want := uint32(25)
					if tc.charge == 0 {
						// The current-value floor is 1; FISTP nearest-even rounds 99.5 to 100.
						want = 100
					}
					if p != unit.UpdateDataPlayer().Player || got != item || cost != want {
						t.Fatalf("quote = %p/%p/%d, want player/item/%d", p, got, cost, want)
					}
				},
				PlayRejectSound: func(*Object) { rejects++ },
			})
			want := ShopRepairNotDamaged5108D0
			wantQuotes, wantRejects := 0, 1
			if tc.quote {
				want = ShopRepairQuoted5108D0
				wantQuotes, wantRejects = 1, 0
			}
			if got != want || quotes != wantQuotes || rejects != wantRejects {
				t.Fatalf("repair quote = %d quotes=%d rejects=%d, want %d", got, quotes, rejects, want)
			}
		})
	}
}

func TestShopRepairWandCompletion510AE0(t *testing.T) {
	for _, tc := range []struct {
		name     string
		health   uint16
		charge   uint8
		progress uint32
		cost     uint32
	}{
		{"charge only", 100, 3, 75, 25},
		{"empty", 100, 0, 0, 100},
		{"durability and charges", 50, 3, 75, 62},
		{"durability only", 50, 4, 100, 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, item, session := shopWandRepairFixture(t, tc.health, tc.charge, 4, tc.progress)
			player := unit.UpdateDataPlayer().Player
			events := []string{}
			got := s.RepairShopItemNative510AE0(unit, session, uint16(item.NetCode), ShopRepairRuntime5108D0{
				RepairCoefficient: 0.5,
				ProtectGold: func(token uint32, delta int32) {
					events = append(events, "protect")
					if token != 0x89abcdef || delta != -int32(tc.cost) || player.GoldVal != 200-tc.cost {
						t.Fatalf("protection = %#x/%d gold=%d", token, delta, player.GoldVal)
					}
				},
				SetHealth: func(got *Object, amount uint16) {
					events = append(events, "health")
					got.HealthData.Cur = amount
				},
				ReportHealth: func(*Player, *Object) { events = append(events, "report-health") },
				ReportCharges: func(p *Player, got *Object, charge, maximum uint8) {
					events = append(events, "charges")
					if p != player || got != item || charge != 4 || maximum != 4 || item.HealthData.Cur != 100 || player.GoldVal != 200-tc.cost {
						t.Fatalf("charge report = %p/%p/%d/%d health=%d gold=%d", p, got, charge, maximum, item.HealthData.Cur, player.GoldVal)
					}
				},
				ReportGold:      func(*Player, *Object) { events = append(events, "gold") },
				PlayRepairSound: func(*Object) { events = append(events, "sound") },
			})
			data := item.UseData.AsWand()
			if got != ShopRepairComplete5108D0 || item.HealthData.Cur != 100 || data.Charge != 4 || data.Progress != 100 || player.GoldVal != 200-tc.cost {
				t.Fatalf("repair = %d health=%d charge=%d/4 progress=%d gold=%d", got, item.HealthData.Cur, data.Charge, data.Progress, player.GoldVal)
			}
			want := []string{"protect", "health", "report-health"}
			if tc.charge < 4 {
				want = append(want, "charges")
			}
			want = append(want, "gold", "sound")
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("repair events = %v, want %v", events, want)
			}
			if got := s.QuoteShopRepairNative5108D0(unit, session, uint16(item.NetCode), ShopRepairRuntime5108D0{}); got != ShopRepairNotDamaged5108D0 {
				t.Fatalf("repaired wand quoted again: %d", got)
			}
		})
	}
}
