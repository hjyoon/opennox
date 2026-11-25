package opennox

import (
	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

func (c *Client) handleNativeClientPacket(op netmsg.Op, data []byte) (int, bool) {
	switch op {
	case netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT:
		return c.handleAudioEventPacketNative48EA70(op, data), true
	case netmsg.MSG_PLAYER_DIED:
		netCode, ok := playerDiedNetCode48EA70(data)
		if !ok {
			return -1, true
		}
		if !nox_client_isConnected() {
			return 3, true
		}
		player := c.srv.Players.ByID(int(netCode))
		if player == nil {
			return 3, true
		}
		if player == getCurPlayer() {
			legacy.ItemAmountCancel4BFE40()
			legacy.Sub_478000()
		}
		if drawable := c.Objs.ByNetCodeDynamic(int(netCode)); drawable != nil {
			c.Objs.RemoveHealthBar(drawable, 3)
		}
		if !noxflags.HasGame(noxflags.GameModeQuest) {
			clearClientPlayerEquipmentOnDeath48EA70(player)
		}
		return 3, true
	case netmsg.MSG_REPORT_ACQUIRE_CREATURE:
		return c.handleCreatureAcquirePacketNative48EA70(data), true
	case netmsg.MSG_REPORT_LOSE_CREATURE:
		return c.handleCreatureLosePacketNative48EA70(data), true
	case netmsg.MSG_REPORT_MONITOR_CREATURE:
		return c.handleCreatureMonitorPacketNative48EA70(data), true
	case netmsg.MSG_REPORT_UNMONITOR_CREATURE:
		return c.handleCreatureUnmonitorPacketNative48EA70(data), true
	case netmsg.MSG_REPORT_STATS:
		return c.handlePlayerStatsPacketNative48EA70(data), true
	case netmsg.MSG_REPORT_X_STATUS,
		netmsg.MSG_REPORT_PLAYER_STATUS,
		netmsg.MSG_REPORT_MODIFIER,
		netmsg.MSG_REPORT_STAT_MODIFIER,
		netmsg.MSG_REPORT_NPC,
		netmsg.MSG_REPORT_CLIENT_STATUS,
		netmsg.MSG_REPORT_ANIMATION_FRAME:
		return c.handleObjectStatePacketNative48EA70(op, data), true
	case netmsg.MSG_JOURNAL_MSG:
		return c.handleJournalPacketNative48EA70(data), true
	case netmsg.MSG_INTERESTING_ID:
		return c.handleInterestingIDPacketNative48EA70(data), true
	case netmsg.MSG_SIMPLE_OBJ:
		return c.handleSimpleObjectPacketNative519410(data), true
	case netmsg.MSG_COMPLEX_OBJ:
		return c.handleComplexObjectPacketNative519410(data), true
	case netmsg.MSG_DOOR_ANGLE,
		netmsg.MSG_OBELISK_CHARGE,
		netmsg.MSG_PENTAGRAM_ACTIVATE:
		return c.handleFixedObjectStatePacketNative48EA70(op, data), true
	case netmsg.MSG_REPORT_Z_PLUS:
		return c.handleObjectZPacketNative48EA70(data, false), true
	case netmsg.MSG_REPORT_Z_MINUS:
		return c.handleObjectZPacketNative48EA70(data, true), true
	case netmsg.MSG_CLIENT_PREDICT_LINEAR:
		return c.handleClientPredictLinearPacketNative48EA70(data), true
	case netmsg.MSG_FX_PLASMA,
		netmsg.MSG_FX_LIGHTNING,
		netmsg.MSG_FX_ENERGY_BOLT,
		netmsg.MSG_FX_CHAIN_LIGHTNING_BOLT,
		netmsg.MSG_FX_DRAIN_MANA,
		netmsg.MSG_FX_CHARM,
		netmsg.MSG_FX_GREATER_HEAL:
		return c.handleRayFXPacketNative48EA70(op, data), true
	case netmsg.MSG_FX_EXPLOSION,
		netmsg.MSG_FX_LESSER_EXPLOSION,
		netmsg.MSG_FX_COUNTERSPELL_EXPLOSION,
		netmsg.MSG_FX_THIN_EXPLOSION,
		netmsg.MSG_FX_TELEPORT,
		netmsg.MSG_FX_DAMAGE_POOF:
		return c.handlePointSpriteFXPacketNative48EA70(op, data), true
	case netmsg.MSG_FX_BLUE_SPARKS,
		netmsg.MSG_FX_YELLOW_SPARKS,
		netmsg.MSG_FX_CYAN_SPARKS,
		netmsg.MSG_FX_VIOLET_SPARKS:
		return c.handlePointSparkFXPacketNative48EA70(op, data), true
	case netmsg.MSG_FX_SMOKE_BLAST:
		return c.handleSmokeBlastPacketNative48EA70(data), true
	case netmsg.MSG_FX_SPARK_EXPLOSION:
		return c.handleSparkExplosionFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_SENTRY_RAY:
		return c.handleSentryRayFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_RICOCHET:
		return c.handleRicochetFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_GREEN_BOLT:
		return c.handleGreenBoltFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_DELTAZ_SPELL_START:
		return c.handleDeltaZFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_ARROW_TRAP:
		return c.handleArrowTrapFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_VAMPIRISM:
		return c.handleVampirismFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_MANA_BOMB_CANCEL:
		return c.handleManaBombCancelPacketNative48EA70(data), true
	case netmsg.MSG_FX_DURATION_SPELL:
		return c.handleDurationRayPacketNative48EA70(data), true
	case netmsg.MSG_FX_TURN_UNDEAD:
		return c.handleTurnUndeadFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_SUMMON:
		return c.handleSummonFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_SUMMON_CANCEL:
		return c.handleSummonCancelFXPacketNative48EA70(data), true
	case netmsg.MSG_FX_SHIELD:
		return c.handleShieldFXPacketNative48EA70(data), true
	case netmsg.MSG_OBJECT_OUT_OF_SIGHT:
		return c.handleObjectOutOfSightPacketNative48EA70(data), true
	case netmsg.MSG_OBJECT_IN_SHADOWS:
		return c.handleObjectInShadowsPacketNative48EA70(data), true
	case netmsg.MSG_ENABLE_OBJECT,
		netmsg.MSG_DISABLE_OBJECT,
		netmsg.MSG_DRAW_FRAME:
		return c.handleObjectActivationPacketNative48EA70(op, data), true
	case netmsg.MSG_OPEN_WALL,
		netmsg.MSG_CLOSE_WALL,
		netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC,
		netmsg.MSG_REMOVE_WALL_MAGIC:
		return c.handleWallStatePacketNative48EA70(op, data), true
	case netmsg.MSG_REPORT_ENCHANTMENT:
		return c.handleObjectEnchantPacketNative48EA70(data), true
	case netmsg.MSG_REPORT_LIGHT_COLOR,
		netmsg.MSG_REPORT_LIGHT_INTENSITY:
		return c.handleObjectLightPacketNative48EA70(op, data), true
	}
	return 0, false
}
