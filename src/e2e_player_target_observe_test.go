package opennox

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerTargetObservationCopiesLiveRecordsWithoutMutation(t *testing.T) {
	hp := server.HealthData{Cur: 15, Max: 20}
	pl := server.Player{}
	pl.BeastScrollLvl[3], pl.BeastScrollLvl[9], pl.SpellLvl[spell.SPELL_CHARM] = 1, 1, 2
	monsterHP := server.HealthData{Cur: 16, Max: 20}
	monster := server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled,
		Extent: 2132, ScriptIDVal: 1000173065, TypeInd: 1355, NetCode: 0x9123,
		PosVec: types.Ptf(927, 1149), HealthData: &monsterHP}
	weapon := server.Object{ObjClass: object.ClassWand, ObjFlags: object.FlagEquipped, TypeInd: 221, ScriptIDVal: 1000003083}
	ud := server.PlayerUpdateData{Player: &pl, State: server.PlayerState13, ManaCur: 71, ManaMax: 78,
		Stamina: 45, CursorObj: &monster, EquippedWeapon: &weapon}
	player := server.Object{ObjClass: object.ClassPlayer, ScriptIDVal: 1000000001,
		PosVec: types.Ptf(902, 1065), HealthData: &hp, Poison540: 3, UpdateData: unsafe.Pointer(&ud)}
	sd := server.DurSpell{ID: 7, Spell: uint32(spell.SPELL_CHARM), Level: 2, Caster16: &player,
		Target48: &monster, Flag20: 5, Pos: types.Ptf(930, 1150), Pos2: types.Ptf(928, 1148),
		Frame60: 100, Frame64: 105, Frame68: 190}
	playerBefore, udBefore, plBefore, hpBefore := player, ud, pl, hp
	monsterBefore, monsterHPBefore, weaponBefore, sdBefore := monster, monsterHP, weapon, sd
	got, err := e2eReadPlayerTarget(105, &player, &sd)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(player, playerBefore) || !reflect.DeepEqual(ud, udBefore) || pl != plBefore || hp != hpBefore ||
		!reflect.DeepEqual(monster, monsterBefore) || monsterHP != monsterHPBefore || !reflect.DeepEqual(weapon, weaponBefore) || sd != sdBefore {
		t.Fatal("target observation must not change any live player, target, weapon, guide or spell record")
	}
	if got.Frame != 105 || got.Player.Health != 15 || got.Poison != 3 || got.Mana != 71 || got.MaximumMana != 78 ||
		got.Stamina != 45 || got.State != 13 || got.EquippedWeapon.ScriptID != 1000003083 ||
		!got.ServerGuidePresent || got.ServerGuideLevels[9] != 1 || got.CharmLevel != 2 ||
		got.CursorTarget.ScriptID != 1000173065 || got.CursorTarget.TypeIndex != 1355 || got.CursorTarget.NetCode != 0x9123 ||
		got.OwnedCharmCount != 1 || got.OwnedCharms[0].Target.Health != 16 || got.OwnedCharms[0].Frame68 != 190 ||
		got.CharmOutputTruncated || got.DurationScanTruncated {
		t.Fatalf("missing actual native targeting values: %+v", got)
	}
	monster.PosVec.X, monsterHP.Cur, hp.Cur, pl.BeastScrollLvl[9], ud.ManaCur, sd.Frame68 = 0, 0, 0, 0, 0, 0
	if got.CursorTarget.Position.X != 927 || got.OwnedCharms[0].Target.Health != 16 ||
		got.Player.Health != 15 || got.ServerGuideLevels[9] != 1 || got.Mana != 71 || got.OwnedCharms[0].Frame68 != 190 {
		t.Fatal("later original updates must not alter a copied observation")
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerTargetObservationRejectsMissingOrWrongPlayer(t *testing.T) {
	ud := server.PlayerUpdateData{}
	for _, player := range []*server.Object{nil, {ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&ud)}} {
		if _, err := e2eReadPlayerTarget(1, player, nil); err == nil {
			t.Fatal("missing or monster update data must not be interpreted as a player")
		}
	}
}

func TestPlayerTargetObservationAllowsAbsentOptionalRecords(t *testing.T) {
	ud := server.PlayerUpdateData{ManaCur: 8}
	player := server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&ud)}
	got, err := e2eReadPlayerTarget(2, &player, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Player.Present || got.Player.HealthPresent || got.ServerGuidePresent || got.CursorTarget.Present ||
		got.EquippedWeapon.Present || got.OwnedCharmCount != 0 || got.Mana != 8 || got.DurationScanTruncated {
		t.Fatalf("absence must remain distinct from a selected or learned target: %+v", got)
	}
}

