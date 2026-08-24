package mainmenu

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	gameDataVA     = uint32(0x587000)
	eyeTableOff    = 168832
	eyeRecordSize  = 48
	eyeTableSHA256 = "3cf911696d89f78bc2cc30205dab4bb5ac91e3d4f195eb60c80a2ba75edbe882"
)

func TestEyeSpecsMatchGAMEEXEData(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "common", "memmap", "nox", "blobdata", "blob_587000.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	table := data[eyeTableOff : eyeTableOff+len(eyeSpecs)*eyeRecordSize]
	if got := fmt.Sprintf("%x", sha256.Sum256(table)); got != eyeTableSHA256 {
		t.Fatalf("GAME.EXE menu table digest = %s, want %s", got, eyeTableSHA256)
	}

	u32 := func(rec []byte, off int) uint32 {
		return binary.LittleEndian.Uint32(rec[off : off+4])
	}
	for i, want := range eyeSpecs {
		rec := table[i*eyeRecordSize : (i+1)*eyeRecordSize]
		nameVA := u32(rec, 0)
		nameOff := int(nameVA - gameDataVA)
		nameEnd := bytes.IndexByte(data[nameOff:], 0)
		if nameEnd < 0 {
			t.Fatalf("record %d name is not terminated", i)
		}
		got := EyeSpec{
			Name:                 string(data[nameOff : nameOff+nameEnd]),
			X:                    int(u32(rec, 8)),
			Y:                    int(u32(rec, 12)),
			Hidden:               u32(rec, 16) != 0,
			VisibleMin:           u32(rec, 20),
			VisibleMax:           u32(rec, 24),
			HiddenMin:            u32(rec, 28),
			HiddenMax:            u32(rec, 32),
			InitialPhaseTicks:    u32(rec, 36),
			InitialBlinkTicks:    u32(rec, 40),
			InitialBlinkCooldown: u32(rec, 44),
		}
		if got != want {
			t.Fatalf("record %d = %+v, want %+v", i, got, want)
		}
	}
	if next := u32(data[eyeTableOff+len(eyeSpecs)*eyeRecordSize:], 0); next != 0 {
		t.Fatalf("menu eye table sentinel = %#x, want 0", next)
	}
}
