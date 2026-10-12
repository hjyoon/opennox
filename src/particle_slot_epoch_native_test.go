package opennox

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Native consumers retain opaque keys, not Go object addresses. Replacing
// the owner's active map must not resurrect an old key for a new generation.
func TestParticleSlotEpochNativeResetInvalidatesOldHandle(t *testing.T) {
	handles.Init()
	for _, kind := range []string{"Free", "Init"} {
		t.Run(kind, func(t *testing.T) {
			var owner partFXes
			owner.Init(nil)
			defer owner.Free()
			p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(3, 5), 1, 4, color.White)
			assertParticleRegistration4B01B0(t, &owner, p)
			old := p.chnd
			dr, free := alloc.New(client.Drawable{})
			defer free()
			dr.Field_124 = old // explicit stale native consumer input
			before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))))
			if kind == "Free" {
				owner.Free()
			} else {
				owner.Init(nil)
			}
			fresh := owner.nox_xxx_particleFxNew_4AF990(image.Pt(7, 11), 2, 9, color.White)
			assertParticleRegistration4B01B0(t, &owner, fresh)
			if fresh.chnd == old || owner.asParticlefx(old) != nil || owner.asParticlefx(dr.Field_124) != nil {
				t.Fatal("owner reset aliased an old native handle to a newly allocated effect")
			}
			if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))) {
				t.Fatal("allocator changed the stale native consumer record")
			}
		})
	}
}

func TestParticleSlotEpochNativeInitClearsRawRecord(t *testing.T) {
	handles.Init()
	var owner partFXes
	owner.Init(nil)
	defer owner.Free()
	p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(3, 5), 1, 4, color.White)
	p.field_68, p.field_72, p.field_76 = 17, 19, 23
	p.free()
	owner.Init(nil)
	raw := owner.newParticlefx()
	assertParticleRegistration4B01B0(t, &owner, raw)
	if raw.pos != (image.Point{}) || raw.z != 0 || raw.ticksLeft != 0 || raw.flags != 0 || raw.field_68 != 0 || raw.field_72 != 0 || raw.field_76 != 0 || raw.drawFunc != nil || raw.updateFunc != nil {
		t.Fatal("new arena generation retained old raw record contents")
	}
}
