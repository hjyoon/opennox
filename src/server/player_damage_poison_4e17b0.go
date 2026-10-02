package server

import "github.com/opennox/libs/object"

// playerDamageMonsterPoison4E17B0 is stock type 5's no-armor path after the
// monster NPC/flags/invulnerability gates. Reflect and physical shield blocks
// require an attack object, so neither runs for source-less poison ticks.
func playerDamageMonsterPoison4E17B0(
	target *Object, damage int32, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, nil, nil, damage, object.DamagePoison)
	}
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, nil, nil, damage, object.DamagePoison)
	}
	update := target.UpdateDataMonster()
	update.Field547 = 2
	update.Field546 = uint32(object.DamagePoison)
	if quest {
		before := damage
		// GAME.EXE spills the binary64 product to binary32 before FISTP.
		damage = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(damage)))
		if before > 0 && damage < 1 {
			damage = 1
		}
	}
	return true, runtime.DefaultDamage(target, nil, nil, damage, object.DamagePoison)
}
