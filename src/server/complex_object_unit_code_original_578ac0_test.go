package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/opennox/libs/object"
)

// The original sender runs the real 578AC0 helper before publishing each
// frozen packet. Its result is a zero-extended network WORD (including 0).
func TestComplexObjectUnitCodeOriginal578AC0(t *testing.T) {
	root := os.Getenv("NOX_COMPLEX_OBJECT_SEND_REFERENCE_DIR")
	if root == "" {
		root = "testdata/complex_object_send_518960"
	}
	raw, err := os.ReadFile(root + "/phaseA-startup-v2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "f80a75388bee334824454bfc763f2819dea1573e822e6de5ddf1b639daf7f174" {
		t.Fatal("original sender seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Rows []struct {
			Name  string `json:"name"`
			Input struct {
				Class     uint32 `json:"object_class"`
				Primary   uint32 `json:"primary_code"`
				Secondary uint32 `json:"secondary_code"`
			} `json:"input"`
			Output struct {
				Events []struct {
					Service string `json:"service"`
					Packet  string `json:"packet"`
				} `json:"events"`
			} `json:"output"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Rows) != 14110 {
		t.Fatal("original network-code dimensions")
	}
	s := new(Server)
	for _, r := range ref.Rows {
		t.Run(r.Name, func(t *testing.T) {
			expected := -1
			for _, e := range r.Output.Events {
				if e.Service == "0x40ef40" {
					packet, err := hex.DecodeString(e.Packet)
					if err != nil || len(packet) != 11 {
						t.Fatal("original packet", err)
					}
					expected = int(binary.LittleEndian.Uint16(packet[1:3]))
				}
			}
			if expected < 0 {
				t.Fatal("original network publication missing")
			}
			unit := Object{ObjClass: object.Class(r.Input.Class), NetCode: r.Input.Primary, Extent: r.Input.Secondary}
			if got := s.GetUnitNetCode(&unit); got != expected {
				t.Errorf("full helper return=%#x original zero-extended WORD=%#x class=%#x primary=%#x secondary=%#x", got, expected, r.Input.Class, r.Input.Primary, r.Input.Secondary)
			}
		})
	}
}
