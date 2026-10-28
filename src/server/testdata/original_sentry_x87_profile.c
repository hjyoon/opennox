#include <stdint.h>
#include <stdio.h>
#include <string.h>
int main(void) {
 unsigned angle,cwInput;
 while(scanf("%x %x",&cwInput,&angle)==2) {
  float value;memcpy(&value,&angle,4);unsigned short old,cw=(unsigned short)cwInput,sw;unsigned char out[16];
  for(int cosine=0;cosine<2;cosine++) {
   memset(out,0,sizeof(out));__asm__ volatile("fnstcw %0":"=m"(old));__asm__ volatile("fldcw %0"::"m"(cw));
   if(cosine)__asm__ volatile("flds %2; fcos; fnstsw %1; fstpt %0":"=m"(out),"=m"(sw):"m"(value):"st","memory");
   else __asm__ volatile("flds %2; fsin; fnstsw %1; fstpt %0":"=m"(out),"=m"(sw):"m"(value):"st","memory");
   __asm__ volatile("fldcw %0"::"m"(old));
   printf("%x %08x %d %04x ",cwInput,angle,cosine,sw);for(int n=9;n>=0;n--)printf("%02x",out[n]);printf("\n");
  }
 }
 return 0;
}
