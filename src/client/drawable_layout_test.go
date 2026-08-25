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

func TestDrawableItemUnionPreservesNativeModifierPointers(t *testing.T) {
	var item DrawableUnionItem
	ptrSize := unsafe.Sizeof(uintptr(0))
	if got, want := unsafe.Sizeof(item), unsafe.Sizeof(DrawableUnion{}); got != want {
		t.Fatalf("item union size = %d, drawable union = %d", got, want)
	}
	if got, want := unsafe.Offsetof(item.Field_109), ptrSize; got != want {
		t.Fatalf("item modifier 1 offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(item.Field_112_0), 4*ptrSize; got != want {
		t.Fatalf("item state offset = %d, want %d", got, want)
	}
}

func TestDrawableFXNativeLayout(t *testing.T) {
	var fx DrawableFX
	wantOwner := uintptr(60)
	wantNext := uintptr(64)
	wantPrev := uintptr(68)
	wantGlobalNext := uintptr(72)
	wantGlobalPrev := uintptr(76)
	wantSize := uintptr(80)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantOwner = 64
		wantNext = 72
		wantPrev = 80
		wantGlobalNext = 88
		wantGlobalPrev = 96
		wantSize = 104
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Owner", unsafe.Offsetof(fx.Owner), wantOwner},
		{"Next", unsafe.Offsetof(fx.Next), wantNext},
		{"Prev", unsafe.Offsetof(fx.Prev), wantPrev},
		{"GlobalNext", unsafe.Offsetof(fx.GlobalNext), wantGlobalNext},
		{"GlobalPrev", unsafe.Offsetof(fx.GlobalPrev), wantGlobalPrev},
		{"size", unsafe.Sizeof(fx), wantSize},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("DrawableFX %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDrawableNativeLayout(t *testing.T) {
	var dr Drawable
	checks := []struct {
		name string
		off  uintptr
	}{
		{"Field_0", unsafe.Offsetof(dr.Field_0)},
		{"Field_2", unsafe.Offsetof(dr.Field_2)},
		{"PosVec", unsafe.Offsetof(dr.PosVec)},
		{"Field_5", unsafe.Offsetof(dr.Field_5)},
		{"TeamVal", unsafe.Offsetof(dr.TeamVal)},
		{"Field_8", unsafe.Offsetof(dr.Field_8)},
		{"Shape", unsafe.Offsetof(dr.Shape)},
		{"ZSizeMin", unsafe.Offsetof(dr.ZSizeMin)},
		{"TypeIDVal", unsafe.Offsetof(dr.TypeIDVal)},
		{"LightColor", unsafe.Offsetof(dr.LightColor)},
		{"Field_42", unsafe.Offsetof(dr.Field_42)},
		{"Field_65", unsafe.Offsetof(dr.Field_65)},
		{"Field_74_4", unsafe.Offsetof(dr.Field_74_4)},
		{"DrawFuncPtr", unsafe.Offsetof(dr.DrawFuncPtr)},
		{"DrawData", unsafe.Offsetof(dr.DrawData)},
		{"Field_83", unsafe.Offsetof(dr.Field_83)},
		{"Field_84", unsafe.Offsetof(dr.Field_84)},
		{"Field_85", unsafe.Offsetof(dr.Field_85)},
		{"Field_87", unsafe.Offsetof(dr.Field_87)},
		{"Field_88", unsafe.Offsetof(dr.Field_88)},
		{"Deadline", unsafe.Offsetof(dr.Deadline)},
		{"Field_90", unsafe.Offsetof(dr.Field_90)},
		{"Field_94", unsafe.Offsetof(dr.Field_94)},
		{"Field_95", unsafe.Offsetof(dr.Field_95)},
		{"InClientUpdateList", unsafe.Offsetof(dr.InClientUpdateList)},
		{"Field_97", unsafe.Offsetof(dr.Field_97)},
		{"Field_99", unsafe.Offsetof(dr.Field_99)},
		{"Field_100", unsafe.Offsetof(dr.Field_100)},
		{"Field_102", unsafe.Offsetof(dr.Field_102)},
		{"Field_104", unsafe.Offsetof(dr.Field_104)},
		{"Field_106", unsafe.Offsetof(dr.Field_106)},
		{"Union", unsafe.Offsetof(dr.Union)},
		{"Field_113", unsafe.Offsetof(dr.Field_113)},
		{"Field_114", unsafe.Offsetof(dr.Field_114)},
		{"ClientUpdateFuncPtr", unsafe.Offsetof(dr.ClientUpdateFuncPtr)},
		{"Field_117", unsafe.Offsetof(dr.Field_117)},
		{"Field_124", unsafe.Offsetof(dr.Field_124)},
		{"Field_125", unsafe.Offsetof(dr.Field_125)},
		{"clientHandle", unsafe.Offsetof(dr.clientHandle)},
	}
	for _, check := range checks {
		t.Logf("%s offset = %d", check.name, check.off)
	}
	t.Logf("Drawable size = %d, align = %d", unsafe.Sizeof(dr), unsafe.Alignof(dr))
}
