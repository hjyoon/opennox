//go:build darwin && arm64

package legacy

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type ddiWireToken4E0B30 struct {
	Kind       uint32
	Allocation uint32
	Offset     uint64
}
type ddiWorkerExpected4E0B30 struct {
	Name            string
	Probes          []ddiProbe4E0B30
	Tokens          map[uint32]ddiWireToken4E0B30
	AllocationCount int
	Bases           []int
}
type ddiFileFact4E0B30 struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

func ddiFile4E0B30(path string) (ddiFileFact4E0B30, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return ddiFileFact4E0B30{}, e
	}
	return ddiFileFact4E0B30{path, len(b), ddiSHA4E0B30(b)}, nil
}
func ddiSaveJSON4E0B30(path string, value any) error {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		return e
	}
	return f.Sync()
}
func ddiOpenGzip4E0B30(path string) (*os.File, *gzip.Writer, error) {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return nil, nil, e
	}
	return f, gzip.NewWriter(f), nil
}
func ddiNativeValue4E0B30(p ddiProbe4E0B30, tokens map[uint32]ddiWireToken4E0B30, bases []uint64, callbacks [2]uint64) (uint64, error) {
	if !p.Pointer {
		return p.Canonical, nil
	}
	v, ok := tokens[uint32(p.Canonical)]
	if !ok {
		return 0, fmt.Errorf("unknown canonical token")
	}
	switch v.Kind {
	case 0:
		return 0, nil
	case 1:
		if int(v.Allocation) >= len(bases) {
			return 0, fmt.Errorf("allocation token")
		}
		return bases[v.Allocation] + v.Offset, nil
	case 2:
		return callbacks[0], nil
	case 3:
		return callbacks[1], nil
	}
	return 0, fmt.Errorf("token kind")
}
func ddiWorkerExpectedFromGraph4E0B30(g *ddiGraph4E0B30) ddiWorkerExpected4E0B30 {
	out := ddiWorkerExpected4E0B30{Name: g.Case.Input.Name, Probes: append([]ddiProbe4E0B30(nil), g.Probes...), Tokens: map[uint32]ddiWireToken4E0B30{}, AllocationCount: len(g.Allocations)}
	for word, v := range g.Tokens {
		out.Tokens[word] = ddiWireToken4E0B30{v.Kind, v.Allocation, v.Offset}
	}
	for _, a := range g.Allocations {
		out.Bases = append(out.Bases, a.Base)
	}
	return out
}

func TestRawDefaultDamageFullInitialLayout4E0B30(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" || unsafe.Sizeof(uintptr(0)) != 8 {
		t.Fatal("this full guarded native fixture requires local Darwin ARM64")
	}
	goOffset, cOffset := rawParentInitialGoOffsets4E0B30(), rawParentInitialCOffsets4E0B30()
	for id, v := range goOffset {
		if id >= 21 && id <= 24 {
			if cOffset[id] != ^uintptr(0) {
				t.Fatal("missing C monster typedef must remain explicit")
			}
		} else if cOffset[id] != v || ddiCOffset4E0B30(id) != v {
			t.Fatal("actual C/Go offsetof", id)
		}
	}
	sizes := rawParentInitialCSizes4E0B30()
	if sizes[1] != 8 || sizes[2] != 8 || unsafe.Sizeof(server.Object{}) != sizes[0]+16 {
		t.Fatal("actual C prefix and two Go EXT handles")
	}
	playerPrefix := unsafe.Offsetof(server.PlayerUpdateData{}.QuestWarpGate) + unsafe.Sizeof(server.PlayerUpdateData{}.QuestWarpGate)
	if sizes[3] != playerPrefix || unsafe.Sizeof(server.WandUseData{}) != 116 || unsafe.Sizeof(server.Object{}.UseData) != 8 || unsafe.Offsetof(server.UseDataPtr{}.Ptr) != 0 {
		t.Fatal("actual partial player prefix/Wand/UseData pointer layout")
	}
	if os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE") == "1" {
		TestRawDefaultDamageABIInitialBefore4E0B30(t)
	}
}

