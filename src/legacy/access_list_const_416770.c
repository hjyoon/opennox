#include "GAME1.h"
extern int* nox_native_access_add_banned_416770(int, wchar2_t*, void*);
int* sub_416770(int minutes,wchar2_t* name,const char* serial) {
    return nox_native_access_add_banned_416770(minutes,name,(void*)serial);
}
