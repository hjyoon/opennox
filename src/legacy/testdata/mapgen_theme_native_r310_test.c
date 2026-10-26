#include <ctype.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include <time.h>
#include <limits.h>

#include "mapgen_theme_native.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_legacy_ptr.c"

#define CHECK(x) do { if (!(x)) { fprintf(stderr, "check failed line %d: %s\n", __LINE__, #x); exit(90); } } while (0)
struct guarded_theme { uint32_t before[8]; uint32_t words[279]; uint32_t after[8]; };
static struct guarded_theme* storage;
static uint8_t* theme;
static FILE* file_handle;
static char* theme_name;
static uint8_t keyword[256], packed_slots[12];
static const char* skeleton_slots[3];
static unsigned slot_reads[32], slot_read_count;
static uint32_t global_reset;
static const char* const* stream;
static unsigned stream_count, stream_index, read_calls;
static int capacity_guard, allow_open, allow_crypto, settings_ok;
static unsigned open_calls, crypto_calls, close_calls, setting_calls, formatter_calls, lookup_calls;
static unsigned dispatch_calls[4], case_count;
static char observed_path[256], expected_path[256], lookup_names[8][60];

static void reset_stream(const char* const* tokens, unsigned count) {
 stream=tokens;stream_count=count;stream_index=read_calls=slot_read_count=0;
 memset(keyword,0,sizeof(keyword));memset(slot_reads,0,sizeof(slot_reads));
}
static void check_guards(void) {
 for (unsigned i=0;i<8;++i) CHECK(storage->before[i]==UINT32_C(0xA5A5A5A5) && storage->after[i]==UINT32_C(0xA5A5A5A5));
}
static void reset_theme(void) {memset(storage,0xA5,sizeof(*storage));memset(theme,0,1116);}
static void reset_services(void) {
 open_calls=crypto_calls=close_calls=setting_calls=formatter_calls=lookup_calls=0;
 memset(dispatch_calls,0,sizeof(dispatch_calls));memset(lookup_names,0,sizeof(lookup_names));
 allow_open=allow_crypto=settings_ok=1;global_reset=0xdeadbeef;
}
static uint8_t* getMemAt(uintptr_t base, uintptr_t offset) {
 if (base==0x5D4594 && offset==2487264) return keyword;
 if (base==0x587000 && offset>=253244 && offset<253256) return packed_slots+offset-253244;
 CHECK(0);return NULL;
}
static uint32_t* getMemU32Ptr(uintptr_t base, uintptr_t offset) {
 if (base==0x5D4594 && offset==2487520) return &global_reset;
 return (uint32_t*)getMemAt(base,offset);
}
static void* getMemPtr(uintptr_t base, uintptr_t offset) {
 CHECK(base==0x587000 && offset>=253244 && offset<=253252 && (offset-253244)%4==0);
 CHECK(slot_read_count<32);slot_reads[slot_read_count++]=(unsigned)(offset-253244)/4;
 return (void*)skeleton_slots[(offset-253244)/4];
}
static int nox_strcmpi(const char* a, const char* b) {return strcasecmp(a,b);}
static time_t fixed_time(time_t* out) {if (out) *out=(time_t)1700000000;return (time_t)1700000000;}
static int nox_sprintf(char* dest,const char* fmt,...) {
 va_list ap;va_start(ap,fmt);const char* arg=va_arg(ap,const char*);va_end(ap);
 if (!strcmp(fmt,"mapgen/%s.thm")) {
  ++formatter_calls;
  if (arg!=theme_name) {fprintf(stderr,"NAME_WIDTH: received=%p expected=%p\n",(void*)arg,(void*)theme_name);exit(41);}
  return sprintf(dest,"mapgen/%s.thm",arg);
 }
 CHECK(!strcmp(fmt,"SPELL_%s"));return sprintf(dest,"SPELL_%s",arg);
}
static FILE* nox_binfile_open_408CC0(char* path,int mode) {
 ++open_calls;CHECK(mode==0 && !strcmp(path,expected_path));strcpy(observed_path,path);
 return allow_open?file_handle:NULL;
}
static int nox_binfile_cryptSet_408D40(FILE* file,int key) {
 ++crypto_calls;CHECK(key==1);
 if (file!=file_handle) {fprintf(stderr,"FILE_WIDTH: received=%p expected=%p\n",(void*)file,(void*)file_handle);exit(42);}
 return allow_crypto;
}
static int nox_binfile_close_408D90(FILE* file) {++close_calls;CHECK(file==file_handle);return 1;}
static int fixture_read_line(FILE* file,uint8_t* dest,size_t capacity) {
 CHECK(file==file_handle && ++read_calls<=512);
 if (stream_index==stream_count) return 0;
 const char* value=stream[stream_index++];size_t bytes=strlen(value)+1;
 CHECK(bytes<=60 || dest==keyword);
 if (capacity_guard && capacity!=(size_t)-1 && capacity<bytes) {fprintf(stderr,"VALUE_CAPACITY: capacity=%zu required=%zu\n",capacity,bytes);exit(43);}
 memcpy(dest,value,bytes);return 1;
}
#define nox_xxx_mapGenReadLine_51E540(file,dest) fixture_read_line((file),(uint8_t*)(dest),__builtin_object_size((dest),0))
static int nox_xxx_spellNameToN_4243F0(const char* name) {
 CHECK(lookup_calls<8);strcpy(lookup_names[lookup_calls++],name);
 return !strcmp(name,"SPELL_FIREBALL")?42:0;
}
static int nox_xxx_mapgenCheckSettings_520AD0(int* settings) {
 ++setting_calls;CHECK(settings==(int*)theme+22 || settings==(int*)theme+30 || settings==(int*)theme+46);
 return settings_ok;
}
static int dispatch_stub(unsigned kind,uint8_t* value,FILE* file) {
 CHECK(value==theme && file==file_handle);++dispatch_calls[kind];
 uint8_t token[60];CHECK(nox_xxx_mapGenReadLine_51E540(file,token));CHECK(!strcmp((char*)token,"END"));return 1;
}
static int nox_xxx_genReadPrefab_520BF0(uint8_t* value,FILE* file) {return dispatch_stub(0,value,file);}
static int nox_xxx_genReadDecor_51F9F0(uint32_t* value,FILE* file) {return dispatch_stub(1,(uint8_t*)value,file);}
static int nox_xxx_genReadWeaponSet_51F030(uint8_t* value,FILE* file) {return dispatch_stub(2,value,file);}
static int nox_xxx_genReadArmorSet_51F640(uint8_t* value,FILE* file) {return dispatch_stub(3,value,file);}
int nox_xxx_genReadAlgData_51EBB0(int,FILE*);
int nox_xxx_genReadSpellSet_51EFB0(int,FILE*);
int nox_xxx_genReadExit_51F800(int,FILE*);
#define time fixed_time
// PRODUCTION_THEME_PARSERS_R310
#undef time

