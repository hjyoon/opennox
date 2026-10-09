package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestE2ESpellUnitMatrixArena(t *testing.T) {
	for _, spectator := range []bool{false, true} {
		original := types.Ptf(100, 200)
		origin, direction, err := e2eSpellUnitMatrixArena(original, 160, spectator, func(a, b types.Pointf) bool {
			if spectator {
				// An unobstructed room admits both the caster-target lane
				// and the perpendicular observer's two sight lines.
				return a.X >= 68 && a.X <= 292 && b.X >= 68 && b.X <= 292 && a.Y >= 168 && a.Y <= 424 && b.Y >= 168 && b.Y <= 424
			}
			// A 64-unit-wide corridor admits separated bodies, but no
			// 80-unit-radius circle. Its walls remain part of the check.
			return a.Y >= 168 && a.Y <= 232 && b.Y >= 168 && b.Y <= 232
		})
		if err != nil || origin != original || direction.Y != 0 || direction.X != 1 {
			t.Fatalf("spectator=%t arena=%v/%v/%v", spectator, origin, direction, err)
		}
	}
	if _, _, err := e2eSpellUnitMatrixArena(types.Pointf{}, 160, true, func(types.Pointf, types.Pointf) bool { return false }); err == nil {
		t.Fatal("accepted blocked spell lane")
	}
}

func TestE2ESpellUnitMatrixObserver(t *testing.T) {
	for _, direction := range []types.Pointf{types.Ptf(1, 0), types.Ptf(-1, 0), types.Ptf(0, 1), types.Ptf(0, -1)} {
		origin := types.Ptf(100, 200)
		target := origin.Add(direction.Mul(160))
		observer := e2eSpellUnitMatrixObserver(origin, direction)
		if observer.Sub(origin).Len() != 192 || observer.Sub(origin).Len() <= target.Sub(origin).Len() || observer.Sub(target).Len() >= 251 {
			t.Errorf("observer loses nearest-enemy or camera margin: origin=%v target=%v observer=%v", origin, target, observer)
		}
	}
}

func TestE2ESpellUnitMatrixPairs(t *testing.T) {
	seen := make(map[[2]string]bool)
	from, to := make(map[string]int), make(map[string]int)
	for _, pair := range e2eSpellUnitMatrixPairs() {
		if pair[0] == pair[1] || seen[pair] {
			t.Fatalf("self or duplicate pair: %v", pair)
		}
		seen[pair] = true
		from[pair[0]]++
		to[pair[1]]++
	}
	for _, kind := range []string{"player", "monster", "NPC"} {
		if from[kind] != 2 || to[kind] != 2 {
			t.Errorf("%s outgoing/incoming coverage=%d/%d, want 2/2", kind, from[kind], to[kind])
		}
	}
}

func TestE2ESpellUnitMatrixKinds(t *testing.T) {
	for kind, want := range map[string]spell.ID{
		"fireball": spell.SPELL_FIREBALL, "magic-missile": spell.SPELL_MAGIC_MISSILE,
		"death-ray": spell.SPELL_DEATH_RAY,
		"poison":    spell.SPELL_POISON, "confused": spell.SPELL_CONFUSE,
		"stun": spell.SPELL_STUN, "slow": spell.SPELL_SLOW,
		"freeze": spell.SPELL_FREEZE, "blind": spell.SPELL_BLIND,
	} {
		if got, ok := e2eSpellUnitMatrixID(kind); !ok || got != want {
			t.Errorf("kind %s: %d/%t, want %d", kind, got, ok, want)
		}
	}
	for _, kind := range []string{"", "Fireball", "NPC", "charm", "slow/player-to-npc", "unknown"} {
		if _, ok := e2eSpellUnitMatrixID(kind); ok {
			t.Errorf("accepted unsupported matrix family %q", kind)
		}
	}
}

func TestE2ESpellUnitMatrixExplosionMarker(t *testing.T) {
	// Original DefaultDamage's 004E0B8F clears the Monster marker before
	// nil-weapon EXPLOSION reaches the raw-type fallback at 004E0FC9.
	// PlayerDamage's 004E18C4/004E18D1 entry also clears Player/NPC markers.
	for _, kind := range []string{"player", "monster", "NPC"} {
		for _, splash := range []bool{false, true} {
			marker, typ := e2eSpellUnitMatrixExplosionMarker(kind, splash, 0xfedcba98)
			wantMarker, wantType := uint32(1), uint32(0xfedcba98)
			if splash {
				wantMarker, wantType = 2, 7
			}
			if marker != wantMarker || typ != wantType {
				t.Errorf("%s splash=%t marker/type=%x/%x, want %x/%x", kind, splash, marker, typ, wantMarker, wantType)
			}
		}
	}
}

func TestE2ESpellUnitMatrixRayParticles(t *testing.T) {
	for _, tc := range []struct {
		to   image.Point
		want int
	}{{image.Pt(0, 0), 1}, {image.Pt(1, 0), 2}, {image.Pt(3, 4), 4}, {image.Pt(160, 0), 81}, {image.Pt(-160, 0), 81}} {
		if got := e2eSpellUnitMatrixRayParticles(image.Point{}, tc.to); got != tc.want {
			t.Errorf("ray to %v count=%d want=%d", tc.to, got, tc.want)
		}
	}
}

