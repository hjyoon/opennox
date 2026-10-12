package opennox

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Fixed workloads measure the real factory/default draw/update. They do not
// assert hardware-specific time limits or replace production callbacks.
func BenchmarkParticleSparseSweep4AFD40(b *testing.B) {
	for _, count := range []int{17, 512, 1025} {
		b.Run(fmt.Sprintf("records-%d", count), func(b *testing.B) {
			for _, gap := range []int{0, 16, 256, 1024} {
				b.Run(fmt.Sprintf("gap-%d", gap), func(b *testing.B) {
					b.StopTimer()
					handles.Init()
					b.Cleanup(handles.Release)
					owner := new(partFXes)
					owner.Init(nil)
					b.Cleanup(owner.Free)
					for i := 0; i < count; i++ {
						owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, math.MaxInt32, color.White)
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
		})
	}
}
