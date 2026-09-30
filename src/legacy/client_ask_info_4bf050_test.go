package legacy

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type tooltipFixture4BF050 struct {
	t     *testing.T
	d     clientAskInfoDeps4BF050
	dr    *client.Drawable
	buf   []uint16
	trace []string
}

func newTooltipFixture4BF050(t *testing.T) *tooltipFixture4BF050 {
	t.Helper()
	dr, freeDr := alloc.New(client.Drawable{})
	t.Cleanup(freeDr)
	buf, freeBuf := alloc.Make([]uint16{}, 1025)
	t.Cleanup(freeBuf)
	buf[1024] = 0x55AA
	f := &tooltipFixture4BF050{t: t, dr: dr, buf: buf, d: clientAskInfoDepsFactory4BF050()}
	f.d.scratch, f.d.initial = &buf[0], alloc.InternCString16("")
	copyText, appendText, formatText := f.d.copyText, f.d.appendText, f.d.formatText
	f.d.copyText = func(dst, src *uint16) {
		if *dst != 0 {
			t.Fatal("scratch was not cleared before the initial copy")
		}
		f.trace = append(f.trace, "copy")
		copyText(dst, src)
	}
	f.d.appendText = func(dst, src *uint16) {
		f.trace = append(f.trace, "append:"+alloc.GoString16(src))
		appendText(dst, src)
	}
	f.d.formatText = func(dst, format *uint16, arg unsafe.Pointer) {
		f.trace = append(f.trace, "format:"+alloc.GoString16(format))
		formatText(dst, format, arg)
	}
	f.d.prettyName = func(id uint32) *uint16 {
		f.trace = append(f.trace, fmt.Sprintf("pretty:%d", id))
		return alloc.InternCString16("사과")
	}
	f.d.typeName = func(id uint32) *byte {
		f.trace = append(f.trace, fmt.Sprintf("type:%d", id))
		return alloc.InternCString("UnknownWeapon")
	}
	f.d.weaponDef = func(id uint32) *server.Modifier {
		f.trace = append(f.trace, fmt.Sprintf("weapon:%d", id))
		return nil
	}
	f.d.armorDef = func(id uint32) *server.Modifier {
		f.trace = append(f.trace, fmt.Sprintf("armor:%d", id))
		return nil
	}
	f.d.language = func() int { f.trace = append(f.trace, "lang"); return 8 }
	f.d.loadString = func(id string, line int) *uint16 {
		f.trace = append(f.trace, fmt.Sprintf("load:%s:%d", id, line))
		s := map[string]string{"BookOf": "BOOK", "LoreScroll": "LORE", "NoArmsInfo": "No info: %S"}[id]
		if s == "" {
			t.Fatalf("unexpected string key %q", id)
		}
		return alloc.InternCString16(s)
	}
	for kind, dst := range map[string]*func(uint32) *uint16{
		"spell": &f.d.spellTitle, "creature": &f.d.creatureName, "ability": &f.d.abilityName,
	} {
		*dst = func(id uint32) *uint16 {
			f.trace = append(f.trace, fmt.Sprintf("%s:%d", kind, id))
			return alloc.InternCString16(strings.ToUpper(kind))
		}
	}
	f.d.unitCode = func(got *client.Drawable) uint16 {
		if got != dr {
			t.Fatal("drawable identity narrowed")
		}
		f.trace = append(f.trace, "wire")
		return 0xFE32
	}
	f.d.send = func(int, []byte, *server.Object, int) int {
		t.Fatal("unexpected network request")
		return 0
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{dr.C(), unsafe.Pointer(&buf[0]), unsafe.Pointer(f.d.space)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("fixture pointer = %p, want above 4 GiB", p)
			}
		}
	}
	return f
}

func (f *tooltipFixture4BF050) run() *uint16 {
	f.t.Helper()
	got := clientAskInfoNative4BF050(f.dr, f.d)
	if f.buf[1024] != 0x55AA {
		f.t.Fatal("tooltip scratch guard overwritten")
	}
	return got
}

func (f *tooltipFixture4BF050) modifier(primary, secondary, identify string) *server.ModifierEff {
	f.t.Helper()
	m, free := alloc.New(server.ModifierEff{})
	f.t.Cleanup(free)
	// Test-only native metadata access: the production function uses typed
	// accessors, never these PE32 offsets or reflection.
	for name, value := range map[string]string{"desc8": primary, "secdesc12": secondary, "identdesc16": identify} {
		field, ok := reflect.TypeOf(*m).FieldByName(name)
		if !ok {
			f.t.Fatal("modifier description metadata missing")
		}
		*(**uint16)(unsafe.Add(m.C(), field.Offset)) = alloc.InternCString16(value)
	}
	return m
}

