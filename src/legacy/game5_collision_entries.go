package legacy

/*
#include "defs.h"
extern void* nox_alloc_hit_2491548;
extern uint32_t dword_5d4594_2490508;
extern uint32_t dword_5d4594_2491544;
extern uint32_t dword_5d4594_2491552;
void sub_548220(nox_object_t*, nox_object_t*);
void nox_xxx_getUnitsInRect_517C10(float4*, void*, void*);
void nox_script_callByIndex_507310(int, void*, void*);
int nox_float2int(float);
static void* game5_word_pointer(uintptr_t p) { return (void*)p; }
static void* game5_pentagram_callback(void) { return (void*)nox_xxx_collidePentagram_4EAB20; }
static void game5_collision_query(nox_object_t* obj) {
 nox_xxx_getUnitsInRect_517C10((float4*)&obj->collide_x1, (void*)sub_548220, obj);
}
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Only the second word has the historical mixed domain: 0..6 are collision
// tokens and larger values are native object addresses. No pointer is narrowed.
type game5CollisionHit struct {
	bucketNext, allNext *game5CollisionHit
	first               *server.Object
	second              uintptr
	normal              types.Pointf
	bucket              uint32
}

var game5Collisions struct {
	buckets [256]*game5CollisionHit
	head    *game5CollisionHit
	owners  map[*game5CollisionHit][2]*server.Object
}

func game5CollisionPeer(word uintptr) *server.Object {
	if word <= 6 {
		return nil
	}
	return (*server.Object)(C.game5_word_pointer(C.uintptr_t(word)))
}

//export nox_xxx_collSysAddCollision_548630
func nox_xxx_collSysAddCollision_548630(raw *nox_object_t, word C.uintptr_t, normal *C.float2) {
	first, second := asObjectS(raw), game5CollisionPeer(uintptr(word))
	hash := first.NetCode
	if second != nil {
		hash += second.NetCode
	}
	bucket := hash % 256
	for it := game5Collisions.buckets[bucket]; it != nil; it = it.bucketNext {
		if it.first == first && it.second == uintptr(word) || uintptr(unsafe.Pointer(it.first)) == uintptr(word) && it.second == uintptr(unsafe.Pointer(first)) {
			return
		}
	}
	hit := (*game5CollisionHit)(alloc.AsClass(C.nox_alloc_hit_2491548).NewObject())
	if hit == nil {
		return
	}
	hit.first, hit.second = first, uintptr(word)
	hit.normal, hit.bucket = *(*types.Pointf)(unsafe.Pointer(normal)), bucket
	hit.bucketNext, hit.allNext = game5Collisions.buckets[bucket], game5Collisions.head
	if game5Collisions.owners == nil {
		game5Collisions.owners = make(map[*game5CollisionHit][2]*server.Object)
	}
	game5Collisions.owners[hit] = [2]*server.Object{first, second}
	game5Collisions.buckets[bucket], game5Collisions.head = hit, hit
}

//export nox_xxx_allocHitArray_5486D0
func nox_xxx_allocHitArray_5486D0() {
	if C.nox_alloc_hit_2491548 == nil {
		C.nox_alloc_hit_2491548 = alloc.NewClass("Hit", unsafe.Sizeof(game5CollisionHit{}), 1024).UPtr()
	}
	game5Collisions.buckets = [256]*game5CollisionHit{}
	game5Collisions.head = nil
	clear(game5Collisions.owners)
	alloc.AsClass(C.nox_alloc_hit_2491548).FreeAllObjects()
	clear(unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 2490520)), 0x400))
	C.dword_5d4594_2491544 = 0
}

//export nox_xxx_collide_548740
func nox_xxx_collide_548740() {
	for hit := game5Collisions.head; hit != nil; hit = hit.allNext {
		first, second := hit.first, game5CollisionPeer(hit.second)
		if hit.second > 6 || hit.second == 0 {
			server.CallObjectCollide(first.Collide, first, second, unsafe.Pointer(&hit.normal))
			if second != nil {
				Nox_xxx_collide_4FDF90(first, second)
			}
		}
		if hit.second == 6 {
			if first.Damage != nil {
				first.CallDamage(nil, nil, 2, object.DamageType(12))
			}
			Nox_xxx_unitHasCollideOrUpdateFn_537610(first)
		} else if second != nil {
			reverse := types.Ptf(-hit.normal.X, -hit.normal.Y)
			server.CallObjectCollide(second.Collide, second, first, unsafe.Pointer(&reverse))
			Nox_xxx_collide_4FDF90(second, first)
			if first.Flags()&8 != 0 {
				Nox_xxx_unitHasCollideOrUpdateFn_537610(second)
			} else if second.Flags()&8 != 0 {
				Nox_xxx_unitHasCollideOrUpdateFn_537610(first)
			}
		}
	}
}

//export sub_548100
func sub_548100(grid *C.int2, raw *nox_object_t) {
	w := GetServer().S().Walls.GetWallAtGrid(image.Pt(int(grid.field_0), int(grid.field_4)))
	if w == nil || w.Flags4&4 == 0 || asObjectS(raw).ObjClass&6 == 0 {
		return
	}
	data := (*server.SecretWall)(w.Data)
	if data == nil || data.State != 1 || data.Flags&2 == 0 {
		return
	}
	data.State = 4
	x := uint32(data.X)
	data.OpenDelay = 0
	y := uint32(data.Y)
	pos := types.Ptf(float32(float64(int32(23*x))+11.5), float32(float64(int32(23*y))+11.5))
	name := GetServer().S().Walls.DefByInd(int(w.Tile1)).OpenSound()
	GetServer().S().Audio.EventPos(sound.ByName(name), pos, 0, 0)
}

//export sub_5481C0
func sub_5481C0(raw *nox_object_t) {
	obj := asObjectS(raw)
	obj.Pos24 = types.Pointf{}
	if obj.Flags()&0x60 != 0 {
		return
	}
	C.game5_collision_query(raw)
	if obj.Flags()&8 == 0 {
		switch obj.Shape.Kind {
		case 2:
			sub_54FEF0_go(raw)
		case 3:
			sub_5504B0_go(raw)
		}
	}
}

//export sub_548220
func sub_548220(candidateRaw, movingRaw *nox_object_t) {
	if sub_548360(movingRaw, candidateRaw) == 0 {
		return
	}
	moving, candidate := asObjectS(movingRaw), asObjectS(candidateRaw)
	switch {
	case moving.ObjClass&0x4000 != 0:
		sub_551AE0_go(movingRaw, candidateRaw, 0)
	case candidate.ObjClass&0x4000 != 0:
		sub_551AE0_go(candidateRaw, movingRaw, 1)
	case moving.ObjClass&0x8000 != 0:
		sub_551C40_go(movingRaw, candidateRaw)
	case candidate.ObjClass&0x8000 != 0:
		sub_551C40_go(candidateRaw, movingRaw)
	case moving.ObjClass&0x80 != 0:
		if candidate.Shape.Kind == 2 {
			sub_5488B0_go(movingRaw, candidateRaw, 0)
		} else if candidate.Shape.Kind == 3 {
			sub_551250_go(movingRaw, candidateRaw, 0)
		}
	case moving.Shape.Kind == 2:
		if candidate.ObjClass&0x80 != 0 {
			sub_5488B0_go(candidateRaw, movingRaw, 1)
		} else if candidate.Shape.Kind == 2 {
			nox_xxx_collisionCheckCircleCircle_550D00_go(movingRaw, candidateRaw)
		} else if candidate.Shape.Kind == 3 {
			sub_54AD50_go(movingRaw, candidateRaw, 0)
		}
	case moving.Shape.Kind == 3:
		if candidate.ObjClass&0x80 != 0 {
			sub_551250_go(candidateRaw, movingRaw, 1)
		} else if candidate.Shape.Kind == 2 {
			sub_54AD50_go(candidateRaw, movingRaw, 1)
		} else if candidate.Shape.Kind == 3 {
			sub_550F80_go(movingRaw, candidateRaw)
		}
	}
}

//export sub_548360
func sub_548360(firstRaw, secondRaw *nox_object_t) C.int {
	first, second := asObjectS(firstRaw), asObjectS(secondRaw)
	if C.dword_5d4594_2490508 == 0 {
		C.dword_5d4594_2490508 = C.uint32_t(GetServer().S().Types.IndByID("Trigger"))
		*memmap.PtrUint32(0x5D4594, 2490512) = uint32(GetServer().S().Types.IndByID("BlackPowder"))
		*memmap.PtrUint32(0x5D4594, 2490516) = uint32(GetServer().S().Types.IndByID("TelekinesisHand"))
	}
	telekinesis := memmap.Uint32(0x5D4594, 2490516)
	fc, sc := first.Collide, second.Collide
	if fc == C.game5_pentagram_callback() && uint32(second.TypeInd) == telekinesis || sc == C.game5_pentagram_callback() && uint32(first.TypeInd) == telekinesis {
		return 0
	}
	fclass, sclass := first.ObjClass, second.ObjClass
	if fclass&0x80 != 0 && second.Flags()&9 != 0 || sclass&0x80 != 0 && first.Flags()&9 != 0 {
		return 0
	}
	ff, sf := first.Flags(), second.Flags()
	if ff&0x60 != 0 || sf&0x60 != 0 || first == second || fc == nil || sc == nil {
		return 0
	}
	if fclass&2 != 0 && ff&0x4000 != 0 && sclass&0xCC00 != 0 {
		return sub_5485B0_go(firstRaw, secondRaw)
	}
	if sclass&2 != 0 && sf&0x4000 != 0 && fclass&0xCC00 != 0 {
		return sub_5485B0_go(secondRaw, firstRaw)
	}
	powder := memmap.Uint32(0x5D4594, 2490512)
	if fclass&0x2000 != 0 && uint32(second.TypeInd) == powder || sclass&0x2000 != 0 && uint32(first.TypeInd) == powder {
		return 1
	}
	trigger := uint32(C.dword_5d4594_2490508)
	if ff&8 != 0 && sf&8 != 0 || fclass&0x80 != 0 && sf&8 != 0 || sclass&0x80 != 0 && ff&8 != 0 ||
		uint32(first.TypeInd) != trigger && uint32(second.TypeInd) != trigger && (ff&0x11 != 0 && sf&0x24000 != 0 || sf&0x11 != 0 && ff&0x24000 != 0) ||
		fclass&sclass&0xC080 != 0 || ff&0x400 != 0 && server.UnitsHaveSameTeam4EC520(first, second) {
		return 0
	}
	return 1
}

type game5DoorQueueNode struct {
	data *server.DoorUpdateData
	next *game5DoorQueueNode
}

var game5DoorQueueHead *game5DoorQueueNode

func game5EnqueueDoor(data *server.DoorUpdateData) {
	if data == nil || data.Queued != 0 {
		return
	}
	game5DoorQueueHead = &game5DoorQueueNode{data: data, next: game5DoorQueueHead}
	data.Queued = 1
	C.dword_5d4594_2491552 = 1
}

//export sub_548830
func sub_548830(word C.uintptr_t) {
	game5EnqueueDoor((*server.DoorUpdateData)(C.game5_word_pointer(word)))
}

//export sub_548860
func sub_548860(raw *nox_object_t, delta C.short) {
	data := asObjectS(raw).UpdateDataDoor()
	direction := int32(data.FractionalDir) + int32(delta)
	data.FractionalDir = int16((direction%256 + 256) % 256)
	game5EnqueueDoor(data)
}

//export sub_548B60
func sub_548B60() {
	node := game5DoorQueueHead
	game5DoorQueueHead = nil
	for ; node != nil; node = node.next {
		data := node.data
		speed := float32(float32(data.AngularVelocity+data.AngularVelocity) + 0.5)
		step := int32(C.nox_float2int(C.float(speed)))
		if step < 0 {
			step = -step
		}
		if step > 4 {
			step = 4
		}
		negativeLimit := data.TargetDirection - 12
		if negativeLimit < 0 {
			negativeLimit += 32
		}
		if data.AngularVelocity > float32(0.0099999998) && data.CurrentDirection != (data.TargetDirection+12)%32 {
			direction := int32(data.FractionalDir) + step
			if direction >= 256 {
				direction -= 256
			}
			data.FractionalDir = int16(direction)
		} else if data.AngularVelocity < -float32(0.0099999998) && data.CurrentDirection != negativeLimit {
			direction := int32(data.FractionalDir) - step
			if direction < 0 {
				direction += 256
			}
			data.FractionalDir = int16(direction)
		}
		direction := 32 * int32(data.FractionalDir) / 256
		data.CurrentDirection = (direction%32 + 32) % 32
		data.Queued, data.AngularVelocity = 0, 0
	}
	C.dword_5d4594_2491552 = 0
}

//export nox_xxx_script_forcedialog_548CD0
func nox_xxx_script_forcedialog_548CD0(playerRaw, monsterRaw *nox_object_t) {
	player, monster := asObjectS(playerRaw), asObjectS(monsterRaw)
	if player == nil || monster == nil || player.ObjClass&4 == 0 || monster.ObjClass&2 == 0 || player.Flags()&0x8020 != 0 {
		return
	}
	data := monster.UpdateDataMonster()
	if monster.Field5&0x10 != 0 && data.DialogStartFunc != -1 && data.DialogEndFunc != -1 {
		C.nox_script_callByIndex_507310(C.int(data.DialogStartFunc), unsafe.Pointer(player), unsafe.Pointer(monster))
	}
}

//export sub_54E730
func sub_54E730(firstRaw, secondRaw *nox_object_t) C.int {
	first, second := asObjectS(firstRaw), asObjectS(secondRaw)
	if second.ObjClass&1 != 0 || first.Flags()&0x20 != 0 {
		return 0
	}
	ff, sf := first.Flags(), second.Flags()
	if sf&0x60 != 0 || first.Collide == nil || second.Collide == nil {
		return 0
	}
	if sf&0x80 != 0 {
		return 1
	}
	owner := first.ObjOwner
	if ff&0x11 != 0 && sf&0x4000 != 0 || sf&0x11 != 0 && ff&0x4000 != 0 ||
		(ff|sf)&0x400 != 0 && server.UnitsHaveSameTeam4EC520(second, first) ||
		owner != nil && first.ObjClass&1 != 0 && first.ObjSubClass&2 == 0 && owner.ObjClass&2 != 0 && second.ObjClass&2 != 0 &&
			(!owner.IsEnemyTo(second) || server.UnitsHaveSameTeam4EC520(second, owner)) {
		return 0
	}
	return 1
}

//export sub_54E6F0
func sub_54E6F0(first, second *nox_object_t) C.int {
	if sub_54E730(second, first) == 0 {
		return 0
	}
	return C.int(bool2int(!server.UnitsHaveSameTeam4EC520(asObjectS(first), asObjectS(second)) || noxflags.HasGamePlay(1)))
}

//export sub_550480
func sub_550480(raw *nox_object_t) C.int {
	id := memmap.PtrUint32(0x5D4594, 2491788)
	if *id == 0 {
		*id = uint32(GetServer().S().Types.IndByID("GameBall"))
	}
	return C.int(bool2int(uint32(asObjectS(raw).TypeInd) == *id))
}

//export sub_551BF0
func sub_551BF0() {
	for i, name := range [...]string{"SmallFist", "MediumFist", "LargeFist", "Meteor"} {
		*memmap.PtrUint32(0x5D4594, 2491792+4*uintptr(i)) = uint32(GetServer().S().Types.IndByID(name))
	}
	*memmap.PtrUint32(0x5D4594, 2491808) = 1
}
