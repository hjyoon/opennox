package legacy

import (
	"encoding/binary"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Frame observation reads the actual foreign clock in a controlled call.
// The ordinary path reads the live server frame at each original read site.
func rawDefaultDamageNativeFrame4E0B30(site uint32, width int) uint32 {
	if value, handled := rawDefaultDamageObservedFrame4E0B30(site, width); handled {
		return value
	}
	value := GetServer().S().Frame()
	switch width {
	case 1:
		return uint32(uint8(value))
	case 4:
		return value
	default:
		panic("undeclared raw default damage frame width")
	}
}

func rawDefaultDamageConvertNative4E0B30(site, bits uint32, cw uint16) int32 {
	rawDefaultDamageObserveConversion4E0B30(site, false, bits, 0)
	value := rawDefaultDamageFloatConvert32Bits4E0B30(bits, cw)
	rawDefaultDamageObserveConversion4E0B30(site, true, bits, uint32(value))
	return value
}

func rawDefaultDamageNativeFloatReply4E0B30(value float64) rawDefaultDamageNativeReply4E0B30 {
	reply := rawDefaultDamageNativeReply4E0B30{FloatFormat: "f64"}
	binary.LittleEndian.PutUint64(reply.FloatPayload[:8], math.Float64bits(value))
	return reply
}

// Controlled services return only independently declared inputs. No reference
// row, expected arithmetic result or branch schedule reaches the raw engine.
// Product services use existing native implementations. Their own admission,
// callback-body and Go-heap lifetime qualifications remain separate gates.
func rawDefaultDamageInvokeService4E0B30(call rawDefaultDamageNativeService4E0B30) rawDefaultDamageNativeReply4E0B30 {
	if reply, handled := rawDefaultDamageDispatchObservedService4E0B30(call); handled {
		return reply
	}
	a := call.Arguments
	var reply rawDefaultDamageNativeReply4E0B30
	// The actual callback must precede any typed Go object cast or world
	// lookup: nil and unknown targets retain their attempted-call fault.
	switch call.Site {
	case 0x4e0d05:
		rawDefaultDamageCallDamageNativeCallback4E0B30(call.Callback, a[0].Pointer, a[1].Pointer, a[2].Pointer, int32(a[3].Word), int32(a[4].Word))
		return reply
	case 0x4e1049, 0x4e104f:
		rawDefaultDamageCallSoundNativeCallback4E0B30(call.Callback, a[0].Pointer, a[1].Pointer)
		return reply
	case 0x4e0c45, 0x4e1168:
		reply.Word = rawDefaultDamageGameFlagsNative4E0B30(a[0].Word)
		return reply
	case 0x4e0c05:
		reply.Word = rawDefaultDamageGameplayFlagsNative4E0B30(int32(a[0].Word))
		return reply
	}
	unit := func(i int) *server.Object { return (*server.Object)(a[i].Pointer) }
	world := func() *server.Server { return GetServer().S() }
	switch call.Site {
	case 0x4e0b46, 0x4e0ca5, 0x4e1079, 0x4e11c2:
		reply.Word = uint32(unitBuffTestLegacy4FF350(unit(0), int32(a[1].Word)))
	case 0x4e109f:
		reply.Word = uint32(unitBuffPowerLegacy4FF570(unit(0), int32(a[1].Word)))
	case 0x4e0cdb, 0x4e0f33, 0x4e0f51:
		reply.Word = uint32(spellBuffOffLegacy4FF5B0(unit(0), int32(a[1].Word)))
	case 0x4e0c16:
		reply.Pointer = unsafe.Pointer(unit(0).FindOwnerChainPlayer())
	case 0x4e0bc2:
		reply.Word = server.UnitIsZombie534A40(unit(0), zombieEntryTypes534A10())
	case 0x4e0c2c, 0x4e0c63, 0x4e11aa:
		if world().IsEnemyTo(unit(0), unit(1)) {
			reply.Word = 1
		}
	case 0x4e0cec, 0x4e10b0:
		reply = rawDefaultDamageNativeFloatReply4E0B30(world().Balance.FloatInd(a[0].String, int(int32(a[1].Word))))
	case 0x4e0d9d:
		reply = rawDefaultDamageNativeFloatReply4E0B30(fireProtectionCall4DFE40(world(), unit(0)))
	case 0x4e0e06:
		reply = rawDefaultDamageNativeFloatReply4E0B30(electricProtectionCall4DFF40(world(), unit(0)))
	case 0x4e0b66, 0x4e0cd3, 0x4e0dc5, 0x4e0e2e, 0x4e1097:
		world().Audio.EventObj(sound.ID(int32(a[0].Word)), unit(1), int(int32(a[2].Word)), a[3].Word)
	case 0x4e0d18, 0x4e115b:
		if Nox_xxx_playerSetState_4FA020(unit(0), server.PlayerState(a[1].Word)) {
			reply.Word = 1
		}
	case 0x4e0f77:
		reply.Word = uint32(Nox_xxx_itemApplyDefendEffect2_4E1320(unit(0), unit(1), unit(2), a[3].Damage, object.DamageType(a[4].Word)))
	case 0x4e0fef:
		// The original caller exposes five words, while the current native
		// callee has four formal arguments. The observer above retains all
		// five; the actual helper receives its declared four arguments.
		reply.Word = uint32(Nox_xxx_itemApplyPreDamageEffect_4E13B0(unit(0), unit(1), unit(2), a[3].Damage))
	case 0x4e1019:
		reply.Pointer = Nox_xxx_monsterGetSoundSet_424300(unit(0))
	case 0x4e102d:
		reply.Word = uint32(world().Audio.Field12(sound.ID(int32(a[0].Word))))
	case 0x4e105c:
		reply.Word = uint32(server.Nox_xxx_soundDefaultDamageSound_532E20(unit(0), unit(1)))
	case 0x4e10de:
		unitAdjustHPCall4EE460(unit(0), int32(a[1].Word))
	case 0x4e1122:
		// Observe all four DWORDs and the full amount before the endpoint's
		// unsigned WORD projection, including the recipient center.
		coordinates := a[1].Coordinates
		from := image.Point{X: int(uint16(coordinates[0])), Y: int(uint16(coordinates[1]))}
		to := image.Point{X: int(uint16(coordinates[2])), Y: int(uint16(coordinates[3]))}
		world().Nox_xxx_netSendVampFx_523270(netmsg.Op(a[0].Word), from, to, int(uint16(a[2].Word)))
	case 0x4e1131:
		s := world()
		s.GameBallOnPlayerDamage4E1230(unit(0), unit(1), int32(a[2].Word), gameBallPlayerDamageRuntime4E1230(s))
	case 0x4e117c:
		*a[2].Damage = world().DefaultDamageFieldGuide4E0B30(unit(0), unit(1), *a[2].Damage)
	case 0x4e11b7:
		s := world()
		reply.Pointer = unsafe.Pointer(server.MonsterHitLatch532880(unit(0), s.Frame))
	case 0x4e11f2:
		spellShieldReduceDamageNative52F710(world(), unit(0), a[1].Damage, object.DamageType(a[2].Word), unit(3))
	case 0x4e1212:
		unitDamageClearCall4EE5E0(unit(0), int32(a[1].Word))
	default:
		panic("undeclared raw default damage service site")
	}
	return reply
}
