package ccall

import (
	"unsafe"
)

func NewFuncs[
	F any,
](wrap func(cfnc unsafe.Pointer) F) *Funcs[F] {
	return &Funcs[F]{
		wrap:  wrap,
		byPtr: make(map[unsafe.Pointer]F),
	}
}

type Funcs[F any] struct {
	wrap  func(p unsafe.Pointer) F
	byPtr map[unsafe.Pointer]F
}

func (r *Funcs[F]) Register(cfnc unsafe.Pointer, fnc F) {
	r.byPtr[cfnc] = fnc
}

// Lookup returns the registered implementation without wrapping the C pointer.
func (r *Funcs[F]) Lookup(cfnc unsafe.Pointer) (F, bool) {
	fnc, ok := r.byPtr[cfnc]
	return fnc, ok
}

func (r *Funcs[F]) Get(cfnc unsafe.Pointer) F {
	if cfnc == nil {
		var zero F
		return zero
	}
	fnc, ok := r.Lookup(cfnc)
	if ok {
		return fnc
	}
	return r.wrap(cfnc)
}
