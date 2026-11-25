#include <stdint.h>

// The historical return is an opaque word; actual callers discard it.
// Preserve char** ABI here, including scalar-ready returns, outside Go.
extern uintptr_t nox_native_loadLook_415D50(void);
char** nox_xxx_loadLook_415D50() {
	return (char**)(uintptr_t)nox_native_loadLook_415D50();
}