func TestPlayerTargetObservationFiltersActualHostCharm(t *testing.T) {
	ud := server.PlayerUpdateData{}
	player := server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&ud)}
	other := server.Object{ObjClass: object.ClassPlayer}
	third := server.DurSpell{ID: 3, Spell: uint32(spell.SPELL_CHARM), Caster16: &player}
	second := server.DurSpell{ID: 2, Spell: uint32(spell.SPELL_CHARM), Caster16: &other, Next: &third}
	first := server.DurSpell{ID: 1, Spell: uint32(spell.SPELL_LESSER_HEAL), Caster16: &player, Next: &second}
	got, err := e2eReadPlayerTarget(3, &player, &first)
	if err != nil {
		t.Fatal(err)
	}
	if got.OwnedCharmCount != 1 || got.OwnedCharms[0].ID != 3 || got.OwnedCharms[0].Target.Present ||
		first.Next != &second || second.Next != &third {
		t.Fatal("another caster or spell must not be reported as the player's selected Charm")
	}
}

func TestPlayerTargetObservationBoundsDurationOutput(t *testing.T) {
	ud := server.PlayerUpdateData{}
	player := server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&ud)}
	var records [12]server.DurSpell
	for i := range records {
		records[i] = server.DurSpell{ID: uint16(i + 1), Spell: uint32(spell.SPELL_CHARM), Caster16: &player}
		if i != 0 {
			records[i-1].Next = &records[i]
		}
	}
	before := records
	got, err := e2eReadPlayerTarget(4, &player, &records[0])
	if err != nil {
		t.Fatal(err)
	}
	if got.OwnedCharmCount != 12 || !got.CharmOutputTruncated || got.DurationScanTruncated ||
		got.OwnedCharms[0].ID != 1 || got.OwnedCharms[7].ID != 8 || records != before {
		t.Fatal("bounded output must retain original order/count and leave the duration list untouched")
	}
}

func TestPlayerTargetObservationTerminatesCyclicDurationList(t *testing.T) {
	ud := server.PlayerUpdateData{}
	player := server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&ud)}
	sd := server.DurSpell{Spell: uint32(spell.SPELL_CHARM), Caster16: &player}
	sd.Next = &sd
	got, err := e2eReadPlayerTarget(5, &player, &sd)
	if err != nil {
		t.Fatal(err)
	}
	if !got.DurationScanTruncated || !got.CharmOutputTruncated || got.OwnedCharmCount != 1024 || sd.Next != &sd {
		t.Fatal("a cyclic list must be bounded and reported, never repaired by the observer")
	}
}

func TestPlayerTargetObservationRetainsNoPointers(t *testing.T) {
	var check func(reflect.Type)
	check = func(typ reflect.Type) {
		switch typ.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Interface, reflect.Slice, reflect.Func, reflect.Chan:
			t.Fatalf("player-target snapshot retains a live reference: %v", typ)
		case reflect.Struct:
			for i := 0; i < typ.NumField(); i++ {
				check(typ.Field(i).Type)
			}
		case reflect.Array:
			check(typ.Elem())
		}
	}
	check(reflect.TypeOf(e2ePlayerTargetObservation{}))
}

func TestPlayerTargetObservationRuntimeCallsOnlyReaders(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_player_target_observe.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, name := range []string{"add", "Errorf", "e2eError", "e2eReadPlayerTarget", "Frame", "HostUnit", "S",
		"GetMousePos", "uint32", "Nox_client_getCursorType", "Nox_xxx_clientGetSpriteAtCursor_476F90",
		"Get_dword_8531A0_2576", "Marshal", "Printf"} {
		allowed[name] = true
	}
	var runtime *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "ObservePlayerTarget" {
			runtime = fn
		}
	}
	if runtime == nil {
		t.Fatal("missing registered observation runtime")
	}
	ast.Inspect(runtime, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			if !allowed[name] {
				t.Errorf("observation runtime invokes unreviewed operation %q", name)
			}
		}
		return true
	})
}

func TestPlayerTargetObservationDoesNotAssignLiveState(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_player_target_observe.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var copiedRoot func(ast.Expr) bool
	copiedRoot = func(expr ast.Expr) bool {
		switch expr := expr.(type) {
		case *ast.Ident:
			return expr.Name == "out"
		case *ast.SelectorExpr:
			return copiedRoot(expr.X)
		case *ast.IndexExpr:
			return copiedRoot(expr.X)
		}
		return false
	}
	check := func(expr ast.Expr) {
		if _, local := expr.(*ast.Ident); !local && !copiedRoot(expr) {
			t.Errorf("observer assigns through a live record or pointer at %v", expr.Pos())
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				check(lhs)
			}
		case *ast.IncDecStmt:
			check(node.X)
		}
		return true
	})
}
