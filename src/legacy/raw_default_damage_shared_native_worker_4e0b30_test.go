//go:build darwin && arm64

package legacy

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

// Keep all default outputs within the prepared runner's unique archived report.
// Explicit child/control directories remain authoritative when supplied.
func rddNativeArtifactsRoot4E0B30() (string, error) {
	if explicit := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ADAPTER_ARTIFACTS"); explicit != "" {
		return filepath.Abs(explicit)
	}
	initial := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	if initial == "" {
		return "", fmt.Errorf("unique prepared INITIAL_ARTIFACTS root is required")
	}
	return filepath.Abs(filepath.Join(initial, "native-adapter"))
}

const rddNativeRowChildEnv4E0B30 = "OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ROW_CHILD"
const rddNativeRowDirectoryEnv4E0B30 = "OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ROW_DIRECTORY"

var rddNativeRawMu4E0B30 sync.Mutex
var rddNativeRawRows4E0B30 = map[*rddRow4E0B30]json.RawMessage{}

// Bind the exact immutable row line supplied by bundle.rows, not a lossy
// re-marshalling of selected Go fields. Only the currently executed row is held.
func rddBindNativeRowSource4E0B30(row *rddRow4E0B30, raw json.RawMessage) {
	rddNativeRawMu4E0B30.Lock()
	defer rddNativeRawMu4E0B30.Unlock()
	rddNativeRawRows4E0B30[row] = append(json.RawMessage(nil), raw...)
}

