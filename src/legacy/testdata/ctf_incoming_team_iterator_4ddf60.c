#include <assert.h>
#include <stdio.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5_2.h"
#include "common__system__team.h"

static nox_playerInfo player;
static nox_object_t unit;
static nox_team_t teams[3];
static nox_game_ball_status_t ball;
static nox_team_flag_status_t flags[3];
static unsigned game_flags;
static int player_index,team_count,team_position,flag_calls,ball_calls,next_calls;
static char events[1024];
uint32_t dword_5d4594_2649712;
uint32_t nox_server_sendMotd_108752;

static void event(const char* name) {
    assert(strlen(events)+strlen(name)+2<sizeof(events));
    if (events[0]) strcat(events,"|");
    strcat(events,name);
}
bool nox_common_gameFlags_check_40A5C0(unsigned f) { return (game_flags & f)!=0; }
nox_playerInfo* nox_common_playerInfoFromNum_417090(int index) {
    assert(index==player_index);event("lookup");return &player;
}
int sub_57B920(void* data) { assert(data==player.netData16);event("setup");return 0; }
char* nox_xxx_newPlayerSendAllPlayers_4DE300(int index) { assert(index==player_index);event("all-players");return NULL; }
static void init_unit(nox_object_t* object,uint32_t arg) { assert(object==&unit && arg==0);event("init-unit"); }
void nox_xxx_playerSendMOTD_4DD140(int index) { assert(index==player_index);event("motd"); }
nox_playerInfo* nox_common_playerInfoGetFirst_416EA0(void) { return NULL; }
nox_playerInfo* nox_common_playerInfoGetNext_416EE0(nox_playerInfo* p) { abort(); }
int nox_xxx_servMinimapRevealFlag_4DE380(int index) { assert(index==player_index);event("reveal");return 0; }
void sub_4DF2E0(int index) { assert(index==player_index);event("teams"); }
int sub_40AA70(nox_playerInfo* p) { assert(p==&player);event("elim-check");return 0; }
int nox_xxx_netNeedTimestampStatus_4174F0(nox_playerInfo* p,int mask) { assert(p==&player && mask==256);event("timestamp");return 0; }
nox_object_t* sub_4E8110(int32_t index) { assert(index==player_index);event("state-init");return &unit; }
nox_team_t* nox_server_teamFirst_418B10(void) { event("first");team_position=0;return team_count ? &teams[0] : NULL; }
nox_team_t* nox_server_teamNext_418B60(nox_team_t* team) {
    if (team!=&teams[team_position]) {
        fprintf(stderr,"CTF incoming team pointer changed/truncated\n");abort();
    }
    next_calls++;event("next");team_position++;
    return team_position<team_count ? &teams[team_position] : NULL;
}
nox_team_flag_status_t* sub_4E8320(uint8_t id) {
    assert(team_position<team_count && id==teams[team_position].field_57);
    event("flag-read");return &flags[team_position];
}
int nox_xxx_netSendFlagStatus_4D95A0(int index,char id,char state,char flag_index,short carrier) {
    assert(index==player_index && flag_calls<team_count);
    nox_team_flag_status_t* wanted=&flags[flag_calls];
    assert((uint8_t)id==wanted->team_id && (uint8_t)state==wanted->status &&
        (uint8_t)flag_index==wanted->flag_index && (uint16_t)carrier==wanted->carrier_net_code);
    flag_calls++;event("flag-send");return 0;
}
nox_game_ball_status_t* sub_4E8310(void) { event("ball-read");return &ball; }
int nox_xxx_netSendBallStatus_4D95F0(int index,char state,short code) {
    assert(index==player_index && (uint8_t)state==ball.state && (uint16_t)code==ball.net_code);
    ball_calls++;event("ball-send");return 0;
}
char* nox_xxx_sendAllClientStatus_4175C0(nox_playerInfo* p) { assert(p==&player);event("all-status");return NULL; }
int sub_409F40(int mode) { assert(mode==0x2000);event("ids-check");return (game_flags & mode)!=0; }
int nox_xxx_sendAllPlayerIDs_4DE270(nox_playerInfo* p) { assert(p==&player);event("ids-send");return 0; }

/* These branches are outside the objective-mode cases; reaching one is an
 * error, not a permissive replacement of another gameplay path. */
void nox_xxx_playerIncomingQuest_native_4DDF60(nox_object_t* unit) { abort(); }
int sub_4D9CF0(int index) { abort(); }
int sub_4D6000(nox_object_t* p) { abort(); }
void nox_xxx_netMarkMinimapObject_417190(int index,nox_object_t* p,unsigned arg) { abort(); }
int nox_xxx_netSendSimpleObject2_4DF360(int index,nox_object_t* p) { abort(); }
int nox_xxx_netSendTeam_4D8670(int index,uint32_t* p) { abort(); }
nox_object_t* nox_xxx_getFirstPlayerUnit_4DA7C0(void) { abort(); }
nox_object_t* nox_xxx_getNextPlayerUnit_4DA7F0(const nox_object_t* p) { abort(); }
int sub_4D9D20(int index,nox_object_t* p) { abort(); }
int sub_4D7A60(int index) { abort(); }

