//go:build !server

package opennox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

const game2AudioBridgeChild = "NOX_TEST_GAME2_AUDIO_BRIDGE_CHILD"

func game2AudioBridgeFile(t *testing.T, data []byte) *binfile.MemFile {
	t.Helper()
	p, _ := alloc.MallocHigh(uintptr(len(data) + 1))
	copy(unsafe.Slice((*byte)(p), len(data)), data)
	f := binfile.NewMemFile(p, len(data))
	t.Cleanup(f.Free)
	return f
}
func game2AudioBridgeAUD() []byte {
	b := appendAudioU32(nil, 1)
	b = appendAudioString8(b, sound.SoundShellClick.String())
	b = appendAudioI16(b, -7)
	b = append(b, 77)
	b = appendAudioI16(b, 10)
	b = append(b, 1, 0xfe, 3, 1)
	b = appendAudioString8(b, "sample0.wav")
	return append(b, 0)
}
func game2AudioBridgeAVNT() []byte {
	b := appendAudioString8(nil, sound.SoundShellClick.String())
	b = append(b, 2, 4, 3, 90, 4, 3, 5, 2, 6, 0xfd, 8, 7)
	b = appendAudioString8(b, "sample0")
	b = append(b, 0, 8)
	b = appendAudioU32(b, 33)
	b = appendAudioU32(b, 44)
	b = append(b, 9)
	b = appendAudioI16(b, 25)
	b = append(b, 10)
	b = appendAudioI16(b, -3)
	return append(b, 0)
}
func game2AudioBridgeParse(t *testing.T, s *nativeAudioEffectsState, aud bool, input []byte) {
	t.Helper()
	tail := []byte{0xa9, 0x7e, 0x5c}
	all := append(append([]byte(nil), input...), tail...)
	f := game2AudioBridgeFile(t, all)
	before := append([]byte(nil), f.RawData()...)
	var err error
	if aud {
		err = s.readAUD(f)
	} else {
		err = s.readAVNT(f)
	}
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(f.Data(), tail) || !bytes.Equal(f.RawData(), before) {
		t.Fatal("actual readAUD/readAVNT cursor or caller bytes changed")
	}
}

func TestGame2AudioBridgeDisabledMirrors(t *testing.T) {
	before := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if before.Initialized {
		t.Fatal("disabled-path test requires untouched inactive service")
	}
	var s nativeAudioEffectsState
	s.resetDefsLocked()
	game2AudioBridgeParse(t, &s, true, game2AudioBridgeAUD())
	if !s.defs[sound.SoundShellClick].enabled || s.defs[sound.SoundShellClick].priority != -7 {
		t.Fatal("ordinary AUD decode")
	}
	game2AudioBridgeParse(t, &s, false, game2AudioBridgeAVNT())
	if s.defs[sound.SoundShellClick].behavior != 4 || s.defs[sound.SoundShellClick].priority != -3 {
		t.Fatal("ordinary AVNT decode")
	}
	after := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if after.Initialized || after.Context != before.Context || after.Bank != before.Bank || after.Events != before.Events || after.Banks != before.Banks || after.Contexts != before.Contexts || after.Samples != before.Samples {
		t.Fatal("inactive mirror created or changed native resources")
	}
}

