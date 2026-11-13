package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type regenerationRawDefaultServer4E01D0 struct {
	Server
	srv *server.Server
}

func (s *regenerationRawDefaultServer4E01D0) S() *server.Server { return s.srv }

func regenerationRawRequireDefaults4E01D0(t *testing.T, hpObserver bool) {
	t.Helper()
	if regenerationRawABI4E01D0() != 2 || regenerationRawEngine4E01D0 == nil || reflect.ValueOf(regenerationRawEngine4E01D0).Pointer() != reflect.ValueOf(regenerationRawNative4E01D0).Pointer() {
		t.Fatal("actual raw native C ABI and production engine required")
	}
	if regenerationRawObserveRead4E01D0 != nil || regenerationRawObserveCall4E01D0 != nil || regenerationRawObserveArithmetic4E01D0 != nil || (!hpObserver && server.ObserveUnitHPRead4EE780 != nil) {
		t.Fatal("ordinary default read/call/arithmetic observers must remain nil")
	}
	for name, fn := range map[string]any{"frame": regenerationRawFrame4E01D0, "fps": regenerationRawFPS4E01D0, "armor": regenerationRawArmorService4E01D0} {
		value := reflect.ValueOf(fn)
		if value.IsNil() {
			t.Fatal("default service absent", name)
		}
		symbol := runtime.FuncForPC(value.Pointer())
		if symbol == nil {
			t.Fatal("default service symbol absent", name)
		}
		file, _ := symbol.FileLine(value.Pointer())
		if filepath.Base(file) != "regeneration_raw_runtime_4e01d0.go" {
			t.Fatal("default service replaced", name, file)
		}
	}
	if reflect.ValueOf(regenerationRawAdjustService4E01D0).Pointer() != reflect.ValueOf(unitAdjustHPCall4EE460).Pointer() {
		t.Fatal("actual production UnitAdjustHP dispatch required")
	}
}

func regenerationRawDefaultWorld4E01D0(t *testing.T) (*server.Server, map[string]uint16) {
	t.Helper()
	regenerationRawRequireDefaults4E01D0(t, false)
	s := server.NewHigh(nil, nil, strman.New())
	s.FreeObjectTypes() // Establish the reserved index0 and writable type map.
	t.Cleanup(s.Close)
	t.Cleanup(s.FreeObjectTypes)
	oldServer, oldFlags := GetServer, noxflags.GetGame()
	GetServer = func() Server { return &regenerationRawDefaultServer4E01D0{srv: s} }
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		GetServer = oldServer
	})
	// Prepared named type records initialize the real production armor table;
	// no armor function/table result is replaced. No stock asset/gameplay credit.
	types := make(map[string]uint16)
	for _, name := range []string{
		"NativeRegenerationOwner", "NativeRegenerationWeapon",
		"LeatherArmor", "ChainTunic", "Breastplate", "LeatherHelm", "SteelHelm",
		"SteelShield", "WoodenShield", "StreetShirt", "StreetPants", "StreetSneakers",
		"LeatherBoots", "LeatherArmoredBoots", "PlateBoots", "MedievalPants", "LeatherLeggings",
		"ChainLeggings", "PlateLeggings", "MedievalShirt", "WizardRobe", "LeatherArmbands",
		"PlateArms", "MedievalCloak", "ChainCoif", "WizardHelm", "ConjurerHelm", "OrnateHelm",
	} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
		index := s.Types.IndByID(name)
		if index <= 0 || index > math.MaxUint16 {
			t.Fatal("real prepared type identity", name, index)
		}
		types[name] = uint16(index)
	}
	s.Nox_xxx_equipArmor_415AB0()
	if s.Armor.Sub_415DF0("WizardRobe") != 0x4000 || s.Armor.Sub_415CD0(0x4000) != int(types["WizardRobe"]) {
		t.Fatal("actual WizardRobe production armor mapping")
	}
	return s, types
}

type regenerationRawDefaultGraph4E01D0 struct {
	owner, item *server.Object
	health      *server.HealthData
	effect      *server.ModifierEff
}

