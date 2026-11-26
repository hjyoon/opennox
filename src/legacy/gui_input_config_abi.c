#include "client__gui__guiinput.h"
#include "client__shell__inputcfg__inputcfg.h"

int nox_gameInputApplyBindings(void);
int nox_shellInputApplyBindings(void);

// The original entries encode the parser's integer status in the return pointer.
char* sub_4C3620(void) { return (char*)(intptr_t)nox_gameInputApplyBindings(); }
char* sub_4CBD30(void) { return (char*)(intptr_t)nox_shellInputApplyBindings(); }
