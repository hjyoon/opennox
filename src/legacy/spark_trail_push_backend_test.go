package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestProjectileTrailNativeDirectionTable53AEC0(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/spark_trail_push_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Table struct {
			Words []uint32 `json:"words"`
		} `json:"direction_table"`
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Table.Words) != 512 {
		t.Fatal("original direction table length")
	}
	stock, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	const off = 0x5b6658 - 0x587000
	want := make([]byte, 2048)
	for n, v := range fixture.Table.Words {
		binary.LittleEndian.PutUint32(want[4*n:], v)
	}
	if !bytes.Equal(stock[off:off+len(want)], want) {
		t.Fatal("stock direction table differs from original")
	}
	// Standalone package tests omit application blob loading. Restore all bytes.
	actual := unsafe.Slice((*byte)(memmap.Ptr(0x5b6658)), len(want))
	before := append([]byte(nil), actual...)
	t.Cleanup(func() { copy(actual, before) })
	copy(actual, stock[off:off+len(want)])
	read := sparkTrailPushRawRuntime().Trail.DirectionFloat
	for n, v := range fixture.Table.Words {
		cos, sin := server.SinCosDir(byte(n / 2))
		public := cos
		if n%2 != 0 {
			public = sin
		}
		if math.Float32bits(read(int32(n))) != v || math.Float32bits(public) != v {
			t.Fatalf("native/public direction word %d", n)
		}
	}
}

func TestProjectileTrailNativeWideLogicRandom53AEC0(t *testing.T) {
	for _, bounds := range [][2]float32{{-2, 2}, {-4, 4}} {
		world, lookups := commonRandomFloatFixture416030(t, 0)
		h := sparkTrailPushRawRuntime()
		if *lookups != 0 {
			t.Fatal("runtime construction eagerly resolved world")
		}
		table := prand.New(0)
		wide := 0
		for n := 0; n < 4096; n++ {
			want := commonRandomFloatReference416030(bounds[0], bounds[1], table.Int(0, 0x7fff))
			got := h.Trail.RandomFloat(bounds[0], bounds[1])
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatalf("retained original RNG at %d", n)
			}
			if float64(float32(got)) != got {
				wide++
			}
			if world.Rand.Logic.Index() != (n+1)%4096 || world.Rand.Other.Index() != 93 || *lookups != n+1 {
				t.Fatal("logic/other stream or service order")
			}
		}
		if wide == 0 {
			t.Fatal("wide RNG cases missing")
		}
	}
}
