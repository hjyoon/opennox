package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestE2ETrapProjectileStockKinds(t *testing.T) {
	for _, tc := range []struct {
		kind, projectile string
		dir              server.Dir16
	}{
		{"ArrowTrap1", "MercArcherArrow", 32}, {"ArrowTrap2", "MercArcherArrow", 96},
		{"Skull1", "StrongFireball", 160}, {"Skull2", "StrongFireball", 224},
		{"Skull3", "StrongFireball", 96}, {"Skull4", "StrongFireball", 32},
		{"Polyp", "ToxicCloud", 0}, {"chapter8-tower", "DeathBall", 0},
	} {
		projectile, dir, ok := e2eTrapProjectileKind(tc.kind)
		if !ok || projectile != tc.projectile || dir != tc.dir {
			t.Errorf("%s: %s/%d/%t", tc.kind, projectile, dir, ok)
		}
	}
	for _, kind := range []string{"Skull", "Tower", "PolypFireball", "", "arrowtrap1"} {
		if _, _, ok := e2eTrapProjectileKind(kind); ok {
			t.Errorf("accepted non-shooting/misnamed kind %q", kind)
		}
	}
}

func TestE2ETrapProjectileArmorExpectation(t *testing.T) {
	for _, tc := range []struct {
		typ       object.DamageType
		armored   bool
		want      int32
		remainder float32
	}{
		{object.DamageImpale, true, 5, -.1}, {object.DamageCrush, true, 7, .15},
		{object.DamageFlame, true, 9, .4}, {object.DamagePoison, true, 9, .4},
		{object.DamageImpale, false, 9, .4}, {object.DamageCrush, false, 9, .4},
	} {
		got, bits := e2eTrapProjectileUnitDamage(9, tc.typ, tc.armored, .5, .4)
		if got != tc.want || math.Abs(float64(math.Float32frombits(bits)-tc.remainder)) > 1e-6 {
			t.Errorf("%d/%t: %d/%g", tc.typ, tc.armored, got, math.Float32frombits(bits))
		}
	}
}

func TestE2ETrapProjectileFixedDirectionLane(t *testing.T) {
	for _, direction := range []server.Dir16{32, 96, 160, 224} {
		origin, err := e2eTrapProjectileArena(types.Ptf(100, 100), direction, func(types.Pointf, types.Pointf) bool { return true })
		if err != nil || origin != types.Ptf(100, 100) {
			t.Errorf("stock direction %d: %v/%v", direction, origin, err)
		}
	}
	if _, err := e2eTrapProjectileArena(types.Pointf{}, 32, func(types.Pointf, types.Pointf) bool { return false }); err == nil {
		t.Fatal("accepted blocked original direction")
	}
}

func TestE2ETrapProjectilePoisonIncludesNaturalRegeneration(t *testing.T) {
	for _, tc := range []struct {
		frame  uint32
		player bool
		health uint16
		dot    bool
	}{
		{588, true, 1991, false}, {620, true, 1991, false},
		{621, true, 1992, false}, {625, true, 1993, false},
		{653, true, 2000, false}, {768, true, 2000, false},
		{769, true, 1999, true}, {800, true, 1999, true},
		{801, true, 2000, true},
		{618, false, 1991, false}, {619, false, 1992, false},
		{621, false, 1993, false}, {635, false, 2000, false},
		{768, false, 2000, false}, {769, false, 1999, true},
		{801, false, 2000, true},
	} {
		got, dot, ok := e2eTrapProjectilePoisonHP(587, tc.frame, 30, 2000, 1991, tc.player)
		if !ok || got != tc.health || dot != tc.dot {
			t.Errorf("frame=%d player=%t: HP=%d DOT=%t valid=%t want=%d/%t", tc.frame, tc.player, got, dot, ok, tc.health, tc.dot)
		}
	}
	// Injury, modulo and observation arithmetic all retain unsigned wrap.
	applied := uint32(math.MaxUint32 - 20)
	if got, dot, ok := e2eTrapProjectilePoisonHP(applied, 129, 30, 75, 70, true); !ok || got != 70 || !dot {
		t.Errorf("wrapped player natural regen + DOT: %d/%t/%t", got, dot, ok)
	}
	for _, tc := range []struct {
		applied, frame, fps uint32
		maximum, initial    uint16
	}{
		{0, 100, 30, 2000, 1991}, {587, 769, 0, 2000, 1991},
		{587, 769, 30, 0, 1}, {587, 769, 30, 2000, 2001},
		{587, 769, 30, 2000, 1}, {587, 1488, 30, 2000, 1991},
		{587, 897, 30, 2000, 1991},
	} {
		if _, _, ok := e2eTrapProjectilePoisonHP(tc.applied, tc.frame, tc.fps, tc.maximum, tc.initial, true); ok {
			t.Errorf("accepted invalid first-DOT range: %+v", tc)
		}
	}
}

