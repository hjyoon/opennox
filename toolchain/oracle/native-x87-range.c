// Independent hardware reference for GAME.EXE 00534470.
// On i386, execute all 20 unchanged bytes with PE32 records and the cdecl ABI.
// On x86_64, execute its unchanged three-byte FLD as a separate control.
// No Go implementation or computed floating-point expected value is used.
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

#if !defined(__i386__) && !defined(__x86_64__)
#error This reference requires native x86 instructions.
#endif

static const unsigned char original[] = {
	0x8b,0x44,0x24,0x04,0x8b,0x88,0xec,0x02,0x00,0x00,
	0x8b,0x91,0xe4,0x01,0x00,0x00,0xd9,0x42,0x70,0xc3
};

static void fail(const char *message) {
	fprintf(stderr,"native x87 reference: %s\n",message);
	exit(1);
}

static int compare_u32(const void *left,const void *right) {
	uint32_t a=*(const uint32_t *)left,b=*(const uint32_t *)right;
	return (a>b)-(a<b);
}

int main(void) {
	uint32_t inputs[119]={
		0,0x80000000,1,0x80000001,0x007fffff,0x807fffff,0x00800000,0x80800000,
		0x3f800000,0xbf800000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,
		0x7fc00000,0xffc00000,0x7fc12345,0xffc12345,0x7f800001,0xff800001
	};
	const uint32_t centers[]={0x3da3d70a,0x3ea8f5c3,0x3f28f5c3};
	const uint16_t controls[]={0x0e3f,0x023f};
	size_t count=20;
	for (size_t i=0;i<sizeof(centers)/sizeof(centers[0]);i++) {
		for (int delta=-16;delta<=16;delta++) inputs[count++]=centers[i]+delta;
	}
	if (count!=119) fail("input corpus length");
	qsort(inputs,count,sizeof(inputs[0]),compare_u32);
	for (size_t i=1;i<count;i++) if(inputs[i]==inputs[i-1]) fail("duplicate input");
	unsigned char unit[1024],update[2304],definition[256];
	memset(unit,0xa5,sizeof(unit));memset(update,0x5a,sizeof(update));
	memset(definition,0x39,sizeof(definition));
	const long page=sysconf(_SC_PAGESIZE);
	if(page<=0) fail("page size");
	void *code=mmap(NULL,(size_t)page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
	if(code==MAP_FAILED) fail("allocate original code");
	memcpy(code,original,sizeof(original));
	if(mprotect(code,(size_t)page,PROT_READ|PROT_EXEC)) fail("protect original code");
#if defined(__i386__)
	uint32_t update_address=(uint32_t)(uintptr_t)update;
	uint32_t definition_address=(uint32_t)(uintptr_t)definition;
	memcpy(unit+748,&update_address,4);memcpy(update+484,&definition_address,4);
	const char *mode="full-original-pe32";
	const unsigned executed_bytes=20;
#else
	const char *mode="original-fld-only-control";
	const unsigned executed_bytes=3;
#endif
	printf("{\"schema\":\"native-original-x87-range-v1\",\"pointer_bits\":%u,\"mode\":\"%s\",\"entry\":\"0x534470\",\"executed_instruction_bytes\":%u,\"original_hex\":\"",
	       (unsigned)(8*sizeof(void *)),mode,executed_bytes);
	for(size_t i=0;i<sizeof(original);i++) printf("%02x",original[i]);
	printf("\",\"controls\":2,\"inputs_per_control\":119}\n");
	uint16_t saved_control;
	__asm__ volatile("fnstcw %0":"=m"(saved_control));
	for(size_t c=0;c<sizeof(controls)/sizeof(controls[0]);c++) {
		for(size_t i=0;i<count;i++) {
			memcpy(definition+112,&inputs[i],4);
			unsigned char before_unit[sizeof(unit)],before_update[sizeof(update)],before_def[sizeof(definition)];
			memcpy(before_unit,unit,sizeof(unit));memcpy(before_update,update,sizeof(update));
			memcpy(before_def,definition,sizeof(definition));
			uint16_t returned_status=0,stored_status=0,loaded_control=0,after_control=0;
			uint64_t result=0;
			__asm__ volatile("fninit\n\tfldcw %1\n\tfnstcw %0":"=m"(loaded_control):"m"(controls[c]):"st","memory");
			// Reserved CW bit 6 reads as one on hardware; it is not a mode.
			if((loaded_control & (uint16_t)~0x0040)!=controls[c]) fail("requested control mode not loaded");
#if defined(__i386__)
			__asm__ volatile(
				"pushl %[unit]\n\tcall *%[code]\n\taddl $4,%%esp\n\t"
				"fnstsw %[returned]\n\tfstpl %[result]\n\tfnstsw %[stored]\n\tfnstcw %[control]"
				: [returned]"=m"(returned_status),[result]"=m"(result),
				  [stored]"=m"(stored_status),[control]"=m"(after_control)
				: [unit]"r"(unit),[code]"r"(code)
				: "eax","ecx","edx","cc","memory","st");
#else
			__asm__ volatile(
				".byte 0xd9,0x42,0x70\n\tfnstsw %[returned]\n\t"
				"fstpl %[result]\n\tfnstsw %[stored]\n\tfnstcw %[control]"
				: [returned]"=m"(returned_status),[result]"=m"(result),
				  [stored]"=m"(stored_status),[control]"=m"(after_control)
				: "d"(definition):"memory","st");
#endif
			if(after_control!=loaded_control || ((returned_status>>11)&7)!=7 || ((stored_status>>11)&7)!=0)
			{
				fprintf(stderr,"requested-control=%04x actual-control=%04x returned-status=%04x stored-status=%04x input=%08" PRIx32 "\n",controls[c],after_control,returned_status,stored_status,inputs[i]);
				fail("control word or x87 stack depth changed");
			}
			if(memcmp(unit,before_unit,sizeof(unit)) || memcmp(update,before_update,sizeof(update)) ||
			   memcmp(definition,before_def,sizeof(definition)) || memcmp(code,original,sizeof(original)))
				fail("original code or input records changed");
			printf("{\"control\":\"%04x\",\"loaded_control\":\"%04x\",\"input\":\"%08" PRIx32 "\",\"output\":\"%016" PRIx64 "\",\"returned_status\":\"%04x\",\"stored_status\":\"%04x\"}\n",
			       controls[c],loaded_control,inputs[i],result,returned_status,stored_status);
		}
	}
	__asm__ volatile("fninit\n\tfldcw %0"::"m"(saved_control):"st","memory");
	if(munmap(code,(size_t)page)) fail("unmap original code");
	return ferror(stdout)?1:0;
}
