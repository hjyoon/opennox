package legacy

// cgoABIIntPtr adapts a C int pointer to a native Go int without aliasing
// storage of different widths. The returned closure copies an output value
// back to the C-sized slot.
func cgoABIIntPtr(src *int32) (*int, func()) {
	if src == nil {
		return nil, func() {}
	}
	dst := int(*src)
	return &dst, func() { *src = int32(dst) }
}

// cgoABIUintPtr is the unsigned counterpart of cgoABIIntPtr.
func cgoABIUintPtr(src *uint32) (*uint, func()) {
	if src == nil {
		return nil, func() {}
	}
	dst := uint(*src)
	return &dst, func() { *src = uint32(dst) }
}
