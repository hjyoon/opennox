package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2EEnvironmentLavaArena(t *testing.T) {
	first := types.Ptf(100, 100)
	tile := func(p types.Pointf) int {
		if math.Abs(float64(p.X-first.X)) < 32 && math.Abs(float64(p.Y-first.Y)) < 32 {
			return 6
		}
		return 0
	}
	clear := func(types.Pointf, types.Pointf) bool { return true }
	contact, outside, err := e2eEnvironmentLavaArena(first, 19, tile, clear)
	if err != nil || contact != first || tile(outside) == 6 || outside.Sub(contact).Len() < 79 {
		t.Fatalf("lava with a visible safe footprint: %v/%v/%v", contact, outside, err)
	}
	for _, radius := range []float32{-1, 0, 33, float32(math.NaN()), float32(math.Inf(1))} {
		if _, _, err := e2eEnvironmentLavaArena(first, radius, tile, clear); err == nil {
			t.Errorf("accepted invalid radius %g", radius)
		}
	}
	for _, tc := range []struct {
		name  string
		tile  func(types.Pointf) int
		clear func(types.Pointf, types.Pointf) bool
	}{
		{"nil floor", nil, clear}, {"nil trace", tile, nil},
		{"no lava", func(types.Pointf) int { return 0 }, clear},
		{"no safe floor", func(types.Pointf) int { return 6 }, clear},
		{"blocked view", tile, func(types.Pointf, types.Pointf) bool { return false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := e2eEnvironmentLavaArena(first, 19, tc.tile, tc.clear); err == nil {
				t.Fatal("accepted missing lava/safe/view requirement")
			}
		})
	}
}

func TestE2EEnvironmentCollisionDamage(t *testing.T) {
	for _, tc := range []struct {
		stockByte uint8
		even, odd int32
	}{
		{0, 0, 0}, {1, 0, 1}, {2, 1, 1}, {3, 1, 1},
		{8, 4, 4}, {127, 63, 63}, {128, 64, 64}, {255, 127, 127},
	} {
		for _, frame := range []uint32{0, 1, 0xfffffffe, 0xffffffff} {
			want := tc.even
			if frame&1 != 0 {
				want = tc.odd
			}
			if got := e2eEnvironmentCollisionDamage(tc.stockByte, frame); got != want {
				t.Errorf("stock byte %d frame %#x: got %d, want %d", tc.stockByte, frame, got, want)
			}
		}
	}
}

func TestE2EEnvironmentDamageObservesNaturalCollision(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_environment_damage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name := selector.Sel.Name
			// GAME.EXE 004E1E49/004E1E78 copy the raw damage-type
			// DWORD for LAVA as well as FLAME, not its float32 bits.
			if fn.Name.Name == "observe" && name == "Float32bits" {
				t.Error("environment hit observer reinterprets the raw damage type as float32")
			}
			if strings.Contains(name, "CallDamage") || name == "DoDamage" || strings.Contains(name, "DamageCollide") || strings.Contains(name, "SendPacket") {
				t.Errorf("fixture injects damage/collision/network: %s.%s", fn.Name.Name, name)
			}
			if (fn.Name.Name == "observe" || fn.Name.Name == "startContact" || fn.Name.Name == "clientResult") && (name == "SetHealth" || name == "SetMaxHealth") {
				t.Errorf("fixture rewrites HP after contact: %s.%s", fn.Name.Name, name)
			}
			return true
		})
	}
	b, err := os.ReadFile("../scripts/e2e/host-game-environment-damage.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"Flame", "Spike", "PeriodicSpike", "SpikeBlock", "SpikeBlockImmobile", "RotatingSpikes", "RotatingSpikesImmobile", "Lava"} {
		if !strings.Contains(string(b), "item: "+kind+"\n") {
			t.Errorf("scenario omits %s", kind)
		}
	}
}
