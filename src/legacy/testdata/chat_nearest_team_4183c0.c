#include <assert.h>
#include <math.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "common__system__team.h"

uint32_t nox_player_netCode_85319C;
static unsigned engine_flags;
static nox_team_t teams[3];
static nox_object_t units[3], flags[3];
static nox_player_update_data_t updates[3];
static nox_playerInfo players[3];
static unsigned team_count, unit_count;
static char events[32768];
static size_t used;

/* CHAT_MATH_HELPERS */
/* COMPLETE_CHAT_BODY */

static void event(const char* format, ...) {
    if (used) events[used++]=',';
    va_list args;va_start(args,format);
    int n=vsnprintf(events+used,sizeof(events)-used,format,args);va_end(args);
    assert(n>=0 && (size_t)n<sizeof(events)-used);used+=(size_t)n;
}
static int unit_index(const nox_object_t* value) {
    for (int i=0;i<3;i++) if (value==&units[i]) return i;
    fprintf(stderr,"Chat player iterator changed/truncated: %p\n",(const void*)value);abort();
}
static int team_index(const nox_team_t* value) {
    for (int i=0;i<3;i++) if (value==&teams[i]) return i;
    fprintf(stderr,"Chat team iterator changed/truncated: %p\n",(const void*)value);abort();
}
static int member_index(const void* value) {
    for (int i=0;i<3;i++) if (value==&units[i].field_12) return i;
    fprintf(stderr,"Chat object-team address changed/truncated: %p\n",value);abort();
}
nox_object_t* nox_xxx_getFirstPlayerUnit_4DA7C0(void) {
    event("[\"unit-first\"]");return unit_count ? &units[0] : NULL;
}
nox_object_t* nox_xxx_getNextPlayerUnit_4DA7F0(const nox_object_t* value) {
    int i=unit_index(value);event("[\"unit-next\",%d]",i);
    return i+1<(int)unit_count ? &units[i+1] : NULL;
}
nox_team_t* nox_xxx_getTeamByID_418AB0(int id) {
    event("[\"by-id\",%u]",(unsigned)id);
    for (unsigned i=0;i<team_count;i++) if (teams[i].field_57==(unsigned)id) return &teams[i];
    return NULL;
}
nox_team_t* nox_server_teamFirst_418B10(void) {
    event("[\"team-first\"]");return team_count ? &teams[0] : NULL;
}
nox_team_t* nox_server_teamNext_418B60(nox_team_t* value) {
    int i=team_index(value);event("[\"team-next\",%d]",i);
    return i+1<(int)team_count ? &teams[i+1] : NULL;
}
bool nox_common_getEngineFlag(const nox_engine_flag bits) { return (engine_flags&bits)!=0; }
int sub_4196D0(void* member,void* team,int code,int flags) {
    event("[\"change\",%d,%d,%u,%d]",member_index(member),team_index(team),(uint32_t)code,flags);
    return 1;
}
void nox_xxx_createAtImpl_4191D0(unsigned char id,nox_object_team_t* member,int active,int code,int flags) {
    event("[\"attach\",%u,%d,%d,%u,%d]",(unsigned)id,member_index(member),active,(uint32_t)code,flags);
}
static void set_float(float* value,uint32_t word) { memcpy(value,&word,4); }

int main(int argc,char** argv) {
    assert(argc==2);unsigned pattern=(unsigned)strtoul(argv[1],NULL,10);assert(pattern<=255);
    if (sizeof(void*)>4) assert((uintptr_t)units>UINT32_MAX && (uintptr_t)teams>UINT32_MAX && (uintptr_t)updates>UINT32_MAX && (uintptr_t)players>UINT32_MAX && (uintptr_t)flags>UINT32_MAX);
    memset(teams,pattern,sizeof(teams));memset(units,pattern,sizeof(units));
    memset(flags,pattern,sizeof(flags));memset(updates,pattern,sizeof(updates));memset(players,pattern,sizeof(players));
    unsigned local;
    assert(scanf("%u %u %u %u",&team_count,&unit_count,&local,&engine_flags)==4);
    assert(team_count<=3 && unit_count<=3);nox_player_netCode_85319C=local;
    for (unsigned i=0;i<team_count;i++) {
        unsigned id,x,y,present;assert(scanf("%u %u %u %u",&id,&x,&y,&present)==4);
        assert(id<=255);teams[i].field_57=(unsigned char)id;
        teams[i].field_72=present ? &flags[i] : NULL;set_float(&flags[i].x,x);set_float(&flags[i].y,y);
    }
    for (unsigned i=0;i<unit_count;i++) {
        unsigned x,y,id,status,code,present;
        assert(scanf("%u %u %u %u %u %u",&x,&y,&id,&status,&code,&present)==6);assert(id<=255);
        set_float(&units[i].x,x);set_float(&units[i].y,y);
        ((nox_object_team_t*)&units[i].field_12)->id=(unsigned char)id;
        units[i].net_code=code;units[i].data_update=present ? &updates[i] : NULL;
        updates[i].player=&players[i];players[i].netCode=code;players[i].field_3680=status;
    }
    nox_team_t team_before[3];nox_object_t unit_before[3],flag_before[3];
    nox_player_update_data_t update_before[3];nox_playerInfo player_before[3];
    memcpy(team_before,teams,sizeof(teams));memcpy(unit_before,units,sizeof(units));
    memcpy(flag_before,flags,sizeof(flags));memcpy(update_before,updates,sizeof(updates));memcpy(player_before,players,sizeof(players));
    assert(sub_4183C0()==0);
    assert(memcmp(team_before,teams,sizeof(teams))==0 && memcmp(unit_before,units,sizeof(units))==0);
    assert(memcmp(flag_before,flags,sizeof(flags))==0 && memcmp(update_before,updates,sizeof(updates))==0 && memcmp(player_before,players,sizeof(players))==0);
    printf("[%s]\n",events);return 0;
}