func TestClientAskInfo4BF050NativeLayoutAndPrettyFallback(t *testing.T) {
	f := newTooltipFixture4BF050(t)
	dr := f.dr
	wantOffsets := []uintptr{108, 112, 116, 432, 8}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantOffsets = []uintptr{120, 124, 128, 560, 16}
	}
	gotOffsets := []uintptr{unsafe.Offsetof(dr.TypeIDVal), unsafe.Offsetof(dr.ObjClass),
		unsafe.Offsetof(dr.ObjSubClass), unsafe.Offsetof(dr.Union), unsafe.Offsetof(server.Modifier{}.Desc8)}
	if !reflect.DeepEqual(gotOffsets, wantOffsets) {
		t.Fatalf("native fields = %v, want %v", gotOffsets, wantOffsets)
	}
	f.d.initial = alloc.InternCString16("initial")
	f.dr = nil
	if p := f.run(); p != f.d.scratch || alloc.GoString16(p) != "initial" || !reflect.DeepEqual(f.trace, []string{"copy"}) {
		t.Fatalf("nil drawable result/trace = %p/%q/%v", p, alloc.GoString16(p), f.trace)
	}
	f.dr, dr.TypeIDVal = dr, 0x80000001
	for _, class := range []object.Class{0, object.ClassFood, object.ClassInfoBook} {
		dr.ObjClass = class
		for _, p := range []*uint16{nil, alloc.InternCString16(""), alloc.InternCString16("사과")} {
			f.trace = nil
			f.d.prettyName = func(id uint32) *uint16 {
				if id != dr.TypeIDVal {
					t.Fatal("type ID narrowed or read from PE32 offset")
				}
				f.trace = append(f.trace, "pretty")
				return p
			}
			before := *dr
			got := f.run()
			want := p
			if want == nil {
				want = f.d.scratch
			}
			if got != want || *dr != before || !reflect.DeepEqual(f.trace, []string{"copy", "pretty"}) {
				t.Fatalf("class=%#x pointer/trace = %p/%v, want %p", class, got, f.trace, want)
			}
		}
	}
}

func TestClientAskInfo4BF050EquipmentLanguageOrderAndDescriptions(t *testing.T) {
	for _, tc := range []struct {
		lang int
		want string
	}{
		{0, "A B SWORD C D"}, {1, "A B SWORD C D"}, {2, "SWORD C DB  A"},
		{3, "A SWORD B C D"}, {4, "A B SWORD C D"}, {5, "SWORD B A C D"},
		{6, "A C D  BSWORD"}, {7, "A B SWORD C D"}, {8, "A B SWORD C D"}, {-1, "A B SWORD C D"},
	} {
		t.Run(fmt.Sprintf("lang%d", tc.lang), func(t *testing.T) {
			f := newTooltipFixture4BF050(t)
			f.dr.TypeIDVal, f.dr.ObjClass = 1170, object.ClassWeapon
			def, free := alloc.New(server.Modifier{})
			t.Cleanup(free)
			def.Desc8 = alloc.InternCString16("SWORD")
			f.d.weaponDef = func(uint32) *server.Modifier { return def }
			f.d.language = func() int { return tc.lang }
			item := f.dr.UnionItem()
			item.Field_108 = f.modifier("A", "wrong-0", "identify-0").C()
			item.Field_109 = f.modifier("B", "wrong-1", "identify-1").C()
			item.Field_110 = f.modifier("C", "wrong-2", "identify-2").C()
			item.Field_111 = f.modifier("wrong-3", "D", "identify-3").C()
			before := *f.dr
			if p := f.run(); p != f.d.scratch || alloc.GoString16(p) != tc.want || *f.dr != before {
				t.Fatalf("tooltip = %q, want %q", alloc.GoString16(p), tc.want)
			}
		})
	}
	for _, lang := range []int{2, 3, 5, 6, 8} {
		for mask := 0; mask < 16; mask++ {
			f := newTooltipFixture4BF050(t)
			f.dr.ObjClass = object.ClassArmor
			f.d.armorDef = func(uint32) *server.Modifier { return &server.Modifier{Desc8: alloc.InternCString16("BASE")} }
			f.d.language = func() int { return lang }
			m := f.modifier("", "", "WRONG")
			fields := []*unsafe.Pointer{&f.dr.UnionItem().Field_108, &f.dr.UnionItem().Field_109,
				&f.dr.UnionItem().Field_110, &f.dr.UnionItem().Field_111}
			for i, p := range fields {
				if mask&(1<<i) != 0 {
					*p = m.C()
				}
			}
			want, before, after := "BASE", 0, 0
			for i := 0; i < 4; i++ {
				if mask&(1<<i) == 0 {
					continue
				}
				prefix := lang == 6 || (lang == 3 && i == 0) || (lang == 8 && i < 2)
				if prefix {
					before++
				} else {
					after++
				}
			}
			want = strings.Repeat(" ", before) + want + strings.Repeat(" ", after)
			if got := alloc.GoString16(f.run()); got != want {
				t.Fatalf("empty descriptions lang=%d mask=%#x got=%q want=%q", lang, mask, got, want)
			}
		}
	}
	f := newTooltipFixture4BF050(t)
	f.dr.ObjClass = object.ClassArmor
	f.d.armorDef = func(uint32) *server.Modifier { return &server.Modifier{Desc8: alloc.InternCString16("BASE")} }
	m, free := alloc.New(server.ModifierEff{})
	t.Cleanup(free)
	u := f.dr.UnionItem()
	u.Field_108, u.Field_109, u.Field_110, u.Field_111 = m.C(), m.C(), m.C(), m.C()
	if got := alloc.GoString16(f.run()); got != "BASE" {
		t.Fatalf("non-null modifiers with null descriptions = %q, want BASE", got)
	}
}

