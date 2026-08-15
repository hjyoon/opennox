//go:build amd64 || arm64

package legacy

import "unsafe"

var (
	_ [104 - int(unsafe.Sizeof(respawnRecord4EC5E0{}))]struct{}
	_ [int(unsafe.Sizeof(respawnRecord4EC5E0{})) - 104]struct{}
	_ [8 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Object))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Object)) - 8]struct{}
	_ [40 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Attrs))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Attrs)) - 40]struct{}
	_ [80 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge1))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge1)) - 80]struct{}
	_ [81 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge0))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Charge0)) - 81]struct{}
	_ [88 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Next))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Next)) - 88]struct{}
	_ [96 - int(unsafe.Offsetof(respawnRecord4EC5E0{}.Prev))]struct{}
	_ [int(unsafe.Offsetof(respawnRecord4EC5E0{}.Prev)) - 96]struct{}
)
