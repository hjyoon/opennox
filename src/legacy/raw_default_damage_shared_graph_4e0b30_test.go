//go:build darwin && arm64

package legacy

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"unsafe"
)

type rddSharedAllocation4E0B30 struct {
	Label          string  `json:"label"`
	ActualOffset   uint64  `json:"actual_offset"`
	ExpectedOffset uint64  `json:"expected_offset"`
	NativeBase     uintptr `json:"native_actual_base"`
	ExpectedBase   uintptr `json:"native_expected_base"`
	Base           int     `json:"base"`
	BeforeSHA      string  `json:"before_sha256"`
}
type rddSharedGraph4E0B30 struct {
	File         *os.File
	Header       unsafe.Pointer
	Graph        *ddiGraph4E0B30
	Rows         []rddSharedAllocation4E0B30
	Path         string
	ExpectedMaps []unsafe.Pointer
}

func (g *rddSharedGraph4E0B30) close() {
	// Actual allocations remain owned by Graph.close. The expected-shadow
	// mappings and header are never exposed as actual graph pointer targets.
	for _, p := range g.ExpectedMaps {
		rawDefaultDamageSharedUnmap4E0B30(p, g.Graph.Page)
	}
	g.ExpectedMaps = nil
	if g.Header != nil {
		rawDefaultDamageSharedUnmap4E0B30(g.Header, g.Graph.Page)
		g.Header = nil
	}
	if g.File != nil {
		g.File.Close()
		g.File = nil
	}
}

