package legacy

import "unsafe"

// VideoModeRowsC exposes the published native row table to ownership tests.
var VideoModeRowsC = (***uint16)(unsafe.Pointer(&legacyPointers.videoRows))
