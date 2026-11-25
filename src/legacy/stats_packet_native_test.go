package legacy

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

const statsPacketReferenceC = "testdata/stats_packet_c_reference.c"

func testStatsPacketRLEOnLinkedGo(t *testing.T, transcript []byte) {
	t.Helper()
	var count int
	scanner := bufio.NewScanner(bytes.NewReader(transcript))
	for scanner.Scan() {
		var row struct {
			Length uint32 `json:"length"`
			Kind   uint32 `json:"kind"`
			Value  uint32 `json:"value"`
			Output string `json:"output"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		data := make([]byte, max(0, int(int32(row.Length))))
		for i := range data {
			switch row.Kind {
			case 0:
				data[i] = byte(row.Value)
			case 1:
				data[i] = byte(i)
			case 2:
				data[i] = byte(i / 7)
			case 3:
				data[i] = byte(255 - i)
			case 4:
				if tail := i % 16; tail >= 8 {
					data[i] = byte(tail)
				}
			}
		}
		want, err := hex.DecodeString(row.Output)
		if err != nil {
			t.Fatal(err)
		}
		if got := statsCompressRLE(data); !bytes.Equal(got, want) {
			t.Fatalf("row %d, length %#x, kind %d: %x, want %x", count, row.Length, row.Kind, got, want)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 2879 {
		t.Fatalf("checked %d statistics compression cases, want 2879", count)
	}
}

func preserveStatsRandomState(t *testing.T) (*[224]byte, *uint32, *uint32, *uint32) {
	t.Helper()
	InitBlobData()
	table := memmap.PtrT[[224]byte](0x5D4594, 741384)
	ready := memmap.PtrUint32(0x5D4594, 741656)
	first, second := statsRandomIndexCells()
	oldTable, oldReady, oldFirst, oldSecond := *table, *ready, *first, *second
	t.Cleanup(func() {
		*table, *ready, *first, *second = oldTable, oldReady, oldFirst, oldSecond
	})
	return table, ready, first, second
}

func testStatsRandomOnLinkedGo(t *testing.T, original statsRNGOriginal262) {
	t.Helper()
	table, ready, first, second := preserveStatsRandomState(t)
	for i, row := range original.Rows {
		state, err := hex.DecodeString(row.Initial)
		if err != nil {
			t.Fatal(err)
		}
		copy(table[:], state[16:240])
		*ready, *first, *second = row.Ready, row.I, row.J
		seed := int32(row.Seed)
		values, states := sha256.New(), sha256.New()
		for range row.Draws {
			var out [8]byte
			binary.LittleEndian.PutUint64(out[:], math.Float64bits(statsNextRandom(&seed)))
			values.Write(out[:])
			var word [4]byte
			binary.LittleEndian.PutUint32(word[:], uint32(seed))
			states.Write(word[:])
			copy(state[16:240], table[:])
			binary.LittleEndian.PutUint32(state[280:284], *first)
			binary.LittleEndian.PutUint32(state[284:288], *second)
			binary.LittleEndian.PutUint32(state[288:292], *ready)
			states.Write(state)
		}
		if fmt.Sprintf("%x", values.Sum(nil)) != row.Values ||
			fmt.Sprintf("%x", states.Sum(nil)) != row.States ||
			uint32(seed) != row.FinalSeed || hex.EncodeToString(state) != row.Final {
			t.Fatalf("statistics RNG row %d failed for seed %#x, ready %#x, indexes %d/%d", i, row.Seed, row.Ready, row.I, row.J)
		}
	}
}

func TestStatsPacketScrambleLinkedGoMatchesTwoClockDraws(t *testing.T) {
	table, ready, first, second := preserveStatsRandomState(t)
	// Digests come from the preserved C scrambler and RNG with two consecutive
	// timestamps and the same byte pattern.
	cases := []struct {
		length   int
		selector byte
		digest   string
	}{
		{15, 10, "b68a4e82ca4ac07e8759977ba12437c3a58d26e5c7162f7eed8ffbcff2807834"},
		{16, 11, "e398e774fd19fe60c7c143c60eb0ad079d5da90fba425aa342eb1c5ace4151d8"},
		{240, 212, "995873d038e9cd3c0518eaef938800ab6679148ce08c8aead0fc882ce084c5c5"},
		{241, 226, "b75e7289c1c3ae8805451359dce6f260985ab565493ed9cd07d573c5062159a4"},
		{512, 226, "74dbff0af184733d6cc928ddc6cf99f9a461ebdf2189ea5263ea3a32df37316c"},
	}
	for _, tc := range cases {
		*table, *ready, *first, *second = [224]byte{}, 0, 0, 0
		data := make([]byte, tc.length)
		for i := range data {
			data[i] = byte(i*73 + 11)
		}
		clock := int64(1700000000)
		packet := statsPacketScramble(data, func() int64 {
			value := clock
			clock++
			return value
		})
		if len(packet) != tc.length+5 || packet[5] != tc.selector || clock != 1700000002 ||
			fmt.Sprintf("%x", sha256.Sum256(packet)) != tc.digest {
			t.Fatalf("statistics scramble changed for length %d", tc.length)
		}
	}
}
