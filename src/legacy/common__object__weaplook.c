#include <stdint.h>

// The historical return is an opaque word; actual callers discard it.
// Preserve char** ABI here, including scalar-ready returns, outside Go.
extern uintptr_t nox_native_loadModifyers_4158C0(void);
char** nox_xxx_loadModifyers_4158C0() {
	return (char**)(uintptr_t)nox_native_loadModifyers_4158C0();
}
