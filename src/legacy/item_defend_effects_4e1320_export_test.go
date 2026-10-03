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

// Execute the real C entry in a fresh process so a regressed PE32 address
// truncation is reported as a test failure, not a fatal crash of its parent.
func TestItemDefendEffectsCEntry4E1320(t *testing.T) {
	if os.Getenv("NOX_TEST_ITEM_DEFEND_4E1320_CHILD") != "1" {
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(exe, "-test.run=^TestItemDefendEffectsCEntry4E1320$", "-test.count=1", "-test.v=true")
		cmd.Env = append(os.Environ(), "NOX_TEST_ITEM_DEFEND_4E1320_CHILD=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("real 004E1320 C entry failed: %v\n%s", err, output)
		}
		t.Logf("real 004E1320 C entry passed:\n%s", output)
		return
	}
	target, freeTarget := alloc.New(server.Object{})
	source, freeSource := alloc.New(server.Object{})
	weapon, freeWeapon := alloc.New(server.Object{})
	item, freeItem := alloc.New(server.Object{})
	initData, freeInit := alloc.New(server.ModifierInitData{})
	armor, freeArmor := alloc.New(server.ModifierEff{})
	grip, freeGrip := alloc.New(server.ModifierEff{})
	damage, freeDamage := alloc.New(int32(0))
	for _, free := range []func(){freeTarget, freeSource, freeWeapon, freeItem, freeInit, freeArmor, freeGrip, freeDamage} {
		t.Cleanup(free)
	}
	*target, *source, *weapon, *item = server.Object{}, server.Object{}, server.Object{}, server.Object{}
	*initData = server.ModifierInitData{}
	fns := playerDamageLateDefendFunctionsNative4E1320()
	*armor = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.armorMultiplier, Valf: 0.25}}
	*grip = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.grip}, DefendCollide88: server.ModifierEffFnc{Val: 0}}
	*damage = int32(math.Float32bits(8))
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(item), unsafe.Pointer(initData), unsafe.Pointer(armor), unsafe.Pointer(grip), unsafe.Pointer(damage)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want above 4 GiB", pointer)
		}
	}
	oldGetServer := GetServer
	GetServer = func() Server { t.Fatal("known/empty effects consulted the global server"); return nil }
	t.Cleanup(func() { GetServer = oldGetServer })
	if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, nil, nil, nil, object.DamageType(math.MinInt32)); got != int32(uint32(uintptr(unsafe.Pointer(target)))) {
		t.Fatalf("empty return=%#x, want target low DWORD", uint32(got))
	}
	target.InvFirstItem = item
	item.ObjFlags = object.FlagEquipped
	item.InitData = unsafe.Pointer(initData)
	// The original gate is flags-only: an equipped item of class zero still
	// reaches slots two and three. It needs neither health nor update data.
	initData.Modifiers[2] = armor
	if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon, damage, object.DamageType(math.MinInt32)); got != 0 || *damage != int32(math.Float32bits(2)) {
		t.Fatalf("armor callback return=%d damage=%#x, want 0/%#x", got, uint32(*damage), math.Float32bits(2))
	}
	initData.Modifiers[3] = grip
	*damage = int32(math.Float32bits(8))
	if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon, damage, object.DamagePoison); got != 0 || *damage != 1 {
		t.Fatalf("armor then Grip return=%d damage=%d, want 0/1", got, *damage)
	}
	item.ObjFlags = object.Flags(0xf0000000)
	if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, nil, nil, nil, object.DamageType(math.MaxInt32)); uint32(got) != 0xf0000000 {
		t.Fatalf("unequipped return=%#x, want flags DWORD", uint32(got))
	}
	item.ObjFlags = object.FlagEquipped
	initData.Modifiers[3] = nil
	for _, tc := range []struct {
		name               string
		function           unsafe.Pointer
		value              float32
		collision          int32
		damage, wantDamage int32
	}{
		{"armor", fns.armorMultiplier, 0.25, 0, int32(math.Float32bits(8)), int32(math.Float32bits(2))},
		{"durability", fns.durabilityMultiplier, 1.25, 0, int32(math.Float32bits(8)), int32(math.Float32bits(6))},
		{"resilience", fns.resilience, 0, 0, math.MinInt32, math.MinInt32},
		{"breaking", fns.breaking, 0, 0, math.MaxInt32, math.MaxInt32},
		{"puncture prone", fns.punctureProne, 0, 0, -1, -1},
		{"inversion", fns.inversion, 0, 0, 99, 0},
		{"inversion unsigned", fns.inversion, 0, math.MinInt32, -99, 1},
		{"grip", fns.grip, 0, 0, -99, 1},
		{"grip unsigned", fns.grip, 0, math.MinInt32, 99, 0},
	} {
		*armor = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: tc.function, Valf: tc.value}, DefendCollide88: server.ModifierEffFnc{Val: tc.collision}}
		*damage = tc.damage
		if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon, damage, object.DamageType(math.MinInt32)); got != 0 || *damage != tc.wantDamage {
			t.Fatalf("%s C entry return=%d damage=%#x, want 0/%#x", tc.name, got, uint32(*damage), uint32(tc.wantDamage))
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		// A foreign PE32 callback is reported but must never receive LP64
		// records. The following supported slot still executes normally.
		var logs bytes.Buffer
		srv := &server.Server{Log: slog.New(slog.NewTextHandler(&logs, nil))}
		GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
		*armor = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: unsafe.Pointer(source)}}
		initData.Modifiers[3] = grip
		*damage = math.MinInt32
		if got := Nox_xxx_itemApplyDefendEffect2_4E1320(target, source, weapon, damage, object.DamageType(math.MinInt32)); got != 0 || *damage != 1 {
			t.Fatalf("unsupported then Grip return=%d damage=%d, want 0/1", got, *damage)
		}
		for _, message := range []string{"equipped Defend native callback is not ported", "procedure=004E1320", "damage=-2147483648", "damage_type=-2147483648"} {
			if !strings.Contains(logs.String(), message) {
				t.Fatalf("unsupported callback log=%q, missing %q", logs.String(), message)
			}
		}
	}
	t.Logf("C entry retained native target/source/weapon/item/context: %p %p %p %p %p", target, source, weapon, item, damage)
}
