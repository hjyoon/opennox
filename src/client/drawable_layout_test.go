package client

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrawableAnimationNativeLayout(t *testing.T) {
	pointerSize := unsafe.Sizeof(uintptr(0))
	want := struct {
		vector, state, monster, equip, player, drawData uintptr
	}{
		vector: 48, state: 148, monster: 772,
		equip: 40, player: 264, drawData: 14524,
	}
	if pointerSize == 8 {
		want = struct {
			vector, state, monster, equip, player, drawData uintptr
		}{
			vector: 88, state: 272, monster: 1416,
			equip: 80, player: 520, drawData: 28608,
		}
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"AnimationVector", unsafe.Sizeof(AnimationVector{}), want.vector},
		{"AnimationStateDrawData", unsafe.Sizeof(AnimationStateDrawData{}), want.state},
		{"MonsterDrawData", unsafe.Sizeof(MonsterDrawData{}), want.monster},
		{"PlayerEquipAnimation", unsafe.Sizeof(PlayerEquipAnimation{}), want.equip},
		{"PlayerAnimation", unsafe.Sizeof(PlayerAnimation{}), want.player},
		{"PlayerDrawData", unsafe.Sizeof(PlayerDrawData{}), want.drawData},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s size = %d, want %d", check.name, check.got, check.want)
		}
	}

	var state AnimationStateDrawData
	stride := uintptr(unsafe.Pointer(&state.Anim[1])) - uintptr(unsafe.Pointer(&state.Anim[0]))
	if stride != unsafe.Sizeof(AnimationVector{}) {
		t.Fatalf("animation state stride = %d, want native vector size %d", stride, unsafe.Sizeof(AnimationVector{}))
	}
}

func TestDrawableDoorUnionPreservesNativeHandles(t *testing.T) {
	if got, want := unsafe.Sizeof(DrawableUnion{}), unsafe.Sizeof(DrawableUnionDoor{}); got != want {
		t.Fatalf("drawable union size = %d, door view = %d", got, want)
	}
	if got, want := unsafe.Alignof(DrawableUnion{}), unsafe.Alignof(uintptr(0)); got != want {
		t.Fatalf("drawable union alignment = %d, want native pointer alignment %d", got, want)
	}
	var (
		union DrawableUnion
		bytes [4]byte
	)
	door := (*DrawableUnionDoor)(unsafe.Pointer(&union))
	door.Field_109 = noxrender.ImageHandle(unsafe.Pointer(&bytes[0]))
	door.Field_110 = noxrender.ImageHandle(unsafe.Pointer(&bytes[1]))
	door.Field_111 = noxrender.ImageHandle(unsafe.Pointer(&bytes[2]))
	door.Field_112 = noxrender.ImageHandle(unsafe.Pointer(&bytes[3]))
	if door.Field_109 != noxrender.ImageHandle(unsafe.Pointer(&bytes[0])) ||
		door.Field_110 != noxrender.ImageHandle(unsafe.Pointer(&bytes[1])) ||
		door.Field_111 != noxrender.ImageHandle(unsafe.Pointer(&bytes[2])) ||
		door.Field_112 != noxrender.ImageHandle(unsafe.Pointer(&bytes[3])) {
		t.Fatal("door image handles did not round-trip at native pointer width")
	}
	runtime.KeepAlive(bytes)
}
