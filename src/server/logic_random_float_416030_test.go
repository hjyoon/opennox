package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/prand"
)

const (
	logicRandomFloatTableSHA256416030    = "7dfccdce34d74e348fae4181630a4c41a20a560432658ba2526969bb7000605f"
	logicRandomFloatPiOutputSHA256416030 = "3c278b2efdbe095d912b42ae8fb459963f76ec819a72d3962bb1425badffaf25"
)

func logicRandomFloatWordsSHA256_416030(words []uint32) string {
	data := make([]byte, 4*len(words))
	for i, word := range words {
		binary.LittleEndian.PutUint32(data[4*i:], word)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func TestLogicRandomFloat416030ScaleAndFirstValue(t *testing.T) {
	if got := math.Float32bits(logicRandomFloatTableScale416030); got != 0x38000100 {
		t.Fatalf("scale bits = %#08x, want 0x38000100", got)
	}
	random := prand.New(0)
	got := logicRandomFloat416030(
		random,
		-randomReachablePointPi4ED970,
		randomReachablePointPi4ED970,
	)
	if bits := math.Float64bits(got); bits != 0x40040900c579a06c {
		t.Fatalf("first result bits = %#016x, want 0x40040900c579a06c", bits)
	}
	if bits := math.Float32bits(float32(got)); bits != 0x40204806 {
		t.Fatalf("first spill bits = %#08x, want 0x40204806", bits)
	}
	if index := random.Index(); index != 1 {
		t.Fatalf("index = %d, want 1", index)
	}
}

func TestLogicRandomFloat416030TableMatchesSealedOracle(t *testing.T) {
	random := prand.New(0)
	words := make([]uint32, 4096)
	maximum := 0
	for i := range words {
		value := random.Int(0, logicRandomFloatTableMax416030)
		if value > maximum {
			maximum = value
		}
		words[i] = uint32(value)
	}
	if got := logicRandomFloatWordsSHA256_416030(words); got != logicRandomFloatTableSHA256416030 {
		t.Fatalf("table SHA-256 = %s, want %s", got, logicRandomFloatTableSHA256416030)
	}
	if maximum != logicRandomFloatTableMax416030 {
		t.Fatalf("table maximum = %#x, want %#x", maximum, logicRandomFloatTableMax416030)
	}
	if index := random.Index(); index != 0 {
		t.Fatalf("wrapped index = %d, want 0", index)
	}
}

func TestLogicRandomFloat416030PiSpillsMatchOracleForWholeTable(t *testing.T) {
	random := prand.New(0)
	words := make([]uint32, 4096)
	for i := range words {
		value := logicRandomFloat416030(
			random,
			-randomReachablePointPi4ED970,
			randomReachablePointPi4ED970,
		)
		words[i] = math.Float32bits(float32(value))
	}
	if got := logicRandomFloatWordsSHA256_416030(words); got != logicRandomFloatPiOutputSHA256416030 {
		t.Fatalf("pi-output SHA-256 = %s, want %s", got, logicRandomFloatPiOutputSHA256416030)
	}
	if index := random.Index(); index != 0 {
		t.Fatalf("wrapped index = %d, want 0", index)
	}
}

func TestLogicRandomFloat416030ZeroAndUnorderedRangesDoNotStep(t *testing.T) {
	random := prand.New(17)
	negativeZero := math.Float32frombits(0x80000000)
	if got := logicRandomFloat416030(random, negativeZero, negativeZero); math.Float64bits(got) != 0x8000000000000000 {
		t.Fatalf("zero-range result bits = %#016x, want negative zero", math.Float64bits(got))
	}
	if index := random.Index(); index != 17 {
		t.Fatalf("zero-range index = %d, want 17", index)
	}
	if got := logicRandomFloat416030(random, math.Float32frombits(0x7fc12345), 2); got != 2 {
		t.Fatalf("unordered-range result = %v, want 2", got)
	}
	if index := random.Index(); index != 17 {
		t.Fatalf("unordered-range index = %d, want 17", index)
	}
}

func TestLogicRandomFloat416030WrapsAt4096(t *testing.T) {
	random := prand.New(4095)
	_ = logicRandomFloat416030(random, -1, 1)
	if index := random.Index(); index != 0 {
		t.Fatalf("first index = %d, want 0", index)
	}
	_ = logicRandomFloat416030(random, -1, 1)
	if index := random.Index(); index != 1 {
		t.Fatalf("second index = %d, want 1", index)
	}
}
