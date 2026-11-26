//go:build !safe

package legacy

// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

var cgoSafe = os.Getenv("NOX_SAFE") == "true"

func legacyCMapPointer(base, off, size uintptr) unsafe.Pointer {
	if p := legacyMemoryPointer(base, off, size); p != nil {
		return p
	}
	fmt.Fprintf(os.Stderr, "Invalid memory access! Requested = %x+%d[%d]\n", base, off, size)
	C.abort()
	return nil
}
