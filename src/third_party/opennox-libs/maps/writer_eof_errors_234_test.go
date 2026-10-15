package maps_test

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/opennox/libs/maps"
)

func TestMapWriterEOFErrors234(t *testing.T) {
	fault := errors.New("private EOF writer I/O fault")
	for _, magic := range []uint32{maps.MagicOld, maps.Magic} {
		for _, n := range []int{0, 1, 7, 8} {
			t.Run(fmt.Sprintf("%08x/close-write-error/len%d", magic, n), func(t *testing.T) {
				var out memory234
				w, err := maps.NewWriter(&out, maps.Header{Magic: magic})
				if err != nil {
					t.Fatal(err)
				}
				if err = w.WriteRawSection(maps.RawSection{Name: "Last", Data: make([]byte, n)}); err != nil {
					t.Fatal(err)
				}
				patches := out.patches
				out.writeErr = fault
				if err = w.Close(); !errors.Is(err, fault) {
					t.Fatalf("Close lost final block error: %v", err)
				}
				if out.patches != patches {
					t.Fatal("checksum patched after failed final block")
				}
			})
		}
		t.Run(fmt.Sprintf("%08x/flush-error", magic), func(t *testing.T) {
			var out memory234
			w, err := maps.NewWriter(&out, maps.Header{Magic: magic})
			if err != nil {
				t.Fatal(err)
			}
			if err = w.WriteRawSection(maps.RawSection{Name: "Last", Data: []byte{1}}); err != nil {
				t.Fatal(err)
			}
			out.writeErr = fault
			if err = w.Flush(); !errors.Is(err, fault) {
				t.Fatalf("Flush lost I/O error: %v", err)
			}
		})
		t.Run(fmt.Sprintf("%08x/invalid-name-after-close", magic), func(t *testing.T) {
			var out memory234
			w, err := maps.NewWriter(&out, maps.Header{Magic: magic, Offs: image.Pt(7, 11)})
			if err != nil {
				t.Fatal(err)
			}
			sections := []maps.RawSection{{Name: "Last", Data: []byte{1, 2, 3}}}
			if err = w.WriteRawSections(sections); err != nil {
				t.Fatal(err)
			}
			if err = w.Close(); err != nil {
				t.Fatal(err)
			}
			before := append([]byte(nil), out.data...)
			if err = w.WriteRawSection(maps.RawSection{Name: strings.Repeat("x", 254)}); err == nil {
				t.Fatal("invalid section name accepted")
			}
			if err = w.Close(); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out.data, before) {
				t.Fatal("failed name validation changed a finalized map")
			}
			requireWireEOF234(t, out.data, magic, sections)
		})
	}
	for _, n := range []int{0, 1, 7, 8} {
		t.Run(fmt.Sprintf("checksum-retry/len%d", n), func(t *testing.T) {
			var out memory234
			w, err := maps.NewWriter(&out, maps.Header{Magic: maps.Magic, Offs: image.Pt(7, 11)})
			if err != nil {
				t.Fatal(err)
			}
			sections := []maps.RawSection{{Name: "Last", Data: make([]byte, n)}}
			if err = w.WriteRawSections(sections); err != nil {
				t.Fatal(err)
			}
			out.patchErr = fault
			if err = w.Close(); !errors.Is(err, fault) {
				t.Fatalf("checksum patch error lost: %v", err)
			}
			size, writes := len(out.data), out.writes
			out.patchErr = nil
			if err = w.Close(); err != nil {
				t.Fatal(err)
			}
			if len(out.data) != size || out.writes != writes {
				t.Fatal("checksum retry wrote another EOF block")
			}
			requireWireEOF234(t, out.data, maps.Magic, sections)
		})
	}
}