func TestClientAskInfo4BF050EquipmentMasksLiveReadsAndMissingDefinition(t *testing.T) {
	for _, class := range []object.Class{object.ClassWeapon, object.ClassWand, object.ClassFlag, object.ClassArmor} {
		f := newTooltipFixture4BF050(t)
		f.dr.ObjClass, f.dr.TypeIDVal = class, 17
		lookup := func(id uint32) *server.Modifier {
			if id != 17 {
				t.Fatal("wrong initial type")
			}
			f.dr.TypeIDVal = 23
			return nil
		}
		if uint32(class)&0x11001000 != 0 {
			f.d.weaponDef = lookup
		} else {
			f.d.armorDef = lookup
		}
		if got := alloc.GoString16(f.run()); got != "No info: UnknownWeapon" ||
			!reflect.DeepEqual(f.trace, []string{"copy", "type:23", "load:NoArmsInfo:53", "format:No info: %S"}) {
			t.Fatalf("class=%#x missing definition = %q/%v", class, got, f.trace)
		}
	}
	for _, class := range []object.Class{object.ClassWeapon, object.ClassWand, object.ClassFlag, object.ClassArmor} {
		for _, masked := range []bool{false, true} {
			f := newTooltipFixture4BF050(t)
			f.dr.ObjClass = class
			if masked {
				f.dr.ObjSubClass = object.SubClass(0x7800000)
			}
			def := &server.Modifier{Desc8: alloc.InternCString16("BASE")}
			f.d.weaponDef = func(uint32) *server.Modifier { return def }
			f.d.armorDef = f.d.weaponDef
			f.dr.UnionItem().Field_108 = f.modifier("POWER", "", "").C()
			want := "POWER BASE"
			if masked && class == object.ClassWeapon {
				want = "BASE"
			}
			if got := alloc.GoString16(f.run()); got != want {
				t.Fatalf("class=%#x masked=%t name=%q want=%q", class, masked, got, want)
			}
		}
	}
	f := newTooltipFixture4BF050(t)
	f.dr.ObjClass, f.dr.ObjSubClass = object.ClassWeapon, object.SubClass(0x7800000)
	def := &server.Modifier{Desc8: alloc.InternCString16("BASE")}
	f.d.weaponDef = func(uint32) *server.Modifier { f.dr.ObjClass = 0; return def }
	f.dr.UnionItem().Field_108 = f.modifier("LIVE", "", "").C()
	f.d.language = func() int {
		def.Desc8 = alloc.InternCString16("REPLACEMENT")
		(*client.DrawableUnionItem)(unsafe.Pointer(&f.dr.Union)).Field_108 = nil
		return 8
	}
	if got := alloc.GoString16(f.run()); got != "LIVE BASE" {
		t.Fatalf("live class and cached descriptions = %q", got)
	}
}