func regenerationRawDefaultCGraph4E01D0(t *testing.T) regenerationRawDefaultGraph4E01D0 {
	t.Helper()
	owner, freeOwner := alloc.NewHigh(server.Object{})
	item, freeItem := alloc.NewHigh(server.Object{})
	health, freeHealth := alloc.NewHigh(server.HealthData{})
	effect, freeEffect := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(freeOwner)
	t.Cleanup(freeItem)
	t.Cleanup(freeHealth)
	t.Cleanup(freeEffect)
	return regenerationRawDefaultGraph4E01D0{owner: owner, item: item, health: health, effect: effect}
}

func (g regenerationRawDefaultGraph4E01D0) prepare(t *testing.T, types map[string]uint16, cur, max uint16, last uint32, flags object.Flags, armor string, rate int32) {
	t.Helper()
	*g.health = server.HealthData{Cur: cur, Field2: 0x1357, Max: max, Field16: 0x2468ace0}
	*g.owner = server.Object{TypeInd: types["NativeRegenerationOwner"], ObjClass: object.ClassMonster, ObjFlags: flags, HealthData: g.health, Frame134: last, Field38: 0x13579bdf}
	*g.item = server.Object{TypeInd: types["NativeRegenerationWeapon"], ObjClass: object.ClassWeapon, InvHolder: g.owner}
	if armor != "" {
		g.item.TypeInd, g.item.ObjClass = types[armor], object.ClassArmor
	}
	*g.effect = server.ModifierEff{Update100: server.ModifierEffFnc{Val: rate}}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(g.owner), unsafe.Pointer(g.item), unsafe.Pointer(g.health), unsafe.Pointer(g.effect)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatal("actual C API graph must exercise native high addresses", pointer)
			}
		}
	}
}

func (g regenerationRawDefaultGraph4E01D0) verifyCall(t *testing.T, s *server.Server, wantHP uint16, hpObserver bool) {
	t.Helper()
	regenerationRawRequireDefaults4E01D0(t, hpObserver)
	frame, fps, flags := s.Frame(), s.TickRate(), noxflags.GetGame()
	beforeHP := g.health.Cur
	objects := []struct {
		pointer unsafe.Pointer
		bytes   int
	}{
		{unsafe.Pointer(g.owner), int(unsafe.Sizeof(*g.owner))}, {unsafe.Pointer(g.item), int(unsafe.Sizeof(*g.item))},
		{unsafe.Pointer(g.health), int(unsafe.Sizeof(*g.health))}, {unsafe.Pointer(g.effect), int(unsafe.Sizeof(*g.effect))},
	}
	want := make([][]byte, len(objects))
	for i, record := range objects {
		want[i] = bytes.Clone(unsafe.Slice((*byte)(record.pointer), record.bytes))
	}
	if wantHP != g.health.Cur {
		binary.LittleEndian.PutUint16(want[2], wantHP)
		binary.LittleEndian.PutUint32(want[0][unsafe.Offsetof(g.owner.Field38):], math.MaxUint32)
	}
	// Direct public C API; all root/clock/armor/HP-adjust services stay real.
	regenerationRawPublic4E01D0(unsafe.Pointer(g.effect), unsafe.Pointer(g.item))
	if g.health.Cur != wantHP || s.Frame() != frame || s.TickRate() != fps || noxflags.GetGame() != flags {
		t.Fatal("actual HP effect with unchanged server clock/flags", g.health.Cur, wantHP)
	}
	for i, record := range objects {
		if !bytes.Equal(unsafe.Slice((*byte)(record.pointer), record.bytes), want[i]) {
			t.Fatal("complete actual native graph record changed outside HP/sync word", i)
		}
	}
	regenerationRawRequireDefaults4E01D0(t, hpObserver)
	t.Logf("REGENERATION_RAW_ACTUAL_DEFAULT_C owner=%p item=%p health=%p effect=%p HP=%d->%d frame=%d fps=%d real-clock-and-default-bindings=true HP-read-observer=%t", g.owner, g.item, g.health, g.effect, beforeHP, g.health.Cur, frame, fps, hpObserver)
}

