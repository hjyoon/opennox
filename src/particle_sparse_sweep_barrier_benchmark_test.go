package opennox

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// An unknown update function delegates to the real default implementation.
// This measures the conservative arbitrary-callback path, not a stock type-1
// emitter or a replacement oracle for the default update result.
func BenchmarkParticleSparseBarrier4AFD40(b *testing.B) {
	for _, gap := range []int{256, 1024} {
		b.Run(fmt.Sprintf("records-1025/gap-%d", gap), func(b *testing.B) {
			b.StopTimer()
			handles.Init()
			b.Cleanup(handles.Release)
			owner := new(partFXes)
			owner.Init(nil)
			b.Cleanup(owner.Free)
			for i := 0; i < 1025; i++ {
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, math.MaxInt32, color.White)
				p.updateFunc = func(p *particleFx) bool { return partfxUpdateDef(p) }
				if reflect.ValueOf(p.updateFunc).Pointer() == reflect.ValueOf(partfxUpdateDef).Pointer() {
					b.Fatal("callback fixture did not use an unknown function entry")
				}
				particleSparseGap4AFD40(gap)
			}
			b.ReportAllocs()
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				owner.Draw()
			}
		})
	}
}
