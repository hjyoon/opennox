package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func compareStatsBulkPacketReference(t *testing.T, reference string) {
	t.Helper()
	output, err := exec.Command(reference, "--packet", "--json").Output()
	if err != nil {
		t.Fatalf("read C reference packet bytes: %v", err)
	}
	tags := memmap.PtrT[[464]byte](0x587000, 71480)
	sequence := memmap.PtrUint32(0x5D4594, 741668)
	index := memmap.PtrUint32(0x5D4594, 741660)
	oldTags, oldSequence, oldIndex := *tags, *sequence, *index
	t.Cleanup(func() { *tags, *sequence, *index = oldTags, oldSequence, oldIndex })
	raw := json.NewDecoder(bytes.NewReader(output))
	cases := 0
	for {
		var row struct {
			Mode           int
			Count          int
			Pairs          int16
			Sequence       uint32
			Variant        uint32
			ResultSequence uint32 `json:"result_sequence"`
			Length         uint32
			Output         string
		}
		err := raw.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		clear(tags[:])
		for i, label := range [...]string{"CNTL", "LGL?", "IPL?", "CLL?", "CNL?", "CMP?", "DUR?", "PAR?"} {
			copy(tags[8*i:], label)
		}
		for _, off := range [...]int{392, 416, 448} {
			copy(tags[off:], "PLRS")
		}
		for _, off := range [...]int{424, 456} {
			copy(tags[off:], "KILS")
		}
		header := make([]byte, 640)
		for i := range uint32(24) {
			binary.NativeEndian.PutUint32(header[4*i:], (i*0x1234567)^(0x89abcdef+row.Variant))
		}
		binary.NativeEndian.PutUint16(header[6:], uint16(row.Count))
		if row.Variant == 0 {
			copy(header[96:], "m")
			copy(header[352:], "s")
		} else {
			copy(header[96:], "map-x")
			copy(header[352:], "server-name")
		}
		columns := nativeStatsColumns{
			names: make([]string, row.Count), ips: make([]uint32, row.Count),
			teams: make([]uint32, row.Count), classes: make([]uint8, row.Count),
			active: make([]uint8, row.Count), durations: make([]uint32, row.Count),
			participants: make([]uint8, row.Count), pairs: bytes.Repeat([]byte{0x6b}, 65536),
		}
		for i := range row.Count {
			var name [9]byte
			for j := range name {
				name[j] = byte('a' + (i+j)%26)
			}
			columns.names[i] = string(name[:])
			columns.ips[i] = 0x80123456 ^ (uint32(i) * 0x10203)
			columns.teams[i] = 0xabcdef00 + uint32(i)
			columns.classes[i], columns.active[i] = byte(i+0x80), byte(i^0xfe)
			columns.durations[i] = 0x22222222
			columns.participants[i] = byte(i ^ 0x81)
		}
		want, err := hex.DecodeString(row.Output)
		if err != nil {
			t.Fatal(err)
		}
		*sequence = row.Sequence
		var length uint32
		packet := statsBuildPacket(&columns, header, row.Mode, row.Pairs, &length)
		packetBytes := unsafe.Slice((*byte)(unsafe.Pointer(packet)), int(length))
		equal := bytes.Equal(packetBytes, want)
		freeStatsPacket(packet)
		if !equal || length != row.Length || *sequence != row.ResultSequence {
			t.Fatalf("mode=%d players=%d pairs=%d case=%d: length=%d/%d sequence=%#x/%#x bytes match=%t",
				row.Mode, row.Count, row.Pairs, cases, length, row.Length, *sequence, row.ResultSequence, equal)
		}
		cases++
	}
	if cases != 34 {
		t.Fatalf("compared %d C reference packets, want 34", cases)
	}
	t.Logf("%d serialized reports match the C reference byte for byte", cases)
}
