package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObeliskCreateNative54CA10(t *testing.T) {
	update := &ObeliskUpdateData{Mana: -7}
	obj := &Object{UpdateData: unsafe.Pointer(update), Field38: 123}

	ObeliskCreateNative54CA10(obj)

	if update.Mana != 50 || obj.Field38 != math.MaxUint32 {
		t.Fatalf("obelisk create = mana %d sync %#x, want 50/%#x", update.Mana, obj.Field38, uint32(math.MaxUint32))
	}
}

func TestAnimCreateNative54CA50(t *testing.T) {
	obj := &Object{ObjClass: object.ClassImmobile, Field5: 0x10, Field38: 123}
	for i := range obj.Field140 {
		obj.Field140[i] = 0x12345067
	}

	AnimCreateNative54CA50(obj)

	if obj.Field5 != 0x12 || obj.Field38 != math.MaxUint32 {
		t.Fatalf("anim create = status %#x sync %#x", obj.Field5, obj.Field38)
	}
	wantSyncField := uint32(0x12345067)&0xfffff000 | 0x80000
	for i, got := range obj.Field140 {
		if got != wantSyncField {
			t.Fatalf("sync field %d = %#x, want %#x", i, got, wantSyncField)
		}
	}
}

func TestTriggerCreateNative54CA60(t *testing.T) {
	update := &TriggerUpdateData{
		Flags:         0x11223344,
		TeamExclude:   0xaa,
		SoundActivate: 0x55667788,
		Colors:        [6]uint8{1, 2, 3, 4, 5, 6},
	}
	obj := &Object{UpdateData: unsafe.Pointer(update)}

	TriggerCreateNative54CA60(obj)

	if update.Colors != [6]uint8{90, 90, 90, 10, 10, 10} {
		t.Fatalf("trigger colors = %v", update.Colors)
	}
	if update.Flags != 0x11223344 || update.TeamExclude != 0xaa || update.SoundActivate != 0x55667788 {
		t.Fatalf("neighboring trigger fields changed: %+v", update)
	}
}

func TestRewardMarkerCreateNative54CAC0(t *testing.T) {
	data := &RewardMarkerInitData{
		CategoryMask: 0xaabbccdd,
		RewardFlags:  0x11,
		Field208:     0x22334455,
		ChanceMode:   0x66778899,
		Field216:     0xaabbccdd,
	}
	obj := &Object{InitData: unsafe.Pointer(data)}

	RewardMarkerCreateNative54CAC0(obj)

	if data.CategoryMask != 255 || data.ChanceMode != 0 {
		t.Fatalf("reward marker create = mask %#x chance %#x", data.CategoryMask, data.ChanceMode)
	}
	if data.RewardFlags != 0x11 || data.Field208 != 0x22334455 || data.Field216 != 0xaabbccdd {
		t.Fatalf("neighboring reward fields changed: %+v", data)
	}
}
