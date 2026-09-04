package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func TestSpellPrecheckNative4FD0E0Layouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerInfo := uintptr(2185)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantUpdateSize = 656
		wantUpdatePlayer = 336
		wantPlayerSize = 6160
		wantPlayerInfo = 2189
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantPlayerInfo},
		{"Player class byte", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantPlayerInfo + 66},
		{"Object.UpdateData width", unsafe.Sizeof(Object{}.UpdateData), unsafe.Sizeof(uintptr(0))},
		{"PlayerUpdateData.Player width", unsafe.Sizeof(PlayerUpdateData{}.Player), unsafe.Sizeof(uintptr(0))},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellPrecheckNative4FD0E0PreservesPointersAndLivePlayerData(t *testing.T) {
	oldPlayer := &Player{}
	oldPlayer.Info().SetPlayerClass(playerlib.Wizard)
	oldUpdate := &PlayerUpdateData{Player: oldPlayer}
	newPlayer := &Player{}
	newPlayer.Info().SetPlayerClass(playerlib.Conjurer)
	newUpdate := &PlayerUpdateData{Player: newPlayer}
	unit := &Object{UpdateData: unsafe.Pointer(oldUpdate)}

	var events []string
	got := spellPrecheckNative4FD0E0(unit, 0x12345678, spellPrecheckNativeDeps4FD0E0{
		spellFlags: func(gotSpell int32) uint32 {
			events = append(events, "flags")
			if gotSpell != 0x12345678 {
				t.Fatalf("flags spell = %#x", gotSpell)
			}
			unit.UpdateData = unsafe.Pointer(newUpdate)
			return math.MaxUint32
		},
		spellEnabled: func(gotSpell int32) int32 {
			events = append(events, "enabled")
			if gotSpell != 0x12345678 {
				t.Fatalf("enabled spell = %#x", gotSpell)
			}
			unit.ObjClass = object.Class(0xffffff04)
			return math.MinInt32
		},
		checkPlayerSpellClass: func(class uint8, gotSpell int32) int32 {
			events = append(events, "class-check")
			if class != uint8(playerlib.Conjurer) || gotSpell != 0x12345678 {
				t.Fatalf("class check = %d/%#x, want %d/0x12345678", class, gotSpell, playerlib.Conjurer)
			}
			return math.MinInt32
		},
		summonAllowed: func(int32, *Object) int32 {
			t.Fatal("Player path reached summon check")
			return 0
		},
	})
	if got != math.MinInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MinInt32))
	}
	if want := []string{"flags", "enabled", "class-check"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if unit.UpdateData != unsafe.Pointer(newUpdate) {
		t.Fatal("live update-data mutation was not observed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"unit":       uintptr(unsafe.Pointer(unit)),
			"old update": uintptr(unsafe.Pointer(oldUpdate)),
			"new update": uintptr(unsafe.Pointer(newUpdate)),
			"old Player": uintptr(unsafe.Pointer(oldPlayer)),
			"new Player": uintptr(unsafe.Pointer(newPlayer)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(oldPlayer)
	runtime.KeepAlive(oldUpdate)
	runtime.KeepAlive(newPlayer)
	runtime.KeepAlive(newUpdate)
	runtime.KeepAlive(unit)
}

func TestSpellPrecheckServer4FD0E0ClassEnablementAndSummonCapacity(t *testing.T) {
	const ordinary = spell.ID(7)
	s := new(Server)
	s.Spells.byID = map[spell.ID]*SpellDef{
		ordinary: {
			ID:      ordinary,
			Enabled: true,
			Def:     things.Spell{Flags: things.SpellClassWizard},
		},
	}
	player := &Player{}
	player.Info().SetPlayerClass(playerlib.Wizard)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}

	if got := s.SpellPrecheck4FD0E0(unit, ordinary); got != 0 {
		t.Fatalf("Wizard-compatible result = %d, want 0", got)
	}
	s.Spells.byID[ordinary].Def.Flags = things.SpellClassConjurer
	if got := s.SpellPrecheck4FD0E0(unit, ordinary); got != spellPrecheckBadSkill4FD0E0 {
		t.Fatalf("wrong-class result = %d, want %d", got, spellPrecheckBadSkill4FD0E0)
	}
	s.Spells.byID[ordinary].Def.Flags = things.SpellClassAny
	player.Info().SetPlayerClass(playerlib.Warrior)
	if got := s.SpellPrecheck4FD0E0(unit, ordinary); got != spellPrecheckBadSkill4FD0E0 {
		t.Fatalf("invalid Player class result = %d, want %d", got, spellPrecheckBadSkill4FD0E0)
	}
	player.Info().SetPlayerClass(playerlib.Conjurer)
	s.Spells.byID[ordinary].Def.Flags = things.SpellClassConjurer
	if got := s.SpellPrecheck4FD0E0(unit, ordinary); got != 0 {
		t.Fatalf("Conjurer-compatible result = %d, want 0", got)
	}

	s.Spells.byID[ordinary].Enabled = false
	if got := s.SpellPrecheck4FD0E0(nil, ordinary); got != spellPrecheckIllegal4FD0E0 {
		t.Fatalf("disabled nil-unit result = %d, want %d", got, spellPrecheckIllegal4FD0E0)
	}
	if got := s.SpellPrecheck4FD0E0(nil, ordinary+1); got != spellPrecheckIllegal4FD0E0 {
		t.Fatalf("undefined spell result = %d, want %d", got, spellPrecheckIllegal4FD0E0)
	}

	s.Spells.byID[ordinary].Enabled = true
	nonPlayer := &Object{ObjClass: object.ClassMissile}
	if got := s.SpellPrecheck4FD0E0(nonPlayer, ordinary); got != 0 {
		t.Fatalf("ordinary non-Player result = %d, want 0", got)
	}

	summon := spell.SPELL_SUMMON_BAT
	s.Spells.byID[summon] = &SpellDef{ID: summon, Enabled: true}
	owner := &Object{ObjClass: object.ClassPlayer}
	effect := &Object{ObjClass: object.ClassMissile, ObjOwner: owner}
	if got := s.SpellPrecheck4FD0E0(effect, summon); got != 0 {
		t.Fatalf("empty summon-capacity result = %d, want 0", got)
	}
	firstUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	secondUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	first := &Object{ObjClass: object.ClassMonster, ObjOwner: owner, UpdateData: unsafe.Pointer(firstUpdate)}
	second := &Object{ObjClass: object.ClassMonster, ObjOwner: owner, UpdateData: unsafe.Pointer(secondUpdate)}
	owner.Field129 = first
	first.Field128 = second
	if got := s.SpellPrecheck4FD0E0(effect, summon); got != spellPrecheckIllegal4FD0E0 {
		t.Fatalf("over-capacity summon result = %d, want %d", got, spellPrecheckIllegal4FD0E0)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(firstUpdate)
	runtime.KeepAlive(secondUpdate)
}

func TestSpellPrecheckNative4FD0E0FaultBoundaries(t *testing.T) {
	baseDeps := func(enabled int32) spellPrecheckNativeDeps4FD0E0 {
		return spellPrecheckNativeDeps4FD0E0{
			spellFlags:            func(int32) uint32 { return 0 },
			spellEnabled:          func(int32) int32 { return enabled },
			checkPlayerSpellClass: func(uint8, int32) int32 { return 0 },
			summonAllowed:         func(int32, *Object) int32 { return 1 },
		}
	}

	t.Run("disabled nil unit returns before class", func(t *testing.T) {
		if got := spellPrecheckNative4FD0E0(nil, 1, baseDeps(0)); got != spellPrecheckIllegal4FD0E0 {
			t.Fatalf("result = %d, want %d", got, spellPrecheckIllegal4FD0E0)
		}
	})

	for _, test := range []struct {
		name string
		unit *Object
	}{
		{name: "enabled nil unit", unit: nil},
		{name: "nil update data", unit: &Object{ObjClass: object.ClassPlayer}},
		{name: "nil Player", unit: &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{})}},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("fault boundary did not panic")
				}
			}()
			_ = spellPrecheckNative4FD0E0(test.unit, 1, baseDeps(1))
		})
	}
}
