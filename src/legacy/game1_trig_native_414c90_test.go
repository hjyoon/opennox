package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestGAME1Trig414C90ActualEntryFullOriginalBytesAndReturn(t *testing.T) {
	original := trigOracle260(t)
	gold, err := hex.DecodeString(original.Table)
	if err != nil {
		t.Fatal(err)
	}
	constants, err := hex.DecodeString(original.Constants)
	if err != nil {
		t.Fatal(err)
	}
	c := memmap.Slice(0x581450, 7184)[:28]
	sine := memmap.Slice(0x85B3FC, 12260)[:16384]
	arc := memmap.Slice(0x5D4594, 338472)[:32768]
	flag := memmap.PtrUint8(0x5D4594, 371240)
	oldC, oldS, oldA, oldF := bytes.Clone(c), bytes.Clone(sine), bytes.Clone(arc), *flag
	t.Cleanup(func() { copy(c, oldC); copy(sine, oldS); copy(arc, oldA); *flag = oldF })
	copy(c, constants)
	for _, row := range original.Rows {
		t.Run(fmt.Sprintf("%02x_%02x", row.Pattern, row.Ready), func(t *testing.T) {
			for i := range sine {
				sine[i] = row.Pattern
			}
			for i := range arc {
				arc[i] = row.Pattern
			}
			*flag = row.Ready
			beforeS, beforeA := bytes.Clone(sine), bytes.Clone(arc)
			got := trigFixturePublic414C90()
			if got != row.Returned || *flag != row.Final {
				t.Fatalf("return/flag=%02x/%02x want%02x/%02x", got, *flag, row.Returned, row.Final)
			}
			if row.Ready != 0 {
				if !bytes.Equal(sine, beforeS) || !bytes.Equal(arc, beforeA) {
					t.Fatal("ready table modified")
				}
				return
			}
			if !bytes.Equal(sine, gold[:16384]) || !bytes.Equal(arc, gold[16384:]) {
				badS, badA := 0, 0
				for i := 0; i < 4096; i++ {
					if binary.LittleEndian.Uint32(sine[i*4:]) != binary.LittleEndian.Uint32(gold[i*4:]) {
						badS++
					}
				}
				for i := 0; i < 8192; i++ {
					if binary.LittleEndian.Uint32(arc[i*4:]) != binary.LittleEndian.Uint32(gold[16384+i*4:]) {
						badA++
					}
				}
				t.Fatalf("complete original table differs: sine%d acos%d", badS, badA)
			}
		})
	}
}
