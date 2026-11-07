//go:build darwin && arm64

package legacy

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
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
)

const rddSharedChildEnv4E0B30 = "OPENNOX_RAW_DEFAULT_DAMAGE_SHARED_CHILD"
const rddSharedChildTimeout4E0B30 = 30 * time.Second

func rddSealChildMain4E0B30(executable, directory string) (string, string, error) {
	path := filepath.Join(directory, "actual-child-main-before-first-exec.zip")
	source, e := os.Open(executable)
	if e != nil {
		return "", "", e
	}
	defer source.Close()
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return "", "", e
	}
	z := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "actual-test-main", Method: zip.Deflate}
	header.SetMode(0700)
	member, e := z.CreateHeader(header)
	if e != nil {
		file.Close()
		return "", "", e
	}
	hash := sha256.New()
	if _, e = io.Copy(io.MultiWriter(member, hash), source); e != nil {
		file.Close()
		return "", "", e
	}
	if e = z.Close(); e != nil {
		file.Close()
		return "", "", e
	}
	if e = file.Sync(); e != nil {
		file.Close()
		return "", "", e
	}
	if e = file.Close(); e != nil {
		return "", "", e
	}
	want := fmt.Sprintf("%x", hash.Sum(nil))
	archive, e := zip.OpenReader(path)
	if e != nil {
		return "", "", e
	}
	defer archive.Close()
	if len(archive.File) != 1 || archive.File[0].Name != "actual-test-main" {
		return "", "", fmt.Errorf("exact one PREEXEC main")
	}
	input, e := archive.File[0].Open()
	if e != nil {
		return "", "", e
	}
	verified := sha256.New()
	_, e = io.Copy(verified, input)
	input.Close()
	if e != nil || fmt.Sprintf("%x", verified.Sum(nil)) != want {
		return "", "", fmt.Errorf("child canonical main byte proof")
	}
	return path, want, nil
}

type rddSharedChildTerminal4E0B30 struct {
	PID, Exit        int
	TimedOut, Forced bool
	RunError         string
}

func rddRunBoundedSharedChild4E0B30(executable, archive, mainSHA, directory, testName string, extraEnv []string) (rddSharedChildTerminal4E0B30, error) {
	out, e := os.OpenFile(filepath.Join(directory, "child.stdout.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return rddSharedChildTerminal4E0B30{}, e
	}
	defer out.Close()
	stderr, e := os.OpenFile(filepath.Join(directory, "child.stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return rddSharedChildTerminal4E0B30{}, e
	}
	defer stderr.Close()
	python := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_PYTHON")
	if python == "" || !filepath.IsAbs(python) {
		return rddSharedChildTerminal4E0B30{}, fmt.Errorf("absolute preserved Python is required")
	}
	wrapper, e := filepath.Abs("testdata/raw_default_damage_shared_child_preexec_4e0b30.py")
	if e != nil {
		return rddSharedChildTerminal4E0B30{}, e
	}
	preexec := filepath.Join(directory, "actual-child-PREEXEC.json")
	ctx, cancel := context.WithTimeout(context.Background(), rddSharedChildTimeout4E0B30)
	defer cancel()
	command := exec.CommandContext(ctx, python, "-I", "-S", "-B", wrapper, executable, mainSHA, preexec, "-test.run=^"+testName+"$", "-test.v")
	command.Env = append(os.Environ(), extraEnv...)
	command.Stdout = out
	command.Stderr = stderr
	command.WaitDelay = 5 * time.Second
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "parent-intent-before-child-exec.json"), map[string]any{
		"argv": command.Args, "main_SHA256": mainSHA, "canonical_PREEXEC_main_archive": archive,
		"timeout_ns": int64(rddSharedChildTimeout4E0B30), "root_algorithm_dispatched_by_supervisor": false,
		"retention": "All raw buffers/logs/PREEXEC/terminal retained on success or failure"}); e != nil {
		return rddSharedChildTerminal4E0B30{}, e
	}
	runError := command.Run()
	terminal := rddSharedChildTerminal4E0B30{Exit: -1, TimedOut: ctx.Err() == context.DeadlineExceeded, RunError: fmt.Sprint(runError)}
	terminal.Forced = terminal.TimedOut
	if command.ProcessState != nil {
		terminal.Exit = command.ProcessState.ExitCode()
	}
	if command.Process != nil {
		terminal.PID = command.Process.Pid
	}
	out.Sync()
	stderr.Sync()
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-child-terminal.json"), terminal); e != nil {
		return terminal, e
	}
	if terminal.TimedOut {
		return terminal, fmt.Errorf("dedicated child timeout; full raw evidence retained")
	}
	var before struct {
		Before bool `json:"before_exec"`
		PID    int  `json:"pid"`
		Binary struct {
			SHA string `json:"sha256"`
		}
	}
	data, e := os.ReadFile(preexec)
	if e != nil {
		return terminal, e
	}
	if e = json.Unmarshal(data, &before); e != nil {
		return terminal, e
	}
	if !before.Before || before.PID != terminal.PID || before.Binary.SHA != mainSHA {
		return terminal, fmt.Errorf("actual same-PID PREEXEC byte identity")
	}
	return terminal, nil
}

