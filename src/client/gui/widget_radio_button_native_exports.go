package gui

// These entry points share the ordinary typed implementation with legacy C
// callers. Native window and draw-data pointers never pass through a DWORD.
func RadioButtonProcNative(win *Window, event WindowEvent) WindowEventResp {
	return radioButtonProc(win, event)
}

func RadioButtonDrawNoImageNative(win *Window, draw *WindowData) int {
	return radioButtonDrawNoImg(win, draw)
}

func RadioButtonDrawImageNative(win *Window, draw *WindowData) int {
	return radioButtonDrawImg(win, draw)
}
