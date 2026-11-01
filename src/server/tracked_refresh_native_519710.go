package server

// Count is the separate 417270 service. Frame replaces the original direct
// DWORD global read and must remain lazy on the zero/large-count paths.
type TrackedRefreshRuntime519710 struct {
	Count func(int32) int32
	Frame func() uint32
}

// TrackedRefreshDue519710 preserves the signed quotient and unsigned elapsed
// comparison. The ordinary scalar adapter also accepts native int counts.
func TrackedRefreshDue519710(frame, last uint32, tracked int) bool {
	if tracked == 0 {
		return false
	}
	return tracked > 60 || frame-last > uint32(60/tracked)
}

// TrackedRefreshNative519710 requires a valid native update and Player link.
// The original caches the byte index before Count, reloads last after Count,
// then loads the frame. Count may replace the Player link without a reread.
func TrackedRefreshNative519710(update *PlayerUpdateData, runtime TrackedRefreshRuntime519710) int32 {
	count := runtime.Count(int32(update.Player.PlayerInd))
	if count == 0 {
		return 0
	}
	if count > 60 {
		return 1
	}
	last := update.Field67
	frame := runtime.Frame()
	if TrackedRefreshDue519710(frame, last, int(count)) {
		return 1
	}
	return 0
}
