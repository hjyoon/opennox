#ifndef NOX_MAPGEN_THEME_NATIVE_H
#define NOX_MAPGEN_THEME_NATIVE_H

#include <stdint.h>
#include <stdio.h>

// Native entry points retain the fixed PE32 theme layout while transporting
// live names, FILE handles and theme addresses at host pointer width.
// Integer entry points in GAME4_2.h resolve registered legacy tokens; native
// callers must use these entries before any pointer-to-int truncation.
int nox_mapgenReadThemeNative_51E260(int* theme, const char* name);
int nox_mapgenReadAlgDataNative_51EBB0(uint8_t* theme, FILE* file);
int nox_mapgenReadSpellSetNative_51EFB0(uint8_t* theme, FILE* file);
int nox_mapgenReadExitNative_51F800(uint8_t* theme, FILE* file);

#endif // NOX_MAPGEN_THEME_NATIVE_H
