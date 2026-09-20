package server

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestSkullUpdateNative54F9A0ScansFiresAndCreatesProjectile(t *testing.T) {
	update := &SkullUpdateData{ScanDelay: 91, FireDelay: 92, ProjectileType: 77}
	source := &Object{
		TypeInd:    21,
		ObjFlags:   object.FlagEnabled,
		PosVec:     types.Ptf(100, 200),
		Direction1: 37,
		UpdateData: unsafe.Pointer(update),
	}
	source.Shape.Circle.R = 6
	target := &Object{ObjClass: object.ClassPlayer}
	projectileData := &ProjectileCollideData{Damage: -1, Field4: -2}
	projectile := &Object{SpeedCur: 7.25, CollideData: unsafe.Pointer(projectileData)}

	var gotRect types.Rectf
	var gotType int
	var gotObject, gotOwner *Object
	var gotPosition types.Pointf
	var gotFXPosition types.Pointf
	var gotFXVariant byte
	var gotAudio sound.ID
	deps := skullUpdateNativeDeps54F9A0{
		tickRate: func() uint32 { return 30 },
		eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			gotRect = rect
			if !fn(target) {
				t.Fatal("target callback stopped enumeration")
			}
		},
		isEnemy:  func(gotSource, gotTarget *Object) bool { return gotSource == source && gotTarget == target },
		mapCheck: func(gotTarget, gotSource *Object) bool { return gotSource == source && gotTarget == target },
		inFront:  func(gotSource, gotTarget *Object) bool { return gotSource == source && gotTarget == target },
		typeByName: func(name string) int {
			switch name {
			case "MercArcherArrow":
				return 77
			case "ArrowTrap1":
				return 21
			case "ArrowTrap2":
				return 22
			default:
				t.Fatalf("unexpected type lookup %q", name)
				return 0
			}
		},
		newObject: func(typ int) *Object {
			gotType = typ
			return projectile
		},
		createAt: func(object, owner *Object, position types.Pointf) {
			gotObject, gotOwner, gotPosition = object, owner, position
		},
		balanceFloat: func(name string) float64 {
			if name != "ArrowTrapDamage" {
				t.Fatalf("balance key = %q", name)
			}
			return 13.5
		},
		floatToInt: playerCollideRound4E8460,
		arrowTrapFX: func(position types.Pointf, variant byte) {
			gotFXPosition, gotFXVariant = position, variant
		},
		audioEvent: func(id sound.ID, gotSource *Object) {
			if gotSource != source {
				t.Fatalf("audio source = %p, want %p", gotSource, source)
			}
			gotAudio = id
		},
	}

	if got := skullUpdateNative54F9A0(source, deps); got != 29 {
		t.Fatalf("result = %d, want 29", got)
	}
	if update.Enabled != 1 || update.TargetReady != 1 || update.ScanDelay != 29 || update.FireDelay != 29 {
		t.Fatalf("update state = %+v", update)
	}
	wantRect := types.Rectf{Min: types.Ptf(-250, -150), Max: types.Ptf(450, 550)}
	if gotRect != wantRect {
		t.Fatalf("scan rect = %+v, want %+v", gotRect, wantRect)
	}
	if gotType != 77 || gotObject != projectile || gotOwner != source {
		t.Fatalf("create = type:%d object:%p owner:%p", gotType, gotObject, gotOwner)
	}
	cosine, sine := SinCosDir(byte(source.Direction1))
	wantPosition := types.Ptf(
		float32(10*float64(cosine)+100),
		float32(10*float64(sine)+200),
	)
	if gotPosition != wantPosition {
		t.Fatalf("projectile position = %+v, want %+v", gotPosition, wantPosition)
	}
	wantVelocity := types.Ptf(
		float32(float64(cosine)*7.25),
		float32(float64(sine)*7.25),
	)
	if projectile.VelVec != wantVelocity || projectile.Direction1 != 37 || projectile.Direction2 != 37 {
		t.Fatalf("projectile velocity/directions = %+v/%d/%d", projectile.VelVec, projectile.Direction1, projectile.Direction2)
	}
	if projectileData.Damage != 14 || projectileData.Field4 != 14 {
		t.Fatalf("projectile damage = %+v, want 14/14", projectileData)
	}
	if gotFXPosition != source.PosVec || gotFXVariant != 1 {
		t.Fatalf("arrow trap FX = %+v/%d", gotFXPosition, gotFXVariant)
	}
	if gotAudio != sound.SoundArrowTrapShoot {
		t.Fatalf("audio = %d, want %d", gotAudio, sound.SoundArrowTrapShoot)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want native address above ABI32 range", source)
	}
}

