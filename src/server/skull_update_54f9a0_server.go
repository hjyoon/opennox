package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

const skullUpdateScanRadius54F9A0 = float32(350)

// SkullUpdateRuntime54F9A0 supplies the two operations that still belong to
// the outer game runtime while keeping all live Object pointers in Go.
type SkullUpdateRuntime54F9A0 struct {
	CreateAt func(object, owner *Object, position types.Pointf)
	InFront  func(source, target *Object) bool
}

type skullUpdateNativeDeps54F9A0 struct {
	tickRate     func() uint32
	eachInRect   func(types.Rectf, func(*Object) bool)
	isEnemy      func(source, target *Object) bool
	mapCheck     func(target, source *Object) bool
	inFront      func(source, target *Object) bool
	typeByName   func(string) int
	newObject    func(int) *Object
	createAt     func(object, owner *Object, position types.Pointf)
	balanceFloat func(string) float64
	floatToInt   func(float32) int32
	arrowTrapFX  func(types.Pointf, byte)
	audioEvent   func(sound.ID, *Object)
}

func skullUpdateTargetReachable54FC50(source *Object, deps skullUpdateNativeDeps54F9A0) bool {
	position := source.PosVec
	radius := skullUpdateScanRadius54F9A0
	rect := types.Rectf{
		Min: types.Ptf(position.X-radius, position.Y-radius),
		Max: types.Ptf(position.X+radius, position.Y+radius),
	}
	found := false
	deps.eachInRect(rect, func(candidate *Object) bool {
		// GAME.EXE reads only the low class byte and accepts players or
		// monsters. The 0x8020 mask rejects destroyed and dead units.
		if uint8(candidate.ObjClass)&uint8(object.MaskUnits) == 0 ||
			uint32(candidate.ObjFlags)&uint32(object.FlagDestroyed|object.FlagDead) != 0 {
			return true
		}
		if deps.isEnemy(source, candidate) &&
			deps.mapCheck(candidate, source) &&
			deps.inFront(source, candidate) {
			found = true
		}
		return true
	})
	return found
}

func skullCreateArrowTrapProjectile54FA80(source *Object, projectileType uint32, deps skullUpdateNativeDeps54F9A0) {
	direction := byte(source.Direction1)
	cosine, sine := SinCosDir(direction)
	distance := float64(source.Shape.Circle.R) + 4.0
	position := types.Ptf(
		float32(distance*float64(cosine)+float64(source.PosVec.X)),
		float32(distance*float64(sine)+float64(source.PosVec.Y)),
	)

	projectile := deps.newObject(int(int32(projectileType)))
	if projectile == nil {
		return
	}
	deps.createAt(projectile, source, position)
	projectile.Direction1 = source.Direction1
	projectile.Direction2 = source.Direction1
	projectile.VelVec.X = float32(float64(cosine) * float64(projectile.SpeedCur))
	projectile.VelVec.Y = float32(float64(sine) * float64(projectile.SpeedCur))

	mercArcherArrow := uint32(int32(deps.typeByName("MercArcherArrow")))
	arrowTrap1 := uint32(int32(deps.typeByName("ArrowTrap1")))
	arrowTrap2 := uint32(int32(deps.typeByName("ArrowTrap2")))
	if uint32(source.TypeInd) == arrowTrap1 || uint32(source.TypeInd) == arrowTrap2 {
		data := (*ProjectileCollideData)(projectile.CollideData)
		damage := deps.floatToInt(float32(deps.balanceFloat("ArrowTrapDamage")))
		data.Damage = damage
		data.Field4 = damage
	}
	if projectileType == mercArcherArrow {
		deps.audioEvent(sound.SoundArrowTrapShoot, source)
	}
}

func skullUpdateNative54F9A0(source *Object, deps skullUpdateNativeDeps54F9A0) uint32 {
	update := source.UpdateDataSkull()
	if !source.Flags().Has(object.FlagEnabled) {
		update.Enabled = 0
		return uint32(source.ObjFlags)
	}

	if update.Enabled == 0 {
		update.ScanDelay = 0
		update.FireDelay = 0
	}
	oldScanDelay := update.ScanDelay
	update.Enabled = 1
	if oldScanDelay == 0 {
		if skullUpdateTargetReachable54FC50(source, deps) {
			if update.TargetReady != 1 {
				update.TargetReady = 1
				update.FireDelay = 0
			}
		} else {
			update.TargetReady = 0
		}
		update.ScanDelay = deps.tickRate()
	}

	if update.TargetReady == 1 && update.FireDelay == 0 {
		// These lookups precede projectile creation in the original update.
		// Object type indices are immutable after thing.bin has loaded, so the
		// native path does not need the two PE32 BSS cache cells.
		arrowTrap1 := uint32(int32(deps.typeByName("ArrowTrap1")))
		arrowTrap2 := uint32(int32(deps.typeByName("ArrowTrap2")))
		skullCreateArrowTrapProjectile54FA80(source, update.ProjectileType, deps)
		switch uint32(source.TypeInd) {
		case arrowTrap1:
			deps.arrowTrapFX(source.PosVec, 1)
		case arrowTrap2:
			deps.arrowTrapFX(source.PosVec, 2)
		}
		update.FireDelay = 30
	}

	if update.ScanDelay != 0 {
		update.ScanDelay--
	}
	if update.FireDelay != 0 {
		update.FireDelay--
	}
	return update.FireDelay
}

func skullArrowTrapFXPacket5238A0(position types.Pointf, variant byte) [6]byte {
	var packet [6]byte
	packet[0] = byte(netmsg.MSG_FX_ARROW_TRAP)
	// nox_float2int16 truncates toward zero and then retains the low word.
	binary.LittleEndian.PutUint16(packet[1:3], uint16(int32(position.X)))
	binary.LittleEndian.PutUint16(packet[3:5], uint16(int32(position.Y)))
	packet[5] = variant
	return packet
}

// SkullUpdate54F9A0 restores GAME.EXE 0054F9A0, 0054FA80, 0054FBB0,
// 0054FBF0, and 0054FC50 without narrowing Object pointers to PE32 ints.
func (s *Server) SkullUpdate54F9A0(source *Object, runtime SkullUpdateRuntime54F9A0) uint32 {
	return skullUpdateNative54F9A0(source, skullUpdateNativeDeps54F9A0{
		tickRate:   s.TickRate,
		eachInRect: s.Map.EachObjInRect,
		isEnemy:    s.IsEnemyTo,
		mapCheck:   s.MapTraceVision,
		inFront:    runtime.InFront,
		typeByName: s.Types.IndByID,
		newObject:  s.NewObjectByTypeInd,
		createAt:   runtime.CreateAt,
		balanceFloat: func(name string) float64 {
			return s.Balance.Float(name)
		},
		floatToInt: playerCollideRound4E8460,
		arrowTrapFX: func(position types.Pointf, variant byte) {
			packet := skullArrowTrapFXPacket5238A0(position, variant)
			s.Nox_xxx_netSendFxAllCli_523030(position, packet[:])
		},
		audioEvent: func(id sound.ID, object *Object) {
			s.Audio.EventObj(id, object, 0, 0)
		},
	})
}
