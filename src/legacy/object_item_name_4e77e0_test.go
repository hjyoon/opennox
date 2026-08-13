package legacy

import (
	"reflect"
	"strings"
	"testing"
)

type itemNameObject4E77E0 struct {
	class   uint32
	typeInd uint16
	attrs   *itemNameAttrs4E77E0
	name    *itemNameString4E77E0
}

type itemNameAttrs4E77E0 struct {
	mods [4]*itemNameModifier4E77E0
}

type itemNameDefinition4E77E0 struct {
	desc *itemNameString4E77E0
}

type itemNameModifier4E77E0 struct {
	desc  *itemNameString4E77E0
	ident *itemNameString4E77E0
}

type itemNameString4E77E0 struct {
	text string
}

type itemNameFixture4E77E0 struct {
	events        []string
	output        strings.Builder
	weaponDef     *itemNameDefinition4E77E0
	armorDef      *itemNameDefinition4E77E0
	noInfo        *itemNameString4E77E0
	noDescription *itemNameString4E77E0
	spaces        int
	onSpace       func(int)
}

func (f *itemNameFixture4E77E0) hooks() objectItemNameHooks4E77E0[*itemNameObject4E77E0, *itemNameAttrs4E77E0, *itemNameDefinition4E77E0, *itemNameModifier4E77E0, *itemNameString4E77E0, *itemNameString4E77E0] {
	return objectItemNameHooks4E77E0[*itemNameObject4E77E0, *itemNameAttrs4E77E0, *itemNameDefinition4E77E0, *itemNameModifier4E77E0, *itemNameString4E77E0, *itemNameString4E77E0]{
		class: func(obj *itemNameObject4E77E0) uint32 {
			f.events = append(f.events, "class")
			return obj.class
		},
		initData: func(obj *itemNameObject4E77E0) *itemNameAttrs4E77E0 {
			f.events = append(f.events, "init")
			return obj.attrs
		},
		typeInd: func(obj *itemNameObject4E77E0) uint16 {
			f.events = append(f.events, "type")
			return obj.typeInd
		},
		weaponDef: func(ind uint16) *itemNameDefinition4E77E0 {
			f.events = append(f.events, "weapon:"+string(rune(ind)))
			return f.weaponDef
		},
		armorDef: func(ind uint16) *itemNameDefinition4E77E0 {
			f.events = append(f.events, "armor:"+string(rune(ind)))
			return f.armorDef
		},
		unitName: func(obj *itemNameObject4E77E0) *itemNameString4E77E0 {
			f.events = append(f.events, "unit-name")
			return obj.name
		},
		noInfo: func() *itemNameString4E77E0 {
			f.events = append(f.events, "no-info")
			return f.noInfo
		},
		noDescription: func() *itemNameString4E77E0 {
			f.events = append(f.events, "no-description")
			return f.noDescription
		},
		clear: func() {
			f.events = append(f.events, "clear")
			f.output.Reset()
		},
		copy: func(s *itemNameString4E77E0) {
			f.events = append(f.events, "copy:"+s.text)
			f.output.Reset()
			f.output.WriteString(s.text)
		},
		formatNoInfo: func(format, name *itemNameString4E77E0) {
			f.events = append(f.events, "format:"+format.text+":"+name.text)
			f.output.Reset()
			f.output.WriteString(strings.ReplaceAll(format.text, "%S", name.text))
		},
		modifier: func(attrs *itemNameAttrs4E77E0, slot int) *itemNameModifier4E77E0 {
			f.events = append(f.events, "mod:"+string(rune('0'+slot)))
			return attrs.mods[slot]
		},
		modifierDesc: func(mod *itemNameModifier4E77E0) *itemNameString4E77E0 {
			if mod.desc == nil {
				f.events = append(f.events, "desc:nil")
				return nil
			}
			f.events = append(f.events, "desc:"+mod.desc.text)
			return mod.desc
		},
		modifierIdent: func(mod *itemNameModifier4E77E0) *itemNameString4E77E0 {
			if mod.ident == nil {
				f.events = append(f.events, "ident:nil")
				return nil
			}
			f.events = append(f.events, "ident:"+mod.ident.text)
			return mod.ident
		},
		definitionDesc: func(def *itemNameDefinition4E77E0) *itemNameString4E77E0 {
			f.events = append(f.events, "definition-desc")
			return def.desc
		},
		append: func(s *itemNameString4E77E0) {
			f.events = append(f.events, "append:"+s.text)
			f.output.WriteString(s.text)
		},
		appendSpace: func() {
			f.spaces++
			f.events = append(f.events, "space")
			f.output.WriteByte(' ')
			if f.onSpace != nil {
				f.onSpace(f.spaces)
			}
		},
	}
}

func itemNameString4E77E0Of(s string) *itemNameString4E77E0 {
	return &itemNameString4E77E0{text: s}
}

func TestObjectItemName4E77E0NoDescriptionShortCircuit(t *testing.T) {
	f := &itemNameFixture4E77E0{noDescription: itemNameString4E77E0Of("none")}
	objectItemName4E77E0(&itemNameObject4E77E0{class: 0xffffffff &^ itemNameClassMask4E77E0}, f.hooks())
	if got := f.output.String(); got != "none" {
		t.Fatalf("output = %q, want %q", got, "none")
	}
	wantEvents := []string{"class", "no-description", "copy:none"}
	if !reflect.DeepEqual(f.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", f.events, wantEvents)
	}
}

