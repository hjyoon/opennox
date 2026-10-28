package server

import (
	"encoding/binary"
	"math"
	"math/big"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// SentryGlobeRayContext511020 is a transient native stack context, not a
// serialized update record. Both pointers remain native-width.
type SentryGlobeRayContext511020 struct {
	Source *Object
	Ray    *types.Rectf
}

var _ = [1]struct{}{}[2*unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(SentryGlobeRayContext511020{})]

type SentryGlobeListRuntime510E60 struct {
	Head    func() *Object
	SetHead func(*Object)
}

type SentryGlobeRawRuntime510E60 struct {
	List        SentryGlobeListRuntime510E60
	TraceRay    func(*types.Rectf, *types.Pointf, uint32) int32
	EachInRect  func(*types.Rectf, *SentryGlobeRayContext511020)
	GameFlags   func(uint32) uint32
	PointOnLine func(*types.Rectf, *types.Pointf, *types.Pointf) int32
	FindParent  func(*Object) *Object
	Damage      func(*Object, *Object, *Object, int, object.DamageType)
	Audio       func(uint32, *Object)
	Player      func(int32) *Player
	Packet      func(int32, int32, []byte) int32
}

func sentryGlobeInsertRaw510E60(source *Object, list SentryGlobeListRuntime510E60) {
	source.Field125 = nil
	source.InvNextItem = list.Head()
	if previous := list.Head(); previous != nil {
		previous.Field125 = source
	}
	list.SetHead(source)
	source.ObjFlags |= object.FlagMarked
}

// SentryGlobeUnlinkRaw510FD0 retains the required pointer and returned object.
// Links are deliberately left in the removed object, as in the original.
func SentryGlobeUnlinkRaw510FD0(source *Object, list SentryGlobeListRuntime510E60) *Object {
	if source.ObjFlags.Has(object.FlagMarked) {
		if previous := source.Field125; previous != nil {
			previous.InvNextItem = source.InvNextItem
		} else {
			list.SetHead(source.InvNextItem)
		}
		if next := source.InvNextItem; next != nil {
			next.Field125 = source.Field125
		}
	}
	source.ObjFlags &^= object.FlagMarked
	return source
}

func sentryGlobeTrig510E60(angle float32, cosine bool) float64 {
	v := float64(angle)
	if math.IsNaN(v) {
		return v
	}
	if math.IsInf(v, 0) {
		return math.Float64frombits(0xfff8000000000000)
	}
	// FSIN/FCOS set C2 and leave ST0 unchanged outside their finite range.
	if math.Abs(v) >= 0x1p63 {
		return v
	}
	// The x87 reduction uses a 66-bit approximation of pi. Ordinary library
	// reduction diverges substantially for large, still in-range float32 angles.
	if math.Abs(v) <= math.Pi/4 {
		if cosine {
			return math.Cos(v)
		}
		return math.Sin(v)
	}
	bits := math.Float32bits(angle) & 0x7fffffff
	mantissa := uint64(bits&0x7fffff | 0x800000)
	exponent := int(bits>>23) - 127 - 23
	numerator := new(big.Int).Lsh(new(big.Int).SetUint64(mantissa), uint(exponent+65))
	pi66, ok := new(big.Int).SetString("3243f6a8885a308d3", 16)
	if !ok {
		panic("invalid x87 pi constant")
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, pi66, remainder)
	if new(big.Int).Lsh(new(big.Int).Set(remainder), 1).Cmp(pi66) > 0 {
		quotient.Add(quotient, big.NewInt(1))
		remainder.Sub(remainder, pi66)
	}
	r, _ := new(big.Float).SetInt(remainder).Float64()
	r = math.Ldexp(r, -65)
	quadrant := quotient.Uint64() & 3
	if cosine {
		quadrant = (quadrant + 1) & 3
	}
	var result float64
	switch quadrant {
	case 0:
		result = math.Sin(r)
	case 1:
		result = math.Cos(r)
	case 2:
		result = -math.Sin(r)
	case 3:
		result = -math.Cos(r)
	}
	if angle < 0 && !cosine {
		result = -result
	}
	return result
}

func sentryGlobeBounds510E60(ray types.Rectf) types.Rectf {
	result := ray
	// The original C0 tests keep the less-than branch for unordered operands.
	if ray.Min.X >= ray.Max.X {
		result.Min.X, result.Max.X = ray.Max.X, ray.Min.X
	}
	if ray.Min.Y >= ray.Max.Y {
		result.Min.Y, result.Max.Y = ray.Max.Y, ray.Min.Y
	}
	return result
}

// SentryGlobeUpdateRaw510E60 shares one native list with unlink/reset/send.
// PC53/ToZero arithmetic and all binary32 stores are local to this routine.
func SentryGlobeUpdateRaw510E60(source *Object, h SentryGlobeRawRuntime510E60) {
	flags := source.ObjFlags
	data := source.UpdateDataSentry()
	if !flags.Has(object.FlagMarked) {
		sentryGlobeInsertRaw510E60(source, h.List)
	}
	if source.ObjFlags.Has(object.FlagDestroyed) {
		SentryGlobeUnlinkRaw510FD0(source, h.List)
	}
	if !source.ObjFlags.Has(object.FlagEnabled) {
		data.Field0 = data.Field4
		return
	}
	ray := types.Rectf{Min: source.PosVec}
	x := objectForceMul53_52DF80(sentryGlobeTrig510E60(math.Float32frombits(data.Field0), true), 600)
	ray.Max.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(source.PosVec.X)))
	y := objectForceMul53_52DF80(sentryGlobeTrig510E60(math.Float32frombits(data.Field0), false), 600)
	ray.Max.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(source.PosVec.Y)))
	var hit types.Pointf
	if h.TraceRay(&ray, &hit, 5) == 0 {
		ray.Max = hit
	}
	source.Pos39 = ray.Max
	// Trace can change the record and object. Retain the entry record pointer
	// while reloading both of its angle words after that service.
	data.Field0 = math.Float32bits(experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(math.Float32frombits(data.Field8)), objectForceMemory32_52DF80(math.Float32frombits(data.Field0)))))
	bounds := sentryGlobeBounds510E60(ray)
	context := SentryGlobeRayContext511020{Source: source, Ray: &ray}
	h.EachInRect(&bounds, &context)
}

