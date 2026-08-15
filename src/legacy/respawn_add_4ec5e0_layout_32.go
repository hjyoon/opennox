//go:build 386 || arm

package legacy

import "unsafe"

var (
	_ [60 - int(unsafe.Sizeof(respawnRecord4EC5E0{}))]struct{}
	_ [int(unsafe.Sizeof(respawnRecord4EC5E0{})) - 60]struct{}
	_ [4 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Object))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Object)) - 4]struct{}
	_ [28 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Attrs))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Attrs)) - 28]struct{}
	_ [48 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge1))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge1)) - 48]struct{}
	_ [49 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge0))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge0)) - 49]struct{}
	_ [52 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Next))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Next)) - 52]struct{}
	_ [56 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Prev))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Prev)) - 56]struct{}
)
