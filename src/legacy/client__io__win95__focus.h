#ifndef NOX_PORT_CLIENT_IO_WIN95_FOCUS
#define NOX_PORT_CLIENT_IO_WIN95_FOCUS

#include "defs.h"

typedef void (*nox_focus_callback)(int);
typedef const char nox_focus_name;
void sub_42EBB0(unsigned int a1, nox_focus_callback fnc, int field_4, nox_focus_name* name);

#endif // NOX_PORT_CLIENT_IO_WIN95_FOCUS
