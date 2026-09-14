package opennox

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// NOX_TRACE_C_DURATIONS=1 prints each unrestored duration callback before
// entering C. A phase (create, update, or destroy) limits the trace to it.
var traceCDurationCalls = os.Getenv("NOX_TRACE_C_DURATIONS")

func traceCDurationCall(phase string, callback unsafe.Pointer, record *server.DurSpell) {
	if traceCDurationCalls != "1" && traceCDurationCalls != phase {
		return
	}
	writeCDurationTrace(os.Stderr, phase, callback, record)
}

func writeCDurationTrace(w io.Writer, phase string, callback unsafe.Pointer, record *server.DurSpell) {
	if record == nil {
		fmt.Fprintf(w, "NOX_C_DURATION phase=%q callback=%p record=%p\n", phase, callback, record)
		return
	}
	fmt.Fprintf(w, "NOX_C_DURATION phase=%q callback=%p record=%p spell=%d level=%d obj12=%p caster=%p obj24=%p target=%p pos=(%g,%g)\n",
		phase, callback, record, record.Spell, record.Level, record.Obj12, record.Caster16,
		record.Obj24, record.Target48, record.Pos.X, record.Pos.Y)
}
