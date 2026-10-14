#include <assert.h>
#include <limits.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"

/* Compile the complete active production bodies here, with their real headers.
 * Only enumeration/membership/descriptor/lookup/transport are controlled. */
/* PRODUCTION_MEMBER_BODY */
/* PRODUCTION_ENUMERATION_BODY */

static nox_team_t* teams[3];
static nox_object_t* units[5];
static unsigned char team_ids[3] = {1, 2, 255};
static unsigned char member_ids[5] = {1, 0, 2, 1, 255};
static uint32_t codes[5] = {0x1234, 0xffff8012, 0x1234abcd, 0x80008000, 0xffffffff};
static uint16_t types[5] = {0x3456, 0x1234, 0xfedc, 0xffff, 0};
static int team_count, unit_count;
static nox_object_t* lookup;
static char events[32768], packets[4096];
static size_t event_size, packet_size;

static void append(char* output, size_t* used, size_t capacity, const char* format, ...) {
    if (*used) {
        assert(*used + 1 < capacity);
        output[(*used)++] = ',';
    }
    va_list args;
    va_start(args, format);
    int n = vsnprintf(output + *used, capacity - *used, format, args);
    va_end(args);
    assert(n >= 0 && (size_t)n < capacity - *used);
    *used += (size_t)n;
}

static nox_object_team_t* member(int i) {
    return (nox_object_team_t*)&units[i]->field_12;
}
static int team_index(const nox_team_t* value) {
    for (int i = 0; i < 3; i++) if (value == teams[i]) return i;
    fprintf(stderr, "team pointer changed/truncated: %p\n", (const void*)value);
    abort();
}
static int unit_index(const nox_object_t* value) {
    for (int i = 0; i < 5; i++) if (value == units[i]) return i;
    fprintf(stderr, "unit pointer changed/truncated: %p\n", (const void*)value);
    abort();
}
static int member_index(const nox_object_team_t* value) {
    for (int i = 0; i < 5; i++) if (value == member(i)) return i;
    fprintf(stderr, "member pointer changed/truncated: %p\n", (const void*)value);
    abort();
}

nox_team_t* nox_server_teamFirst_418B10(void) {
    append(events, &event_size, sizeof(events), "[\"team-first\"]");
    return team_count ? teams[0] : NULL;
}
nox_team_t* nox_server_teamNext_418B60(nox_team_t* value) {
    int i = team_index(value);
    append(events, &event_size, sizeof(events), "[\"team-next\",%d]", i);
    return i + 1 < team_count ? teams[i + 1] : NULL;
}
void sub_4197C0(wchar2_t* title, int recipient) {
    int i = team_index((nox_team_t*)title);
    append(events, &event_size, sizeof(events), "[\"descriptor\",%d,%d]", i, recipient);
}
nox_object_t* nox_xxx_getFirstPlayerUnit_4DA7C0(void) {
    append(events, &event_size, sizeof(events), "[\"unit-first\"]");
    return unit_count ? units[0] : NULL;
}
nox_object_t* nox_xxx_getNextPlayerUnit_4DA7F0(const nox_object_t* value) {
    int i = unit_index(value);
    append(events, &event_size, sizeof(events), "[\"unit-next\",%d]", i);
    return i + 1 < unit_count ? units[i + 1] : NULL;
}
int nox_xxx_teamCompare2_419180(nox_object_team_t* value, unsigned char id) {
    int i = member_index(value);
    append(events, &event_size, sizeof(events), "[\"contains\",%d,%u]", i, (unsigned)id);
    return value->id != 0 && value->id == id;
}
nox_object_t* nox_server_getObjectFromNetCode_4ECCB0(int code) {
    append(events, &event_size, sizeof(events), "[\"lookup\",%u]", (unsigned)(uint32_t)code);
    return lookup;
}
int nox_xxx_netSendPacket1_4E5390(int recipient, const void* payload, int length,
                                 nox_object_t* related, int remove) {
    assert(payload != NULL && length == 10 && related == NULL && remove == 1);
    const unsigned char* bytes = payload;
    char hex[21];
    for (int i = 0; i < 10; i++) snprintf(hex + 2 * i, 3, "%02x", bytes[i]);
    append(events, &event_size, sizeof(events), "[\"packet\",%d,\"%s\",0,%d]", recipient, hex, remove);
    append(packets, &packet_size, sizeof(packets), "[%d,\"%s\",0,%d]", recipient, hex, remove);
    return 77;
}