func TestRawDefaultDamageSharedFaultControls4E0B30(t *testing.T) {
	rddNativeAdapterPhaseGate4E0B30(t)
	if mode := os.Getenv(rddSharedChildEnv4E0B30); mode != "" {
		rddRunSharedFaultChild4E0B30(t, mode)
		return
	}
	root, rootError := rddNativeArtifactsRoot4E0B30()
	if rootError != nil {
		t.Fatal(rootError)
	}
	if e := os.MkdirAll(root, 0700); e != nil {
		t.Fatal(e)
	}
	root = filepath.Join(root, "shared-fault-controls")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	archive, mainSHA, e := rddSealChildMain4E0B30(executable, root)
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"healthy-WORD", "guarded-WORD", "nil-C-callback", "unrelated-user-signal", "other-thread-hardware-fault", "actual-caller-mutate-then-fault"} {
		mode := mode
		t.Run(mode, func(t *testing.T) {
			directory := filepath.Join(root, mode)
			if e := os.Mkdir(directory, 0700); e != nil {
				t.Fatal(e)
			}
			terminal, e := rddRunBoundedSharedChild4E0B30(executable, archive, mainSHA, directory, "TestRawDefaultDamageSharedFaultControls4E0B30", []string{rddSharedChildEnv4E0B30 + "=" + mode, "OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ADAPTER_ARTIFACTS=" + directory})
			if e != nil {
				t.Fatal(e)
			}
			after, e := rddReadSharedAfter4E0B30(directory)
			if e != nil {
				t.Fatal(e)
			}
			words := after.Header
			if words[13] != 0 || words[20] == 0 {
				t.Fatal("exact control progress/owner")
			}
			want := append([]byte(nil), after.Before...)
			operandOffset := len(want) - after.Page + 256
			if mode == "guarded-WORD" {
				operandOffset = len(want) - 1
			}
			physical := "not-a-WORD-fault"
			switch mode {
			case "healthy-WORD":
				if terminal.Exit != 0 || words[2] != 2 || words[3] != 0 || words[10] != rawDefaultDamageSharedWrite4E0B30 || words[8] != 2 || words[9] != 0x4e0e6d {
					t.Fatal("healthy exact C store lifecycle/header")
				}
				want[operandOffset] = 2
				want[operandOffset+1] = 0
				if !after.AllActualNativePagesCompared {
					t.Fatal("healthy complete actual/expected banks")
				}
			case "guarded-WORD":
				if terminal.Exit != 170 || words[2] != 3 || words[23] != 1 || words[24] != 0 || words[10] != rawDefaultDamageSharedWrite4E0B30 || words[8] != 2 || words[9] != 0x4e0e6d {
					t.Fatal("exact guarded store signal classification")
				}
				if words[6] != words[14] || words[16] != words[15] || words[21] != words[20] || words[5] < words[7] || words[5]-words[7] >= 2 {
					t.Fatal("exact actual STRH PC/SP/thread/operand")
				}
				// Preserve a differing bounded physical fragment without converting
				// it into a qualified native WORD/engine equivalence result.
				if after.Actual[operandOffset] == 0xaa {
					physical = "native-AA-unchanged"
				} else if after.Actual[operandOffset] == 2 {
					physical = "native-partial-02-UNQUALIFIED"
					want[operandOffset] = 2
				} else {
					t.Fatal("unexplained physical WORD fragment")
				}
			case "nil-C-callback", "actual-caller-mutate-then-fault":
				if terminal.Exit != 170 || words[2] != 3 || words[23] != 1 || words[10] != rawDefaultDamageSharedCallback4E0B30 || words[8] != 0 || words[9] != 0x4e1049 || words[5] != 0 || words[6] != 0 || words[12] != 0 {
					t.Fatal("actual nil fetch classification")
				}
				if words[18] != words[17]+4 || words[19] != words[18] || words[16] != words[15] || words[21] != words[20] {
					t.Fatal("actual BLR caller LR/SP/thread")
				}
				if !after.AllActualNativePagesCompared {
					t.Fatal("whole native banks before nil fetch")
				}
				if mode == "actual-caller-mutate-then-fault" && (words[25] == 0 || words[26] != 0x80000000 || words[27] != 0x7ffffffd || words[28] != 1 || words[29] != 1 || words[30] != words[20] || words[31] != 1 || words[32] != 0x80000000 || words[35] != 1) {
					t.Fatal("actual pinned caller scalar sampled after C mutation before any after-observer")
				}
			case "other-thread-hardware-fault":
				if terminal.Exit != 171 || words[2] != 4 || !((words[3] == 10 && words[4] >= 1 && words[4] <= 3) || (words[3] == 11 && words[4] >= 1 && words[4] <= 2)) || words[24] != 1 || words[20] == 0 || words[21] == 0 || words[20] == words[21] || words[5] != 0 || words[6] != 0 || words[10] != rawDefaultDamageSharedCallback4E0B30 || words[23] != 0 {
					t.Fatal("other C thread hardware fault must fail owner gate")
				}
				if words[18] != words[17]+4 || words[16] != words[15] || !after.AllActualNativePagesCompared {
					t.Fatal("other thread actual BLR/whole-bank evidence")
				}
			case "unrelated-user-signal":
				if terminal.Exit != 171 || words[2] != 4 || words[23] != 0 || words[24] == 0 {
					t.Fatal("unrelated signal must remain unqualified")
				}
				if !after.AllActualNativePagesCompared {
					t.Fatal("unrelated signal changed native graph")
				}
			}
			if !bytes.Equal(want, after.Actual) {
				t.Fatal("whole actual allocation bytes differ outside the explicit control store")
			}
			if e = ddiSaveJSON4E0B30(filepath.Join(directory, "actual-native-fault-AFTER-observation.json"), map[string]any{
				"mode": mode, "PID": terminal.PID, "exit": terminal.Exit, "header_raw_words": words,
				"AFTER_SHA256": after.FullAFTERArtifactSHA, "allocation_count": after.AllocationCount,
				"all_outside_control_operand_bytes_exact": true, "all_actual_vs_accepted_shadow_pages_equal": after.AllActualNativePagesCompared,
				"physical_WORD_observation": physical, "physical_WORD_equivalence_qualified": false,
				"actual_binary_PC_opcode_qualification_pending": true, "root_engine_leaves": 0,
				"actual_caller_scalar_sampled": mode == "actual-caller-mutate-then-fault", "whole_private_Go_stack_equivalence": false,
			}); e != nil {
				t.Fatal(e)
			}
			t.Logf("RAW_DEFAULT_DAMAGE_SHARED_FAULT_CONTROL_OK mode=%s PID=%d exit=%d native_engine_leaves=0", mode, terminal.PID, terminal.Exit)
		})
	}
}

