package opennox

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestWriteCDurationTracePreservesNativePointers(t *testing.T) {
	var marker byte
	callback := unsafe.Pointer(&marker)
	target := new(server.Object)
	record := &server.DurSpell{
		Spell:    13,
		Level:    2,
		Target48: target,
		Pos:      types.Pointf{X: 1.725, Y: 2.5},
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatalf("target pointer = %p, want above 4 GiB", target)
	}
	var output bytes.Buffer
	writeCDurationTrace(&output, "update", callback, record)
	for _, want := range []string{
		`NOX_C_DURATION phase="update"`,
		fmt.Sprintf("callback=%p", callback),
		fmt.Sprintf("record=%p", record),
		fmt.Sprintf("target=%p", target),
		"spell=13 level=2",
		"pos=(1.725,2.5)",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("trace = %q, missing %q", output.String(), want)
		}
	}
}

func TestWriteCDurationTraceNilRecord(t *testing.T) {
	var output bytes.Buffer
	writeCDurationTrace(&output, "destroy", nil, nil)
	if got := output.String(); got != "NOX_C_DURATION phase=\"destroy\" callback=0x0 record=0x0\n" {
		t.Fatalf("nil trace = %q", got)
	}
}