func rddSaveRawRow4E0B30(row *rddRow4E0B30, directory string) error {
	rddNativeRawMu4E0B30.Lock()
	raw, ok := rddNativeRawRows4E0B30[row]
	delete(rddNativeRawRows4E0B30, row)
	rddNativeRawMu4E0B30.Unlock()
	if !ok {
		return fmt.Errorf("authentic immutable original row source was not bound")
	}
	var identity rddRow4E0B30
	if e := json.Unmarshal(raw, &identity); e != nil {
		return e
	}
	if identity.Name != row.Name || identity.Mode != row.Mode || identity.InputSHA != row.InputSHA {
		return fmt.Errorf("exact original row identity")
	}
	file, e := os.OpenFile(filepath.Join(directory, "chosen-original-row.json.gz"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	z := gzip.NewWriter(file)
	if _, e = z.Write(raw); e != nil {
		file.Close()
		return e
	}
	if e = z.Close(); e != nil {
		file.Close()
		return e
	}
	if e = file.Sync(); e != nil {
		file.Close()
		return e
	}
	if e = file.Close(); e != nil {
		return e
	}
	return ddiSaveJSON4E0B30(filepath.Join(directory, "chosen-original-row-source.json"), map[string]any{"name": row.Name, "mode": row.Mode, "input_SHA256": row.InputSHA, "uncompressed_bytes": len(raw), "uncompressed_SHA256": ddiSHA4E0B30(raw), "lossy_row_reencoding": false})
}

func rddLoadRawRow4E0B30(directory string) (*rddRow4E0B30, error) {
	file, e := os.Open(filepath.Join(directory, "chosen-original-row.json.gz"))
	if e != nil {
		return nil, e
	}
	defer file.Close()
	z, e := gzip.NewReader(file)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	raw, e := io.ReadAll(io.LimitReader(z, 16<<20))
	if e != nil {
		return nil, e
	}
	var source struct {
		Name, Mode string
		InputSHA   string `json:"input_SHA256"`
		Bytes      int    `json:"uncompressed_bytes"`
		SHA        string `json:"uncompressed_SHA256"`
	}
	data, e := os.ReadFile(filepath.Join(directory, "chosen-original-row-source.json"))
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(data, &source); e != nil {
		return nil, e
	}
	if source.Bytes != len(raw) || source.SHA != ddiSHA4E0B30(raw) {
		return nil, fmt.Errorf("whole immutable row bytes")
	}
	var row rddRow4E0B30
	if e = json.Unmarshal(raw, &row); e != nil {
		return nil, e
	}
	if row.Name != source.Name || row.Mode != source.Mode || row.InputSHA != source.InputSHA {
		return nil, fmt.Errorf("child original row identity")
	}
	return &row, nil
}

type rddNativeChildSeal4E0B30 struct {
	Archive, SHA string
	ArchiveSHA   string
	Error        error
}

var rddNativeSealMu4E0B30 sync.Mutex
var rddNativeSeals4E0B30 = map[string]rddNativeChildSeal4E0B30{}

func rddNativeChildSealForRoot4E0B30(root, executable string) (rddNativeChildSeal4E0B30, error) {
	rddNativeSealMu4E0B30.Lock()
	defer rddNativeSealMu4E0B30.Unlock()
	if seal, ok := rddNativeSeals4E0B30[root]; ok {
		return seal, seal.Error
	}
	archive, sha, e := rddSealChildMain4E0B30(executable, root)
	archiveSHA := ""
	if e == nil {
		record, hashError := ddiFile4E0B30(archive)
		if hashError != nil {
			e = hashError
		} else {
			archiveSHA = record.SHA256
		}
	}
	seal := rddNativeChildSeal4E0B30{Archive: archive, SHA: sha, ArchiveSHA: archiveSHA, Error: e}
	rddNativeSeals4E0B30[root] = seal
	return seal, e
}

// One real same-executable child per chosen original row/backend. Faults never
// longjmp into Go, and dead-child addresses remain numeric observations only.
func rddRunSupervisedNativeRow4E0B30(t *testing.T, bundle *rddBundle4E0B30, row *rddRow4E0B30, fixture *ddiFixture4E0B30, backend string) error {
	t.Helper()
	if backend != "Go" && backend != "C" {
		return fmt.Errorf("actual backend identity")
	}
	root, rootError := rddNativeArtifactsRoot4E0B30()
	if rootError != nil {
		return rootError
	}
	root = filepath.Join(root, "actual-native-rows")
	if e := os.MkdirAll(root, 0700); e != nil {
		return e
	}
	executable, e := os.Executable()
	if e != nil {
		return e
	}
	seal, e := rddNativeChildSealForRoot4E0B30(root, executable)
	if e != nil {
		return e
	}
	key := ddiSHA4E0B30([]byte(row.Name + "\x00" + row.Mode + "\x00" + backend))
	directory := filepath.Join(root, key)
	if e = os.Mkdir(directory, 0700); e != nil {
		return e
	}
	if e = rddSaveRawRow4E0B30(row, directory); e != nil {
		return e
	}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "native-row-selection-before-exec.json"), map[string]any{"name": row.Name, "mode": row.Mode, "backend": backend, "input_SHA256": row.InputSHA, "main_SHA256": seal.SHA, "original_bundle_inputs_unchanged": true, "root_dispatched": false}); e != nil {
		return e
	}
	terminal, e := rddRunBoundedSharedChild4E0B30(executable, seal.Archive, seal.SHA, directory, "TestRawDefaultDamageNativeRowWorker4E0B30", []string{rddNativeRowChildEnv4E0B30 + "=" + backend, rddNativeRowDirectoryEnv4E0B30 + "=" + directory})
	if e != nil {
		return e
	}
	after, e := rddReadSharedAfter4E0B30(directory)
	if e != nil {
		return e
	}
	if terminal.Exit == 0 {
		var result struct {
			Qualified                    bool
			Backend, Name, Mode, MainSHA string
			PID                          int
			RootDispatched               bool
			RootEntry, CallerAddress     uintptr
			CallerBits                   uint32
		}
		data, e := os.ReadFile(filepath.Join(directory, "actual-native-returned-verdict.json"))
		if e != nil {
			return e
		}
		if e = json.Unmarshal(data, &result); e != nil {
			return e
		}
		if !result.Qualified || !result.RootDispatched || result.RootEntry == 0 || result.Backend != backend || result.Name != row.Name || result.Mode != row.Mode || !after.AllActualNativePagesCompared || result.MainSHA != seal.SHA || result.PID != terminal.PID {
			return fmt.Errorf("actual returned root/verdict/full-bank evidence")
		}
		backendCode := uint64(2)
		if backend == "Go" {
			backendCode = 1
		}
		header := after.Header
		if header[2] != 2 || header[36] != 0 || header[38] != 1 || uintptr(header[39]) != result.RootEntry || header[40] != backendCode || header[28] != 2 || header[29] != 0 || uintptr(header[25]) != result.CallerAddress || uint32(header[27]) != result.CallerBits {
			return fmt.Errorf("returned verdict must match actual shared root/caller cleanup record")
		}
		return nil
	}
	if terminal.Exit != 170 && terminal.Exit != 172 {
		return fmt.Errorf("actual native child failed unqualified, exit=%d; evidence retained", terminal.Exit)
	}
	// The child has already terminated. This decoder reads retained bytes only;
	// it never executes the native parent again or receives expected-row values.
	if e = rddProduceNativeFaultPC4E0B30(directory, seal, terminal.PID); e != nil {
		return e
	}

	var layout struct {
		Offsets [27]uintptr `json:"native_offsets"`
		Tokens  map[uint32]struct {
			Bits uintptr `json:"native_bits"`
		} `json:"tokens"`
		Allocations []rddSharedAllocation4E0B30 `json:"allocations"`
	}
	data, e := os.ReadFile(filepath.Join(directory, "actual-native-shared-layout-before-entry.json"))
	if e != nil {
		return e
	}
	if e = json.Unmarshal(data, &layout); e != nil {
		return e
	}
	// Original planning uses this process's real graph, but the retained
	// verifier consumes child tokens as numbers and never reifies dead pointers.
	cases := rddFixtureIndex4E0B30(fixture)
	chosen, ok := cases[row.Name]
	if !ok {
		return fmt.Errorf("exact independent initial case missing")
	}
	g, e := ddiBuildGraph4E0B30(fixture, chosen)
	if e != nil {
		return e
	}
	defer g.close()
	plan, e := rddNativePlan4E0B30(bundle, row, fixture, g)
	if e != nil {
		return e
	}
	if plan.Projection.ClockIndex >= len(layout.Allocations) {
		return fmt.Errorf("retained native clock allocation")
	}
	actualTokens := map[uint32]uintptr{}
	for word, token := range layout.Tokens {
		actualTokens[word] = token.Bits
	}
	observed, e := os.ReadFile(filepath.Join(directory, "actual-native-observable.jsonl"))
	if e != nil {
		return e
	}
	pre, e := os.ReadFile(filepath.Join(directory, "actual-child-PREEXEC.json"))
	if e != nil {
		return e
	}
	// A separately reviewed actual-binary decoder must supply this proof. A
	// known fault exit, a shadow match, or a desired reference result is not it.
	pcProof, e := os.ReadFile(filepath.Join(directory, "actual-native-PC-qualified.json"))
	if e != nil {
		return fmt.Errorf("raw native fault retained, actual binary PC qualification is still missing: %w", e)
	}
	var pc struct {
		Qualified     bool
		ExecutableSHA string
		PC, Address   uintptr
		Site          uint32
		RootEntry     uintptr
	}
	if e = json.Unmarshal(pcProof, &pc); e != nil {
		return e
	}
	words := after.Header
	if !pc.Qualified || pc.ExecutableSHA != seal.SHA || pc.PC != uintptr(words[6]) || pc.Address != uintptr(words[5]) || pc.Site != uint32(words[9]) || pc.RootEntry != uintptr(words[39]) {
		return fmt.Errorf("independent exact binary-PC binding")
	}
	clock := layout.Allocations[plan.Projection.ClockIndex]
	proof := rddRetainedNativeFault4E0B30{
		NativeTokens: actualTokens, NativeOffsets: layout.Offsets, ClockAddress: clock.NativeBase + uintptr(clock.Base),
		CallerAddress: uintptr(words[25]), CallerBits: uint32(words[27]), CallerValid: words[28] == 1,
		Signal: int(words[3]), FaultPC: uintptr(words[6]), FaultAddress: uintptr(words[5]), FaultSite: uint32(words[9]),
		FaultOperation: int(words[10]), FaultWidth: int(words[8]), FaultProgress: int(words[13]),
		BinaryPCQualificationSHA: ddiSHA4E0B30(pcProof), ActualExecutableSHA: seal.SHA, SamePIDPREEXECReceiptSHA: ddiSHA4E0B30(pre),
		FullAFTERArtifactSHA: after.FullAFTERArtifactSHA, ObservableJSONL: observed, AllActualNativePagesCompared: after.AllActualNativePagesCompared,
		ValidationFailed: words[36] != 0, ActualRootDispatched: words[38] == 1, Backend: backend, RootEntry: uintptr(words[39]),
	}
	return rddVerifyRetainedNativeFault4E0B30(row, plan, proof)
}