// SentryGlobeHitRaw511020 preserves live context/position reads, lazy Quest
// lookup, ordered radius comparison and owner/damage/audio service order.
func SentryGlobeHitRaw511020(candidate *Object, context *SentryGlobeRayContext511020, h SentryGlobeRawRuntime510E60) {
	flags := candidate.ObjFlags
	if flags.HasAny(object.FlagBelow | object.FlagNoCollide) {
		return
	}
	if flags.Has(object.FlagShort) {
		if h.GameFlags(4096) == 0 || !candidate.ObjClass.Has(object.ClassMonster) || candidate.ObjFlags.Has(object.FlagDead) {
			return
		}
	}
	if candidate.HealthData == nil {
		return
	}
	var point types.Pointf
	if h.PointOnLine(context.Ray, &candidate.PosVec, &point) == 0 {
		return
	}
	dx := objectForceSub53_52DF80(float64(candidate.PosVec.X), objectForceMemory32_52DF80(point.X))
	dy := objectForceSub53_52DF80(float64(candidate.PosVec.Y), objectForceMemory32_52DF80(point.Y))
	distance := objectForceAdd53_52DF80(objectForceMul53_52DF80(dy, dy), objectForceMul53_52DF80(dx, dx))
	radius := float64(candidate.Shape.Circle.R)
	if !(objectForceMul53_52DF80(radius, radius) > distance) {
		return
	}
	source := context.Source
	parent := h.FindParent(source)
	h.Damage(candidate, parent, source, 500, object.DamageZapRay)
	h.Audio(298, candidate)
}

func sentryGlobeRoundRaw419A30(value float32) uint32 {
	if !(value >= 0) {
		return 0
	}
	sum := objectForceAdd53_52DF80(float64(value), 8388608)
	return math.Float32bits(experienceSpill32TowardZero(sum)) & 0x7fffff
}

func sentryGlobePacketBytes511250(source *Object) [9]byte {
	var packet [9]byte
	packet[0] = byte(netmsg.MSG_FX_SENTRY_RAY)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(sentryGlobeRoundRaw419A30(source.PosVec.X)))
	binary.LittleEndian.PutUint16(packet[3:5], uint16(sentryGlobeRoundRaw419A30(source.PosVec.Y)))
	binary.LittleEndian.PutUint16(packet[5:7], uint16(sentryGlobeRoundRaw419A30(source.Pos39.X)))
	binary.LittleEndian.PutUint16(packet[7:9], uint16(sentryGlobeRoundRaw419A30(source.Pos39.Y)))
	return packet
}

func SentryGlobePacketRaw511250(index int32, source *Object, send func(int32, int32, []byte) int32) int32 {
	packet := sentryGlobePacketBytes511250(source)
	return send(index, 1, packet[:])
}

func SentryGlobeSendRaw511100(index int32, h SentryGlobeRawRuntime510E60) {
	player := h.Player(index)
	width, height := float64(player.Field10), float64(player.Field12)
	head := h.List.Head()
	view := types.Rectf{
		Min: types.Ptf(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(player.Pos3632Vec.X), width)), experienceSpill32TowardZero(objectForceSub53_52DF80(float64(player.Pos3632Vec.Y), height))),
		Max: types.Ptf(experienceSpill32TowardZero(objectForceAdd53_52DF80(width, objectForceMemory32_52DF80(player.Pos3632Vec.X))), experienceSpill32TowardZero(objectForceAdd53_52DF80(height, objectForceMemory32_52DF80(player.Pos3632Vec.Y)))),
	}
	for source := head; source != nil; source = source.InvNextItem {
		if !source.ObjFlags.Has(object.FlagEnabled) {
			data := source.UpdateDataSentry()
			data.Field0 = data.Field4
			continue
		}
		beam := sentryGlobeBounds510E60(types.Rectf{Min: source.PosVec, Max: source.Pos39})
		if !(view.Min.X >= beam.Max.X) && view.Max.X > beam.Min.X && !(view.Min.Y >= beam.Max.Y) && view.Max.Y > beam.Min.Y {
			SentryGlobePacketRaw511250(index, source, h.Packet)
		}
	}
}

func SentryGlobeResetRaw510E50(list SentryGlobeListRuntime510E60) { list.SetHead(nil) }