func TestObjectItemName4E77E0DefinitionSelection(t *testing.T) {
	tests := []struct {
		name  string
		class uint32
		want  string
	}{
		{name: "armor", class: 0x02000000, want: "armor:"},
		{name: "weapon", class: 0x01000000, want: "weapon:"},
		{name: "wand", class: 0x00001000, want: "weapon:"},
		{name: "flag", class: 0x10000000, want: "weapon:"},
		{name: "mixed armor and weapon", class: 0x03000000, want: "weapon:"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &itemNameFixture4E77E0{
				weaponDef: &itemNameDefinition4E77E0{},
				armorDef:  &itemNameDefinition4E77E0{},
			}
			objectItemName4E77E0(&itemNameObject4E77E0{class: tc.class, typeInd: 0x8001, attrs: &itemNameAttrs4E77E0{}}, f.hooks())
			if len(f.events) < 4 || !strings.HasPrefix(f.events[3], tc.want) {
				t.Fatalf("lookup events = %#v, want %q lookup", f.events, tc.want)
			}
			if f.events[0] != "class" || f.events[1] != "init" || f.events[2] != "type" {
				t.Fatalf("prefix events = %#v, want class/init/type", f.events[:3])
			}
		})
	}
}

func TestObjectItemName4E77E0NoInfoIgnoresNilInitData(t *testing.T) {
	f := &itemNameFixture4E77E0{noInfo: itemNameString4E77E0Of("missing %S")}
	obj := &itemNameObject4E77E0{
		class:   0x01000000,
		typeInd: 9,
		name:    itemNameString4E77E0Of("Sword"),
	}
	objectItemName4E77E0(obj, f.hooks())
	if got := f.output.String(); got != "missing Sword" {
		t.Fatalf("output = %q, want %q", got, "missing Sword")
	}
	wantEvents := []string{"class", "init", "type", "weapon:\t", "unit-name", "no-info", "format:missing %S:Sword"}
	if !reflect.DeepEqual(f.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", f.events, wantEvents)
	}
}

func TestObjectItemName4E77E0CompositionAndPostSpaceReloads(t *testing.T) {
	primary := &itemNameModifier4E77E0{desc: itemNameString4E77E0Of("Primary")}
	secondary := &itemNameModifier4E77E0{desc: itemNameString4E77E0Of("Secondary")}
	oldTwo := &itemNameModifier4E77E0{desc: itemNameString4E77E0Of("OldTwo")}
	newTwo := &itemNameModifier4E77E0{desc: itemNameString4E77E0Of("NewTwo")}
	oldThree := &itemNameModifier4E77E0{ident: itemNameString4E77E0Of("OldThree")}
	newThree := &itemNameModifier4E77E0{ident: itemNameString4E77E0Of("NewThree")}
	attrs := &itemNameAttrs4E77E0{mods: [4]*itemNameModifier4E77E0{primary, secondary, oldTwo, oldThree}}
	f := &itemNameFixture4E77E0{
		weaponDef: &itemNameDefinition4E77E0{desc: itemNameString4E77E0Of("Sword")},
	}
	f.onSpace = func(n int) {
		switch n {
		case 3:
			attrs.mods[2] = newTwo
		case 4:
			attrs.mods[3] = newThree
		}
	}
	objectItemName4E77E0(&itemNameObject4E77E0{class: 0x01000000, attrs: attrs}, f.hooks())
	if got := f.output.String(); got != "Primary Secondary Sword NewTwo NewThree" {
		t.Fatalf("output = %q, want post-space reload result", got)
	}
	wantTail := []string{
		"mod:2", "desc:OldTwo", "space", "mod:2", "desc:NewTwo", "append:NewTwo",
		"mod:3", "ident:OldThree", "space", "mod:3", "ident:NewThree", "append:NewThree",
	}
	if got := f.events[len(f.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("reload events = %#v, want %#v", got, wantTail)
	}
}

func TestObjectItemName4E77E0NonNilEmptyDescriptionStillAddsSpace(t *testing.T) {
	attrs := &itemNameAttrs4E77E0{
		mods: [4]*itemNameModifier4E77E0{{desc: itemNameString4E77E0Of("")}},
	}
	f := &itemNameFixture4E77E0{weaponDef: &itemNameDefinition4E77E0{}}
	objectItemName4E77E0(&itemNameObject4E77E0{class: 0x01000000, attrs: attrs}, f.hooks())
	if got := f.output.String(); got != " " {
		t.Fatalf("output = %q, want one space for a non-nil empty string", got)
	}
}

func TestObjectItemName4E77E0DefinitionWithNilInitDataFaultsAfterClear(t *testing.T) {
	f := &itemNameFixture4E77E0{weaponDef: &itemNameDefinition4E77E0{}}
	defer func() {
		if recover() == nil {
			t.Fatal("definition with nil init data did not fault")
		}
		wantPrefix := []string{"class", "init", "type", "weapon:\x00", "clear", "mod:0"}
		if !reflect.DeepEqual(f.events, wantPrefix) {
			t.Fatalf("events = %#v, want %#v", f.events, wantPrefix)
		}
	}()
	objectItemName4E77E0(&itemNameObject4E77E0{class: 0x01000000}, f.hooks())
}