func TestRawDefaultDamageNativeRowWorker4E0B30(t *testing.T) {
	before := rddNativeAdapterPhaseGate4E0B30(t)
	backend := os.Getenv(rddNativeRowChildEnv4E0B30)
	if backend == "" {
		return
	}
	if before {
		t.Fatal("full row engine child cannot execute before native engine freeze")
	}
	directory := os.Getenv(rddNativeRowDirectoryEnv4E0B30)
	if directory == "" {
		t.Fatal("dedicated actual row directory")
	}
	row, e := rddLoadRawRow4E0B30(directory)
	if e != nil {
		t.Fatal(e)
	}
	bundle, e := rddOpenBundle4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	defer bundle.close()
	if _, e = rddValidateRow4E0B30(bundle, row); e != nil {
		t.Fatal(e)
	}
	fixture, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	chosen, ok := rddFixtureIndex4E0B30(fixture)[row.Name]
	if !ok || chosen.InputSHA != row.InputSHA {
		t.Fatal("actual immutable row/initial input binding")
	}
	g, e := ddiBuildGraph4E0B30(fixture, chosen)
	if e != nil {
		t.Fatal(e)
	}
	defer g.close()
	comparison, e := rddNativePlan4E0B30(bundle, row, fixture, g)
	if e != nil {
		t.Fatal(e)
	}
	shared, e := rddShareGraph4E0B30(g, directory)
	if e != nil {
		t.Fatal(e)
	}
	defer shared.close()
	trace, e := os.OpenFile(filepath.Join(directory, "actual-native-observable.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer trace.Close()
	var pin runtime.Pinner
	var caller *int32
	comparison.Record = func(value rddNativeObserved4E0B30) error {
		raw, e := json.Marshal(value)
		if e != nil {
			return e
		}
		raw = append(raw, '\n')
		if n, e := trace.Write(raw); e != nil {
			return e
		} else if n != len(raw) {
			return io.ErrShortWrite
		}
		return trace.Sync()
	}
	comparison.ValidationFailure = func(position int, message string) {
		rawDefaultDamageSharedValidationFailed4E0B30(shared.Header, uint64(position))
		_ = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-comparison-validation-failure.json"), map[string]any{"position": position, "message": message})
	}
	comparison.Progress = func(position int, operand rawDefaultDamageNativeOperand4E0B30) {
		operation := rawDefaultDamageSharedRead4E0B30
		if operand.Write {
			operation = rawDefaultDamageSharedWrite4E0B30
		}
		var callback unsafe.Pointer
		if operand.Field == -2 {
			operation = rawDefaultDamageSharedCallback4E0B30
			callback = operand.Address
		}
		rawDefaultDamageSharedProgress4E0B30(shared.Header, uint64(position), operand.Site, operand.Address, operand.Width, operation, callback)
	}
	entry := rawParentEntry4E0B30()
	backendCode := uint64(2)
	if backend == "Go" {
		entry = reflect.ValueOf(rawDefaultDamageNativeEngine4E0B30).Pointer()
		backendCode = 1
	}
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	actualMain, e := ddiFile4E0B30(executable)
	if e != nil {
		t.Fatal(e)
	}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "child-before-root-entry.json"), map[string]any{
		"PID": os.Getpid(), "actual_main_SHA256": actualMain.SHA256, "root_entry": entry, "backend": backend,
		"raw_C_root_entry":               rawParentEntry4E0B30(),
		"generic_read_entry":             reflect.ValueOf(rawDefaultDamageReadNativeOperand4E0B30).Pointer(),
		"generic_write_entry":            reflect.ValueOf(rawDefaultDamageWriteNativeOperand4E0B30).Pointer(),
		"callback_before_observer_entry": reflect.ValueOf(rawDefaultDamageBeforeNativeCallback4E0B30).Pointer(),
		"production_callback_branch_entry_deferred_until_engine_bridge_exists": true,
		"row_name": row.Name, "row_mode": row.Mode, "input_SHA256": row.InputSHA,
	}); e != nil {
		t.Fatal(e)
	}
	comparison.CallerSink = func(position int, event rawDefaultDamageNativeCaller4E0B30) {
		if event.Entry {
			if caller != nil {
				panic("duplicate actual caller entry")
			}
			pin.Pin(event.Pointer)
			caller = event.Pointer
			if e := rawDefaultDamageSharedCallerBegin4E0B30(shared.Header, caller); e != nil {
				panic(e)
			}
			rawDefaultDamageSharedRootDispatched4E0B30(shared.Header, entry, backendCode)
		}
		if caller == nil || caller != event.Pointer {
			panic("actual caller pointer changed")
		}
		if e := rawDefaultDamageSharedCallerObserve4E0B30(shared.Header, uint64(position), event.Site, event.Pointer, event.Value); e != nil {
			panic(e)
		}
	}
	if e = rawDefaultDamageSharedBegin4E0B30(shared.Header); e != nil {
		t.Fatal(e)
	}
	active := true
	defer func() {
		if active {
			if caller != nil {
				_ = rawDefaultDamageSharedCallerEnd4E0B30(shared.Header)
				pin.Unpin()
			}
			_ = rawDefaultDamageSharedEnd4E0B30()
		}
	}()
	result, recovered, runError := comparison.execute(backend)
	if runError != nil {
		t.Fatal(runError)
	}
	if recovered != nil {
		t.Fatalf("native Go panic is not silently a qualified original fault: %v", recovered)
	}
	if e = comparison.finishReturned(result); e != nil {
		t.Fatal(e)
	}
	if caller == nil {
		t.Fatal("actual caller scalar was not observed")
	}
	callerAddress := uintptr(unsafe.Pointer(caller))
	callerBits := uint32(*caller)
	clearError := rawDefaultDamageSharedCallerEnd4E0B30(shared.Header)
	pin.Unpin()
	caller = nil
	endError := rawDefaultDamageSharedEnd4E0B30()
	active = false
	if clearError != nil || endError != nil {
		t.Fatalf("actual root scope cleanup: %v %v", clearError, endError)
	}
	for i, a := range g.Allocations {
		if e = rawDefaultDamageSharedSync4E0B30(a.Map, g.Page); e != nil {
			t.Fatal(e)
		}
		if e = rawDefaultDamageSharedSync4E0B30(shared.ExpectedMaps[i], g.Page); e != nil {
			t.Fatal(e)
		}
	}
	if e = rawDefaultDamageSharedSync4E0B30(shared.Header, g.Page); e != nil {
		t.Fatal(e)
	}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-native-returned-verdict.json"), map[string]any{"Qualified": true, "Backend": backend, "Name": row.Name, "Mode": row.Mode, "RootDispatched": comparison.DispatchedRoot, "RootEntry": entry, "MainSHA": actualMain.SHA256, "PID": os.Getpid(), "CallerAddress": callerAddress, "CallerBits": callerBits, "result_bits": uint32(result), "completed_ordered_steps": comparison.Position, "caller_physical_bytes_compared": 4, "whole_private_Go_stack_equivalence": false}); e != nil {
		t.Fatal(e)
	}
}