func TestClientAskInfo4BF050BookRequestPendingAndReply(t *testing.T) {
	for _, tc := range []struct {
		subclass uint32
		kind     byte
		pending  uint32
	}{{1, 1, 137}, {2, 2, 41}, {4, 4, 6}, {7, 1, 137}, {6, 2, 41}} {
		f := newTooltipFixture4BF050(t)
		f.dr.ObjClass, f.dr.ObjSubClass = object.ClassInfoBook, object.SubClass(tc.subclass)
		f.dr.UnionEffect().Field_109 = 0x11223344
		calls := 0
		f.d.send = func(recipient int, p []byte, related *server.Object, disconnect int) int {
			calls++
			if recipient != 31 || related != nil || disconnect != 1 ||
				!reflect.DeepEqual(p, []byte{0xE2, 0x32, 0xFE, tc.kind}) || f.dr.UnionEffect().Field_108 != tc.pending {
				t.Fatalf("book request route/payload/sentinel = %d/%x/%d", recipient, p, f.dr.UnionEffect().Field_108)
			}
			return -1 // The pending sentinel is retained even if send fails.
		}
		if p := f.run(); p != f.d.scratch || alloc.GoString16(p) != "" || calls != 1 ||
			!reflect.DeepEqual(f.trace, []string{"copy", "wire"}) || f.dr.UnionEffect().Field_109 != 0x11223344 {
			t.Fatalf("book kind=%d result=%q calls=%d trace=%v", tc.kind, alloc.GoString16(p), calls, f.trace)
		}
		f.trace = nil
		f.run()
		if calls != 1 || !reflect.DeepEqual(f.trace, []string{"copy"}) {
			t.Fatalf("pending book requested again: calls=%d trace=%v", calls, f.trace)
		}
		f.dr.UnionEffect().Field_108 = 17
		f.trace = nil
		got := alloc.GoString16(f.run())
		want := map[byte]string{1: "BOOK SPELL", 2: "CREATURE LORE", 4: "BOOK ABILITY"}[tc.kind]
		if got != want || calls != 1 {
			t.Fatalf("server reply title=%q want=%q calls=%d", got, want, calls)
		}
	}
	f := newTooltipFixture4BF050(t)
	f.dr.ObjClass, f.dr.ObjSubClass = object.ClassInfoBook, object.SubClass(1)
	f.d.send = func(int, []byte, *server.Object, int) int { f.dr.UnionEffect().Field_108 = 27; return 1 }
	f.run()
	if f.dr.UnionEffect().Field_108 != 27 {
		t.Fatal("synchronous reply overwritten by pending sentinel")
	}
	f.dr.UnionEffect().Field_108 = 0
	f.d.unitCode = func(dr *client.Drawable) uint16 {
		dr.ObjClass, dr.ObjSubClass = 0, 0
		dr.UnionEffect().Field_108 = 23
		return 0xFE32
	}
	f.d.send = func(_ int, packet []byte, _ *server.Object, _ int) int {
		if f.dr.UnionEffect().Field_108 != 137 || !reflect.DeepEqual(packet, []byte{0xE2, 0x32, 0xFE, 1}) {
			t.Fatalf("cached book classification and live pending store = %d/%x", f.dr.UnionEffect().Field_108, packet)
		}
		return 1
	}
	f.run()
}

func TestClientAskInfo4BF050BookLanguageAndCallbackOrder(t *testing.T) {
	for _, tc := range []struct {
		kind, lang uint32
		want       string
		trace      []string
	}{
		{1, 6, "SPELL BOOK", []string{"copy", "lang", "spell:17", "append:SPELL", "append: ", "load:BookOf:288", "append:BOOK"}},
		{1, 8, "BOOK SPELL", []string{"copy", "lang", "load:BookOf:292", "format:%s ", "spell:17", "append:SPELL"}},
		{2, 3, "LORE CREATURE", []string{"copy", "lang", "load:LoreScroll:313", "append:LORE", "append: ", "creature:17", "append:CREATURE"}},
		{2, 5, "LORE CREATURE", []string{"copy", "lang", "load:LoreScroll:313", "append:LORE", "append: ", "creature:17", "append:CREATURE"}},
		{2, 8, "CREATURE LORE", []string{"copy", "lang", "creature:17", "format:%s ", "load:LoreScroll:320", "append:LORE"}},
		{4, 6, "ABILITY BOOK", []string{"copy", "lang", "ability:17", "append:ABILITY", "append: ", "load:BookOf:342", "append:BOOK"}},
		{4, 8, "BOOK ABILITY", []string{"copy", "lang", "load:BookOf:346", "format:%s ", "ability:17", "append:ABILITY"}},
	} {
		f := newTooltipFixture4BF050(t)
		f.dr.ObjClass, f.dr.ObjSubClass = object.ClassInfoBook, object.SubClass(tc.kind)
		f.dr.UnionEffect().Field_108 = 17
		f.d.language = func() int { f.trace = append(f.trace, "lang"); return int(tc.lang) }
		before := *f.dr
		if got := alloc.GoString16(f.run()); got != tc.want || *f.dr != before || !reflect.DeepEqual(f.trace, tc.trace) {
			t.Fatalf("book kind/lang=%d/%d got=%q trace=%v want=%q/%v", tc.kind, tc.lang, got, f.trace, tc.want, tc.trace)
		}
	}
	for _, tc := range []struct {
		kind   uint32
		lang   int
		wantID uint32
	}{
		{1, 6, 17}, {1, 8, 29}, {2, 3, 29}, {2, 8, 17}, {4, 6, 17}, {4, 8, 29},
	} {
		f := newTooltipFixture4BF050(t)
		f.dr.ObjClass, f.dr.ObjSubClass = object.ClassInfoBook, object.SubClass(tc.kind)
		f.dr.UnionEffect().Field_108 = 17
		f.d.language = func() int { f.dr.UnionEffect().Field_108 = 23; return tc.lang }
		load := f.d.loadString
		f.d.loadString = func(s string, line int) *uint16 { f.dr.UnionEffect().Field_108 = 29; return load(s, line) }
		title := func(id uint32) *uint16 {
			if id != tc.wantID {
				t.Fatalf("kind/lang=%d/%d title ID=%d want=%d", tc.kind, tc.lang, id, tc.wantID)
			}
			return alloc.InternCString16("TITLE")
		}
		f.d.spellTitle, f.d.creatureName, f.d.abilityName = title, title, title
		f.run()
	}
}

