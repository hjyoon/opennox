#include <assert.h>
#include <stdio.h>
#include "GAME1.h"
#include "common__system__team.h"

static nox_team_t teams[16];
static int count,position,first_calls,next_calls;

nox_team_t* nox_server_teamFirst_418B10(void) {
    first_calls++;position=0;return count ? &teams[0] : NULL;
}
nox_team_t* nox_server_teamNext_418B60(nox_team_t* team) {
    if (team!=&teams[position]) {
        fprintf(stderr,"registered team count truncated native iterator\n");abort();
    }
    next_calls++;position++;
    return position<count ? &teams[position] : NULL;
}

/* COMPLETE_COUNT_BODY */

int main(void) {
    unsigned patterns[3]={0,0x5a,0xff};
    int lengths[4]={0,1,3,16};
    int cases=0;
    if (sizeof(void*)>4) assert((uintptr_t)&teams[0]>UINT32_MAX);
    for (int p=0;p<3;p++) for (int n=0;n<4;n++) for (int mode=0;mode<4;mode++) {
        count=lengths[n];first_calls=next_calls=position=0;
        memset(teams,patterns[p],sizeof(teams));
        int wanted=0;
        for (int i=0;i<count;i++) {
            uint32_t value=0;
            if (mode==1 && i%2==0) value=1;
            else if (mode==2) value=0x80000000;
            else if (mode==3) value=0xffffffff;
            teams[i].field_60=value;
            if (value) wanted++;
        }
        nox_team_t before[16];memcpy(before,teams,sizeof(teams));
        unsigned result=(unsigned char)sub_417DE0();
        assert(result==(unsigned)wanted && first_calls==1 && next_calls==count);
        assert(memcmp(before,teams,sizeof(teams))==0);
        cases++;
    }
    printf("complete native registered team count: %d iterator/scalar/guard cases passed\n",cases);
    return 0;
}