// The shadow bank preserves the comparer-accepted prefix across a terminal C
// fault. It is derived bookkeeping, not an independent oracle. Independent
// qualification must reconstruct it from input + original ordered events.
func rddShareGraph4E0B30(g *ddiGraph4E0B30, directory string) (out *rddSharedGraph4E0B30, err error) {
	if len(g.Allocations) == 0 || len(g.Allocations) > 256 {
		return nil, fmt.Errorf("shared allocation count")
	}
	if err = g.checkCanaries(); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, "actual-native-shared-AFTER.bin")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	out = &rddSharedGraph4E0B30{File: file, Graph: g, Path: path}
	defer func() {
		if err != nil {
			out.close()
		}
	}()
	n := len(g.Allocations)
	if err = file.Truncate(int64((1 + 2*n) * g.Page)); err != nil {
		return out, err
	}
	out.Header, err = rawDefaultDamageSharedMap4E0B30(int(file.Fd()), 0, g.Page)
	if err != nil {
		return out, err
	}
	ddiFill4E0B30(out.Header, g.Page, 0)
	*(*uint64)(out.Header) = rawDefaultDamageSharedMagic4E0B30
	*(*uint64)(unsafe.Add(out.Header, 8)) = rawDefaultDamageSharedVersion4E0B30
	old := append([]ddiAllocation4E0B30(nil), g.Allocations...)
	actual := make([]unsafe.Pointer, 0, n)
	shadow := make([]unsafe.Pointer, 0, n)
	committed := false
	defer func() {
		if !committed {
			for _, p := range actual {
				rawDefaultDamageSharedUnmap4E0B30(p, g.Page)
			}
			for _, p := range shadow {
				rawDefaultDamageSharedUnmap4E0B30(p, g.Page)
			}
		} else {
			for _, a := range old {
				ddiUnmap4E0B30(a.Map, g.Page)
			}
		}
	}()
	for i, a := range old {
		p, e := rawDefaultDamageSharedMap4E0B30(int(file.Fd()), uint64((1+i)*g.Page), g.Page)
		if e != nil {
			return out, e
		}
		actual = append(actual, p)
		q, e := rawDefaultDamageSharedMap4E0B30(int(file.Fd()), uint64((1+n+i)*g.Page), g.Page)
		if e != nil {
			return out, e
		}
		shadow = append(shadow, q)
		if uintptr(p) <= math.MaxUint32 || uintptr(q) <= math.MaxUint32 {
			return out, fmt.Errorf("both shared banks require actual addresses above 4 GiB")
		}
		ddiCopyIn4E0B30(p, ddiSnapshot4E0B30(a.Map, g.Page))
		ddiCopyIn4E0B30(q, a.Expected)
	}
	// Commit only once every mapping and initial copy succeeded.
	for i := range old {
		g.Allocations[i].Map = actual[i]
		g.Allocations[i].Expected = unsafe.Slice((*byte)(shadow[i]), g.Page)
	}
	out.ExpectedMaps = shadow
	committed = true
	oldIdentities := map[uintptr]uint32{}
	for word, token := range g.Tokens {
		if previous, ok := oldIdentities[token.Bits]; ok && previous != word {
			return out, fmt.Errorf("ambiguous old native token identity")
		}
		oldIdentities[token.Bits] = word
	}
	for word, token := range g.Tokens {
		if token.Kind == 1 {
			index := int(token.Allocation)
			if index < 0 || index >= len(actual) || token.Offset > uint64(g.Page) {
				return out, fmt.Errorf("shared token extent")
			}
			token.Pointer = unsafe.Add(actual[index], uintptr(token.Offset))
			token.Bits = uintptr(token.Pointer)
			g.Tokens[word] = token
		}
	}
	for i, fixup := range g.Fixups {
		word, ok := oldIdentities[fixup.Token.Bits]
		if !ok {
			return out, fmt.Errorf("shared fixup canonical identity")
		}
		token, ok := g.Tokens[word]
		if !ok || int(fixup.Destination) >= len(g.Allocations) {
			return out, fmt.Errorf("shared fixup target")
		}
		a := &g.Allocations[fixup.Destination]
		at := a.Base + int(fixup.GoOffset)
		if at < 0 || at+8 > g.Page {
			return out, fmt.Errorf("shared pointer fixup would widen a cutoff")
		}
		if !ddiPutPointer4E0B30(g.base(int(fixup.Destination)), int(fixup.Field), uintptr(fixup.GoOffset), token.Pointer) {
			return out, fmt.Errorf("shared typed C fixup")
		}
		copy(a.Expected[at:at+8], ddiBytes4E0B30(uint64(token.Bits), 8))
		g.Fixups[i].Token = token
	}
	if err = g.checkCanaries(); err != nil {
		return out, err
	}
	if _, err = g.roundtrip(); err != nil {
		return out, err
	}
	before, e := os.OpenFile(filepath.Join(directory, "actual-native-shared-BEFORE.bin.gz"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return out, e
	}
	z := gzip.NewWriter(before)
	for i, a := range g.Allocations {
		raw := ddiSnapshot4E0B30(a.Map, g.Page)
		if _, e = z.Write(raw); e != nil {
			z.Close()
			before.Close()
			return out, e
		}
		out.Rows = append(out.Rows, rddSharedAllocation4E0B30{Label: a.Label, ActualOffset: uint64((1 + i) * g.Page), ExpectedOffset: uint64((1 + n + i) * g.Page), NativeBase: uintptr(a.Map), ExpectedBase: uintptr(shadow[i]), Base: a.Base, BeforeSHA: ddiSHA4E0B30(raw)})
		if e = rawDefaultDamageSharedSync4E0B30(a.Map, g.Page); e != nil {
			z.Close()
			before.Close()
			return out, e
		}
		if e = rawDefaultDamageSharedSync4E0B30(shadow[i], g.Page); e != nil {
			z.Close()
			before.Close()
			return out, e
		}
	}
	if e = z.Close(); e != nil {
		before.Close()
		return out, e
	}
	if e = before.Sync(); e != nil {
		before.Close()
		return out, e
	}
	if e = before.Close(); e != nil {
		return out, e
	}
	if e = rawDefaultDamageSharedSync4E0B30(out.Header, g.Page); e != nil {
		return out, e
	}
	if e = file.Sync(); e != nil {
		return out, e
	}
	tokens := map[uint32]map[string]any{}
	for word, token := range g.Tokens {
		tokens[word] = map[string]any{"kind": token.Kind, "allocation": token.Allocation, "offset": token.Offset, "native_bits": token.Bits}
	}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-native-shared-layout-before-entry.json"), map[string]any{
		"schema": "opennox-raw-default-damage-shared-native-layout-v2", "page": g.Page, "allocation_count": n,
		"allocations": out.Rows, "tokens": tokens, "native_offsets": g.Offsets, "initial_fixture_SHA": ddiFixtureSHA4E0B30,
		"header_words": rawDefaultDamageSharedHeaderWords4E0B30, "file_pages": 1 + 2*n,
		"actual_bank_first_page": 1, "expected_bank_first_page": 1 + n, "root_algorithm_executed": false,
		"shadow_is_independent_oracle": false, "whole_private_Go_stack_or_heap_equivalence": false,
	}); e != nil {
		return out, e
	}
	return out, nil
}

