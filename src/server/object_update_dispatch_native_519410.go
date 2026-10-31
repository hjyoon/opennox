package server

import "math"

// Child calls are separate contracts. Words remain unnormalized so the raw
// entry can retain its byte return independently of successful publication.
type ObjectDispatchNativeRuntime519410 struct {
	Friend  func(int32, *Object, uint32) uint32
	Owner   func(*Object, *Object) uint32
	Game    func(uint32) uint32
	Tracks  func(int32, *Object) uint32
	Ray     func([4]uint32, int32, int32, int32) uint32
	Sprite  func(*Object, int32, *Object) uint32
	Complex func(int32, *Object, int32) uint32
	Player  func(*Object, *Object, int32, int32) uint32
	Phantom func(int32, *Object) uint32
	Simple  func(int32, *Object) uint32
	Special func(*Object, *Object) uint32
	Out     func(int32, *Object) uint32
	Shadow  func(int32, *Object) uint32
}

// ObjectDispatchNative519410 shares the original visibility and sender core
// between the public C entry and the ordinary Go adapter. Raw inputs require
// a valid recipient/update/player and an original player index in 0..31.
func ObjectDispatchNative519410(recipient, unit *Object, runtime ObjectDispatchNativeRuntime519410) (byte, bool) {
	update := (*PlayerUpdateData)(recipient.UpdateData)
	index := int32(update.Player.PlayerInd)
	result := byte(unit.ObjFlags)
	if result&0x20 != 0 || uint32(unit.ObjClass)&0x40000000 != 0 {
		return result, false
	}
	bit := uint32(1) << (uint32(index) & 31)
	if unit != recipient && unit.Field35&bit != 0 {
		var add uint32
		if unit.Field36&bit != 0 {
			add = 1
		}
		runtime.Friend(index, unit, add)
		result = byte(^bit)
		unit.Field35 &^= bit
	}
	if unit.Field38&bit == 0 && byte(unit.ObjClass)&1 == 0 {
		return result, false
	}
	if uint32(unit.ObjFlags)&0x800 != 0 && runtime.Owner(unit, recipient) == 0 && runtime.Game(512) == 0 {
		return 0, false
	}
	// The update record is cached before Friend/Owner/Game; its Player link is
	// reloaded here. Copy every ray scalar before Tracks or Ray can mutate it.
	player := update.Player
	ray := [4]uint32{
		math.Float32bits(player.Pos3632Vec.X), math.Float32bits(player.Pos3632Vec.Y),
		math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y),
	}
	word := uint32(unit.ObjClass)
	visible := word&0x20400000 != 0
	if !visible {
		word = runtime.Tracks(index, unit)
		visible = word != 0
		if !visible {
			word = runtime.Ray(ray, 0, 0, 69)
			visible = word != 0
		}
	}
	if !visible {
		if unit.Field37&bit == 0 {
			return byte(word), false
		}
		if byte(unit.ObjClass)&6 != 0 {
			runtime.Out(index, unit)
		} else {
			runtime.Shadow(index, unit)
		}
		pending, seen := unit.Field38|bit, unit.Field37&^bit
		unit.Field38, unit.Field37 = pending, seen
		return byte(seen), false
	}
	if unit.Field37&bit != 0 {
		if byte(unit.Field5)&0x20 != 0 {
			return byte(word), false
		}
	} else if state := unit.Field140[index]; state&0xfff != 0 {
		unit.Field140[index] = state | (state&0xfff)<<16
	}
	var publication uint32
	class := uint32(unit.ObjClass)
	switch {
	case class&0x400000 != 0:
		publication = runtime.Sprite(recipient, index, unit)
	case class&0x200000 != 0 && class&2 != 0:
		publication = runtime.Complex(index, unit, 1)
	case class&0x200000 != 0 && class&4 != 0:
		publication = runtime.Player(recipient, unit, 1, 1)
	case class&0x200000 != 0:
		publication = runtime.Phantom(index, unit)
	case class&0x100000 != 0:
		publication = runtime.Simple(index, unit)
	default:
		unit.Field38 = 0
	}
	result = byte(runtime.Special(recipient, unit))
	if publication != 0 {
		pending, seen := unit.Field38&^bit, unit.Field37|bit
		unit.Field38, unit.Field37 = pending, seen
		result = byte(seen)
	}
	return result, publication != 0
}