func TestRegenerationRawActualDefaults4E01D0(t *testing.T) {
	for _, name := range []string{"heal-clamps-at-maximum", "already-full", "dead-and-recent-gates", "real-robe-rate-and-ordinary-armor-control", "DWORD-frame-and-product-wrap"} {
		t.Run(name, func(t *testing.T) {
			s, types := regenerationRawDefaultWorld4E01D0(t)
			g := regenerationRawDefaultCGraph4E01D0(t)
			s.SetFrame(36)
			s.SetTickRate(30)
			g.prepare(t, types, 99, 100, 0, 0, "", 60)
			switch name {
			case "heal-clamps-at-maximum":
				g.verifyCall(t, s, 100, false)
			case "already-full":
				g.health.Cur = 100
				g.verifyCall(t, s, 100, false)
			case "dead-and-recent-gates":
				g.owner.ObjFlags = object.FlagDead
				g.verifyCall(t, s, 99, false)
				g.owner.ObjFlags = 0
				s.SetFrame(29)
				g.verifyCall(t, s, 99, false)
			case "real-robe-rate-and-ordinary-armor-control":
				s.SetFrame(30)
				g.prepare(t, types, 99, 100, 0, 0, "LeatherArmor", 60)
				if actual := s.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(g.item); actual != 0x8000 {
					t.Fatal("real ordinary armor flag", actual)
				}
				g.verifyCall(t, s, 99, false)
				g.prepare(t, types, 99, 100, 0, 0, "WizardRobe", 60)
				if actual := s.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(g.item); actual != 0x4000 {
					t.Fatal("real special robe flag", actual)
				}
				g.verifyCall(t, s, 100, false)
			case "DWORD-frame-and-product-wrap":
				s.SetFrame(0)
				g.prepare(t, types, 99, 100, 0xffffffe2, 0, "", 60)
				g.verifyCall(t, s, 100, false)
				s.SetFrame(2)
				s.SetTickRate(2)
				g.prepare(t, types, 1, 2, 0, 0, "", -2147483647)
				g.verifyCall(t, s, 2, false)
			}
		})
	}
}

func TestRegenerationRawPinnedGoGraph4E01D0(t *testing.T) {
	for _, armor := range []string{"", "WizardRobe"} {
		name := "weapon"
		if armor != "" {
			name = "robe"
		}
		t.Run(name, func(t *testing.T) {
			s, types := regenerationRawDefaultWorld4E01D0(t)
			s.SetFrame(36)
			s.SetTickRate(30)
			g := regenerationRawDefaultGraph4E01D0{owner: new(server.Object), item: new(server.Object), health: new(server.HealthData), effect: new(server.ModifierEff)}
			g.prepare(t, types, 99, 100, 0, 0, armor, 60)
			var pin runtime.Pinner
			for _, value := range []any{g.owner, g.item, g.health, g.effect} {
				pin.Pin(value)
			}
			defer pin.Unpin()
			previous := server.ObserveUnitHPRead4EE780
			defer func() { server.ObserveUnitHPRead4EE780 = previous }()
			gc, objectReads, healthReads := 0, 0, 0
			server.ObserveUnitHPRead4EE780 = func(read server.UnitHPRead4EE780) {
				if read.Complete {
					return
				}
				if read.Field == 0 {
					if read.Unit != g.owner {
						t.Fatal("actual HP helper must retain pinned owner")
					}
					objectReads++
				} else {
					if (read.Field != 1 && read.Field != 2) || read.Health != g.health {
						t.Fatal("actual HP helper must retain pinned health")
					}
					healthReads++
				}
				runtime.GC()
				gc++
			}
			g.verifyCall(t, s, 100, true)
			if gc != 6 || objectReads != 3 || healthReads != 3 {
				t.Fatal("actual two-max/one-current helper reads and in-call GCs", gc, objectReads, healthReads)
			}
			t.Logf("REGENERATION_RAW_PINNED_GO_GC=%d actual-HP-helper-reads=%d complete-pinned-graph=true", gc, objectReads+healthReads)
			runtime.KeepAlive(g)
		})
	}
}
