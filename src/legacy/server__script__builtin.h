#ifndef NOX_SERVER_SCRIPT_BUILTIN_H
#define NOX_SERVER_SCRIPT_BUILTIN_H

#include "defs.h"
#include "monster_wander_512930.h"

bool nox_script_shouldReadMoreXxx(int fi);
bool nox_script_shouldReadEvenMoreXxx(int fi);

void nox_server_gotoHome(nox_object_t* unit);
int nox_script_SetQuestInt_514BE0(void);
int nox_script_SetQuestFloat_514C10(void);
int nox_script_GiveExp_516190(void);
void nox_script_StartupScreen_516600_A(void);
int nox_script_MakeFriendly_516720(void);
int nox_script_MakeEnemy_516760(void);
int nox_script_builtin_516790(void* defaultValue);
int nox_script_BecomePet_5167D0(void);
int nox_script_BecomeEnemy_516810(void);
int nox_script_builtin_516850(void* defaultValue);
int nox_script_OblivionGive_516890(void);

#endif // NOX_SERVER_SCRIPT_BUILTIN_H