// Uses generated PCM and the actual native AIL driver. OpenAL Soft's null
// backend is allowed for local headless submission; physical audibility is not
// asserted. Parent source/main bytes must be preserved by the root runner.
func TestGame2AudioBridgeActiveParsersAndShutdown(t *testing.T) {
	if os.Getenv("NOX_TEST_NATIVE_FX") != "true" {
		t.Skip("set NOX_TEST_NATIVE_FX=true to run real local OpenAL bridge integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	python := os.Getenv("OPENNOX_TEST_PYTHON")
	if python == "" {
		python = "python3"
	}
	report := os.Getenv("OPENNOX_GAME2_AUDIO_CHILD_REPORT")
	if report == "" {
		report = t.TempDir()
	}
	absolute, err := filepath.Abs(report)
	if err != nil {
		t.Fatal(err)
	}
	report = absolute
	// A private subdirectory prevents another invocation from sharing a journal.
	if err = os.MkdirAll(report, 0700); err != nil {
		t.Fatal(err)
	}
	report, err = os.MkdirTemp(report, "game2-audio-bridge-child-")
	if err != nil {
		t.Fatal(err)
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate root-owned test preserver")
	}
	preserver := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "scripts", "run-preserved-go-test.py"))
	childArgs := []string{os.Args[0], "-test.run=^TestGame2AudioBridgeActiveParsersAndShutdownChild$", "-test.count=1", "-test.v"}
	args := []string{"-I", "-S", "-B", preserver, "--report", report, "--timeout-seconds", "45", "--heartbeat-seconds", "10", "--"}
	args = append(args, childArgs...)
	cmd := exec.CommandContext(ctx, python, args...)
	cmd.Dir = t.TempDir()
	resolvedCWD, err := filepath.EvalSymlinks(cmd.Dir)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Dir = resolvedCWD
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_NATIVE_FX_") || strings.HasPrefix(key, "NOX_E2E") || key == "NOX_DATA" || key == game2AudioBridgeChild {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, game2AudioBridgeChild+"=true")
	out, runErr := cmd.CombinedOutput()
	t.Logf("preserved child report=%s\n%s", report, out)
	if runErr != nil {
		t.Fatalf("preserved bridge subprocess failed: %v\n%s", runErr, out)
	}
	paths, err := filepath.Glob(filepath.Join(report, "test-binary-records", "*.json"))
	if err != nil || len(paths) != 1 {
		t.Fatal("exact one actual child main record", err, paths)
	}
	var rec struct {
		PID       int      `json:"pid"`
		Name      string   `json:"name"`
		Bytes     int      `json:"bytes"`
		SHA       string   `json:"sha256"`
		Retained  string   `json:"retained_path"`
		Command   []string `json:"command"`
		CWD       string   `json:"cwd"`
		Preserved bool     `json:"preservation_before_execution"`
		Started   bool     `json:"execution_started"`
		Forced    bool     `json:"forced_termination"`
		Native    *int     `json:"native_exit_code"`
		Wrapper   *int     `json:"wrapper_exit_code"`
		Phase     string   `json:"phase"`
		Timeout   float64  `json:"timeout_seconds"`
		Heartbeat float64  `json:"heartbeat_seconds"`
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	expectedSHA := sha256.Sum256(original)
	expectedText := fmt.Sprintf("%x", expectedSHA)
	sameCommand := len(rec.Command) == len(childArgs)
	if sameCommand {
		for i := range childArgs {
			sameCommand = sameCommand && rec.Command[i] == childArgs[i]
		}
	}
	if rec.PID <= 0 || !rec.Preserved || !rec.Started || rec.Forced || rec.Native == nil || *rec.Native != 0 || rec.Wrapper == nil || *rec.Wrapper != 0 || rec.Phase != "exited" || rec.Timeout != 45 || rec.Heartbeat != 10 || rec.Bytes != len(original) || rec.SHA != expectedText || rec.CWD != cmd.Dir || !sameCommand {
		t.Fatalf("actual child main terminal/source binding: %+v", rec)
	}
	retained, err := os.ReadFile(rec.Retained)
	if err != nil || len(retained) != len(original) || sha256.Sum256(retained) != expectedSHA {
		t.Fatal("child main fullbytes not preserved", err)
	}
	eventPath := filepath.Join(report, "test-supervisor-events", strings.TrimSuffix(filepath.Base(paths[0]), ".json")+".jsonl")
	events, err := os.ReadFile(eventPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(events), []byte("\n"))
	if len(lines) < 3 {
		t.Fatal("missing child PREEXEC/running/Wait events")
	}
	for i, line := range lines {
		var e struct {
			Phase   string `json:"phase"`
			PID     int    `json:"pid"`
			SHA     string `json:"sha256"`
			Started bool   `json:"execution_started"`
		}
		if json.Unmarshal(line, &e) != nil || e.SHA != expectedText {
			t.Fatal("child event bytes/source")
		}
		if i == 0 {
			if e.Phase != "preserved-before-start" || e.Started {
				t.Fatal("missing actual child PREEXEC")
			}
		} else if e.PID != rec.PID || !e.Started {
			t.Fatal("child actual PID event binding")
		}
		if i == len(lines)-1 && e.Phase != "exited" {
			t.Fatal("child Wait missing")
		}
	}
	marker := "GAME2_BRIDGE_CHILD_COMPLETED pid=" + strconv.Itoa(rec.PID)
	if strings.Contains(string(out), "--- SKIP:") || !strings.Contains(string(out), marker) || !strings.Contains(string(out), "--- PASS: TestGame2AudioBridgeActiveParsersAndShutdownChild") {
		t.Fatal("child skipped or missing actual PID completion marker")
	}

}

func TestGame2AudioBridgeActiveParsersAndShutdownChild(t *testing.T) {
	if os.Getenv(game2AudioBridgeChild) != "true" {
		t.Skip("only the isolated local bridge subprocess")
	}
	if env.IsE2E() {
		t.Fatal("bridge must not use mock audio")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	legacy.InitBlobData()
	handles.InitHigh()
	defer handles.Release()
	if audioDev != 0 || audioMaster != nil || len(nativeAudioFX.voices) != 0 || nativeAudioFX.bank != nil || legacy.Game2AudioBridgeSnapshot(1).Initialized {
		t.Fatal("child requires fresh audio ownership")
	}
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("real OpenAL playback device unavailable")
	}
	audioDev = driver
	shutdown := false
	defer func() {
		if !shutdown {
			shutdownNativeAudio()
		}
	}()
	// Load the same generated GABA through both actual platform and legacy APIs.
	// Legacy initialization must retain the ordinary bank and all16 real voices.
	pcm := bytes.Repeat([]byte{0x80}, 64)
	idx := make([]byte, 48)
	copy(idx, "GABA")
	binary.LittleEndian.PutUint32(idx[4:], 2)
	binary.LittleEndian.PutUint32(idx[8:], 1)
	copy(idx[12:28], "sample0")
	binary.LittleEndian.PutUint32(idx[32:], uint32(len(pcm)))
	binary.LittleEndian.PutUint32(idx[36:], 1000)
	base := filepath.Join(t.TempDir(), "game2-bridge")
	if err := os.WriteFile(base+".idx", idx, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(base+".bag", pcm, 0600); err != nil {
		t.Fatal(err)
	}
	if err := nativeAudioFX.init(driver, base); err != nil {
		t.Fatal(err)
	}
	ordinaryBank := nativeAudioFX.bank
	ordinaryVoices := append([]ail.Sample(nil), nativeAudioFX.voices...)
	if ordinaryBank == nil || len(ordinaryVoices) != nativeAudioVoiceCount {
		t.Fatal("actual ordinary native audio initialization")
	}
	for _, v := range ordinaryVoices {
		if v == 0 || v.GetSource() == nil {
			t.Fatal("ordinary real OpenAL source")
		}
	}
	if err := legacy.Game2AudioInitialize(driver, base); err != nil {
		t.Fatal(err)
	}
	initial := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if !initial.Initialized || initial.Context == nil || initial.Master == nil || initial.Bank == nil || initial.ContextMaster != initial.Master || initial.Driver != uintptr(driver) || initial.Events != 0 || initial.Voices != 16 || initial.Banks != 1 || initial.Contexts != 1 || initial.Samples != 1 {
		t.Fatalf("actual typed resource owners %+v", initial)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{initial.Context, initial.Master, initial.Bank} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("native resource pointer must exceed4GiB")
			}
		}
	}
	if uintptr(initial.Context) == initial.Driver {
		t.Fatal("native context confused with numeric AIL driver")
	}
	assertOrdinary := func() {
		t.Helper()
		if nativeAudioFX.bank != ordinaryBank || len(nativeAudioFX.voices) != len(ordinaryVoices) || len(nativeAudioFX.events) != 0 {
			t.Fatal("legacy mirror duplicated/changed ordinary FX owners")
		}
		for i, v := range ordinaryVoices {
			if nativeAudioFX.voices[i] != v {
				t.Fatal("ordinary voice replaced")
			}
		}
	}
	game2AudioBridgeParse(t, &nativeAudioFX, true, game2AudioBridgeAUD())
	assertOrdinary()
	aud := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	word := func(s legacy.Game2AudioBridgeState, i int) uint32 {
		return binary.LittleEndian.Uint32(s.Definition[4*i:])
	}
	if !aud.DefinitionPresent || word(aud, 0) != 1 || word(aud, 1) != 2 || word(aud, 2) != uint32(0xfffffff9) || word(aud, 5) != uint32(163*77)<<16 || word(aud, 12) != 1 || word(aud, 16) != 150 || word(aud, 48) != 1 {
		t.Fatal("main AUD tap did not update actual200-byte scalar owner")
	}
	game2AudioBridgeParse(t, &nativeAudioFX, false, game2AudioBridgeAVNT())
	assertOrdinary()
	avnt := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if word(avnt, 0) != 1 || word(avnt, 1) != 4 || word(avnt, 2) != uint32(0xfffffffd) || word(avnt, 5) != uint32(163*90)<<16 || word(avnt, 14) != 3 || word(avnt, 15) != 2 || word(avnt, 16) != 375 || word(avnt, 17) != 33 || word(avnt, 18) != 44 || word(avnt, 48) != 1 {
		t.Fatal("main AVNT tap did not update scalar owner")
	}
	event := legacy.Game2AudioBridgeCreateEvent(int(sound.SoundShellClick))
	if event == nil || !legacy.Game2AudioEventSetSample(event, nil) || legacy.Game2AudioBridgeSnapshot(1).Events != 1 {
		t.Fatal("actual public C→Go live event allocation")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(event) <= math.MaxUint32 {
		t.Fatal("native live event above4GiB")
	}
	nativeAudioFX.resetDefs()
	assertOrdinary()
	reset := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if reset.Events != 0 || word(reset, 0) != 0 || reset.Context != initial.Context || reset.Bank != initial.Bank || reset.ContextMaster != initial.Master || legacy.Game2AudioEventSetSample(event, nil) {
		t.Fatal("main reset tap must clear live events and retain owned resources")
	}
	game2AudioBridgeParse(t, &nativeAudioFX, false, game2AudioBridgeAVNT())
	event = legacy.Game2AudioBridgeCreateEvent(int(sound.SoundShellClick))
	if event == nil {
		t.Fatal("second live event")
	}
	assertOrdinary()
	shutdownNativeAudio()
	shutdown = true
	closed := legacy.Game2AudioBridgeSnapshot(int(sound.SoundShellClick))
	if closed.Initialized || closed.Context != nil || closed.Master != nil || closed.Bank != nil || closed.Events != 0 || closed.Banks != 0 || closed.Contexts != 0 || closed.Samples != 0 || legacy.Game2AudioEventSetSample(event, nil) {
		t.Fatal("main shutdown tap left legacy native owners/live handles")
	}
	if audioDev != 0 || audioMaster != nil || nativeAudioFX.bank != nil || len(nativeAudioFX.voices) != 0 || len(nativeAudioFX.events) != 0 || closed.OrdinaryOwner != nil {
		t.Fatal("actual ordinary audio shutdown cleanup")
	}
	t.Logf("GAME2_BRIDGE_CHILD_COMPLETED pid=%d driver=%#x context=%p bank=%p master=%p; AUD+AVNT/reset/shutdown owners verified; physical audibility unqualified", os.Getpid(), driver, initial.Context, initial.Bank, initial.Master)
}
