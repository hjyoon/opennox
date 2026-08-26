package server

import (
	"reflect"
	"testing"
)

func TestMonsterAutoSpellsDefault54C0C0(t *testing.T) {
	update := &MonsterUpdateData{Field509: 0xaabbcc00}
	if got := monsterAutoSpells54C0C0(0x1234, monsterAutoSpellsDefault54C0C0, update, 30, false); got != 0x1234 {
		t.Fatalf("return = %#x, want 0x1234", got)
	}
	if update.Field509 != 0xaabbcc01 {
		t.Fatalf("Field509 = %#x, want upper bytes preserved and low byte 1", update.Field509)
	}
}

func TestMonsterAutoSpellsKinds54C0C0(t *testing.T) {
	tests := []struct {
		name  string
		kind  monsterAutoSpellsKind54C0C0
		quest bool
		ret   uint16
		want  MonsterUpdateData
	}{
		{
			name: "UrchinShaman", kind: monsterAutoSpellsUrchinShaman54C0C0, ret: 30,
			want: MonsterUpdateData{
				Field510: 2, Field410: 0x08000000,
				Field362_2: 15, Field384: 0x20000000,
				Field366_0: 90, Field366_2: 150,
				Field368_2: 90, Field370_0: 30, Field370_2: 90,
				Field430: 0x40000000, Field424: 0x40000000, Field438: 0x80000000,
				Field509: 1,
			},
		},
		{
			name: "Wizard", kind: monsterAutoSpellsWizard54C0C0, ret: 0,
			want: MonsterUpdateData{
				Field510: 3, Field385: 0x08000000, Field410: 0x08000000,
				Field411: 0x10000000, Field444: 0x20000000,
				Field399: 0x40000000, Field422: 0x40000000, Field415: 0x40000000,
				Field396: 0x40000000, Field376: 0x80000000, Field509: 1,
			},
		},
		{
			name: "Beholder regular", kind: monsterAutoSpellsBeholder54C0C0, ret: 0,
			want: MonsterUpdateData{Field510: 3, Field396: 0x40000000, Field376: 0x80000000, Field509: 1},
		},
		{
			name: "Beholder quest", kind: monsterAutoSpellsBeholder54C0C0, quest: true, ret: 1,
			want: MonsterUpdateData{Field510: 3, Field396: 0x40000000, Field509: 1},
		},
		{
			name: "Lich", kind: monsterAutoSpellsLich54C0C0, ret: 0,
			want: MonsterUpdateData{
				Field510: 3, Field385: 0x08000000, Field410: 0x08000000,
				Field443: 0x10000000, Field444: 0x20000000, Field405: 0x20000000,
				Field411: 0x80000000, Field399: 0x40000000, Field415: 0x40000000,
				Field396: 0x40000000, Field509: 1,
			},
		},
		{
			name: "LichLord", kind: monsterAutoSpellsLichLord54C0C0, ret: 30,
			want: MonsterUpdateData{
				Field510: 3, Field385: 0x08000000, Field410: 0x08000000,
				Field443: 0x10000000, Field368_0: 90, Field368_2: 150,
				Field403: 0x40000000, Field411: 0x80000000, Field509: 1,
			},
		},
		{
			name: "Demon", kind: monsterAutoSpellsDemon54C0C0, ret: 0,
			want: MonsterUpdateData{
				Field510: 3, Field385: 0x08000000, Field446: 0x20000000,
				Field399: 0x40000000, Field382: 0x40000000,
				Field411: 0x80000000, Field509: 1,
			},
		},
		{
			name: "WizardGreen", kind: monsterAutoSpellsWizardGreen54C0C0, ret: 0,
			want: MonsterUpdateData{
				Field510: 3, Field410: 0x08000000, Field384: 0x20000000,
				Field430: 0x40000000, Field432: 0x40000000, Field422: 0x40000000,
				Field376: 0x80000000, Field509: 1,
			},
		},
		{
			name: "WillOWisp", kind: monsterAutoSpellsWillOWisp54C0C0, ret: 0x1234,
			want: MonsterUpdateData{Field510: 3, Field415: 0x40000000, Field509: 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var update MonsterUpdateData
			got := monsterAutoSpells54C0C0(0x1234, test.kind, &update, 30, test.quest)
			if got != test.ret {
				t.Fatalf("return = %#x, want %#x", got, test.ret)
			}
			if !reflect.DeepEqual(update, test.want) {
				t.Fatalf("update differs:\n got: %+v\nwant: %+v", update, test.want)
			}
		})
	}
}
