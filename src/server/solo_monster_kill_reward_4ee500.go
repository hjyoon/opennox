package server

import "math"

const (
	soloMonsterKillRewardCoopFlag4EE500    = uint32(0x800)
	soloMonsterKillRewardMonsterBit4EE500  = uint8(0x02)
	soloMonsterKillRewardPlayerBit4EE500   = uint8(0x04)
	soloMonsterKillRewardMessageKey4EE500  = "gainpoints"
	soloMonsterKillRewardMessagePath4EE500 = `C:\NoxPost\src\Server\Object\health.c`
	soloMonsterKillRewardMessageLine4EE500 = 172
)

type soloMonsterKillRewardHooks4EE500[O comparable, M any] struct {
	gameFlag        func(uint32) int32
	loadAttribution func(O) O
	findParent      func(O) O
	loadClassLow    func(O) uint8
	isMonitored     func(O, O) int32
	loadOwner       func(O) O
	loadExperience  func(O) float32
	giveXP          func(O, float32) float64
	loadString      func(string, string, int) M
	sendLineMessage func(O, M, uint32)
}

// soloMonsterKillReward4EE500 preserves GAME.EXE 004EE500. In particular,
// the first Monster in a non-player attribution chain is checked at most once,
// but its live owner is still loaded after the monitored callback. The chain
// deliberately has no cycle or nil-owner guard beyond the original gates.
func soloMonsterKillReward4EE500[O comparable, M any](
	killed O,
	hooks soloMonsterKillRewardHooks4EE500[O, M],
) {
	var zero O
	if killed == zero {
		return
	}
	if hooks.gameFlag(soloMonsterKillRewardCoopFlag4EE500) == 0 {
		return
	}

	attribution := hooks.loadAttribution(killed)
	if attribution == zero {
		return
	}

	monitored := int32(1)
	player := hooks.findParent(attribution)
	if hooks.loadClassLow(player)&soloMonsterKillRewardPlayerBit4EE500 == 0 {
		return
	}

	if attribution != player {
		current := attribution
		checkedMonster := false
		for current != player {
			if checkedMonster {
				break
			}
			if hooks.loadClassLow(current)&soloMonsterKillRewardMonsterBit4EE500 != 0 {
				monitored = hooks.isMonitored(player, current)
				checkedMonster = true
			}
			current = hooks.loadOwner(current)
		}
		if monitored == 0 {
			return
		}
	}

	experience := hooks.loadExperience(killed)
	awarded := hooks.giveXP(player, experience)
	if !(awarded > 0) {
		return
	}

	points := soloMonsterKillRewardPoints4EE500(awarded)
	message := hooks.loadString(
		soloMonsterKillRewardMessageKey4EE500,
		soloMonsterKillRewardMessagePath4EE500,
		soloMonsterKillRewardMessageLine4EE500,
	)
	hooks.sendLineMessage(player, message, points)
}

// soloMonsterKillRewardPoints4EE500 models GAME.EXE 00566DCC: truncate the
// x87 value to signed int64 and return EAX, the low 32 bits. Invalid qword
// conversions yield 0x8000000000000000, whose low 32 bits are zero.
func soloMonsterKillRewardPoints4EE500(value float64) uint32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return uint32(int64(value))
}
