//go:build darwin

package memguard

import "syscall"

func newPage(size int) ([]byte, func()) {
	page := PageSize()
	num := size / page
	if size%page != 0 {
		num++
	}
	b, err := syscall.Mmap(-1, 0, num*page, syscall.PROT_NONE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		panic(err)
	}
	return b[:size], func() {
		if err := syscall.Munmap(b); err != nil {
			panic(err)
		}
	}
}