func TestE2ESpellUnitMatrixRayHealth(t *testing.T) {
	for _, tc := range []struct {
		frame           uint32
		monster, player uint16
	}{{594, 1900, 1900}, {624, 1900, 1900}, {625, 1901, 1901}, {626, 1901, 1901}, {627, 1902, 1901}, {634, 1905, 1903}, {1100, 2000, 2000}} {
		if got := e2eSpellUnitMatrixRayHealth(2000, 2000, 100, 593, tc.frame, 30, true); got != tc.monster {
			t.Errorf("completed monster frame %d HP=%d want=%d", tc.frame, got, tc.monster)
		}
		if got := e2eSpellUnitMatrixRayHealth(2000, 2000, 100, 593, tc.frame, 30, false); got != tc.player {
			t.Errorf("completed player frame %d HP=%d want=%d", tc.frame, got, tc.player)
		}
	}
	if got := e2eSpellUnitMatrixRayHealth(65000, 65000, 100, 100, 132, 30, true); got != 64912 {
		t.Errorf("stock per-frame regeneration HP=%d want=64912", got)
	}
}

func TestE2ESpellUnitMatrixDeathRayUsesScriptTimer(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_spell_unit_matrix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	timer, cast := false, false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "beginCast" && fn.Name.Name != "beginDeathRayCast" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			s, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if fn.Name.Name == "beginCast" && s.Sel.Name == "NewTimer" && len(call.Args) == 2 {
				callback, ok := call.Args[1].(*ast.SelectorExpr)
				timer = ok && callback.Sel.Name == "beginDeathRayCast"
			}
			if fn.Name.Name == "beginDeathRayCast" && s.Sel.Name == "CastSpellLvl" {
				cast = true
			}
			return true
		})
	}
	if !timer || !cast {
		t.Fatal("instant ray must cast through the ordinary in-tick script timer, after net-list reset")
	}
}

func TestE2ESpellUnitMatrixRenewsOrdinaryInputInEveryWait(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_spell_unit_matrix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	wait, motion := false, false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "complete" && len(fn.Body.List) > 0 {
			statement, ok := fn.Body.List[0].(*ast.ExprStmt)
			if ok {
				call, ok := statement.X.(*ast.CallExpr)
				if ok {
					method, ok := call.Fun.(*ast.SelectorExpr)
					wait = ok && method.Sel.Name == "keepAlive"
				}
			}
		}
		if fn.Name.Name == "keepAlive" {
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 3 {
					return true
				}
				method, ok := call.Fun.(*ast.SelectorExpr)
				queue, queued := call.Args[2].(*ast.Ident)
				motion = ok && method.Sel.Name == "observe" && queued && queue.Name == "e2eQueueInput"
				return true
			})
		}
	}
	if !wait || !motion {
		t.Fatalf("every spell-family wait must renew same-position ordinary motion: wait=%t motion=%t", wait, motion)
	}
}

func TestE2ESpellUnitMatrixNativeMarkerLayout(t *testing.T) {
	playerUpdate := server.PlayerUpdateData{Field76: 2, Field75: 12, Field21: 0x3e800000}
	monsterUpdate := server.MonsterUpdateData{Field547: 1, Field546: 0xabcd, Field1: 0xbe800000}
	for _, tc := range []struct {
		class  object.Class
		update unsafe.Pointer
		want   [3]uint32
	}{
		{object.ClassPlayer, unsafe.Pointer(&playerUpdate), [3]uint32{2, 12, 0x3e800000}},
		{object.ClassMonster, unsafe.Pointer(&monsterUpdate), [3]uint32{1, 0xabcd, 0xbe800000}},
	} {
		u := &server.Object{ObjClass: tc.class, UpdateData: tc.update}
		marker, typ, carry := e2eSpellUnitMatrixMarker(u)
		if got := [3]uint32{marker, typ, carry}; got != tc.want {
			t.Errorf("class %x native marker=%x, want %x", uint32(tc.class), got, tc.want)
		}
	}
}

func TestE2ESpellUnitMatrixObservesRealResults(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_spell_unit_matrix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if assignment, ok := n.(*ast.AssignStmt); ok {
				for _, lhs := range assignment.Lhs {
					selector, ok := lhs.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					switch selector.Sel.Name {
					case "Target", "Buffs", "Poison540", "Field542", "Field547", "Field546", "Field75", "Field76", "Obj130", "Field131", "Frame134", "Damage", "Collide", "Cur":
						t.Errorf("matrix writes a live result/callback: %s.%s", fn.Name.Name, selector.Sel.Name)
					}
				}
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name := selector.Sel.Name
			if strings.Contains(name, "CallDamage") || name == "DoDamage" || name == "DefaultDamage" || name == "ApplyEnchant" || name == "Enchant" || strings.Contains(name, "SendPacket") {
				t.Errorf("matrix injects damage/buff/network: %s.%s", fn.Name.Name, name)
			}
			if fn.Name.Name != "prepare" && fn.Name.Name != "cleanup" && (name == "SetHealth" || name == "SetMaxHealth") {
				t.Errorf("matrix rewrites HP after casting: %s.%s", fn.Name.Name, name)
			}
			return true
		})
	}
	b, err := os.ReadFile("../scripts/e2e/host-game-spell-unit-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(b), "action: check-spell-unit-matrix"); got != 9 {
		t.Errorf("scenario has %d spell families, want 9", got)
	}
	for _, kind := range []string{"fireball", "magic-missile", "poison", "death-ray", "confused", "stun", "slow", "freeze", "blind"} {
		if !strings.Contains(string(b), "text: "+kind+"\n") {
			t.Errorf("scenario omits %s", kind)
		}
	}
}