const rddNativePCScriptSHA4E0B30 = "9c1ccc77e9ab39c8cea3da97c2911c51fbc273c74ce1d59d52a6ababf5fb236d"
const rddNativePCContractSHA4E0B30 = "073c5bb61c7ceac89b5d6b452ef92bf1326aa2d74f0e35263eea33e165c09934"

// Run the separately sealed standalone decoder only AFTER an actual native
// fault child has exited. All paths are absolute, no original row/output is an
// argument, and Python is exactly the prepared INITIAL_PYTHON executable.
func rddProduceNativeFaultPC4E0B30(directory string, seal rddNativeChildSeal4E0B30, childPID int) (err error) {
	terminalPath := filepath.Join(directory, "native-PC-producer-terminal.json")
	state := map[string]any{"schema": "opennox-native-fault-PC-producer-terminal-v1", "execution_started": false, "native_reexecuted": false, "child_already_exited_PID": childPID, "expected_script_SHA256": rddNativePCScriptSHA4E0B30, "expected_contract_SHA256": rddNativePCContractSHA4E0B30, "main_archive": seal.Archive, "main_archive_SHA256": seal.ArchiveSHA, "actual_main_SHA256": seal.SHA, "timeout_seconds": 30}
	defer func() {
		state["error"] = fmt.Sprint(err)
		if saveError := ddiSaveJSON4E0B30(terminalPath, state); saveError != nil {
			if err != nil {
				err = fmt.Errorf("PC producer failure: %v; terminal retention: %w", err, saveError)
			} else {
				err = saveError
			}
		}
	}()
	script, e := filepath.Abs(filepath.Join("..", "..", "scripts", "qualify-default-damage-native-fault-PC.py"))
	if e != nil {
		return e
	}
	contract, e := filepath.Abs(filepath.Join("testdata", "default_damage_parent_native_PC_contract_4e0b30.json"))
	if e != nil {
		return e
	}
	python := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_PYTHON")
	if python == "" || !filepath.IsAbs(python) {
		return fmt.Errorf("exact absolute prepared INITIAL_PYTHON is required for retained PC producer")
	}
	scriptFact, e := ddiFile4E0B30(script)
	if e != nil {
		return e
	}
	contractFact, e := ddiFile4E0B30(contract)
	if e != nil {
		return e
	}
	archiveFact, e := ddiFile4E0B30(seal.Archive)
	if e != nil {
		return e
	}
	pythonFact, e := ddiFile4E0B30(python)
	if e != nil {
		return e
	}
	state["script"] = scriptFact
	state["source_contract"] = contractFact
	state["Python_executable"] = pythonFact
	state["actual_archive"] = archiveFact
	if scriptFact.SHA256 != rddNativePCScriptSHA4E0B30 || contractFact.SHA256 != rddNativePCContractSHA4E0B30 || seal.ArchiveSHA == "" || archiveFact.SHA256 != seal.ArchiveSHA {
		return fmt.Errorf("retained PC producer script/contract/archive source binding")
	}
	output := filepath.Join(directory, "actual-native-PC-qualified.json")
	if _, e = os.Lstat(output); e == nil {
		return fmt.Errorf("PC proof already existed before standalone producer")
	} else if !os.IsNotExist(e) {
		return e
	}
	args := []string{"-I", "-S", "-B", script, "--main-archive", seal.Archive, "--archive-sha256", seal.ArchiveSHA, "--main-sha256", seal.SHA, "--child-directory", directory, "--source-contract", contract, "--source-contract-sha256", rddNativePCContractSHA4E0B30, "--output", output}
	intent := map[string]any{"schema": "opennox-native-fault-PC-producer-intent-v1", "argv": append([]string{python}, args...), "cwd": directory, "script": scriptFact, "source_contract": contractFact, "main_archive": archiveFact, "main_SHA256": seal.SHA, "Python_executable": pythonFact, "timeout_seconds": 30, "native_reexecuted": false, "original_or_expected_values_supplied": false}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "native-PC-producer-intent-before-exec.json"), intent); e != nil {
		return e
	}
	stdout, e := os.OpenFile(filepath.Join(directory, "native-PC-producer.stdout.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stdout.Close()
	stderr, e := os.OpenFile(filepath.Join(directory, "native-PC-producer.stderr.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer stderr.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, python, args...)
	command.Dir = directory
	command.Stdout = stdout
	command.Stderr = stderr
	command.WaitDelay = 5 * time.Second
	var forcedRequest atomic.Bool
	defaultCancel := command.Cancel
	command.Cancel = func() error { forcedRequest.Store(true); return defaultCancel() }
	state["execution_attempted"] = true
	state["argv"] = command.Args
	start := time.Now()
	runError := command.Run()
	contextError := ctx.Err()
	state["elapsed_milliseconds"] = time.Since(start).Milliseconds()
	state["timed_out"] = contextError == context.DeadlineExceeded
	state["deadline_or_cancellation_observed"] = contextError != nil
	state["decoder_unforced_exit_zero"] = runError == nil && contextError == nil && !forcedRequest.Load()
	state["exit_code"] = -1
	if command.ProcessState != nil {
		state["exit_code"] = command.ProcessState.ExitCode()
	}
	state["execution_started"] = command.Process != nil
	if command.Process != nil {
		state["PID"] = command.Process.Pid
	}
	state["run_error"] = fmt.Sprint(runError)
	state["context_error"] = fmt.Sprint(contextError)
	state["forced_termination_requested"] = forcedRequest.Load()
	// Preserve after-Wait source facts for rejection, timeout and Start failure
	// too. Missing/changed bytes must remain evidence, not bypass this record.
	postErrors := []string{}
	for _, binding := range []struct{ Name, Path, SHA string }{{"script", script, rddNativePCScriptSHA4E0B30}, {"source_contract", contract, rddNativePCContractSHA4E0B30}, {"main_archive", seal.Archive, seal.ArchiveSHA}} {
		record, hashError := ddiFile4E0B30(binding.Path)
		observed := map[string]any{"path": binding.Path, "expected_SHA256": binding.SHA, "available": hashError == nil, "error": fmt.Sprint(hashError)}
		if hashError == nil {
			observed["fact"] = record
			observed["bytes_exact"] = record.SHA256 == binding.SHA
			if record.SHA256 != binding.SHA {
				postErrors = append(postErrors, binding.Name+" changed")
			}
		} else {
			postErrors = append(postErrors, binding.Name+": "+hashError.Error())
		}
		state["after_"+binding.Name] = observed
	}
	if syncError := stdout.Sync(); syncError != nil {
		postErrors = append(postErrors, "stdout Sync: "+syncError.Error())
	}
	if syncError := stderr.Sync(); syncError != nil {
		postErrors = append(postErrors, "stderr Sync: "+syncError.Error())
	}
	state["post_Wait_binding_and_retention_errors"] = postErrors
	if runError != nil || contextError != nil || forcedRequest.Load() || len(postErrors) != 0 {
		return fmt.Errorf("standalone retained native PC producer unqualified: run=%v context=%v post=%v; raw child/AFTER/stdout/stderr retained", runError, contextError, postErrors)
	}
	proofFact, e := ddiFile4E0B30(output)
	if e != nil {
		return e
	}
	state["proof"] = proofFact
	state["decoder_exit_zero"] = true
	return nil
}
