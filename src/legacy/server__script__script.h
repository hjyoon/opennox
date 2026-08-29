#ifndef NOX_SERVER_SCRIPT_SCRIPT_H
#define NOX_SERVER_SCRIPT_SCRIPT_H

#include "defs.h"

void nox_script_callByIndex_507310(int index, void* a1, void* a2);
int nox_script_indexByEvent(char* a1);

int32_t nox_xxx_xferReadScriptHandler_4F5580(
	nox_script_callback_t* handler,
	char* context);
void* nox_xxx_scriptCallByEventBlock_502490(void* a1, void* a2, void* a3, int eventCode);
char* nox_script_objCallbackName_508CB0(nox_object_t* a1, int a2);

#endif // NOX_SERVER_SCRIPT_SCRIPT_H
