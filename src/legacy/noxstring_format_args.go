//go:build 386 || amd64 || arm || arm64

package legacy

import (
	"runtime"
	"unsafe"
)

type sysVVariadicArgs struct {
	generalOffset uint32
	_             uint32 // Floating-point offset, advanced by the C formatter.
	overflow      unsafe.Pointer
	registers     unsafe.Pointer
}

type aapcs64VariadicArgs struct {
	stack         unsafe.Pointer
	generalTop    unsafe.Pointer
	_             unsafe.Pointer // Floating-point register save area.
	generalOffset int32
	_             int32 // Floating-point offset, advanced by the C formatter.
}

func nativeFormatInteger(args *nativeFormatArgs, _ bool) int32 {
	return *(*int32)(nextNativeFormatWord(args))
}

func nativeFormatStringPointer(args *nativeFormatArgs, _ bool) unsafe.Pointer {
	return *(*unsafe.Pointer)(nextNativeFormatWord(args))
}

func nextNativeFormatWord(args *nativeFormatArgs) unsafe.Pointer {
	const registerSize = 8
	address := unsafe.Pointer(args)
	// The C va_list remains live through the formatting callback. Update the
	// same cursor that the C floating-point formatter uses for mixed lists.
	switch {
	case runtime.GOARCH == "amd64" && runtime.GOOS != "windows":
		list := (*sysVVariadicArgs)(address)
		if list.generalOffset <= (6-1)*registerSize {
			slot := unsafe.Add(list.registers, list.generalOffset)
			list.generalOffset += registerSize
			return slot
		}
		return nextNativeFormatStackWord(&list.overflow)
	case runtime.GOARCH == "arm64" && runtime.GOOS != "darwin" && runtime.GOOS != "ios" && runtime.GOOS != "windows":
		list := (*aapcs64VariadicArgs)(address)
		offset := list.generalOffset
		if offset < 0 {
			list.generalOffset += registerSize
			if list.generalOffset <= 0 {
				return unsafe.Add(list.generalTop, offset)
			}
		}
		return nextNativeFormatStackWord(&list.stack)
	default:
		return nextNativeFormatStackWord((*unsafe.Pointer)(address))
	}
}

func nextNativeFormatStackWord(cursor *unsafe.Pointer) unsafe.Pointer {
	slot := *cursor
	*cursor = unsafe.Add(slot, unsafe.Sizeof(slot))
	return slot
}
