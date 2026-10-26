#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int dword_5d4594_2489420;
static uint32_t return_opcode;
static unsigned code_calls;

static int nox_fs_fread(FILE* file, void* data, size_t size) {
	size_t n=fread(data,1,size,file);
	assert(n==size);
	return (int)n;
}
static int nox_fs_fwrite(FILE* file, const void* data, size_t size) {
	size_t n=fwrite(data,1,size,file);
	assert(n==size);
	return (int)n;
}
static int nox_fs_fseek_cur(FILE* file, int32_t offset) {
	return fseek(file,offset,SEEK_CUR);
}

/* The child bytecode service is declared: these inputs contain one RETURN
 * instruction per GLOBAL function. All metadata and name copies execute in
 * the complete production 541E40 body with real binary FILE streams. */
static int nox_script_readWriteYyy_542380(FILE* src, FILE* dst, int relocated) {
	uint32_t op;
	assert(relocated==(int)code_calls);
	code_calls++;
	nox_fs_fread(src,&op,4);
	assert(op==return_opcode);
	nox_fs_fwrite(dst,&op,4);
	return (int)code_calls;
}

// PRODUCTION_GLOBAL_MERGE_541E40

static uint32_t word(void) {
	uint32_t value;
	assert(fread(&value,4,1,stdin)==1);
	return value;
}
static uint8_t* bytes(uint32_t* size) {
	*size=word();
	assert(*size>0 && *size<=16384);
	uint8_t* out=malloc(*size);
	assert(out);
	assert(fread(out,1,*size,stdin)==*size);
	return out;
}
static FILE* binary_temp(void) {
	FILE* out=tmpfile();
	assert(out);
#ifdef _WIN32
	assert(_setmode(_fileno(out),_O_BINARY)!=-1);
#endif
	return out;
}
int main(void) {
	return_opcode=word();
	uint32_t count=word();
	assert(count==3);
	for (uint32_t row=0;row<count;row++) {
		uint32_t asize,bsize,wsize;
		uint8_t *a=bytes(&asize), *b=bytes(&bsize), *want=bytes(&wsize);
		FILE *af=binary_temp(), *bf=binary_temp(), *dst=binary_temp();
		nox_fs_fwrite(af,a,asize);rewind(af);
		nox_fs_fwrite(bf,b,bsize);rewind(bf);
		code_calls=0;dword_5d4594_2489420=0;
		assert(nox_script_readWriteVvv_541E40(af,bf,dst)==2);
		assert(code_calls==2 && dword_5d4594_2489420==4);
		assert(ftell(af)==asize && ftell(bf)==bsize && ftell(dst)==wsize);
		size_t capacity=wsize>asize ? wsize : asize;
		if (bsize>capacity) capacity=bsize;
		uint8_t* got=malloc(capacity);assert(got);rewind(dst);
		assert(fread(got,1,wsize,dst)==wsize && fgetc(dst)==EOF);
		assert(memcmp(got,want,wsize)==0);
		rewind(af);assert(fread(got,1,asize,af)==asize && fgetc(af)==EOF);
		assert(memcmp(got,a,asize)==0);
		rewind(bf);assert(fread(got,1,bsize,bf)==bsize && fgetc(bf)==EOF);
		assert(memcmp(got,b,bsize)==0);
		free(got);free(a);free(b);free(want);
		assert(fclose(af)==0 && fclose(bf)==0 && fclose(dst)==0);
	}
	assert(fgetc(stdin)==EOF);
	puts("native global merge: 3 exact metadata/name/code/return cases");
	return 0;
}