func rddRunSharedFaultChild4E0B30(t *testing.T, mode string) {
	switch mode {
	case "healthy-WORD", "guarded-WORD", "nil-C-callback", "unrelated-user-signal", "other-thread-hardware-fault", "actual-caller-mutate-then-fault":
	default:
		t.Fatal("unknown shared control")
	}
	directory := os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_NATIVE_ADAPTER_ARTIFACTS")
	if directory == "" {
		t.Fatal("child retention root")
	}
	f, e := ddiLoadFixture4E0B30()
	if e != nil {
		t.Fatal(e)
	}
	g, e := ddiBuildGraph4E0B30(f, f.Cases[0])
	if e != nil {
		t.Fatal(e)
	}
	defer g.close()
	index, e := g.alloc("shared-fault-control-operand", 256, 0xaa)
	if e != nil {
		t.Fatal(e)
	}
	shared, e := rddShareGraph4E0B30(g, directory)
	if e != nil {
		t.Fatal(e)
	}
	defer shared.close()
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	main, e := ddiFile4E0B30(executable)
	if e != nil {
		t.Fatal(e)
	}
	if e = ddiSaveJSON4E0B30(filepath.Join(directory, "child-before-C-entry.json"), map[string]any{"PID": os.Getpid(), "mode": mode, "actual_main_SHA256": main.SHA256, "raw_C_root_entry": rawParentEntry4E0B30(), "fixture_SHA256": ddiFixtureSHA4E0B30, "native_root_algorithm_present": false, "shared_page_count": len(g.Allocations), "registered_parent_invoked": false}); e != nil {
		t.Fatal(e)
	}
	address := g.base(index)
	if mode == "guarded-WORD" {
		address = unsafe.Add(g.Allocations[index].Map, g.Page-1)
	}
	operation, site, width := rawDefaultDamageSharedWrite4E0B30, uint32(0x4e0e6d), 2
	if mode == "nil-C-callback" || mode == "other-thread-hardware-fault" || mode == "actual-caller-mutate-then-fault" {
		operation = rawDefaultDamageSharedCallback4E0B30
		site = 0x4e1049
		width = 0
		address = nil
	}
	rawDefaultDamageSharedProgress4E0B30(shared.Header, 0, site, address, width, operation, nil)
	if e = rawDefaultDamageSharedBegin4E0B30(shared.Header); e != nil {
		t.Fatal(e)
	}
	active := true
	defer func() {
		if active {
			_ = rawDefaultDamageSharedEnd4E0B30()
		}
	}()
	if mode == "actual-caller-mutate-then-fault" {
		actual := int32(-2147483648)
		var pin runtime.Pinner
		pin.Pin(&actual)
		if e = rawDefaultDamageSharedCallerBegin4E0B30(shared.Header, &actual); e != nil {
			pin.Unpin()
			t.Fatal(e)
		}
		rawDefaultDamageSharedCallerMutationControl4E0B30(shared.Header, &actual, 0x7ffffffd)
		// Any return is a failure. Clear the C pointer before releasing the pin.
		clearError := rawDefaultDamageSharedCallerEnd4E0B30(shared.Header)
		pin.Unpin()
		t.Fatalf("mutation/fetch control unexpectedly returned: %v", clearError)
	}
	if mode == "unrelated-user-signal" {
		rawDefaultDamageSharedUnrelatedSignalControl4E0B30(shared.Header)
	} else if mode == "other-thread-hardware-fault" {
		if e = rawDefaultDamageSharedOtherThreadControl4E0B30(shared.Header); e != nil {
			t.Fatal(e)
		}
	} else {
		rawDefaultDamageSharedControl4E0B30(shared.Header, address, 2, mode == "nil-C-callback")
	}
	e = rawDefaultDamageSharedEnd4E0B30()
	active = false
	if e != nil {
		t.Fatal(e)
	}
	if mode != "healthy-WORD" {
		t.Fatal("fault control unexpectedly returned")
	}
	// Independent input value of this completed primitive, never read back from
	// native actual bytes to manufacture an expected result.
	copy(g.Allocations[index].Expected[g.Allocations[index].Base:], []byte{2, 0})
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
}
