package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectDataWandParsersKeepPartialStores5361B0(t *testing.T) {
	oldFPS := gameFPSHook
	gameFPSHook = func() uint32 { return 60 }
	t.Cleanup(func() { gameFPSHook = oldFPS })

	projectile := &server.WandUseData{ProjectileType: 99, Flags: 0x40, LastUsed: 27}
	args := CString("300 Fireball 2.5 MULTI_SHOT SmallFireballWand 6")
	defer StrFree(args)
	if sub_536260(args, projectile.UseDataPtr()) != 1 ||
		projectile.Kind != 0 || projectile.Charge != 44 || projectile.MaxCharge != 44 ||
		projectile.Projectile() != "Fireball" || projectile.ProjectileType != 0 ||
		projectile.Progress != 100 || projectile.Cooldown != 24 || projectile.Flags != 0x41 ||
		projectile.Sound != uint32(sound.ByName("SmallFireballWand")) || projectile.LastUsed != 27 {
		t.Fatalf("projectile wand = %+v", projectile)
	}
	if got := GoString(args); got != "300" {
		t.Fatalf("first token was left %q", got)
	}

	cast := &server.WandUseData{Kind: 42, Cooldown: 67, Spell: 78}
	incomplete := CString("5 2.5")
	defer StrFree(incomplete)
	if sub_5361B0(incomplete, cast.UseDataPtr()) != 0 || cast.Kind != 1 || cast.Charge != 5 ||
		cast.MaxCharge != 5 || cast.Cooldown != 24 || cast.Spell != 78 || cast.Progress != 100 {
		t.Fatalf("partial casting wand = %+v", cast)
	}
	name := CString("5 -3 SPELL_DEATH_RAY")
	defer StrFree(name)
	if sub_5361B0(name, cast.UseDataPtr()) != 1 || cast.Cooldown != 0xffffffec ||
		cast.Spell != uint32(spell.ParseID("SPELL_DEATH_RAY")) {
		t.Fatalf("casting wand = %+v", cast)
	}
	// InfinitePainWand's decimal rate from the stock thing.bin. Executing
	// GAME.EXE 00536204 through 00536236 with startup control 0x0e3f
	// produces 299 at 60 FPS.
	stock := CString("3 0.2 SPELL_FORCE_OF_NATURE")
	defer StrFree(stock)
	if sub_5361B0(stock, cast.UseDataPtr()) != 1 || cast.Cooldown != 299 ||
		cast.Spell != uint32(spell.ParseID("SPELL_FORCE_OF_NATURE")) {
		t.Fatalf("stock InfinitePainWand = %+v", cast)
	}
}

func TestObjectDataTextParsersKeepNativeRecords5364E0(t *testing.T) {
	skull := server.SkullUpdateData{ScanDelay: 33, ProjectileType: 77}
	arrow := CString(" \tArrow  other")
	defer StrFree(arrow)
	if sub_5364E0(arrow, unsafe.Pointer(&skull)) != 1 || skull.ProjectileType != 0 ||
		alloc.GoStringS(skull.ProjectileName[:]) != "Arrow" || skull.ScanDelay != 33 {
		t.Fatalf("skull data = %+v", skull)
	}

	trigger := server.TriggerUpdateData{SoundDeactivate: 0xbaadf00d, ClassExclude: 0x1234}
	name := CString("SmallFireballWand")
	defer StrFree(name)
	if sub_5365B0(name, unsafe.Pointer(&trigger)) != 1 ||
		trigger.SoundActivate != uint32(sound.ByName("SmallFireballWand")) ||
		trigger.SoundDeactivate != 0xbaadf00d || trigger.ClassExclude != 0x1234 {
		t.Fatalf("trigger data = %+v", trigger)
	}

	deathArgs := CString("RedPotion\tFireballCast")
	defer StrFree(deathArgs)
	var death server.CreateSpawnObjectDeathData54E010
	if sub_536B40(deathArgs, unsafe.Pointer(&death)) != 1 ||
		alloc.GoStringS(death.TypeID[:]) != "RedPotion" || death.Sound != uint32(sound.ByName("FireballCast")) {
		t.Fatalf("death object data = %+v", death)
	}
}
