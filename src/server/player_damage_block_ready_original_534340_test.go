package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestPlayerDamageBlockReadyOriginal534340(t *testing.T) {
	path := os.Getenv("NOX_BLOCK_READY_ORIGINAL_REFERENCE")
	if path == "" {
		path = "testdata/monster_combat_predicates_original_534020.json.gz"
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(blob)
	if hex.EncodeToString(sum[:]) != "3e22140d34890c7309037a368e91a10d5cf580a3868c28bbac491ad55b776bc3" {
		t.Fatal("unsealed original state reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var reference struct {
		Schema      string `json:"schema"`
		OriginalSHA string `json:"original_sha256"`
		Control     uint16 `json:"control_word"`
		Rows        []struct {
			Name  string `json:"name"`
			Entry uint32 `json:"entry"`
			Input struct {
				Class  uint32 `json:"cls"`
				Status uint32 `json:"status"`
				Head   int8   `json:"head"`
				Action uint32 `json:"action"`
				Prefix uint32 `json:"prefix528"`
			} `json:"input"`
			Return uint32 `json:"return_u32"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&reference); err != nil {
		t.Fatal(err)
	}
	if reference.Schema != "original-monster-combat-predicates-534020-v2" || reference.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || reference.Control != 0xe3f || len(reference.Rows) != 11835 {
		t.Fatal("original state provenance")
	}
	type unitGuard struct {
		Before [64]byte
		Value  Object
		After  [64]byte
	}
	type updateGuard struct {
		Before [64]byte
		Value  MonsterUpdateData
		After  [64]byte
	}
	ug, freeU := alloc.New(unitGuard{})
	t.Cleanup(freeU)
	dg, freeD := alloc.New(updateGuard{})
	t.Cleanup(freeD)
	snapshot := func(ptr unsafe.Pointer, size uintptr) []byte {
		return append([]byte(nil), unsafe.Slice((*byte)(ptr), int(size))...)
	}
	for _, region := range []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(ug), unsafe.Sizeof(*ug)}, {unsafe.Pointer(dg), unsafe.Sizeof(*dg)}} {
		for i := range unsafe.Slice((*byte)(region.ptr), int(region.size)) {
			unsafe.Slice((*byte)(region.ptr), int(region.size))[i] = byte(i*31 + 19)
		}
	}
	controlOnly := os.Getenv("NOX_BLOCK_READY_ORIGINAL_REFERENCE_CONTROL_ONLY") == "true"
	count := 0
	for _, row := range reference.Rows {
		if row.Entry != 0x534340 {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			c := row.Input
			ug.Value = Object{ObjClass: object.Class(c.Class), UpdateData: unsafe.Pointer(&dg.Value)}
			dg.Value = MonsterUpdateData{StatusFlags: object.MonsterStatus(c.Status), AIStackInd: c.Head, Field132: c.Prefix}
			if c.Head >= 0 {
				dg.Value.AIStack[c.Head].Action = c.Action
			}
			beforeU, beforeD := snapshot(unsafe.Pointer(ug), unsafe.Sizeof(*ug)), snapshot(unsafe.Pointer(dg), unsafe.Sizeof(*dg))
			if !controlOnly {
				var got uint32
				if playerDamageMonsterBlockReady534340(&ug.Value) {
					got = 1
				}
				if got != row.Return {
					t.Errorf("block-readiness got=%d original=%d status=%08x head=%d action=%d prefix=%d", got, row.Return, c.Status, c.Head, c.Action, c.Prefix)
				}
			}
			if !bytes.Equal(beforeU, snapshot(unsafe.Pointer(ug), unsafe.Sizeof(*ug))) || !bytes.Equal(beforeD, snapshot(unsafe.Pointer(dg), unsafe.Sizeof(*dg))) {
				t.Error("block-readiness changed native input or guard")
			}
		})
	}
	if count != 342 {
		t.Fatal("original hold context count", count)
	}
	t.Logf("original Go block-readiness contexts=%d control-only=%t", count, controlOnly)
}
