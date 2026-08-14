package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestHarpoonCollide4EB6A0NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	type layoutCase struct {
		name string
		got  uintptr
		w32  uintptr
		w64  uintptr
	}
	cases := []layoutCase{
		{"collide size", unsafe.Sizeof(HarpoonCollideData{}), 8, 16},
		{"collide field zero", unsafe.Offsetof(HarpoonCollideData{}.Field0), 0, 0},
		{"collide owner", unsafe.Offsetof(HarpoonCollideData{}.Owner), 4, 8},
		{"player harpoon target", unsafe.Offsetof(PlayerUpdateData{}.HarpoonTarg), 132, 152},
		{"player harpoon bolt", unsafe.Offsetof(PlayerUpdateData{}.HarpoonBolt), 136, 160},
		{"player harpoon field 35", unsafe.Offsetof(PlayerUpdateData{}.Harpoon35), 140, 168},
		{"player harpoon target X", unsafe.Offsetof(PlayerUpdateData{}.HarpoonTargX), 144, 172},
		{"player harpoon target Y", unsafe.Offsetof(PlayerUpdateData{}.HarpoonTargY), 148, 176},
		{"player harpoon frame", unsafe.Offsetof(PlayerUpdateData{}.HarpoonFrame), 152, 180},
	}
	for _, tc := range cases {
		want := tc.w64
		if ptrSize == 4 {
			want = tc.w32
		}
		if tc.got != want {
			t.Errorf("%s = %d, want %d on pointer size %d", tc.name, tc.got, want, ptrSize)
		}
	}
}

func TestHarpoonCollideNative4EB6A0UsesCachedNativePlayerData(t *testing.T) {
	oldData := &PlayerUpdateData{}
	replacement := &PlayerUpdateData{}
	owner := &Object{UpdateData: unsafe.Pointer(oldData)}
	liveOwner := &Object{}
	source := &Object{ObjOwner: owner, ObjFlags: object.FlagShadow}
	target := &Object{PosVec: types.Pointf{X: 12.5, Y: -4.25}}

	var relationOwner, audioOwner *Object
	harpoonCollideNative4EB6A0(source, target, nil, harpoonCollideNativeDeps4EB6A0{
		loadDamage:  func() int32 { return 17 },
		loadBalance: func() float32 { t.Fatal("balance loaded with populated cache"); return 0 },
		floatToInt:  harpoonRoundFloat32ToInt32_4EB6A0,
		storeDamage: func(int32) { t.Fatal("damage cache stored") },
		damageMap: func(int32, int32, int32, object.DamageType, *Object) {
			t.Fatal("map damaged on target path")
		},
		disableAbility: func(*Object, Ability) { t.Fatal("ability disabled on attach") },
		delayedDelete:  func(*Object) { t.Fatal("source deleted on attach") },
		markRelation: func(gotOwner, gotTarget *Object) {
			relationOwner = gotOwner
			if gotTarget != target {
				t.Fatalf("relation target = %p, want %p", gotTarget, target)
			}
		},
		findParentPlayer: func(got *Object) *Object {
			if got != source {
				t.Fatalf("parent source = %p, want %p", got, source)
			}
			return owner
		},
		targetDamage: func(gotTarget, parent, gotSource *Object, damage int32, typ object.DamageType) int32 {
			if gotTarget != target || parent != owner || gotSource != source || damage != 17 || typ != object.DamageImpact {
				t.Fatalf("damage args = (%p,%p,%p,%d,%d)", gotTarget, parent, gotSource, damage, typ)
			}
			owner.UpdateData = unsafe.Pointer(replacement)
			target.PosVec = types.Pointf{X: 19.5, Y: -8.75}
			source.ObjFlags = object.FlagDead
			source.ObjOwner = liveOwner
			return 0x100
		},
		isEnemy:      func(gotOwner, gotTarget *Object) bool { return gotOwner == owner && gotTarget == target },
		gameplayFlag: func(uint32) bool { t.Fatal("gameplay fallback called for enemy"); return false },
		defaultSound: func(*Object, *Object) { t.Fatal("default sound called on attach") },
		frame:        func() uint32 { return 9876 },
		audio: func(id uint32, gotOwner *Object) {
			if id != harpoonReelSound4EB6A0 {
				t.Fatalf("audio ID = %d, want %d", id, harpoonReelSound4EB6A0)
			}
			audioOwner = gotOwner
		},
	})

	if oldData.HarpoonTarg != target || oldData.HarpoonTargX != 19.5 || oldData.HarpoonTargY != -8.75 || oldData.HarpoonFrame != 9876 {
		t.Fatalf("cached player data = %+v", oldData)
	}
	if replacement.HarpoonTarg != nil {
		t.Fatalf("replacement player data was modified: %+v", replacement)
	}
	if source.ObjFlags != object.Flags(uint32(object.FlagDead)|harpoonNoCollideFlag4EB6A0) {
		t.Fatalf("source flags = %#x", source.ObjFlags)
	}
	if relationOwner != liveOwner || audioOwner != owner {
		t.Fatalf("relation/audio owners = (%p,%p), want (%p,%p)", relationOwner, audioOwner, liveOwner, owner)
	}
}

func TestHarpoonRoundFloat32ToInt32_4EB6A0(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {3.5, 4},
		{-0.5, 0}, {-1.5, -2}, {-2.5, -2},
		{math.Float32frombits(0x4effffff), 2147483520},
		{math.Float32frombits(0x4f000000), math.MinInt32},
		{math.Float32frombits(0xcf000000), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, tc := range tests {
		if got := harpoonRoundFloat32ToInt32_4EB6A0(tc.value); got != tc.want {
			t.Errorf("round(%08x) = %d, want %d", math.Float32bits(tc.value), got, tc.want)
		}
	}
	if got := math.Float32bits(harpoonGridInverse4EB6A0); got != 0x3d321643 {
		t.Fatalf("grid inverse bits = %#x, want 0x3d321643", got)
	}
}