/* COMPLETE_INCOMING_BODY */

static void run_case(unsigned mode,int count,int index,unsigned pattern) {
    game_flags=mode;team_count=count;player_index=index;
    team_position=flag_calls=ball_calls=next_calls=0;events[0]=0;
    memset(&player,pattern,sizeof(player));memset(&unit,pattern,sizeof(unit));
    memset(teams,pattern,sizeof(teams));memset(flags,pattern,sizeof(flags));
    player.playerUnit=&unit;player.playerInd=index;
    player.field_4700=0x12345678;player.field_3676=0x87;
    player.pos_x_3632=71;player.pos_y_3636=73;
    unit.func_init=(void*)init_unit;unit.x=128.5f;unit.y=-63.25f;unit.net_code=0xabcd;
    const unsigned char ids[3]={1,9,16};
    for (int i=0;i<3;i++) {
        teams[i].field_57=ids[i];teams[i].ind=i+1;
        flags[i].team_id=0xe0+i;flags[i].flag_index=0xa0+i;
        flags[i].status=0x80+i;flags[i].carrier_net_code=0xf010+i;
    }
    ball.net_code=0xf123;ball.state=0xe7;
    if (sizeof(void*)>4) assert((uintptr_t)&teams[0]>UINT32_MAX);
    nox_playerInfo expected=player;
    expected.field_4700=0;expected.field_3676=3;
    if (!(mode & 512)) { expected.pos_x_3632=unit.x;expected.pos_y_3636=unit.y; }
    nox_object_t unit_before=unit;nox_team_t teams_before[3];
    nox_team_flag_status_t flags_before[3];nox_game_ball_status_t ball_before=ball;
    memcpy(teams_before,teams,sizeof(teams));memcpy(flags_before,flags,sizeof(flags));
    nox_server_sendMotd_108752=1;dword_5d4594_2649712=0x02468000;
    uint32_t mask_before=dword_5d4594_2649712;
    int result=nox_xxx_netPlayerIncomingServ_4DDF60(index);
    assert(result==0xabcd && dword_5d4594_2649712==(mask_before | (UINT32_C(1)<<index)));
    assert(memcmp(&player,&expected,sizeof(player))==0);
    assert(memcmp(&unit,&unit_before,sizeof(unit))==0 && memcmp(teams,teams_before,sizeof(teams))==0);
    assert(memcmp(flags,flags_before,sizeof(flags))==0 && memcmp(&ball,&ball_before,sizeof(ball))==0);
    char wanted[1024]="lookup|setup|all-players|init-unit";
    if (mode & 0x2000) strcat(wanted,"|motd");
    strcat(wanted,"|reveal|teams");
    if (mode & 1024) strcat(wanted,"|elim-check|timestamp");
    strcat(wanted,"|state-init");
    if (mode & 64) {
        strcat(wanted,"|ball-read|ball-send");
        assert(ball_calls==1 && flag_calls==0 && next_calls==0);
    } else if (mode & 32) {
        strcat(wanted,"|first");
        for (int i=0;i<count;i++) strcat(wanted,"|flag-read|flag-send|next");
        assert(ball_calls==0 && flag_calls==count && next_calls==count);
    } else assert(ball_calls==0 && flag_calls==0 && next_calls==0);
    strcat(wanted,"|all-status|ids-check");
    if (mode & 0x2000) strcat(wanted,"|ids-send");
    assert(strcmp(events,wanted)==0);
}
int main(int argc,char** argv) {
    assert(argc==2);
    unsigned patterns[3]={0,0x5a,0xff};
    int cases=0;
    for (int p=0;p<3;p++) {
        if (strcmp(argv[1],"ctf")==0) {
            for (int count=0;count<=3;count+=count==0 ? 1 : 2) {
                run_case(32,count,0,patterns[p]);cases++;
                run_case(32|0x2000|512,count,31,patterns[p]);cases++;
            }
        } else {
            assert(strcmp(argv[1],"unchanged")==0);
            unsigned modes[5]={0,64,64|32,64|32|512|0x2000,1024|0x2000};
            for (int m=0;m<5;m++) { run_case(modes[m],3,7,patterns[p]);cases++; }
        }
    }
    printf("complete native incoming body: %d %s cases passed, native pointers and original status/tail order preserved\n",cases,argv[1]);
    return 0;
}