static unsigned char team_guards[3][sizeof(nox_team_t)];
static unsigned char unit_guards[5][sizeof(nox_object_t)];
static void initialize(unsigned char pattern) {
    for (int i = 0; i < 3; i++) {
        memset(teams[i], pattern, sizeof(*teams[i]));
        teams[i]->field_57 = team_ids[i];
        memcpy(team_guards[i], teams[i], sizeof(*teams[i]));
    }
    for (int i = 0; i < 5; i++) {
        memset(units[i], pattern, sizeof(*units[i]));
        units[i]->typ_ind = types[i];
        units[i]->net_code = codes[i];
        member(i)->id = member_ids[i];
        memcpy(unit_guards[i], units[i], sizeof(*units[i]));
    }
}
static void begin_case(void) {
    events[0] = packets[0] = 0;
    event_size = packet_size = 0;
}
static void end_case(const char* name, int comma) {
    for (int i = 0; i < 3; i++) assert(memcmp(team_guards[i], teams[i], sizeof(*teams[i])) == 0);
    for (int i = 0; i < 5; i++) assert(memcmp(unit_guards[i], units[i], sizeof(*units[i])) == 0);
    printf("%s{\"name\":\"%s\",\"events\":[%s],\"packets\":[%s]}", comma ? "," : "", name, events, packets);
}

int main(int argc, char** argv) {
    assert(argc == 3);
    int enumeration = strcmp(argv[1], "enumeration") == 0;
    assert(enumeration || strcmp(argv[1], "packet") == 0);
    unsigned char pattern = (unsigned char)strtoul(argv[2], NULL, 0);
    for (int i = 0; i < 3; i++) {
        teams[i] = malloc(sizeof(*teams[i]));
        assert(teams[i] && (sizeof(void*) == 4 || (uintptr_t)teams[i] > UINT32_MAX));
    }
    for (int i = 0; i < 5; i++) {
        units[i] = malloc(sizeof(*units[i]));
        assert(units[i] && (sizeof(void*) == 4 || (uintptr_t)units[i] > UINT32_MAX));
    }
    initialize(pattern);
    printf("[");
    if (enumeration) {
        const char* names[5] = {"host-index-31", "empty-teams", "empty-players", "all-members", "last-ordinary-index"};
        int recipients[5] = {31, 0, 30, 0, 30};
        int nt[5] = {3, 0, 3, 3, 3}, nu[5] = {5, 5, 0, 5, 5};
        for (int i = 0; i < 5; i++) {
            begin_case(); team_count = nt[i]; unit_count = nu[i]; lookup = units[4];
            sub_4DF2E0(recipients[i]); end_case(names[i], i != 0);
        }
    } else {
        const char* names[5] = {"nil-member", "missing-object", "zero-team-id", "high-net-code", "signed-net-code"};
        int members[5] = {-1, 0, 1, 2, 4}, recipients[5] = {17, 17, 17, 30, 0};
        uint32_t net_codes[5] = {0xffffffff, 0x1234abcd, 0, 0x1234abcd, 0xffffffff};
        int lookups[5] = {2, -1, 0, 2, 3};
        for (int i = 0; i < 5; i++) {
            begin_case(); lookup = lookups[i] < 0 ? NULL : units[lookups[i]];
            sub_4198A0(members[i] < 0 ? NULL : member(members[i]), recipients[i], (int32_t)net_codes[i]);
            end_case(names[i], i != 0);
        }
    }
    printf("]\n");
    for (int i = 0; i < 5; i++) free(units[i]);
    for (int i = 0; i < 3; i++) free(teams[i]);
    return 0;
}