func TestSkullUpdateTargetReachable54FC50FiltersAndKeepsEnumerating(t *testing.T) {
	source := &Object{PosVec: types.Ptf(7, 9)}
	candidates := []*Object{
		{ObjClass: object.ClassObstacle},
		{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDestroyed},
		{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead},
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassMonster},
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassPlayer},
	}
	names := map[*Object]string{
		candidates[0]: "obstacle",
		candidates[1]: "destroyed",
		candidates[2]: "dead",
		candidates[3]: "friendly",
		candidates[4]: "occluded",
		candidates[5]: "behind",
		candidates[6]: "valid",
		candidates[7]: "after",
	}
	var enemyCalls, mapCalls, frontCalls []string
	deps := skullUpdateNativeDeps54F9A0{
		eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			want := types.Rectf{Min: types.Ptf(-343, -341), Max: types.Ptf(357, 359)}
			if rect != want {
				t.Fatalf("rect = %+v, want %+v", rect, want)
			}
			for _, candidate := range candidates {
				if !fn(candidate) {
					t.Fatalf("enumeration stopped at %q", names[candidate])
				}
			}
		},
		isEnemy: func(gotSource, target *Object) bool {
			if gotSource != source {
				t.Fatalf("enemy source = %p", gotSource)
			}
			name := names[target]
			enemyCalls = append(enemyCalls, name)
			return name != "friendly" && name != "after"
		},
		mapCheck: func(target, gotSource *Object) bool {
			if gotSource != source {
				t.Fatalf("map source = %p", gotSource)
			}
			name := names[target]
			mapCalls = append(mapCalls, name)
			return name != "occluded"
		},
		inFront: func(gotSource, target *Object) bool {
			if gotSource != source {
				t.Fatalf("front source = %p", gotSource)
			}
			name := names[target]
			frontCalls = append(frontCalls, name)
			return name != "behind"
		},
	}

	if !skullUpdateTargetReachable54FC50(source, deps) {
		t.Fatal("valid target was not found")
	}
	if got, want := enemyCalls, []string{"friendly", "occluded", "behind", "valid", "after"}; !equalStrings54F9A0(got, want) {
		t.Fatalf("enemy calls = %v, want %v", got, want)
	}
	if got, want := mapCalls, []string{"occluded", "behind", "valid"}; !equalStrings54F9A0(got, want) {
		t.Fatalf("map calls = %v, want %v", got, want)
	}
	if got, want := frontCalls, []string{"behind", "valid"}; !equalStrings54F9A0(got, want) {
		t.Fatalf("front calls = %v, want %v", got, want)
	}
}

func equalStrings54F9A0(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func TestSkullUpdateNative54F9A0DisableAndReenableResetTimers(t *testing.T) {
	update := &SkullUpdateData{ScanDelay: 11, FireDelay: 12, TargetReady: 1, Enabled: 1}
	source := &Object{ObjFlags: object.FlagDead, UpdateData: unsafe.Pointer(update)}
	deps := skullUpdateNativeDeps54F9A0{
		tickRate:   func() uint32 { return 20 },
		eachInRect: func(types.Rectf, func(*Object) bool) {},
	}

	if got := skullUpdateNative54F9A0(source, deps); got != uint32(object.FlagDead) {
		t.Fatalf("disabled result = %#x, want %#x", got, object.FlagDead)
	}
	if update.Enabled != 0 || update.ScanDelay != 11 || update.FireDelay != 12 || update.TargetReady != 1 {
		t.Fatalf("disabled state = %+v", update)
	}

	source.ObjFlags = object.FlagEnabled
	if got := skullUpdateNative54F9A0(source, deps); got != 0 {
		t.Fatalf("reenabled result = %d, want 0", got)
	}
	if update.Enabled != 1 || update.ScanDelay != 19 || update.FireDelay != 0 || update.TargetReady != 0 {
		t.Fatalf("reenabled state = %+v", update)
	}
}

func TestSkullArrowTrapFXPacket5238A0(t *testing.T) {
	packet := skullArrowTrapFXPacket5238A0(types.Ptf(12.9, -3.9), 2)
	if packet[0] != byte(netmsg.MSG_FX_ARROW_TRAP) || packet[5] != 2 {
		t.Fatalf("packet opcode/variant = %#x/%d", packet[0], packet[5])
	}
	if got := int16(binary.LittleEndian.Uint16(packet[1:3])); got != 12 {
		t.Fatalf("packet X = %d, want 12", got)
	}
	if got := int16(binary.LittleEndian.Uint16(packet[3:5])); got != -3 {
		t.Fatalf("packet Y = %d, want -3", got)
	}
}
