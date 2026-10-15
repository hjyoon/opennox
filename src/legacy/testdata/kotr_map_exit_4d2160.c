#include <assert.h>
#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
#include <unistd.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "common__system__team.h"
#include "common__random.h"
#include "crown_pickup_4f3400.h"

static nox_team_t teams[3], expected_teams[3];
static nox_playerInfo players[4], expected_players[4];
static nox_object_t units[4], crowns[4], expected_units[4], expected_crowns[4];
// Native counterpart of GAME.EXE Team+76, deliberately NOT field_76.
static nox_object_t* flags[3];
static int which, nt, np, cached, rolls[4], ri, roll_count, members[3];
static char transcript[8192];
static size_t used;

static void event(const char* format, ...) {
    va_list args;va_start(args,format);
    int n=vsnprintf(transcript+used,sizeof(transcript)-used,format,args);
    va_end(args);assert(n>0 && (size_t)n<sizeof(transcript)-used);
    used+=(size_t)n;
}
static void fault(int signum) {
    (void)signum;
    // No allocator/stdio in the fault handler. Preserve the observed prefix.
    ssize_t n=write(STDOUT_FILENO,transcript,used);
    _Exit(n==(ssize_t)used ? 86 : 87);
}
static int team_index(nox_team_t* p) {
    for(int i=0;i<nt;i++) if(p==&teams[i]) return i;
    assert(!"truncated or invalid team pointer");return -1;
}
static int unit_index(nox_object_t* p) {
    for(int i=0;i<4;i++) if(p==&units[i]) return i;
    assert(!"truncated or invalid unit pointer");return -1;
}
static int crown_index(nox_object_t* p) {
    for(int i=0;i<4;i++) if(p==&crowns[i]) return i;
    assert(!"truncated or invalid crown pointer");return -1;
}
int nox_common_playerInfoCount_416F40(void) {event("count\n");return cached;}
nox_team_t* nox_server_teamFirst_418B10(void) {event("team-first\n");return nt ? &teams[0] : NULL;}
nox_team_t* nox_server_teamNext_418B60(nox_team_t* p) {
    int i=team_index(p);event("team-next %d\n",i);return i+1<nt ? &teams[i+1] : NULL;
}
int sub_418BC0(nox_team_t* p) {int i=team_index(p);event("members %d\n",i);return members[i];}
nox_object_t* nox_server_teamGetFlagObject_4180D0(nox_team_t* p) {
    int i=team_index(p);event("flag-load %d\n",i);return flags[i];
}
int nox_common_randomInt_415FA0(int lo,int hi) {
    event("random %d %d\n",lo,hi);assert(ri<roll_count);return rolls[ri++];
}
nox_playerInfo* nox_common_playerInfoGetFirst_416EA0(void) {event("player-first\n");return np ? &players[0] : NULL;}
nox_playerInfo* nox_common_playerInfoGetNext_416EE0(nox_playerInfo* p) {
    for(int i=0;i<np;i++) if(p==&players[i]) {
        event("player-next %d\n",i);return i+1<np ? &players[i+1] : NULL;
    }
    assert(!"truncated or invalid player pointer");return NULL;
}
int nox_xxx_teamCompare2_419180(nox_object_team_t* p,unsigned char id) {
    int i=-1;
    for(int n=0;n<4;n++) if(p==(nox_object_team_t*)&units[n].field_12) i=n;
    assert(i>=0);event("contains %d %u\n",i,(unsigned)id);
    int result=p->id && p->id==id;
    if(which==10) {
        flags[0]=&crowns[3];players[0].playerUnit=&units[3];
        expected_players[0].playerUnit=&units[3];
    }
    return result;
}
uint32_t sub_4F3400(nox_object_t* who,nox_object_t* crown,int32_t a,int32_t b) {
    int i=unit_index(who),j=crown_index(crown);
    event("pickup %d %d %d %d\n",i,j,a,b);
    if(which==9) crowns[j].inv_holder=expected_crowns[j].inv_holder=who;
    return which==8 ? 0 : 0x80000001u;
}

/* COMPLETE_KOTR_MAP_EXIT_BODY */

int main(int argc,char** argv) {
    assert(argc==3);which=atoi(argv[1]);int pattern=atoi(argv[2]);
    assert(which>=0 && which<16 && (pattern==0 || pattern==0x5a || pattern==0xff));
    signal(SIGSEGV,fault);signal(SIGBUS,fault);
    if(sizeof(void*)>4) {
        assert((uintptr_t)teams>UINT32_MAX && (uintptr_t)players>UINT32_MAX);
        assert((uintptr_t)units>UINT32_MAX && (uintptr_t)crowns>UINT32_MAX);
    }
    memset(teams,pattern,sizeof(teams));memset(players,pattern,sizeof(players));
    memset(units,pattern,sizeof(units));memset(crowns,pattern,sizeof(crowns));
    nt=np=cached=roll_count=1;rolls[0]=0;
    unsigned char tids[3]={1,2,255},uids[4]={1,2,255,255};
    for(int i=0;i<3;i++) {members[i]=1;teams[i].field_57=tids[i];flags[i]=&crowns[i];}
    for(int i=0;i<4;i++) {players[i].playerUnit=&units[i];crowns[i].inv_holder=NULL;}
    if(which==0 || which==1) {nt=np=0;cached=which==0 ? 0 : 32;}
    if(which==2 || which==12) crowns[0].inv_holder=&units[0];
    if(which==3 || which==12) members[0]=0;
    if(which==5) {np=cached=3;rolls[0]=2;uids[0]=2;uids[1]=255;uids[2]=1;}
    if(which==6) {np=cached=4;roll_count=3;rolls[1]=1;rolls[2]=2;players[0].playerUnit=NULL;uids[2]=1;}
    if(which==7) {nt=np=cached=3;roll_count=3;rolls[1]=1;rolls[2]=2;}
    if(which==9) {nt=np=cached=2;flags[1]=&crowns[0];}
    if(which==11) cached=INT32_MIN;
    if(which==13) flags[0]=NULL;
    if(which==14) np=cached=0;
    if(which==15) {cached=3;rolls[0]=2;}
    for(int i=0;i<4;i++) ((nox_object_team_t*)&units[i].field_12)->id=uids[i];
    memcpy(expected_teams,teams,sizeof(teams));memcpy(expected_players,players,sizeof(players));
    memcpy(expected_units,units,sizeof(units));memcpy(expected_crowns,crowns,sizeof(crowns));
    sub_4D2160();
    assert(which<13);assert(memcmp(expected_teams,teams,sizeof(teams))==0);
    assert(memcmp(expected_players,players,sizeof(players))==0);
    assert(memcmp(expected_units,units,sizeof(units))==0);
    assert(memcmp(expected_crowns,crowns,sizeof(crowns))==0);
    assert(fwrite(transcript,1,used,stdout)==used);return 0;
}