func TestE2ETrapTowerUsesActualLaunchAndNarrowCorridor(t *testing.T) {
	origin, dir := types.Ptf(100, 100), types.Ptf(1, 0)
	contact, observer, ok := e2eTrapTowerPositions(origin, dir, 34, func(a, b types.Pointf) bool {
		// The Tower is part of the wall. Its ordinary launch at +34 and the
		// target's 18-unit footprint are in open corridor space.
		return a.X >= 134 && b.X >= 134
	})
	if !ok || contact != types.Ptf(160, 100) || math.Abs(float64(observer.Y-contact.Y)) < 192 {
		t.Fatalf("original Tower corridor positions: %v %v %t", contact, observer, ok)
	}
	if _, _, ok := e2eTrapTowerPositions(origin, dir, 20, func(a, b types.Pointf) bool { return a.X >= 120 && b.X >= 120 }); !ok {
		t.Fatal("rejected the original ColorLight FON origin's 20-unit launch")
	}
	if _, _, ok := e2eTrapTowerPositions(origin, dir, 20, func(a, b types.Pointf) bool { return a.X >= 134 && b.X >= 134 }); !ok {
		t.Fatal("rejected a stock origin inset into the corridor wall")
	}
	for _, distance := range []float32{0, -1, 60, 61} {
		if _, _, ok := e2eTrapTowerPositions(origin, dir, distance, func(types.Pointf, types.Pointf) bool { return true }); ok {
			t.Fatalf("accepted invalid launch distance %v", distance)
		}
	}
	if _, _, ok := e2eTrapTowerPositions(origin, dir, 34, func(types.Pointf, types.Pointf) bool { return false }); ok {
		t.Fatal("accepted blocked Tower launch/contact")
	}
}

func TestE2ETrapProjectileDoesNotInjectOutputs(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_trap_projectile_damage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if sel, ok := c.Fun.(*ast.SelectorExpr); ok {
					name := sel.Sel.Name
					if strings.Contains(name, "CallDamage") || strings.HasPrefix(name, "CastSpell") || strings.Contains(name, "SendPacket") || name == "ActivatePoison" || name == "SetGame" {
						t.Errorf("%s injects gameplay output: %s", fn.Name.Name, name)
					}
					if name == "SetHealth" && fn.Name.Name != "cleanup" || name == "SetMaxHealth" && fn.Name.Name != "prepare" && fn.Name.Name != "cleanup" {
						t.Errorf("%s rewrites HP after triggering", fn.Name.Name)
					}
				}
			}
			if a, ok := n.(*ast.AssignStmt); ok {
				for _, lhs := range a.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok {
						switch sel.Sel.Name {
						case "Cur", "Damage", "Collide", "CollideData", "UpdateData", "Direction1", "Field32", "Field34", "Field547", "Field546", "Field76", "Field75", "Poison540", "Frame134", "Obj130", "Field518", "Field57", "Field21", "Field1":
							t.Errorf("%s writes protected result %s", fn.Name.Name, sel.Sel.Name)
						}
					}
				}
			}
			return true
		})
	}
	for _, file := range []string{"../scripts/e2e/host-game-trap-projectile-damage.yaml", "../scripts/e2e/solo-wizard-chapter8-tower-damage.yaml"} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "action: check-trap-projectile-damage") {
			t.Errorf("%s omits trap check", file)
		}
	}
	b, _ := os.ReadFile("../scripts/e2e/solo-wizard-chapter8-tower-damage.yaml")
	if !strings.Contains(string(b), "map: Wiz08e") || strings.Contains(string(b), "call-noxscript-function") {
		t.Fatal("chapter-eight fixture bypasses original map initialization/timer")
	}
}
