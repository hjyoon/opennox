package client

func DrawableBackWallQueue(dr *Drawable, staticDraw bool) bool {
	if dr == nil || dr.DrawFuncPtr == nil {
		return false
	}
	flags := uint32(dr.ObjFlags)
	class := uint32(dr.ObjClass)
	return flags&0x1000 == 0 && flags&1 != 0 && staticDraw && class&0x80800000 == 0 && (flags&0x48 != 0 || class&0x400000 != 0) && flags&0x800 == 0
}

func DrawableFrontQueue(dr *Drawable, staticDraw bool) bool {
	if dr == nil || dr.DrawFuncPtr == nil {
		return false
	}
	flags := uint32(dr.ObjFlags)
	if flags&0x1000 != 0 || flags&1 == 0 {
		return false
	}
	class := uint32(dr.ObjClass)
	return !(staticDraw && class&0x80800000 == 0 && flags&0x800 == 0 && (flags&0x48 != 0 || class&0x400000 != 0))
}

func DrawableOverlayQueue(dr *Drawable) bool {
	if dr == nil || dr.DrawFuncPtr == nil {
		return false
	}
	flags := uint32(dr.ObjFlags)
	return flags&0x1000 == 0 && flags&0x4000 != 0 && flags&0x40 != 0
}

func DrawableWorldQueue(dr *Drawable) bool {
	if dr == nil || dr.DrawFuncPtr == nil {
		return false
	}
	flags := uint32(dr.ObjFlags)
	class := uint32(dr.ObjClass)
	return flags&1 == 0 && (class&0x2000 == 0 || flags&0x1000000 != 0) && flags&0x1000 == 0
}