func TestRawDefaultDamageFullInitialMemory4E0B30(t *testing.T) {
	TestRawDefaultDamageFullInitialLayout4E0B30(t)
	fixture, err := ddiLoadFixture4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	worker := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_WORKER")
	workerSHA := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_WORKER_SHA256")
	if worker == "" || workerSHA == "" {
		t.Fatal("reviewed/prebuilt standalone C worker and SHA are required; no in-test compiler or skipped C fault probes")
	}
	worker, err = filepath.Abs(worker)
	if err != nil {
		t.Fatal(err)
	}
	wf, err := ddiFile4E0B30(worker)
	if err != nil || wf.SHA256 != workerSHA {
		t.Fatal("worker identity", err)
	}
	art := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if art == "" {
		t.Fatal("unique owned artifacts directory required")
	}
	art, err = filepath.Abs(art)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(art, 0700); err != nil {
		t.Fatal("refuse artifact overwrite", err)
	}
	planPath := filepath.Join(art, "actual-native-worker-input-plan-v1.bin.gz")
	planFile, plan, err := ddiOpenGzip4E0B30(planPath)
	if err != nil {
		t.Fatal(err)
	}
	planClosed := false
	defer func() {
		if !planClosed {
			_ = plan.Close()
			_ = planFile.Close()
		}
	}()
	if _, err = plan.Write([]byte("DDIPLN1\n")); err != nil {
		t.Fatal(err)
	}
	for _, v := range []uint32{1, 1038, 8, uint32(os.Getpagesize())} {
		if err = ddiWrite32_4E0B30(plan, v); err != nil {
			t.Fatal(err)
		}
	}
	tracePath := filepath.Join(art, "actual-Go-native-initial-field-probes-v1.jsonl.gz")
	traceFile, trace, err := ddiOpenGzip4E0B30(tracePath)
	if err != nil {
		t.Fatal(err)
	}
	traceClosed := false
	defer func() {
		if !traceClosed {
			_ = trace.Close()
			_ = traceFile.Close()
		}
	}()
	enc := json.NewEncoder(trace)
	expected := make([]ddiWorkerExpected4E0B30, 0, 1038)
	goReads, cReadable, goFaults, transportCalls := 0, 0, 0, 0
	for index, c := range fixture.Cases {
		index, c := index, c
		ok := t.Run(c.Input.Name, func(t *testing.T) {
			restore := ddiFaultMode4E0B30()
			defer restore()
			g, e := ddiBuildGraph4E0B30(fixture, c)
			if e != nil {
				_ = ddiSaveJSON4E0B30(filepath.Join(art, fmt.Sprintf("cutoff-or-build-failure-%04d.json", index)), map[string]any{"case": c.Input.Name, "input_sha256": c.InputSHA, "error": e.Error(), "padding_or_skip_applied": false})
				t.Fatal(e)
			}
			defer g.close()
			allocFacts := []map[string]any{}
			for _, a := range g.Allocations {
				allocFacts = append(allocFacts, map[string]any{"name": a.Label, "mapped_base": uintptr(a.Map), "role_offset": a.Base, "mapped_bytes": g.Page, "initial_sha256": ddiSHA4E0B30(a.Expected)})
			}
			if e = enc.Encode(map[string]any{"event": "case", "index": index, "name": c.Input.Name, "reference_input_sha256": c.InputSHA, "allocations": allocFacts, "parent_calls": 0}); e != nil {
				t.Fatal(e)
			}
			for _, p := range g.Probes {
				actual := ddiGoRead4E0B30(g.base(int(p.Allocation)), uintptr(p.GoOffset), int(p.Width))
				goReads++
				if actual.Fault {
					goFaults++
				}
				want := uint64(0)
				if p.Readable {
					want = g.expectedNative(p)
				}
				address := uintptr(g.base(int(p.Allocation))) + uintptr(p.GoOffset)
				bad := actual.Fault == p.Readable || (!actual.Fault && actual.Value != want)
				if !p.Readable && (actual.FaultAddress <= 0xffffffff || actual.FaultAddress < address || actual.FaultAddress >= address+uintptr(p.Width)) {
					bad = true
				}
				if bad {
					t.Errorf("probe%d field%d width%d initial domain/value/address: %+v readable=%t want=%#x operand=%#x", p.ID, p.Field, p.Width, actual, p.Readable, want, address)
				}
				var cv any
				if p.Readable {
					width := int(p.Width)
					if width == 8 {
						width = 0
					}
					v := uint64(rawParentInitialCBits4E0B30(g.base(int(p.Allocation)), uintptr(p.GoOffset), width))
					cv = v
					cReadable++
					if v != want {
						t.Errorf("readable actual C probe%d %#x want%#x", p.ID, v, want)
					}
				}
				if e = enc.Encode(map[string]any{"event": "probe", "case_index": index, "id": p.ID, "role": p.Allocation, "field": p.Field, "native_width": p.Width, "address": address, "expected_initial_readable": p.Readable, "Go": actual, "C_inprocess_readable_only": cv}); e != nil {
					t.Fatal(e)
				}
			}
			args := [3]unsafe.Pointer{}
			for i, word := range c.Input.Args[:3] {
				args[i] = g.Tokens[word].Pointer
			}
			for _, native := range []bool{false, true} {
				actual := rawParentEcho4E0B30(native, args[0], args[1], args[2], int32(c.Input.Args[3]), int32(c.Input.Args[4]))
				want := [5]uintptr{uintptr(args[0]), uintptr(args[1]), uintptr(args[2]), uintptr(c.Input.Args[3]), uintptr(c.Input.Args[4])}
				if !native {
					for i := 0; i < 3; i++ {
						want[i] = uintptr(uint32(want[i]))
					}
				}
				if actual != want {
					t.Errorf("actual five-word echo native%t got%v want%v", native, actual, want)
				}
				transportCalls++
				if e = enc.Encode(map[string]any{"event": "transport-only-echo", "case_index": index, "native": native, "words": actual, "parent_calls": 0}); e != nil {
					t.Fatal(e)
				}
			}
			hashes, e := g.roundtrip()
			if e != nil {
				t.Fatal(e)
			}
			if e = g.checkCanaries(); e != nil {
				t.Fatal(e)
			}
			if e = enc.Encode(map[string]any{"event": "complete-byte-guards", "case_index": index, "canonical_PE_region_SHA256": hashes, "all_native_and_mirror_mapped_bytes_unchanged": true, "postfault_parent_memory_not_tested": true}); e != nil {
				t.Fatal(e)
			}
			if e = g.writeWorkerCase(plan, uint32(index)); e != nil {
				t.Fatal(e)
			}
			expected = append(expected, ddiWorkerExpectedFromGraph4E0B30(g))
		})
		if !ok {
			t.Fatal("full1038 initial fixture is incomplete; no C worker/whole-suite pass is claimed")
		}
	}
	if goReads != 158814 || len(expected) != 1038 || transportCalls != 2076 {
		t.Fatal("actual Go/transport multiplicities", goReads, len(expected), transportCalls)
	}
	if err = plan.Close(); err != nil {
		t.Fatal(err)
	}
	if err = planFile.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = planFile.Close(); err != nil {
		t.Fatal(err)
	}
	planClosed = true
	if err = trace.Close(); err != nil {
		t.Fatal(err)
	}
	if err = traceFile.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = traceFile.Close(); err != nil {
		t.Fatal(err)
	}
	traceClosed = true
	workerProof, err := ddiRunWorker4E0B30(t, art, worker, wf, planPath, expected)
	if err != nil {
		t.Fatal(err)
	}
	proof := map[string]any{"cases": 1038, "native_Go_field_reads": goReads, "native_Go_fault_reads": goFaults, "native_C_inprocess_readable_reads": cReadable, "native_C_worker_field_reads": workerProof, "five_word_transport_calls": transportCalls, "actual_C_named_or_explicit_fixed_view_offsets": 23, "Go_derived_monster_adapter_offsets_without_named_C_type": 4, "native_parent_engine_leaves": 0, "parent_algorithm_calls": 0, "original_output_rows_read": false, "complete_PE_initial_roundtrip_and_native_byte_guards": true, "native_postfault_parent_whole_memory_equivalence": false}
	if err = ddiSaveJSON4E0B30(filepath.Join(art, "full-initial-actual-proof-v1.json"), proof); err != nil {
		t.Fatal(err)
	}
	t.Logf("RAW_DEFAULT_DAMAGE_FULL_INITIAL_OK cases=1038 Go_reads=%d C_worker_reads=%d readable_C_reads=%d Go_faults=%d transport_calls=%d new_engine_leaves=0", goReads, workerProof, cReadable, goFaults, transportCalls)
}

