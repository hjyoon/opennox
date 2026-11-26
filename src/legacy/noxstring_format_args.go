//go:build 386 || amd64 || arm || arm64

package legacy

import (
	"runtime"
	"unsafe"
)

type sysVVariadicArgs struct {
	generalOffset uint32
	floatOffset   uint32
	overflow      unsafe.Pointer
	registers     unsafe.Pointer
}

type aapcs64VariadicArgs struct {
	stack         unsafe.Pointer
	generalTop    unsafe.Pointer
	floatTop      unsafe.Pointer
	generalOffset int32
	floatOffset   int32
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
	// Keep the C va_list cursor live for mixed register and stack arguments.
	switch {
	case runtime.GOARCH == "amd64" && runtime.GOOS != "windows":
		list := (*sysVVariadicArgs)(address)
		if list.generalOffset <= (6-1)*registerSize {
			slot := unsafe.Add(list.registers, list.generalOffset)
			list.generalOffset += registerSize
			return slot
		}
		return nextNativeFormatStackArgument(&list.overflow, unsafe.Sizeof(address))
	case runtime.GOARCH == "arm64" && runtime.GOOS != "darwin" && runtime.GOOS != "ios" && runtime.GOOS != "windows":
		list := (*aapcs64VariadicArgs)(address)
		offset := list.generalOffset
		if offset < 0 {
			list.generalOffset += registerSize
			if list.generalOffset <= 0 {
				return unsafe.Add(list.generalTop, offset)
			}
		}
		return nextNativeFormatStackArgument(&list.stack, unsafe.Sizeof(address))
	default:
		return nextNativeFormatStackArgument((*unsafe.Pointer)(address), unsafe.Sizeof(address))
	}
}

func nativeFormatDoubleBits(args *nativeFormatArgs) uint64 {
	const doubleSize = 8
	const floatRegisterSize = 16
	address := unsafe.Pointer(args)
	switch {
	case runtime.GOARCH == "amd64" && runtime.GOOS != "windows":
		list := (*sysVVariadicArgs)(address)
		const floatRegisterEnd = 6*8 + 8*floatRegisterSize
		if list.floatOffset < floatRegisterEnd {
			slot := unsafe.Add(list.registers, list.floatOffset)
			list.floatOffset += floatRegisterSize
			return *(*uint64)(slot)
		}
		return *(*uint64)(nextNativeFormatStackArgument(&list.overflow, doubleSize))
	case runtime.GOARCH == "arm64" && runtime.GOOS != "darwin" && runtime.GOOS != "ios" && runtime.GOOS != "windows":
		list := (*aapcs64VariadicArgs)(address)
		offset := list.floatOffset
		if offset < 0 {
			list.floatOffset += floatRegisterSize
			if list.floatOffset <= 0 {
				return *(*uint64)(unsafe.Add(list.floatTop, offset))
			}
		}
		return *(*uint64)(nextNativeFormatStackArgument(&list.stack, doubleSize))
	default:
		cursor := (*unsafe.Pointer)(address)
		if runtime.GOARCH == "arm" {
			// AAPCS32 aligns the anonymous double after any preceding word.
			*cursor = unsafe.Add(*cursor, -uintptr(*cursor)&(doubleSize-1))
		}
		return *(*uint64)(nextNativeFormatStackArgument(cursor, doubleSize))
	}
}

func nextNativeFormatStackArgument(cursor *unsafe.Pointer, size uintptr) unsafe.Pointer {
	slot := *cursor
	*cursor = unsafe.Add(slot, size)
	return slot
}