func TestClientAskInfo4BF050CEntryHighPointersAndRawUTF16(t *testing.T) {
	f := newTooltipFixture4BF050(t)
	f.dr.TypeIDVal, f.dr.ObjClass = 649, object.ClassFood
	f.d.prettyName = func(uint32) *uint16 { return alloc.InternCString16("사과") }
	old := clientAskInfoDepsFactory4BF050
	clientAskInfoDepsFactory4BF050 = func() clientAskInfoDeps4BF050 { return f.d }
	t.Cleanup(func() { clientAskInfoDepsFactory4BF050 = old })
	before := *f.dr
	if got := Nox_xxx_clientAskInfoMb_4BF050(f.dr); got != "사과" || *f.dr != before {
		t.Fatalf("real C entry name=%q", got)
	}
	f.dr.ObjClass = object.ClassWeapon
	base, free := alloc.CloneSlice([]uint16{0xD800, 0xAC00, 0})
	t.Cleanup(free)
	f.d.weaponDef = func(uint32) *server.Modifier { return &server.Modifier{Desc8: &base[0]} }
	Nox_xxx_clientAskInfoMb_4BF050(f.dr)
	if !reflect.DeepEqual(f.buf[:3], base) {
		t.Fatalf("raw UTF-16 units changed: %x want=%x", f.buf[:3], base)
	}
	f.dr.ObjClass, f.dr.ObjSubClass = object.ClassInfoBook, object.SubClass(4)
	f.dr.UnionEffect().Field_108 = 17
	if got := Nox_xxx_clientAskInfoMb_4BF050(f.dr); got != "BOOK ABILITY" {
		t.Fatalf("real C entry book title = %q", got)
	}
}

func TestClientAskInfo4BF050CursorUTF16Boundary(t *testing.T) {
	f := newTooltipFixture4BF050(t)
	f.dr.ObjClass = object.ClassFood
	old := clientAskInfoDepsFactory4BF050
	clientAskInfoDepsFactory4BF050 = func() clientAskInfoDeps4BF050 { return f.d }
	t.Cleanup(func() { clientAskInfoDepsFactory4BF050 = old })
	buf := unsafe.Slice(memmap.PtrUint16(0x5D4594, 1096676), 257)
	previous := append([]uint16(nil), buf...)
	t.Cleanup(func() { copy(buf, previous) })
	for _, value := range []string{"", "사과", "버섯", "가죽 튜닉", "책: 화이어볼",
		strings.Repeat("가", 255), strings.Repeat("나", 256), strings.Repeat("다", 300)} {
		f.d.prettyName = func(uint32) *uint16 { return alloc.InternCString16(value) }
		buf[256] = 0x55AA
		name := Nox_xxx_clientAskInfoMb_4BF050(f.dr)
		if name != value {
			t.Fatalf("C entry UTF-16 name = %q, want %q", name, value)
		}
		Nox_xxx_cursorSetTooltip_4776B0(name)
		want := []rune(value)
		if len(want) > 255 {
			want = want[:255]
		}
		if got := alloc.GoString16(&buf[0]); got != string(want) || buf[256] != 0x55AA {
			t.Fatalf("cursor text/guard = %q/%#x, want %q/0x55AA", got, buf[256], string(want))
		}
	}
}