type rddSharedAfter4E0B30 struct {
	Header                       []uint64
	Actual                       []byte
	Expected                     []byte
	Before                       []byte
	Page                         int
	AllocationCount              int
	FullAFTERArtifactSHA         string
	AllActualNativePagesCompared bool
	MismatchedAllocations        []string
}

// Read only after the parent has a real post-Wait receipt. This compares all
// actual graph pages with the accepted-prefix bank; it does not certify the
// latter's independent reconstruction or include private Go stacks/heap.
func rddReadSharedAfter4E0B30(directory string) (*rddSharedAfter4E0B30, error) {
	var layout struct {
		Schema      string                      `json:"schema"`
		Page        int                         `json:"page"`
		Count       int                         `json:"allocation_count"`
		Allocations []rddSharedAllocation4E0B30 `json:"allocations"`
	}
	data, e := os.ReadFile(filepath.Join(directory, "actual-native-shared-layout-before-entry.json"))
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(data, &layout); e != nil {
		return nil, e
	}
	if layout.Schema != "opennox-raw-default-damage-shared-native-layout-v2" || layout.Page != os.Getpagesize() || layout.Count < 1 || layout.Count > 256 || len(layout.Allocations) != layout.Count {
		return nil, fmt.Errorf("shared exact layout")
	}
	file, e := os.Open(filepath.Join(directory, "actual-native-shared-AFTER.bin"))
	if e != nil {
		return nil, e
	}
	defer file.Close()
	stat, e := file.Stat()
	if e != nil {
		return nil, e
	}
	if stat.Size() != int64((1+2*layout.Count)*layout.Page) {
		return nil, fmt.Errorf("whole actual/expected shared file extent")
	}
	raw, e := io.ReadAll(file)
	if e != nil {
		return nil, e
	}
	out := &rddSharedAfter4E0B30{Page: layout.Page, AllocationCount: layout.Count, FullAFTERArtifactSHA: ddiSHA4E0B30(raw)}
	out.Header = make([]uint64, rawDefaultDamageSharedHeaderWords4E0B30)
	for i := range out.Header {
		out.Header[i] = binary.LittleEndian.Uint64(raw[8*i:])
	}
	if out.Header[0] != rawDefaultDamageSharedMagic4E0B30 || out.Header[1] != rawDefaultDamageSharedVersion4E0B30 {
		return nil, fmt.Errorf("shared signal header version")
	}
	out.Actual = raw[layout.Page : (1+layout.Count)*layout.Page]
	out.Expected = raw[(1+layout.Count)*layout.Page:]
	for i, a := range layout.Allocations {
		if a.ActualOffset != uint64((1+i)*layout.Page) || a.ExpectedOffset != uint64((1+layout.Count+i)*layout.Page) {
			return nil, fmt.Errorf("exact shared bank offsets")
		}
		if !bytes.Equal(out.Actual[i*layout.Page:(i+1)*layout.Page], out.Expected[i*layout.Page:(i+1)*layout.Page]) {
			out.MismatchedAllocations = append(out.MismatchedAllocations, a.Label)
		}
	}
	out.AllActualNativePagesCompared = len(out.MismatchedAllocations) == 0
	beforeFile, e := os.Open(filepath.Join(directory, "actual-native-shared-BEFORE.bin.gz"))
	if e != nil {
		return nil, e
	}
	defer beforeFile.Close()
	z, e := gzip.NewReader(beforeFile)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	out.Before, e = io.ReadAll(io.LimitReader(z, int64(layout.Count*layout.Page)+1))
	if e != nil || len(out.Before) != len(out.Actual) {
		return nil, fmt.Errorf("whole native BEFORE extent: %v", e)
	}
	for i, a := range layout.Allocations {
		if ddiSHA4E0B30(out.Before[i*layout.Page:(i+1)*layout.Page]) != a.BeforeSHA {
			return nil, fmt.Errorf("shared BEFORE allocation SHA")
		}
	}
	return out, nil
}
