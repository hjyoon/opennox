package legacy

import (
	"bytes"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Isolate the real C entry so a regressed pointer truncation fails this test
// without terminating unrelated tests in its parent process.
func TestItemPreDamageCEntry4E13B0(t *testing.T) {
	if os.Getenv("NOX_TEST_ITEM_PRE_DAMAGE_4E13B0_CHILD") != "1" {
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(exe, "-test.run=^TestItemPreDamageCEntry4E13B0$", "-test.count=1", "-test.v=true")
		cmd.Env = append(os.Environ(), "NOX_TEST_ITEM_PRE_DAMAGE_4E13B0_CHILD=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("real 004E13B0 C entry failed: %v\n%s", err, output)
		}
		t.Logf("real 004E13B0 C entry passed:\n%s", output)
		return
	}
	weapon, freeWeapon := alloc.New(server.Object{})
	initData, freeInit := alloc.New(server.ModifierInitData{})
	for _, free := range []func(){freeWeapon, freeInit} {
		t.Cleanup(free)
	}
	*weapon = server.Object{InitData: unsafe.Pointer(initData)}
	*initData = server.ModifierInitData{}
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(weapon), unsafe.Pointer(initData)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
		}
	}
	t.Logf("real C pre-Damage: weapon=%p init=%p", weapon, initData)
	if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(nil, nil, weapon, nil); got != 0 {
		t.Fatalf("empty four-slot counter=%d, want 0", got)
	}
	target, freeTarget := alloc.New(server.Object{})
	source, freeSource := alloc.New(server.Object{})
	effect, freeEffect := alloc.New(server.ModifierEff{})
	other, freeOther := alloc.New(server.ModifierEff{})
	empty, freeEmpty := alloc.New(server.ModifierEff{})
	damage, freeDamage := alloc.New(int32(0))
	for _, free := range []func(){freeTarget, freeSource, freeEffect, freeOther, freeEmpty, freeDamage} {
		t.Cleanup(free)
	}
	*target, *source, *empty = server.Object{}, server.Object{}, server.ModifierEff{}
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(effect), unsafe.Pointer(other), unsafe.Pointer(empty), unsafe.Pointer(damage)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
		}
	}
	oldGetServer := GetServer
	GetServer = func() Server { t.Fatal("empty/non-unit stock effects consulted the global server"); return nil }
	t.Cleanup(func() { GetServer = oldGetServer })
	fns := itemPreDamageFunctionsNative4E13B0()
	for _, fnc := range []unsafe.Pointer{fns.drainMana, fns.vampirism, fns.poison, fns.panic, fns.sympathy} {
		*effect = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fnc}}
		initData.Modifiers = [4]*server.ModifierEff{effect, empty, nil, effect}
		*damage = math.MinInt32
		if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, damage); got != 0 || *damage != math.MinInt32 {
			t.Fatalf("stock callback=%p counter=%d damage=%#x", fnc, got, uint32(*damage))
		}
		if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, nil); got != 0 {
			t.Fatalf("stock callback=%p nil-damage counter=%d", fnc, got)
		}
	}

	// Execute actual stock Vampirism/Sympathy and production HP services,
	// not substituted callbacks. Clamping at max HP makes their order visible:
	// Vampirism 19->20, then Sympathy 20->18 (the reverse order ends at 20).
	srv := npcReflectLegacyServer4E17B0(t)
	sourceHP, freeSourceHP := alloc.New(server.HealthData{})
	targetHP, freeTargetHP := alloc.New(server.HealthData{})
	t.Cleanup(freeSourceHP)
	t.Cleanup(freeTargetHP)
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(sourceHP), unsafe.Pointer(targetHP)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native HP pointer=%p, want above 4 GiB", ptr)
		}
	}
	*sourceHP = server.HealthData{Cur: 19, Max: 20, Field2: 20}
	*targetHP = server.HealthData{Cur: 20, Max: 20, Field2: 20}
	source.ObjClass, source.HealthData = object.ClassMonster, sourceHP
	target.ObjClass, target.HealthData = object.ClassMonster, targetHP
	*effect = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.vampirism, Valf: 0.5}}
	*other = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.sympathy, Valf: 0.25}}
	initData.Modifiers = [4]*server.ModifierEff{effect, empty, nil, other}
	*damage = 8
	if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, damage); got != 0 || *damage != 8 || sourceHP.Cur != 18 || targetHP.Cur != 20 {
		t.Fatalf("actual C stock effects counter=%d damage=%d source HP=%d target HP=%d, want 0/8/18/20", got, *damage, sourceHP.Cur, targetHP.Cur)
	}
	// Poison ignores its damage/context argument. A helper-level nil guard
	// must not suppress this original callback when the address is nil.
	*effect = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.poison, Val: 4}}
	initData.Modifiers = [4]*server.ModifierEff{3: effect}
	if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, nil); got != 0 || target.Poison540 != 1 {
		t.Fatalf("nil-context stock Poison counter=%d poison=%d, want 0/1", got, target.Poison540)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		var logs bytes.Buffer
		srv.Log = slog.New(slog.NewTextHandler(&logs, nil))
		*effect = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: unsafe.Pointer(source)}}
		initData.Modifiers = [4]*server.ModifierEff{effect, empty, nil, other}
		if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, damage); got != 0 || sourceHP.Cur != 16 || *damage != 8 {
			t.Fatalf("unsupported then actual Sympathy counter=%d HP=%d damage=%d", got, sourceHP.Cur, *damage)
		}
		for _, message := range []string{"weapon pre-damage native callback is not ported", "procedure=004E13B0", "damage=8"} {
			if !strings.Contains(logs.String(), message) {
				t.Fatalf("unsupported log=%q, missing %q", logs.String(), message)
			}
		}
		*other = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.panic}}
		*damage = math.MinInt32
		if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, damage); got != 0 || *damage != math.MinInt32 {
			t.Fatalf("raw signed unknown counter=%d damage=%#x", got, uint32(*damage))
		}
		if got := Nox_xxx_itemApplyPreDamageEffect_4E13B0(target, source, weapon, nil); got != 0 || !strings.Contains(logs.String(), "damage_present=false") {
			t.Fatalf("nil-context unknown counter=%d log=%q", got, logs.String())
		}
	}
	t.Logf("actual C stock effects retained target=%p source=%p weapon=%p modifier=%p damage=%p; source HP 19->20->18->16, nil-context Poison=1", target, source, weapon, effect, damage)
}
