package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const optionsDialogCycleChild4AA650 = "OPENNOX_TEST_OPTIONS_DIALOG_CYCLE_4AA650"

type optionsDialogLegacyServer4AA650 struct {
	Server
	srv *server.Server
	onS func()
}

func (s *optionsDialogLegacyServer4AA650) S() *server.Server {
	if s.onS != nil {
		fn := s.onS
		s.onS = nil
		fn()
	}
	return s.srv
}

func TestOptionsDialogCycle4AA650NativePointerSlots(t *testing.T) {
	if os.Getenv(optionsDialogCycleChild4AA650) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestOptionsDialogCycle4AA650NativePointerSlots$", "-test.count=1")
		cmd.Env = append(os.Environ(), optionsDialogCycleChild4AA650+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated options-dialog regression failed: %v\n%s", err, out)
		}
		return
	}

	InitBlobData()
	stringsPath := filepath.Join(t.TempDir(), "options-dialog.json")
	stringsJSON := []byte(`{
		"entries": [
			{"id":"War06a:HecThreatJack","vals":[{"str":"zero","str2":"dialog-zero"}]},
			{"id":"Con10B.scr:HecubahLine9","vals":[{"str":"one","str2":"dialog-one"}]},
			{"id":"Wiz11A.scr:HecubahTalk06","vals":[{"str":"two","str2":"dialog-two"}]}
		]
	}`)
	if err := os.WriteFile(stringsPath, stringsJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	strings := strman.New()
	if err := strings.ReadJSON(stringsPath); err != nil {
		t.Fatal(err)
	}
	srv := server.New(nil, nil, strings)
	defer srv.Close()
	bridge := &optionsDialogLegacyServer4AA650{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	defer func() { GetServer = oldGetServer }()
	oldDialogs := Dialogs
	initDialog()
	defer func() { Dialogs = oldDialogs }()

	const (
		tableBase = uintptr(0x587000)
		tableOff  = uintptr(172892)
	)
	wantPacked := []byte{
		0x68, 0x13, 0x5b, 0x00,
		0x80, 0x13, 0x5b, 0x00,
		0x98, 0x13, 0x5b, 0x00,
	}
	if got := memmap.Slice(tableBase, tableOff)[:len(wantPacked)]; !bytes.Equal(got, wantPacked) {
		t.Fatalf("packed PE32 dialog table = %x, want %x", got, wantPacked)
	}

	wantKeys := []string{
		"War06a:HecThreatJack",
		"Con10B.scr:HecubahLine9",
		"Wiz11A.scr:HecubahTalk06",
	}
	wantDialogs := []string{"dialog-zero", "dialog-one", "dialog-two"}
	counter := memmap.PtrUint32(0x5D4594, 1309744)
	*counter = 2
	Dialogs.PlayFile("already-busy", 100)
	optionsDialogCycleNative4AA650()
	if *counter != 2 {
		t.Fatalf("busy dialog changed counter to %d, want 2", *counter)
	}
	if got := Dialogs.FileToRead(); got != "already-busy" {
		t.Fatalf("busy dialog file = %q, want unchanged", got)
	}
	Dialogs.Sub_44D8F0()

	*counter = 0
	for i, wantKey := range wantKeys {
		pointer := *memmap.PtrPtr(tableBase, tableOff+4*uintptr(i))
		if pointer == nil {
			t.Fatalf("dialog table pointer %d is nil", i)
		}
		if unsafe.Sizeof(pointer) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("dialog table pointer %d = %p, want native address above PE32 range", i, pointer)
		}
		if got := alloc.GoString((*byte)(pointer)); got != wantKey {
			t.Fatalf("dialog table key %d = %q, want %q", i, got, wantKey)
		}

		optionsDialogCycleNative4AA650()
		if want := uint32((i + 1) % len(wantKeys)); *counter != want {
			t.Fatalf("dialog counter after call %d = %d, want %d", i, *counter, want)
		}
		if got := Dialogs.FileToRead(); got != wantDialogs[i] {
			t.Fatalf("dialog file after call %d = %q, want %q", i, got, wantDialogs[i])
		}
		Dialogs.Sub_44D8F0()
	}

	*counter = 0
	observedAtLookup := ^uint32(0)
	bridge.onS = func() {
		observedAtLookup = *counter
		negativeCounter := int32(-4)
		*counter = uint32(negativeCounter)
	}
	optionsDialogCycleNative4AA650()
	if observedAtLookup != 1 {
		t.Fatalf("counter observed by string lookup = %d, want incremented value 1", observedAtLookup)
	}
	if got := int32(*counter); got != -1 {
		t.Fatalf("counter after callback mutation = %d, want signed -4 %% 3 remainder -1", got)
	}
	if got := Dialogs.FileToRead(); got != wantDialogs[0] {
		t.Fatalf("dialog after callback counter mutation = %q, want %q", got, wantDialogs[0])
	}
	Dialogs.Sub_44D8F0()

	if got := memmap.Slice(tableBase, tableOff)[:len(wantPacked)]; !bytes.Equal(got, wantPacked) {
		t.Fatalf("native dialog lookups changed packed PE32 table: got %x, want %x", got, wantPacked)
	}
}
