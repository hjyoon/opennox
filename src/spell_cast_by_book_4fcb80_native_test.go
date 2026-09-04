package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellCastByBookNative4FCB80Layouts(t *testing.T) {
	wantMagicSize := uintptr(60)
	wantMagicObject := uintptr(4)
	wantMagicSpells := uintptr(8)
	wantMagicIndex := uintptr(28)
	wantMagicLeaf := uintptr(32)
	wantMagicProgress := uintptr(36)
	wantMagicFrame := uintptr(40)
	wantMagicDelay := uintptr(44)
	wantMagicTarget := uintptr(48)
	wantMagicNext := uintptr(52)
	wantMagicPrev := uintptr(56)
	wantObjectClass := uintptr(8)
	wantObjectFlags := uintptr(16)
	wantObjectUpdate := uintptr(748)
	wantUpdateLeaf := uintptr(184)
	wantUpdateProgress := uintptr(188)
	wantUpdateTraps := uintptr(192)
	wantUpdateTrapCount := uintptr(212)
	wantUpdateCastStart := uintptr(216)
	wantUpdateCastX := uintptr(220)
	wantUpdateCastY := uintptr(224)
	wantUpdatePlayer := uintptr(276)
	wantUpdateCursorObject := uintptr(288)
	wantPlayerIndex := uintptr(2064)
	wantPlayerCursor := uintptr(2284)
	wantPlayerTarget := uintptr(3640)
	wantLeafChildren := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantMagicSize = 80
		wantMagicObject = 8
		wantMagicSpells = 16
		wantMagicIndex = 36
		wantMagicLeaf = 40
		wantMagicProgress = 48
		wantMagicFrame = 52
		wantMagicDelay = 56
		wantMagicTarget = 60
		wantMagicNext = 64
		wantMagicPrev = 72
		wantObjectClass = 12
		wantObjectFlags = 20
		wantObjectUpdate = 872
		wantUpdateLeaf = 232
		wantUpdateProgress = 240
		wantUpdateTraps = 244
		wantUpdateTrapCount = 264
		wantUpdateCastStart = 268
		wantUpdateCastX = 272
		wantUpdateCastY = 280
		wantUpdatePlayer = 336
		wantUpdateCursorObject = 360
		wantPlayerIndex = 2068
		wantPlayerCursor = 2288
		wantPlayerTarget = 4928
		wantLeafChildren = 8
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"MagicEntityClass size", unsafe.Sizeof(server.MagicEntityClass{}), wantMagicSize},
		{"MagicEntityClass.Obj4", unsafe.Offsetof(server.MagicEntityClass{}.Obj4), wantMagicObject},
		{"MagicEntityClass.Spells8", unsafe.Offsetof(server.MagicEntityClass{}.Spells8), wantMagicSpells},
		{"MagicEntityClass.SpellInd28", unsafe.Offsetof(server.MagicEntityClass{}.SpellInd28), wantMagicIndex},
		{"MagicEntityClass.Field29", unsafe.Offsetof(server.MagicEntityClass{}.Field29), wantMagicIndex + 1},
		{"MagicEntityClass.Field30", unsafe.Offsetof(server.MagicEntityClass{}.Field30), wantMagicIndex + 2},
		{"MagicEntityClass.Field32", unsafe.Offsetof(server.MagicEntityClass{}.Field32), wantMagicLeaf},
		{"MagicEntityClass.Field36", unsafe.Offsetof(server.MagicEntityClass{}.Field36), wantMagicProgress},
		{"MagicEntityClass.Frame40", unsafe.Offsetof(server.MagicEntityClass{}.Frame40), wantMagicFrame},
		{"MagicEntityClass.Field44", unsafe.Offsetof(server.MagicEntityClass{}.Field44), wantMagicDelay},
		{"MagicEntityClass.Field48", unsafe.Offsetof(server.MagicEntityClass{}.Field48), wantMagicTarget},
		{"MagicEntityClass.Next52", unsafe.Offsetof(server.MagicEntityClass{}.Next52), wantMagicNext},
		{"MagicEntityClass.Prev56", unsafe.Offsetof(server.MagicEntityClass{}.Prev56), wantMagicPrev},
		{"Object.ObjClass", unsafe.Offsetof(server.Object{}.ObjClass), wantObjectClass},
		{"Object.ObjFlags", unsafe.Offsetof(server.Object{}.ObjFlags), wantObjectFlags},
		{"Object.UpdateData", unsafe.Offsetof(server.Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.SpellPhonemeLeaf", unsafe.Offsetof(server.PlayerUpdateData{}.SpellPhonemeLeaf), wantUpdateLeaf},
		{"PlayerUpdateData.Field47_0", unsafe.Offsetof(server.PlayerUpdateData{}.Field47_0), wantUpdateProgress},
		{"PlayerUpdateData.TrapSpells", unsafe.Offsetof(server.PlayerUpdateData{}.TrapSpells), wantUpdateTraps},
		{"PlayerUpdateData.TrapSpellsCnt", unsafe.Offsetof(server.PlayerUpdateData{}.TrapSpellsCnt), wantUpdateTrapCount},
		{"PlayerUpdateData.SpellCastStart", unsafe.Offsetof(server.PlayerUpdateData{}.SpellCastStart), wantUpdateCastStart},
		{"PlayerUpdateData.Field55", unsafe.Offsetof(server.PlayerUpdateData{}.Field55), wantUpdateCastX},
		{"PlayerUpdateData.Field56", unsafe.Offsetof(server.PlayerUpdateData{}.Field56), wantUpdateCastY},
		{"PlayerUpdateData.Player", unsafe.Offsetof(server.PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"PlayerUpdateData.CursorObj", unsafe.Offsetof(server.PlayerUpdateData{}.CursorObj), wantUpdateCursorObject},
		{"Player.PlayerInd", unsafe.Offsetof(server.Player{}.PlayerInd), wantPlayerIndex},
		{"Player.CursorVec", unsafe.Offsetof(server.Player{}.CursorVec), wantPlayerCursor},
		{"Player.Obj3640", unsafe.Offsetof(server.Player{}.Obj3640), wantPlayerTarget},
		{"PhonemeLeaf.Ind", unsafe.Offsetof(server.PhonemeLeaf{}.Ind), 0},
		{"PhonemeLeaf.Pho", unsafe.Offsetof(server.PhonemeLeaf{}.Pho), wantLeafChildren},
		{"Settings.BroadcastGestures62", unsafe.Offsetof(server.Settings{}.BroadcastGestures62), 62},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellCastByBookNative4FCB80PreservesOnePastAliases(t *testing.T) {
	node := &server.MagicEntityClass{
		Spells8:    [5]int32{1, -2, math.MinInt32, math.MaxInt32, 0x12345678},
		SpellInd28: 0xd4,
		Field29:    0xc3,
		Field30:    0xa1b2,
	}
	for i, want := range node.Spells8 {
		if got := spellBookSpellNative4FCB80(node, i); got != want {
			t.Fatalf("spell slot %d = %#x, want %#x", i, got, want)
		}
	}
	if got, want := uint32(spellBookSpellNative4FCB80(node, 5)), uint32(0xa1b2c3d4); got != want {
		t.Fatalf("packed control slot = %#x, want %#x", got, want)
	}

	update := &server.PlayerUpdateData{
		TrapSpells:    [5]uint32{1, 2, 3, 4, 0xfedcba98},
		TrapSpellsCnt: 0xaabbcc05,
	}
	if got := uint32(spellBookTrapSpellNative4FCB80(update, 4)); got != 0xfedcba98 {
		t.Fatalf("trap slot four = %#x", got)
	}
	if got := uint32(spellBookTrapSpellNative4FCB80(update, 5)); got != 0xaabbcc05 {
		t.Fatalf("trap count alias = %#x", got)
	}
	spellBookStoreTrapSpellNative4FCB80(update, 5, int32(0x11223344))
	if update.TrapSpellsCnt != 0x11223344 {
		t.Fatalf("trap count after aliased store = %#x", update.TrapSpellsCnt)
	}
	hooks := spellCastByBookNativeHooks4FCB80(spellCastByBookNativeDeps4FCB80{})
	hooks.storeTrapCount(update, 0x45)
	if update.TrapSpellsCnt != 0x11223345 {
		t.Fatalf("trap count after low-byte increment = %#x", update.TrapSpellsCnt)
	}

	for _, call := range []func(){
		func() { spellBookSpellNative4FCB80(node, 6) },
		func() { spellBookTrapSpellNative4FCB80(update, 6) },
		func() { spellBookStoreTrapSpellNative4FCB80(update, 6, 1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("native-width boundary crossing did not fault")
				}
			}()
			call()
		}()
	}
}

func spellBookUnexpectedNativeDeps4FCB80(t *testing.T) spellCastByBookNativeDeps4FCB80 {
	t.Helper()
	fail := func(name string) { t.Fatalf("unexpected native dependency: %s", name) }
	return spellCastByBookNativeDeps4FCB80{
		loadHead:  func() *server.MagicEntityClass { fail("loadHead"); return nil },
		storeHead: func(*server.MagicEntityClass) { fail("storeHead") },
		loadAllocator: func() alloc.ClassT[server.MagicEntityClass] {
			fail("loadAllocator")
			return alloc.ClassT[server.MagicEntityClass]{}
		},
		freeFirst: func(alloc.ClassT[server.MagicEntityClass], *server.MagicEntityClass) { fail("freeFirst") },
		loadFrame: func() uint32 { fail("loadFrame"); return 0 },
		reportStart: func(ntype.PlayerInd, uint8, uint8) {
			fail("reportStart")
		},
		loadSettings: func() *server.Settings { fail("loadSettings"); return nil },
		loadPhonemes: func(spell.ID) []spell.Phoneme { fail("loadPhonemes"); return nil },
		loadSuppress: func() uint32 { fail("loadSuppress"); return 0 },
		broadcast:    func(*server.Object, int8) { fail("broadcast") },
		advanceLeaf: func(*server.PhonemeLeaf, spell.Phoneme) *server.PhonemeLeaf {
			fail("advanceLeaf")
			return nil
		},
		phonemeRoot:  func() *server.PhonemeLeaf { fail("phonemeRoot"); return nil },
		informResult: func(ntype.PlayerInd, uint8, int32) { fail("informResult") },
		chargeMana: func(*server.Object, spell.ID, int32) int32 {
			fail("chargeMana")
			return 0
		},
		audioEvent:  func(sound.ID, *server.Object, int32, uint32) { fail("audioEvent") },
		playerSpell: func(*server.Object) { fail("playerSpell") },
		castByUser:  func(spell.ID, int, *server.Object, *server.SpellAcceptArg) { fail("castByUser") },
	}
}

func TestSpellCastByBookNative4FCB80BindsMismatchPointersAndFields(t *testing.T) {
	leafA := &server.PhonemeLeaf{Ind: 3}
	leafB := &server.PhonemeLeaf{Ind: 9}
	player := &server.Player{PlayerInd: 0xe7}
	update := &server.PlayerUpdateData{Player: player}
	objectA := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	objectB := &server.Object{ObjClass: object.ClassPlayer}
	node := &server.MagicEntityClass{
		Obj4:       objectA,
		Spells8:    [5]int32{0x109},
		Field32:    leafA,
		Field44:    math.MaxUint32,
		SpellInd28: 0,
		Field36:    0,
	}
	settings := new(server.Settings)
	binary.LittleEndian.PutUint32(settings.BroadcastGestures62[:], 0x01020304)
	phonemes := []spell.Phoneme{spell.Phoneme(5)}

	deps := spellBookUnexpectedNativeDeps4FCB80(t)
	deps.loadHead = func() *server.MagicEntityClass { return node }
	frameCalls := 0
	deps.loadFrame = func() uint32 {
		frameCalls++
		return uint32(99 + frameCalls)
	}
	deps.reportStart = func(index ntype.PlayerInd, message, spellLow uint8) {
		if index != ntype.PlayerInd(0xe7) || message != 112 || spellLow != 9 {
			t.Fatalf("start report = %d/%d/%d", index, message, spellLow)
		}
	}
	deps.loadSettings = func() *server.Settings { return settings }
	deps.loadPhonemes = func(id spell.ID) []spell.Phoneme {
		if id != spell.ID(0x109) {
			t.Fatalf("phoneme spell = %#x", id)
		}
		return phonemes
	}
	deps.loadSuppress = func() uint32 { return 1 }
	deps.broadcast = func(got *server.Object, phoneme int8) {
		if got != objectA || phoneme != 5 {
			t.Fatalf("broadcast = %p/%d, want %p/5", got, phoneme, objectA)
		}
		node.Obj4 = objectB
	}
	deps.advanceLeaf = func(got *server.PhonemeLeaf, phoneme spell.Phoneme) *server.PhonemeLeaf {
		if got != leafA || phoneme != spell.Phoneme(5) {
			t.Fatalf("advance = %p/%d, want %p/5", got, phoneme, leafA)
		}
		return leafB
	}

	spellCastByBookNative4FCB80(deps)
	if frameCalls != 2 {
		t.Fatalf("frame calls = %d, want 2", frameCalls)
	}
	if node.Field32 != leafB || update.SpellPhonemeLeaf != leafB {
		t.Fatalf("leaf state = %p/%p, want %p", node.Field32, update.SpellPhonemeLeaf, leafB)
	}
	if node.Field36 != 1 || node.Frame40 != 100 {
		t.Fatalf("progress/deadline = %d/%d, want 1/100", node.Field36, node.Frame40)
	}
	if node.Obj4 != objectB {
		t.Fatalf("live object pointer = %p, want %p", node.Obj4, objectB)
	}

	assertSpellBookNativePointers4FCB80(t, map[string]unsafe.Pointer{
		"node": unsafe.Pointer(node), "object-a": unsafe.Pointer(objectA), "object-b": unsafe.Pointer(objectB),
		"update": unsafe.Pointer(update), "player": unsafe.Pointer(player), "leaf-a": unsafe.Pointer(leafA),
		"leaf-b": unsafe.Pointer(leafB), "settings": unsafe.Pointer(settings),
	})
	runtime.KeepAlive(node)
	runtime.KeepAlive(objectA)
	runtime.KeepAlive(objectB)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
	runtime.KeepAlive(leafA)
	runtime.KeepAlive(leafB)
	runtime.KeepAlive(settings)
}

func TestSpellCastByBookNative4FCB80FinalPlayerAndUnlink(t *testing.T) {
	leaf := &server.PhonemeLeaf{Ind: int32(spell.SPELL_GLYPH)}
	target := new(server.Object)
	player := &server.Player{PlayerInd: 19, CursorVec: image.Pt(-123, 456)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		player.CursorVec = image.Pt(int(int64(1)<<32|123), int(-(int64(1)<<32)-456))
	}
	update := &server.PlayerUpdateData{
		Player:         player,
		CursorObj:      target,
		SpellCastStart: 0xaabbccdd,
		Field47_0:      0xee,
		TrapSpellsCnt:  0x55667788,
	}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	node := &server.MagicEntityClass{
		Obj4:    unit,
		Spells8: [5]int32{int32(spell.SPELL_GLYPH)},
		Field32: leaf,
		Field36: 1,
	}
	allocatorClass := new(alloc.Class)
	allocator := alloc.ClassT[server.MagicEntityClass]{Class: allocatorClass}
	head := node
	freed := false

	deps := spellBookUnexpectedNativeDeps4FCB80(t)
	deps.loadHead = func() *server.MagicEntityClass { return head }
	deps.storeHead = func(value *server.MagicEntityClass) { head = value }
	deps.loadAllocator = func() alloc.ClassT[server.MagicEntityClass] { return allocator }
	deps.freeFirst = func(got alloc.ClassT[server.MagicEntityClass], value *server.MagicEntityClass) {
		if got.Class != allocatorClass || value != node {
			t.Fatalf("free = %p/%p, want %p/%p", got.Class, value, allocatorClass, node)
		}
		freed = true
	}
	deps.loadFrame = func() uint32 { return 100 }
	deps.playerSpell = func(got *server.Object) {
		if got != unit || player.Obj3640 != target || update.Field55 != 123 || update.Field56 != -456 {
			t.Fatalf("player spell state = unit %p target %p pos %d/%d", got, player.Obj3640, update.Field55, update.Field56)
		}
		update.SpellCastStart = 0x11223344
		update.Field47_0 = 0x55
		update.TrapSpellsCnt = 0x99aabbcc
	}

	spellCastByBookNative4FCB80(deps)
	if head != nil || !freed {
		t.Fatalf("head/freed = %p/%t, want nil/true", head, freed)
	}
	if player.Obj3640 != target {
		t.Fatalf("target = %p, want %p", player.Obj3640, target)
	}
	if update.Field55 != 123 || update.Field56 != -456 {
		t.Fatalf("stored cast position = %d/%d, want 123/-456", update.Field55, update.Field56)
	}
	if update.SpellCastStart != 0 || update.Field47_0 != 0 || update.TrapSpellsCnt != 0x99aabb00 {
		t.Fatalf("cleared state = start %#x casting %#x trap count %#x", update.SpellCastStart, update.Field47_0, update.TrapSpellsCnt)
	}

	assertSpellBookNativePointers4FCB80(t, map[string]unsafe.Pointer{
		"node": unsafe.Pointer(node), "unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
		"player": unsafe.Pointer(player), "leaf": unsafe.Pointer(leaf), "target": unsafe.Pointer(target),
		"allocator": unsafe.Pointer(allocatorClass),
	})
	runtime.KeepAlive(node)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
	runtime.KeepAlive(leaf)
	runtime.KeepAlive(target)
	runtime.KeepAlive(allocatorClass)
}

func TestSpellCastByBookNative4FCB80NonPlayerUsesNilCastArgument(t *testing.T) {
	leaf := &server.PhonemeLeaf{Ind: 9}
	unit := new(server.Object)
	node := &server.MagicEntityClass{Obj4: unit, Spells8: [5]int32{9}, Field32: leaf, Field36: 1}
	head := node
	deps := spellBookUnexpectedNativeDeps4FCB80(t)
	deps.loadHead = func() *server.MagicEntityClass { return head }
	deps.storeHead = func(value *server.MagicEntityClass) { head = value }
	deps.loadAllocator = func() alloc.ClassT[server.MagicEntityClass] {
		return alloc.ClassT[server.MagicEntityClass]{}
	}
	deps.freeFirst = func(class alloc.ClassT[server.MagicEntityClass], got *server.MagicEntityClass) {
		if class.Class != nil || got != node {
			t.Fatalf("free = %p/%p", class.Class, got)
		}
	}
	deps.loadFrame = func() uint32 { return 1 }
	deps.castByUser = func(id spell.ID, level int, got *server.Object, arg *server.SpellAcceptArg) {
		if id != 9 || level != -1 || got != unit || arg != nil {
			t.Fatalf("cast = %d/%d/%p/%p, want 9/-1/%p/nil", id, level, got, arg, unit)
		}
	}

	spellCastByBookNative4FCB80(deps)
	if head != nil {
		t.Fatalf("head = %p, want nil", head)
	}
	assertSpellBookNativePointers4FCB80(t, map[string]unsafe.Pointer{
		"node": unsafe.Pointer(node), "unit": unsafe.Pointer(unit), "leaf": unsafe.Pointer(leaf),
	})
	runtime.KeepAlive(node)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(leaf)
}

func assertSpellBookNativePointers4FCB80(t *testing.T, pointers map[string]unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for name, pointer := range pointers {
		if value := uintptr(pointer); value <= math.MaxUint32 {
			t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, value)
		}
	}
}