func ddiArchiveWorker4E0B30(path, worker, plan string) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	z := zip.NewWriter(f)
	for _, entry := range []struct{ name, path string }{{"actual-worker", worker}, {"actual-input-plan.bin.gz", plan}, {"source-worker.c", "testdata/default_damage_parent_initial_worker_4e0b30.c"}, {"source-layout.h", "default_damage_parent_full_initial_4e0b30.h"}, {"source-preexec.py", "testdata/default_damage_parent_initial_worker_preexec_4e0b30.py"}, {"frozen-fixture.json.gz", "testdata/default_damage_parent_full_initial_4e0b30.json.gz"}} {
		data, e := os.ReadFile(entry.path)
		if e != nil {
			_ = z.Close()
			return e
		}
		w, e := z.Create(entry.name)
		if e != nil {
			_ = z.Close()
			return e
		}
		if _, e = w.Write(data); e != nil {
			_ = z.Close()
			return e
		}
	}
	if e = z.Close(); e != nil {
		return e
	}
	return f.Sync()
}
func ddiRead32_4E0B30(r io.Reader) (uint32, error) {
	var n uint32
	e := binary.Read(r, binary.LittleEndian, &n)
	return n, e
}
func ddiRead64_4E0B30(r io.Reader) (uint64, error) {
	var n uint64
	e := binary.Read(r, binary.LittleEndian, &n)
	return n, e
}
func ddiRunWorker4E0B30(t *testing.T, art, worker string, wf ddiFileFact4E0B30, plan string, cases []ddiWorkerExpected4E0B30) (int, error) {
	t.Helper()
	planFact, e := ddiFile4E0B30(plan)
	if e != nil {
		return 0, e
	}
	archive := filepath.Join(art, "actual-C-worker-and-plan-PREEXEC-v1.zip")
	if e = ddiArchiveWorker4E0B30(archive, worker, plan); e != nil {
		return 0, e
	}
	// Reopen and verify executable/input bytes before the first exec. Raw C
	// probe output is retained even if parsing, a case, or the process fails.
	z, e := zip.OpenReader(archive)
	if e != nil {
		return 0, e
	}
	found := 0
	for _, m := range z.File {
		if m.Name == "actual-worker" || m.Name == "actual-input-plan.bin.gz" {
			r, e := m.Open()
			if e != nil {
				z.Close()
				return 0, e
			}
			data, e := io.ReadAll(r)
			r.Close()
			if e != nil {
				z.Close()
				return 0, e
			}
			want := wf.SHA256
			if m.Name != "actual-worker" {
				want = planFact.SHA256
			}
			if ddiSHA4E0B30(data) != want {
				z.Close()
				return 0, fmt.Errorf("PREEXEC archive bytes")
			}
			found++
		}
	}
	z.Close()
	if found != 2 {
		return 0, fmt.Errorf("PREEXEC archive members")
	}
	python := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_PYTHON")
	if python == "" {
		python, e = exec.LookPath("python3")
		if e != nil {
			return 0, e
		}
	}
	wrapper, e := filepath.Abs("testdata/default_damage_parent_initial_worker_preexec_4e0b30.py")
	if e != nil {
		return 0, e
	}
	worker, e = filepath.Abs(worker)
	if e != nil {
		return 0, e
	}
	plan, e = filepath.Abs(plan)
	if e != nil {
		return 0, e
	}
	preexec := filepath.Join(art, "actual-C-worker-PREEXEC-v1.json")
	args := []string{"-I", "-S", "-B", wrapper, worker, wf.SHA256, plan, planFact.SHA256, preexec}
	if e = ddiSaveJSON4E0B30(filepath.Join(art, "actual-C-worker-intent-v1.json"), map[string]any{"python": python, "args": args, "worker": wf, "plan": planFact, "archive": archive, "timeout_seconds": 120, "original_or_parent_calls": 0}); e != nil {
		return 0, e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, python, args...)
	stderr, e := os.OpenFile(filepath.Join(art, "actual-C-worker.stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return 0, e
	}
	defer stderr.Close()
	cmd.Stderr = stderr
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		return 0, e
	}
	rawFile, raw, e := ddiOpenGzip4E0B30(filepath.Join(art, "actual-C-worker-probes-v1.bin.gz"))
	if e != nil {
		return 0, e
	}
	if e = cmd.Start(); e != nil {
		raw.Close()
		rawFile.Close()
		return 0, e
	}
	r := io.TeeReader(stdout, raw)
	count, parseError := ddiParseWorker4E0B30(r, cases, cmd.Process.Pid)
	// Drain after any semantic/parse failure so full raw evidence survives and
	// the supervised child cannot block behind a closed/stalled stdout pipe.
	drainedBytes, drainError := io.Copy(io.Discard, r)
	if parseError == nil && drainedBytes != 0 {
		parseError = fmt.Errorf("unexpected bytes after complete worker end frame: %d", drainedBytes)
	}
	waitError := cmd.Wait()
	closeError := raw.Close()
	syncError := rawFile.Sync()
	fileError := rawFile.Close()
	exit := map[string]any{"pid": cmd.Process.Pid, "parse_error": fmt.Sprint(parseError), "drained_bytes": drainedBytes, "drain_error": fmt.Sprint(drainError), "wait_error": fmt.Sprint(waitError), "timeout": ctx.Err() != nil, "raw_close_error": fmt.Sprint(closeError), "raw_sync_error": fmt.Sprint(syncError), "raw_file_error": fmt.Sprint(fileError), "actual_C_read_count": count}
	if exitError := ddiSaveJSON4E0B30(filepath.Join(art, "actual-C-worker-exit-v1.json"), exit); exitError != nil {
		return count, fmt.Errorf("required C worker exit receipt failed: %w; observed exit=%v", exitError, exit)
	}
	if parseError != nil || drainError != nil || waitError != nil || ctx.Err() != nil || closeError != nil || syncError != nil || fileError != nil {
		return count, fmt.Errorf("C worker incomplete/failure: %v", exit)
	}
	preData, e := os.ReadFile(preexec)
	if e != nil {
		return count, e
	}
	var pre struct {
		PID    int               `json:"pid"`
		Before bool              `json:"before_exec"`
		Binary ddiFileFact4E0B30 `json:"binary"`
		Input  ddiFileFact4E0B30 `json:"input_gzip"`
	}
	if e = json.Unmarshal(preData, &pre); e != nil {
		return count, e
	}
	if pre.PID != cmd.Process.Pid || !pre.Before || pre.Binary.SHA256 != wf.SHA256 || pre.Input.SHA256 != planFact.SHA256 {
		return count, fmt.Errorf("same-PID actual C PREEXEC identity")
	}
	return count, nil
}

