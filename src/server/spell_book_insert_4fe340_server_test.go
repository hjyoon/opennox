package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestSpellBookInsertNative4FE340Layouts(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectFlags := uintptr(16)
	wantObjectUpdate := uintptr(748)
	wantCasting := uintptr(188)
	wantCastStart := uintptr(216)
	wantCurTraps := uintptr(244)
	wantUpdatePlayer := uintptr(276)
	wantUpdateTrade := uintptr(280)
	wantPlayerIndex := uintptr(2064)
	wantSpellLevels := uintptr(3696)
	wantEntitySize := uintptr(60)
	wantEntityObject := uintptr(4)
	wantEntitySpells := uintptr(8)
	wantEntitySpellIndex := uintptr(28)
	wantEntityGlyphMode := uintptr(29)
	wantEntityDefinitions := uintptr(32)
	wantEntityField36 := uintptr(36)
	wantEntityFrame := uintptr(40)
	wantEntityDelay := uintptr(44)
	wantEntityTargetMode := uintptr(48)
	wantEntityNext := uintptr(52)
	wantEntityPrev := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectFlags = 20
		wantObjectUpdate = 872
		wantCasting = 240
		wantCastStart = 268
		wantCurTraps = 304
		wantUpdatePlayer = 336
		wantUpdateTrade = 344
		wantPlayerIndex = 2068
		wantSpellLevels = 4992
		wantEntitySize = 80
		wantEntityObject = 8
		wantEntitySpells = 16
		wantEntitySpellIndex = 36
		wantEntityGlyphMode = 37
		wantEntityDefinitions = 40
		wantEntityField36 = 48
		wantEntityFrame = 52
		wantEntityDelay = 56
		wantEntityTargetMode = 60
		wantEntityNext = 64
		wantEntityPrev = 72
	}

	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Field47_0", unsafe.Offsetof(PlayerUpdateData{}.Field47_0), wantCasting},
		{"PlayerUpdateData.SpellCastStart", unsafe.Offsetof(PlayerUpdateData{}.SpellCastStart), wantCastStart},
		{"PlayerUpdateData.CurTraps", unsafe.Offsetof(PlayerUpdateData{}.CurTraps), wantCurTraps},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"PlayerUpdateData.Trade70", unsafe.Offsetof(PlayerUpdateData{}.Trade70), wantUpdateTrade},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"MagicEntityClass size", unsafe.Sizeof(MagicEntityClass{}), wantEntitySize},
		{"MagicEntityClass.Obj4", unsafe.Offsetof(MagicEntityClass{}.Obj4), wantEntityObject},
		{"MagicEntityClass.Spells8", unsafe.Offsetof(MagicEntityClass{}.Spells8), wantEntitySpells},
		{"MagicEntityClass.SpellInd28", unsafe.Offsetof(MagicEntityClass{}.SpellInd28), wantEntitySpellIndex},
		{"MagicEntityClass.Field29", unsafe.Offsetof(MagicEntityClass{}.Field29), wantEntityGlyphMode},
		{"MagicEntityClass.Field32", unsafe.Offsetof(MagicEntityClass{}.Field32), wantEntityDefinitions},
		{"MagicEntityClass.Field36", unsafe.Offsetof(MagicEntityClass{}.Field36), wantEntityField36},
		{"MagicEntityClass.Frame40", unsafe.Offsetof(MagicEntityClass{}.Frame40), wantEntityFrame},
		{"MagicEntityClass.Field44", unsafe.Offsetof(MagicEntityClass{}.Field44), wantEntityDelay},
		{"MagicEntityClass.Field48", unsafe.Offsetof(MagicEntityClass{}.Field48), wantEntityTargetMode},
		{"MagicEntityClass.Next52", unsafe.Offsetof(MagicEntityClass{}.Next52), wantEntityNext},
		{"MagicEntityClass.Prev56", unsafe.Offsetof(MagicEntityClass{}.Prev56), wantEntityPrev},
		{"spell width", unsafe.Sizeof(int32(0)), 4},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellBookInsertNative4FE340PreservesPointersAndFields(t *testing.T) {
	spells := [5]int32{spellBookInsertGlyph4FE340, 0, 0, 0, 0}
	player := &Player{PlayerInd: 0xe1}
	player.SpellLvl[spellBookInsertGlyph4FE340] = 1
	update := &PlayerUpdateData{
		Player:   player,
		CurTraps: 0xffffff01,
	}
	unit := &Object{
		ObjClass:   0x80000004,
		UpdateData: unsafe.Pointer(update),
	}
	definitions := new(PhonemeLeaf)
	oldHead := new(MagicEntityClass)
	entity := &MagicEntityClass{
		Spells8:    [5]int32{-1, -1, -1, -1, -1},
		SpellInd28: 0xff,
		Field29:    0xff,
		Field36:    0xff,
		Prev56:     oldHead,
	}
	currentHead := oldHead

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":        uintptr(unsafe.Pointer(unit)),
			"sequence":    uintptr(unsafe.Pointer(&spells[0])),
			"player":      uintptr(unsafe.Pointer(player)),
			"update":      uintptr(unsafe.Pointer(update)),
			"entity":      uintptr(unsafe.Pointer(entity)),
			"definitions": uintptr(unsafe.Pointer(definitions)),
			"old head":    uintptr(unsafe.Pointer(oldHead)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer %#x does not exercise the native-width path", name, pointer)
			}
		}
	}

	var (
		state             int32 = -1
		frameLoads        int
		headLoads         int
		manaSequence      *int32
		manaCount         int32
		prechecked        []int32
		castGateSpell     int32
		castGateGlyphMode int32
		balanceKeys       []string
	)
	deps := spellBookInsertNativeDeps4FE340{
		manaPreflight: func(gotUnit *Object, sequence *int32, count int32) int32 {
			if gotUnit != unit {
				t.Fatalf("mana unit = %p, want %p", gotUnit, unit)
			}
			manaSequence, manaCount = sequence, count
			return 1
		},
		checkSummoned: func(*Object, int32) int32 {
			t.Fatal("non-conjurer path checked summoned-creature limit")
			return 0
		},
		countSlaves: func(*Object, uint32, uint32) int32 {
			t.Fatal("non-conjurer path counted slaves")
			return 0
		},
		balanceFloat: func(key string) float64 {
			balanceKeys = append(balanceKeys, key)
			return 2
		},
		spellPrecheck: func(gotUnit *Object, spellID int32) int32 {
			if gotUnit != unit {
				t.Fatalf("precheck unit = %p, want %p", gotUnit, unit)
			}
			prechecked = append(prechecked, spellID)
			return 0
		},
		spellCastGate: func(gotUnit *Object, spellID, glyphMode int32) int32 {
			if gotUnit != unit {
				t.Fatalf("cast-gate unit = %p, want %p", gotUnit, unit)
			}
			castGateSpell, castGateGlyphMode = spellID, glyphMode
			return 0
		},
		informResult: func(uint8, uint8, int32) {
			t.Fatal("successful insertion informed a failure")
		},
		audioEvent: func(int32, *Object, int32, uint32) {
			t.Fatal("successful insertion emitted a failure sound")
		},
		setPlayerState: func(gotUnit *Object, gotState int32) {
			if gotUnit != unit {
				t.Fatalf("state unit = %p, want %p", gotUnit, unit)
			}
			state = gotState
		},
		loadFrame: func() uint32 {
			frameLoads++
			if frameLoads == 1 {
				return 0x89abcdef
			}
			return 0x10203040
		},
		loadAllocator: func() SpellBookInsertAllocator4FE340 {
			if state != spellBookInsertState4FE340 || update.Field47_0 != 1 || update.SpellCastStart != 0x89abcdef {
				t.Fatalf("allocator loaded before cast state mutation: state=%d casting=%d start=%#x", state, update.Field47_0, update.SpellCastStart)
			}
			return func() *MagicEntityClass { return entity }
		},
		loadDefinitions: func() *PhonemeLeaf {
			return definitions
		},
		loadHead: func() *MagicEntityClass {
			headLoads++
			return currentHead
		},
		storeHead: func(head *MagicEntityClass) {
			currentHead = head
		},
	}

	if got := spellBookInsertNative4FE340(unit, &spells[0], 1, -2, -3, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if manaSequence != &spells[0] || manaCount != 1 {
		t.Fatalf("mana args = (%p, %d), want (%p, 1)", manaSequence, manaCount, &spells[0])
	}
	if len(balanceKeys) != 1 || balanceKeys[0] != spellBookInsertMaxTrapKey4FE340 {
		t.Fatalf("balance keys = %v, want [%q]", balanceKeys, spellBookInsertMaxTrapKey4FE340)
	}
	if len(prechecked) != 1 || prechecked[0] != spellBookInsertGlyph4FE340 {
		t.Fatalf("prechecks = %v, want [%d]", prechecked, spellBookInsertGlyph4FE340)
	}
	if castGateSpell != spellBookInsertGlyph4FE340 || castGateGlyphMode != 1 {
		t.Fatalf("cast gate = (%d, %d), want (%d, 1)", castGateSpell, castGateGlyphMode, spellBookInsertGlyph4FE340)
	}
	if state != spellBookInsertState4FE340 || frameLoads != 2 {
		t.Fatalf("state/frame loads = (%d, %d), want (%d, 2)", state, frameLoads, spellBookInsertState4FE340)
	}
	if update.Field47_0 != 1 || update.SpellCastStart != 0x89abcdef || update.Player != player {
		t.Fatalf("update mutated incorrectly: casting=%d start=%#x player=%p", update.Field47_0, update.SpellCastStart, update.Player)
	}
	if entity.Obj4 != unit || entity.Field32 != definitions {
		t.Fatalf("entity pointers = (%p, %p), want (%p, %p)", entity.Obj4, entity.Field32, unit, definitions)
	}
	if entity.Spells8 != spells || entity.SpellInd28 != 0 || entity.Field29 != 1 || entity.Field36 != 0 {
		t.Fatalf("entity gesture = (%v, %d, %d, %d), want (%v, 0, 1, 0)", entity.Spells8, entity.SpellInd28, entity.Field29, entity.Field36, spells)
	}
	if entity.Frame40 != 0x10203040 || entity.Field44 != math.MaxUint32-1 || entity.Field48 != math.MaxUint32-2 {
		t.Fatalf("entity scalars = (%#x, %#x, %#x)", entity.Frame40, entity.Field44, entity.Field48)
	}
	if headLoads != 2 || entity.Next52 != oldHead || entity.Prev56 != nil || oldHead.Prev56 != entity || currentHead != entity {
		t.Fatalf("queue mismatch: loads=%d next=%p prev=%p old.prev=%p head=%p", headLoads, entity.Next52, entity.Prev56, oldHead.Prev56, currentHead)
	}
}