static uint32_t high_token(void* value, void** collision) {
#if UINTPTR_MAX > UINT32_MAX
 uint32_t low=(uint32_t)(uintptr_t)value;*collision=(void*)(uintptr_t)low;
 CHECK(nox_mapgenLegacyPtrRegister(*collision)==low);
 uint32_t token=nox_mapgenLegacyPtrRegister(value);CHECK(token>=UINT32_C(0xF0000000));return token;
#else
 *collision=NULL;uint32_t token=nox_mapgenLegacyPtrRegister(value);CHECK(token!=0);return token;
#endif
}
static void forget_token(void* value,void* collision) {nox_mapgenLegacyPtrForget(value);nox_mapgenLegacyPtrForget(collision);}
static int read_theme(int legacy) {
 if (!legacy) return nox_mapgenReadThemeNative_51E260((int*)theme,theme_name);
 void* collision;uint32_t token=high_token(theme_name,&collision);
 int result=nox_xxx_mapGenReadTheme_51E260((int*)theme,(int)token);
 forget_token(theme_name,collision);return result;

}
static int read_alg(int legacy) {
 if (!legacy) return nox_mapgenReadAlgDataNative_51EBB0(theme,file_handle);
 void* collision;uint32_t token=high_token(theme,&collision);
 int result=nox_xxx_genReadAlgData_51EBB0((int)token,file_handle);
 forget_token(theme,collision);return result;

}
static int read_spells(int legacy) {
 if (!legacy) return nox_mapgenReadSpellSetNative_51EFB0(theme,file_handle);
 void* collision;uint32_t token=high_token(theme,&collision);int r=nox_xxx_genReadSpellSet_51EFB0((int)token,file_handle);forget_token(theme,collision);return r;

}
static int read_exit(int legacy) {
 if (!legacy) return nox_mapgenReadExitNative_51F800(theme,file_handle);
 void* collision;uint32_t token=high_token(theme,&collision);int r=nox_xxx_genReadExit_51F800((int)token,file_handle);forget_token(theme,collision);return r;

}
static void default_words(uint32_t expected[279]) {
 memset(expected,0,1116);
 const unsigned indices[]={3,5,1,2,4,6,7,8,9,10,16,18,19,14,12,13};
 const uint32_t values[]={3,3,5,2,1,40,20,10,5,20,1157234688,100,1700000000,1,100,100};
 for (unsigned i=0;i<sizeof(indices)/sizeof(indices[0]);++i) expected[indices[i]]=values[i];
}
static void boundary(int legacy,int open_ok,int crypto_ok,int settings_valid) {
 reset_theme();reset_services();reset_stream(NULL,0);allow_open=open_ok;allow_crypto=crypto_ok;settings_ok=settings_valid;
 uint32_t expected[279];default_words(expected);int result=read_theme(legacy);
 CHECK(result==(open_ok && crypto_ok && settings_valid));CHECK(!memcmp(theme,expected,1116));CHECK(global_reset==0);
 CHECK(formatter_calls==1 && open_calls==1 && crypto_calls==(unsigned)open_ok);
 CHECK(close_calls==(unsigned)(open_ok && crypto_ok));
 CHECK(read_calls==(unsigned)(open_ok && crypto_ok));
 CHECK(setting_calls==(unsigned)(open_ok && crypto_ok ? (settings_valid?2:1):0));
 check_guards();nox_mapgenLegacyPtrForget(theme);++case_count;
}
static void algorithm_cases(void) {
 const char* end[]={"END"};const char* incomplete[]={"midHallLength"};const char* partial[]={"midHallLength","17"};
 for (int legacy=0;legacy<2;++legacy) {
  reset_theme();reset_services();reset_stream(end,1);CHECK(read_alg(legacy)==1 && stream_index==1);CHECK(storage->words[1]==0);check_guards();++case_count;
  reset_stream(NULL,0);CHECK(read_alg(legacy)==0 && read_calls==1);++case_count;
  reset_stream(incomplete,1);CHECK(read_alg(legacy)==0 && read_calls==2 && storage->words[1]==0);++case_count;
  reset_stream(partial,2);CHECK(read_alg(legacy)==0 && read_calls==3 && storage->words[1]==17);check_guards();++case_count;
  char value[60];
  for (unsigned length=1;length<=59;++length) {
   memset(value,'X',length);value[length]=0;const char* tokens[]={"ignored",value,"END"};uint8_t before[1116];memcpy(before,theme,1116);
   reset_stream(tokens,3);CHECK(read_alg(legacy)==1 && stream_index==3);CHECK(!memcmp(before,theme,1116));check_guards();++case_count;
  }
  const struct {const char* key;unsigned offset;} fields[]={
   {"midHallLength",4},{"hallLengthVariance",8},{"midHallWidth",12},{"hallWidthVariance",16},{"hallLimit",20},{"hallBranchRate",24},{"hallRoomRate",28},{"midRoomSize",32},{"roomVariance",36},{"irregularRoomRate",40},{"emptyRoomRate",44},{"mergeRate",48},{"adjacentPortalRate",52},{"recursionLimit",72},{"seed",76}};
  for (unsigned i=0;i<sizeof(fields)/sizeof(fields[0]);++i) {
   reset_theme();const char* tokens[]={fields[i].key,"-17","END"};uint8_t expected[1116]={0};uint32_t want=UINT32_C(0xFFFFFFEF);memcpy(expected+fields[i].offset,&want,4);
   reset_stream(tokens,3);CHECK(read_alg(legacy)==1 && !memcmp(theme,expected,1116));check_guards();++case_count;
  }
  const struct {const char* text;uint32_t bits;} floats[]={{"2.75",0x40300000},{"-0.125",0xBE000000},{"1e3",0x447A0000}};
  for (unsigned i=0;i<3;++i) {reset_theme();const char* tokens[]={"mapSize",floats[i].text,"END"};reset_stream(tokens,3);CHECK(read_alg(legacy)==1 && storage->words[16]==floats[i].bits);check_guards();++case_count;}
  const char* booleans[]={"true","TRUE","True","false","1","yes"};
  for (unsigned field=0;field<2;++field) for (unsigned i=0;i<6;++i) {
   reset_theme();const char* tokens[]={field?"debug":"useDoors",booleans[i],"END"};reset_stream(tokens,3);CHECK(read_alg(legacy)==1 && storage->words[field?15:14]==(i<3?1u:0u));check_guards();++case_count;
  }
  const char* values[]={"NONE","hall_ring","UNKNOWN"};
  for (unsigned i=0;i<3;++i) {reset_theme();const char* tokens[]={"skeleton",values[i],"END"};reset_stream(tokens,3);CHECK(read_alg(legacy)==1 && storage->words[0]==(i==1?1u:0u));CHECK(slot_read_count==i+1);for(unsigned j=0;j<slot_read_count;++j) CHECK(slot_reads[j]==j);check_guards();++case_count;}
  const char* first=skeleton_slots[0];skeleton_slots[0]=NULL;reset_theme();const char* unknown[]={"skeleton","HALL_RING","END"};reset_stream(unknown,3);CHECK(read_alg(legacy)==1 && storage->words[0]==0 && slot_read_count==1);skeleton_slots[0]=first;check_guards();++case_count;
 }
}
static void spell_cases(void) {
 for (int legacy=0;legacy<2;++legacy) {
  reset_theme();reset_services();const char* tokens[]={"fireball","unknown","fireball","END"};reset_stream(tokens,4);CHECK(read_spells(legacy)==1 && storage->words[274]==2 && storage->words[137]==42 && storage->words[138]==42);CHECK(lookup_calls==3 && !strcmp(lookup_names[1],"SPELL_UNKNOWN"));check_guards();++case_count;
  reset_theme();reset_services();storage->words[274]=136;const char* last[]={"fireball","END"};reset_stream(last,2);CHECK(read_spells(legacy)==1 && storage->words[274]==137 && storage->words[273]==42 && lookup_calls==1);check_guards();++case_count;
  reset_theme();reset_services();storage->words[274]=137;reset_stream(last,2);CHECK(read_spells(legacy)==1 && storage->words[274]==137 && stream_index==1 && lookup_calls==0);check_guards();++case_count;
  reset_theme();reset_services();const char* end[]={"END"};reset_stream(end,1);CHECK(read_spells(legacy)==1 && stream_index==1 && lookup_calls==0);++case_count;
  reset_stream(NULL,0);CHECK(read_spells(legacy)==0);check_guards();++case_count;
 }
}
static void exit_cases(void) {
 for (int legacy=0;legacy<2;++legacy) {
  reset_theme();const char* tokens[]={"OBJECT","ExitN","NORTH","OBJECT","ExitS","SOUTH","OBJECT","ExitE","EAST","OBJECT","ExitW","WEST","LINKDATA","Gate42","END"};reset_stream(tokens,15);
  CHECK(read_exit(legacy)==1 && storage->words[118]==4 && !strcmp((char*)theme+476,"Gate42"));
  const char* names[]={"ExitN","ExitS","ExitE","ExitW"};for(unsigned i=0;i<4;++i) {CHECK(!strcmp((char*)theme+216+64*i,names[i]));uint32_t direction;memcpy(&direction,theme+276+64*i,4);CHECK(direction==i);}check_guards();++case_count;
  reset_theme();storage->words[69]=0xDEADBEEF;const char* unknown[]={"OBJECT","ExitX","UNKNOWN","END"};reset_stream(unknown,4);CHECK(read_exit(legacy)==1 && storage->words[118]==1 && storage->words[69]==0xDEADBEEF);check_guards();++case_count;
  reset_theme();const char* partial[]={"OBJECT","ExitN"};reset_stream(partial,2);CHECK(read_exit(legacy)==0 && !strcmp((char*)theme+216,"ExitN") && storage->words[118]==0);check_guards();++case_count;
  reset_theme();const char* bad[]={"INVALID"};reset_stream(bad,1);CHECK(read_exit(legacy)==0 && stream_index==1);check_guards();++case_count;
  reset_theme();const char* end[]={"END"};reset_stream(end,1);CHECK(read_exit(legacy)==1);reset_stream(NULL,0);CHECK(read_exit(legacy)==0);check_guards();++case_count;
 }
}
static void theme_dispatch(int legacy) {
 reset_theme();reset_services();
 const char* tokens[]={"ALGORITHM_DATA","midHallLength","17","END","AMBIENT_LIGHT","12","34","56","EXIT","LINKDATA","Gate42","END","SPELL_SET","fireball","END","PREFAB","END","WEAPON_SET","END","ARMOR_SET","END","DECOR","END"};
 reset_stream(tokens,23);CHECK(read_theme(legacy)==1 && stream_index==23);
 CHECK(storage->words[1]==17 && storage->words[134]==12 && storage->words[135]==34 && storage->words[136]==56);
 CHECK(!strcmp((char*)theme+476,"Gate42") && storage->words[274]==1 && storage->words[137]==42);
 for (unsigned i=0;i<4;++i) CHECK(dispatch_calls[i]==1);
 CHECK(close_calls==1 && setting_calls==2);check_guards();nox_mapgenLegacyPtrForget(theme);++case_count;
 reset_theme();reset_services();const char* incomplete[]={"AMBIENT_LIGHT","12","34"};reset_stream(incomplete,3);CHECK(read_theme(legacy)==0 && storage->words[134]==12 && storage->words[135]==34 && storage->words[136]==0 && close_calls==1 && setting_calls==0);check_guards();nox_mapgenLegacyPtrForget(theme);++case_count;
}
int main(int argc,char** argv) {
 CHECK(argc==2);storage=malloc(sizeof(*storage));file_handle=malloc(sizeof(*file_handle));theme_name=malloc(245);theme=(uint8_t*)storage->words;
 CHECK(storage && file_handle && theme_name);
#if UINTPTR_MAX>UINT32_MAX
 CHECK((uintptr_t)theme>UINT32_MAX && (uintptr_t)file_handle>UINT32_MAX && (uintptr_t)theme_name>UINT32_MAX);
#endif
 strcpy(theme_name,"audit_theme");sprintf(expected_path,"mapgen/%s.thm",theme_name);
 skeleton_slots[0]=strdup("NONE");skeleton_slots[1]=strdup("HALL_RING");skeleton_slots[2]=NULL;CHECK(skeleton_slots[0] && skeleton_slots[1]);memset(packed_slots,0xA5,sizeof(packed_slots));
 CHECK(!strcmp(argv[1],"capacity") || !strcmp(argv[1],"native"));
 capacity_guard=!strcmp(argv[1],"capacity");
 {
  for (int legacy=0;legacy<2;++legacy) {
   for(int op=0;op<2;++op) for(int cr=0;cr<2;++cr) for(int valid=0;valid<2;++valid) boundary(legacy,op,cr,valid);
   memset(theme_name,'Q',244);theme_name[244]=0;sprintf(expected_path,"mapgen/%s.thm",theme_name);boundary(legacy,1,1,1);
   strcpy(theme_name,"audit_theme");sprintf(expected_path,"mapgen/%s.thm",theme_name);theme_dispatch(legacy);
  }
  algorithm_cases();spell_cases();exit_cases();
 }
 CHECK(!memcmp(packed_slots,(uint8_t[12]){0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5,0xA5},12));
 free((void*)skeleton_slots[0]);free((void*)skeleton_slots[1]);free(theme_name);free(file_handle);free(storage);
 printf("Theme native R310: %u bounded cases; native pointers, registry adapters, original scalar/layout/partial-write contracts; no complete generator claim.\n",case_count);return 0;
}
