package legacy

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

func objectReplayChecksumNative4E7700(obj *server.Object) uint32 {
	var in objectReplayChecksumInput4E7700
	// Keep these assignments in the original object-field load order. Named
	// fields make the same contract valid for both native pointer widths.
	in.teamID = uint8(obj.TeamVal.ID)
	in.typeInd = obj.TypeInd
	in.scriptID = obj.ScriptIDVal
	in.posXBits = math.Float32bits(obj.PosVec.X)
	in.extent = obj.Extent
	in.netCode = obj.NetCode
	in.field5 = obj.Field5
	in.objFlags = uint32(obj.ObjFlags)
	in.posYBits = math.Float32bits(obj.PosVec.Y)
	in.newPosXBits = math.Float32bits(obj.NewPos.X)
	in.newPosYBits = math.Float32bits(obj.NewPos.Y)
	in.prevPosXBits = math.Float32bits(obj.PrevPos.X)
	in.prevPosYBits = math.Float32bits(obj.PrevPos.Y)
	in.velXBits = math.Float32bits(obj.VelVec.X)
	in.velYBits = math.Float32bits(obj.VelVec.Y)
	in.forceXBits = math.Float32bits(obj.ForceVec.X)
	in.forceYBits = math.Float32bits(obj.ForceVec.Y)
	in.pos24XBits = math.Float32bits(obj.Pos24.X)
	in.pos24YBits = math.Float32bits(obj.Pos24.Y)
	in.zBits = math.Float32bits(obj.ZVal)
	in.field27Bits = math.Float32bits(obj.Field27)
	in.direction1 = int16(obj.Direction1)
	in.direction2 = int16(obj.Direction2)
	in.field38 = obj.Field38
	in.field37 = obj.Field37
	in.field34 = obj.Field34
	in.field33 = obj.Field33
	in.field32 = obj.Field32
	in.massBits = math.Float32bits(obj.Mass)
	in.buffs = obj.Buffs
	in.field62 = obj.Field62[0]
	if health := obj.HealthData; health != nil {
		in.healthPresent = true
		in.healthMax = health.Max
		in.healthField2 = health.Field2
		in.healthCur = health.Cur
	}
	return objectReplayChecksum4E7700(in)
}

// Sub_4E76F0 is the original one-byte no-op callback.
func Sub_4E76F0(*server.Object) {}

// Sub_4E7700 returns the replay checksum for one object.
//
// The pass at 004E76C0 discards the accumulated checksum, but the original
// still performs each object read. Keep this call boundary so those reads and
// their possible fault remain observable to the traversal.
//
//go:noinline
func Sub_4E7700(obj *server.Object) uint32 {
	return objectReplayChecksumNative4E7700(obj)
}

func objectReplayChecksumPassNative4E76C0(first func() *server.Object) *server.Object {
	return objectReplayChecksumPass4E76C0(objectReplayChecksumPassHooks4E76C0[*server.Object]{
		first:    first,
		noop:     Sub_4E76F0,
		checksum: Sub_4E7700,
		next: func(obj *server.Object) *server.Object {
			return obj.Next()
		},
	})
}