func ddiParseWorker4E0B30(r io.Reader, cases []ddiWorkerExpected4E0B30, pid int) (int, error) {
	var magic [8]byte
	if _, e := io.ReadFull(r, magic[:]); e != nil {
		return 0, e
	}
	if string(magic[:]) != "DDIOUT1\n" {
		return 0, fmt.Errorf("worker magic")
	}
	for _, want := range []uint32{1, 1038, 8, uint32(os.Getpagesize()), uint32(pid)} {
		v, e := ddiRead32_4E0B30(r)
		if e != nil || v != want {
			return 0, fmt.Errorf("worker actual profile/PID got%d want%d: %v", v, want, e)
		}
	}
	prefix, e := ddiRead64_4E0B30(r)
	if e != nil || prefix != uint64(rawParentInitialCSizes4E0B30()[0]) {
		return 0, fmt.Errorf("worker C prefix")
	}
	offsets := rawParentInitialGoOffsets4E0B30()
	for id := 0; id < 27; id++ {
		v, e := ddiRead64_4E0B30(r)
		want := uint64(offsets[id])
		if id >= 21 && id <= 24 {
			want = ^uint64(0)
		}
		if e != nil || v != want {
			return 0, fmt.Errorf("worker independently named C offset %d", id)
		}
	}
	var callbacks [2]uint64
	for i := range callbacks {
		v, e := ddiRead64_4E0B30(r)
		if e != nil || v <= 0xffffffff {
			return 0, fmt.Errorf("worker actual high callback")
		}
		callbacks[i] = v
	}
	loaders := map[uint32]uint64{}
	for _, width := range []uint32{1, 2, 4, 8} {
		v, e := ddiRead64_4E0B30(r)
		if e != nil || v <= 0xffffffff {
			return 0, fmt.Errorf("worker load instruction identity")
		}
		loaders[width] = v
	}
	count := 0
	var firstSemantic error
	for index, c := range cases {
		for _, want := range []uint32{uint32(index), uint32(c.AllocationCount), 153} {
			v, e := ddiRead32_4E0B30(r)
			if e != nil || v != want {
				return count, fmt.Errorf("worker case dimensions %s", c.Name)
			}
		}
		bases := make([]uint64, c.AllocationCount)
		seen := map[uint64]bool{}
		for i := range bases {
			v, e := ddiRead64_4E0B30(r)
			if e != nil || v <= 0xffffffff || seen[v] || v%uint64(os.Getpagesize()) != 0 {
				return count, fmt.Errorf("worker distinct high C mapping")
			}
			bases[i] = v
			seen[v] = true
		}
		for _, p := range c.Probes {
			id, e := ddiRead32_4E0B30(r)
			if e != nil || id != p.ID {
				return count, fmt.Errorf("worker probe id")
			}
			status, e := ddiRead32_4E0B30(r)
			if e != nil {
				return count, e
			}
			value, e := ddiRead64_4E0B30(r)
			if e != nil {
				return count, e
			}
			signal, e := ddiRead32_4E0B30(r)
			if e != nil {
				return count, e
			}
			code, e := ddiRead32_4E0B30(r)
			if e != nil {
				return count, e
			}
			fault, e := ddiRead64_4E0B30(r)
			if e != nil {
				return count, e
			}
			pc, e := ddiRead64_4E0B30(r)
			if e != nil {
				return count, e
			}
			count++
			want, e := ddiNativeValue4E0B30(p, c.Tokens, bases, callbacks)
			if e != nil {
				return count, e
			}
			bad := false
			if p.Readable {
				bad = status != 0 || value != want || signal != 0 || pc != 0 || fault != 0
			} else {
				address := bases[p.Allocation] + uint64(c.Bases[p.Allocation]) + p.GoOffset
				bad = status != 1 || (signal != 10 && signal != 11) || pc != loaders[p.Width] || fault < address || fault >= address+uint64(p.Width)
			}
			if bad && firstSemantic == nil {
				firstSemantic = fmt.Errorf("actual C initial probe mismatch case=%s id=%d field=%d width=%d readable=%t status=%d value=%#x wanted=%#x signal=%d code=%d fault=%#x pc=%#x", c.Name, p.ID, p.Field, p.Width, p.Readable, status, value, want, signal, int32(code), fault, pc)
			}
		}
		n, e := ddiRead32_4E0B30(r)
		if e != nil || int(n) != c.AllocationCount {
			return count, fmt.Errorf("worker full canary count")
		}
		for i := 0; i < int(n); i++ {
			equal, e := ddiRead32_4E0B30(r)
			if e != nil {
				return count, e
			}
			if equal != 1 {
				size, e := ddiRead32_4E0B30(r)
				if e != nil {
					return count, e
				}
				if size != uint32(os.Getpagesize()) {
					return count, fmt.Errorf("worker changed buffer evidence size")
				}
				if _, e = io.CopyN(io.Discard, r, int64(2*size)); e != nil {
					return count, e
				}
				if firstSemantic == nil {
					firstSemantic = fmt.Errorf("worker complete native buffer changed case%s allocation%d", c.Name, i)
				}
			}
		}
	}
	if _, e := io.ReadFull(r, magic[:]); e != nil {
		return count, e
	}
	if string(magic[:]) != "DDIEND1\n" {
		return count, fmt.Errorf("worker end magic")
	}
	total, e := ddiRead32_4E0B30(r)
	if e != nil || total != 158814 || count != 158814 {
		return count, fmt.Errorf("worker actual full read multiplicity")
	}
	equal, e := ddiRead32_4E0B30(r)
	if e != nil || equal != 1 {
		return count, fmt.Errorf("worker whole native canary")
	}
	return count, firstSemantic
}
