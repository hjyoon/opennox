#ifndef NOX_PORT_QUEST_JOURNAL_500540_H
#define NOX_PORT_QUEST_JOURNAL_500540_H

#include <stdint.h>

typedef struct nox_quest_journal_native {
	char name[132];
	uint32_t kind;
	uint32_t value;
	struct nox_quest_journal_native* next;
	struct nox_quest_journal_native* prev;
} nox_quest_journal_native;

nox_quest_journal_native* nox_xxx_journalQuestSet_500540(char* name, int32_t value);
nox_quest_journal_native* nox_xxx_journalQuestSetBool_5006B0(char* name, int32_t value);
int32_t sub_500750(char* name);
double sub_500770(char* name);
void sub_500790(nox_quest_journal_native* entry);
void sub_5007E0(char* pattern);
uint32_t sub_5009B0(char* name);
int32_t sub_500A60(void);
int32_t sub_500B70(void);

#endif // NOX_PORT_QUEST_JOURNAL_500540_H
