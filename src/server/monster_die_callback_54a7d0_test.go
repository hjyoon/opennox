package server

import "testing"

func TestMonsterDieCallbackNative54A7D0ExactDropTables(t *testing.T) {
	tests := []struct {
		name string
		kind MonsterDieCallbackKind54A7D0
		roll int
		want MonsterDieDrop54A390
		drop bool
	}{
		{name: "swordsman lower no-drop boundary", kind: MonsterDieCallbackSwordsman54A7D0, roll: 20},
		{
			name: "swordsman sword lower boundary", kind: MonsterDieCallbackSwordsman54A7D0, roll: 21, drop: true,
			want: MonsterDieDrop54A390{TypeID: "Sword", Modifiers: [4]string{"WeaponPower1", "Material1"}},
		},
		{
			name: "swordsman sword upper boundary", kind: MonsterDieCallbackSwordsman54A7D0, roll: 50, drop: true,
			want: MonsterDieDrop54A390{TypeID: "Sword", Modifiers: [4]string{"WeaponPower1", "Material1"}},
		},
		{
			name: "swordsman shield boundary", kind: MonsterDieCallbackSwordsman54A7D0, roll: 51, drop: true,
			want: MonsterDieDrop54A390{TypeID: "WoodenShield", Modifiers: [4]string{"", "Material1"}},
		},
		{name: "urchin shaman no-drop boundary", kind: MonsterDieCallbackUrchinShaman54A850, roll: 25},
		{
			name: "urchin shaman staff boundary", kind: MonsterDieCallbackUrchinShaman54A850, roll: 26, drop: true,
			want: MonsterDieDrop54A390{TypeID: "StaffWooden", Modifiers: [4]string{"WeaponPower1"}},
		},
		{name: "archer no-drop boundary", kind: MonsterDieCallbackArcher54A890, roll: 20},
		{name: "archer bow boundary", kind: MonsterDieCallbackArcher54A890, roll: 21, drop: true, want: MonsterDieDrop54A390{TypeID: "Bow"}},
		{name: "archer quiver boundary", kind: MonsterDieCallbackArcher54A890, roll: 51, drop: true, want: MonsterDieDrop54A390{TypeID: "Quiver"}},
		{name: "ogre no-drop boundary", kind: MonsterDieCallbackOgre54A900, roll: 25},
		{
			name: "ogre axe boundary", kind: MonsterDieCallbackOgre54A900, roll: 26, drop: true,
			want: MonsterDieDrop54A390{TypeID: "OgreAxe", Modifiers: [4]string{"WeaponPower1", "Material2"}},
		},
		{name: "ogre warlord no-drop boundary", kind: MonsterDieCallbackOgreWarlord54A950, roll: 25},
		{
			name: "ogre warlord chakram boundary", kind: MonsterDieCallbackOgreWarlord54A950, roll: 26, drop: true,
			want: MonsterDieDrop54A390{TypeID: "FanChakram", Charge: 5},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit := new(Object)
			randomCalls := 0
			coopCalls := 0
			dropCalls := 0
			if !MonsterDieCallbackNative54A7D0(unit, test.kind, MonsterDieCallbackRuntime54A7D0{
				RandomInt: func(minimum, maximum int) int {
					randomCalls++
					if minimum != 0 || maximum != 100 {
						t.Fatalf("random bounds = %d..%d, want 0..100", minimum, maximum)
					}
					return test.roll
				},
				CoopMode: func() bool {
					coopCalls++
					return true
				},
				DropItem: func(gotUnit *Object, gotDrop MonsterDieDrop54A390) {
					dropCalls++
					if gotUnit != unit || gotDrop != test.want {
						t.Fatalf("drop = %p/%#v, want %p/%#v", gotUnit, gotDrop, unit, test.want)
					}
				},
			}) {
				t.Fatal("registered monster DIE_FUNCTION was not handled")
			}
			if randomCalls != 1 {
				t.Fatalf("random calls = %d, want 1", randomCalls)
			}
			wantCalls := 0
			if test.drop {
				wantCalls = 1
			}
			if coopCalls != wantCalls || dropCalls != wantCalls {
				t.Fatalf("coop/drop calls = %d/%d, want %d/%d", coopCalls, dropCalls, wantCalls, wantCalls)
			}
		})
	}
}

func TestMonsterDieCallbackNative54A7D0SuppressesDropOutsideCoop(t *testing.T) {
	if !MonsterDieCallbackNative54A7D0(new(Object), MonsterDieCallbackOgre54A900, MonsterDieCallbackRuntime54A7D0{
		RandomInt: func(int, int) int { return 100 },
		CoopMode:  func() bool { return false },
		DropItem: func(*Object, MonsterDieDrop54A390) {
			t.Fatal("non-cooperative game created a monster death drop")
		},
	}) {
		t.Fatal("non-cooperative callback was not handled")
	}
}

func TestMonsterDieCallbackNative54A7D0RejectsInvalidDependencies(t *testing.T) {
	randomCalls := 0
	randomInt := func(int, int) int {
		randomCalls++
		return 100
	}
	if MonsterDieCallbackNative54A7D0(nil, MonsterDieCallbackArcher54A890, MonsterDieCallbackRuntime54A7D0{RandomInt: randomInt}) {
		t.Fatal("nil unit was accepted")
	}
	if MonsterDieCallbackNative54A7D0(new(Object), 0xff, MonsterDieCallbackRuntime54A7D0{RandomInt: randomInt}) {
		t.Fatal("unknown DIE_FUNCTION was accepted")
	}
	if randomCalls != 0 {
		t.Fatalf("invalid inputs consumed %d random rolls", randomCalls)
	}
	if MonsterDieCallbackNative54A7D0(new(Object), MonsterDieCallbackArcher54A890, MonsterDieCallbackRuntime54A7D0{}) {
		t.Fatal("missing random source was accepted")
	}
	if MonsterDieCallbackNative54A7D0(new(Object), MonsterDieCallbackArcher54A890, MonsterDieCallbackRuntime54A7D0{RandomInt: randomInt}) {
		t.Fatal("selected drop without cooperative-mode query was accepted")
	}
	if MonsterDieCallbackNative54A7D0(new(Object), MonsterDieCallbackArcher54A890, MonsterDieCallbackRuntime54A7D0{
		RandomInt: randomInt,
		CoopMode:  func() bool { return true },
	}) {
		t.Fatal("selected cooperative drop without item callback was accepted")
	}
	if !MonsterDieCallbackNative54A7D0(new(Object), MonsterDieCallbackArcher54A890, MonsterDieCallbackRuntime54A7D0{
		RandomInt: func(int, int) int { return 0 },
	}) {
		t.Fatal("no-drop roll unnecessarily required drop dependencies")
	}
}
