#ifndef MINIMP3_H
#define MINIMP3_H
/*
	https://github.com/lieff/minimp3
	To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights to
   this software to the public domain worldwide. This software is distributed without any warranty. See
   <http://creativecommons.org/publicdomain/zero/1.0/>.
*/
#include <stdint.h>

typedef float mp3dec_spectral_frame_t[2][2][576];

typedef struct {
	uint8_t sfbtab[40];
	uint16_t part_23_length, big_values, scalefac_compress;
	uint8_t global_gain, block_type, mixed_block_flag, n_long_sfb, n_short_sfb;
	uint8_t table_select[3], region_count[3], subblock_gain[3];
	uint8_t preflag, scalefac_scale, count1_table, scfsi;
} mp3dec_granule_t;

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

// Unpack the coefficients from an assembled Layer III payload and return the bits consumed.
int mp3dec_unpack_layer3(const uint8_t* hdr, const mp3dec_granule_t* granules, const uint8_t* main_data,
						int data_bytes, mp3dec_spectral_frame_t* spectrum);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* MINIMP3_H */
#if defined(MINIMP3_IMPLEMENTATION) && !defined(_MINIMP3_IMPLEMENTATION_GUARD)
#define _MINIMP3_IMPLEMENTATION_GUARD

#include <stdlib.h>
#include <string.h>

#define MAX_FREE_FORMAT_FRAME_SIZE 2304 /* more than ISO spec's */
#define MAX_L3_FRAME_PAYLOAD_BYTES MAX_FREE_FORMAT_FRAME_SIZE /* MUST be >= 320000/8/32000*1152 = 1440 */

#define MAX_BITRESERVOIR_BYTES 511
#define MODE_MONO 3
#define MODE_JOINT_STEREO 1
#define HDR_IS_MONO(h) (((h[3]) & 0xC0) == 0xC0)
#define HDR_IS_MS_STEREO(h) (((h[3]) & 0xE0) == 0x60)
#define HDR_TEST_MPEG1(h) ((h[1]) & 0x8)
#define HDR_TEST_NOT_MPEG25(h) ((h[1]) & 0x10)
#define HDR_TEST_I_STEREO(h) ((h[3]) & 0x10)
#define HDR_TEST_MS_STEREO(h) ((h[3]) & 0x20)
#define HDR_GET_STEREO_MODE(h) (((h[3]) >> 6) & 3)
#define HDR_GET_STEREO_MODE_EXT(h) (((h[3]) >> 4) & 3)
#define HDR_GET_LAYER(h) (((h[1]) >> 1) & 3)
#define HDR_GET_BITRATE(h) ((h[2]) >> 4)
#define HDR_GET_SAMPLE_RATE(h) (((h[2]) >> 2) & 3)
#define HDR_GET_MY_SAMPLE_RATE(h) (HDR_GET_SAMPLE_RATE(h) + (((h[1] >> 3) & 1) + ((h[1] >> 4) & 1)) * 3)
#define HDR_IS_LAYER_1(h) ((h[1] & 6) == 6)

#define BITS_DEQUANTIZER_OUT -1
#define MAX_SCF (255 + BITS_DEQUANTIZER_OUT * 4 - 210)
#define MAX_SCFI ((MAX_SCF + 3) & ~3)

#define MINIMP3_MIN(a, b) ((a) > (b) ? (b) : (a))
#define MINIMP3_MAX(a, b) ((a) < (b) ? (b) : (a))

#if !defined(MINIMP3_NO_SIMD)

#if !defined(MINIMP3_ONLY_SIMD) && (defined(_M_X64) || defined(_M_ARM64) || defined(__x86_64__) || defined(__aarch64__))
/* x64 always have SSE2, arm64 always have neon, no need for generic code */
#define MINIMP3_ONLY_SIMD
#endif /* SIMD checks... */

#if (defined(_MSC_VER) && (defined(_M_IX86) || defined(_M_X64))) ||                                                    \
	((defined(__i386__) || defined(__x86_64__)) && defined(__SSE2__))
#if defined(_MSC_VER)
#include <intrin.h>
#endif /* defined(_MSC_VER) */
#include <immintrin.h>
#define HAVE_SSE 1
#define HAVE_SIMD 1
#define VSTORE _mm_storeu_ps
#define VLD _mm_loadu_ps
#define VSET _mm_set1_ps
#define VADD _mm_add_ps
#define VSUB _mm_sub_ps
#define VMUL _mm_mul_ps
#define VMAC(a, x, y) _mm_add_ps(a, _mm_mul_ps(x, y))
#define VMSB(a, x, y) _mm_sub_ps(a, _mm_mul_ps(x, y))
#define VMUL_S(x, s) _mm_mul_ps(x, _mm_set1_ps(s))
#define VREV(x) _mm_shuffle_ps(x, x, _MM_SHUFFLE(0, 1, 2, 3))
typedef __m128 f4;
#if defined(_MSC_VER) || defined(MINIMP3_ONLY_SIMD)
#define minimp3_cpuid __cpuid
#else /* defined(_MSC_VER) || defined(MINIMP3_ONLY_SIMD) */
static __inline__ __attribute__((always_inline)) void minimp3_cpuid(int CPUInfo[], const int InfoType) {
#if defined(__PIC__)
	__asm__ __volatile__(
#if defined(__x86_64__)
		"push %%rbx\n"
		"cpuid\n"
		"xchgl %%ebx, %1\n"
		"pop  %%rbx\n"
#else  /* defined(__x86_64__) */
		"xchgl %%ebx, %1\n"
		"cpuid\n"
		"xchgl %%ebx, %1\n"
#endif /* defined(__x86_64__) */
		: "=a"(CPUInfo[0]), "=r"(CPUInfo[1]), "=c"(CPUInfo[2]), "=d"(CPUInfo[3])
		: "a"(InfoType));
#else  /* defined(__PIC__) */
	__asm__ __volatile__("cpuid"
						 : "=a"(CPUInfo[0]), "=b"(CPUInfo[1]), "=c"(CPUInfo[2]), "=d"(CPUInfo[3])
						 : "a"(InfoType));
#endif /* defined(__PIC__)*/
}
#endif /* defined(_MSC_VER) || defined(MINIMP3_ONLY_SIMD) */
static int have_simd() {
#ifdef MINIMP3_ONLY_SIMD
	return 1;
#else /* MINIMP3_ONLY_SIMD */
	static int g_have_simd;
	int CPUInfo[4];
#ifdef MINIMP3_TEST
	static int g_counter;
	if (g_counter++ > 100)
		return 0;
#endif /* MINIMP3_TEST */
	if (g_have_simd)
		goto end;
	minimp3_cpuid(CPUInfo, 0);
	g_have_simd = 1;
	if (CPUInfo[0] > 0) {
		minimp3_cpuid(CPUInfo, 1);
		g_have_simd = (CPUInfo[3] & (1 << 26)) + 1; /* SSE2 */
	}
end:
	return g_have_simd - 1;
#endif /* MINIMP3_ONLY_SIMD */
}
#elif defined(__ARM_NEON) || defined(__aarch64__)
#include <arm_neon.h>
#define HAVE_SIMD 1
#define VSTORE vst1q_f32
#define VLD vld1q_f32
#define VSET vmovq_n_f32
#define VADD vaddq_f32
#define VSUB vsubq_f32
#define VMUL vmulq_f32
#define VMAC(a, x, y) vmlaq_f32(a, x, y)
#define VMSB(a, x, y) vmlsq_f32(a, x, y)
#define VMUL_S(x, s) vmulq_f32(x, vmovq_n_f32(s))
#define VREV(x) vcombine_f32(vget_high_f32(vrev64q_f32(x)), vget_low_f32(vrev64q_f32(x)))
typedef float32x4_t f4;
static int have_simd() { /* TODO: detect neon for !MINIMP3_ONLY_SIMD */
	return 1;
}
#else /* SIMD checks... */
#define HAVE_SIMD 0
#ifdef MINIMP3_ONLY_SIMD
#error MINIMP3_ONLY_SIMD used, but SSE/NEON not enabled
#endif /* MINIMP3_ONLY_SIMD */
#endif /* SIMD checks... */
#else  /* !defined(MINIMP3_NO_SIMD) */
#define HAVE_SIMD 0
#endif /* !defined(MINIMP3_NO_SIMD) */

typedef struct {
	const uint8_t* buf;
	int pos, limit;
} bs_t;

typedef struct {
	float scf[3 * 64];
	uint8_t total_bands, stereo_bands, bitalloc[64], scfcod[64];
} L12_scale_info;

typedef struct {
	uint8_t tab_offset, code_tab_width, band_count;
} L12_subband_alloc_t;

typedef struct {
	bs_t bs;
	uint8_t maindata[MAX_BITRESERVOIR_BYTES + MAX_L3_FRAME_PAYLOAD_BYTES];
	float grbuf[2][576], scf[40], reorder[576];
	uint8_t ist_pos[2][39];
} mp3dec_scratch_t;

static void bs_init(bs_t* bs, const uint8_t* data, int bytes) {
	bs->buf = data;
	bs->pos = 0;
	bs->limit = bytes * 8;
}

static uint32_t get_bits(bs_t* bs, int n) {
	uint32_t next, cache = 0, s = bs->pos & 7;
	int shl = n + s;
	const uint8_t* p = bs->buf + (bs->pos >> 3);
	if ((bs->pos += n) > bs->limit)
		return 0;
	next = *p++ & (255 >> s);
	while ((shl -= 8) > 0) {
		cache |= next << shl;
		next = *p++;
	}
	return cache | (next >> -shl);
}

#ifndef MINIMP3_ONLY_MP3
static unsigned hdr_bitrate_kbps(const uint8_t* h) {
	static const uint8_t halfrate[2][3][15] = {
		{{0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 72, 80},
		 {0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 72, 80},
		 {0, 16, 24, 28, 32, 40, 48, 56, 64, 72, 80, 88, 96, 112, 128}},
		{{0, 16, 20, 24, 28, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160},
		 {0, 16, 24, 28, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192},
		 {0, 16, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224}},
	};
	return 2 * halfrate[!!HDR_TEST_MPEG1(h)][HDR_GET_LAYER(h) - 1][HDR_GET_BITRATE(h)];
}

static const L12_subband_alloc_t* L12_subband_alloc_table(const uint8_t* hdr, L12_scale_info* sci) {
	const L12_subband_alloc_t* alloc;
	int mode = HDR_GET_STEREO_MODE(hdr);
	int nbands, stereo_bands = (mode == MODE_MONO)           ? 0
							   : (mode == MODE_JOINT_STEREO) ? (HDR_GET_STEREO_MODE_EXT(hdr) << 2) + 4
															 : 32;

	if (HDR_IS_LAYER_1(hdr)) {
		static const L12_subband_alloc_t g_alloc_L1[] = {{76, 4, 32}};
		alloc = g_alloc_L1;
		nbands = 32;
	} else if (!HDR_TEST_MPEG1(hdr)) {
		static const L12_subband_alloc_t g_alloc_L2M2[] = {{60, 4, 4}, {44, 3, 7}, {44, 2, 19}};
		alloc = g_alloc_L2M2;
		nbands = 30;
	} else {
		static const L12_subband_alloc_t g_alloc_L2M1[] = {{0, 4, 3}, {16, 4, 8}, {32, 3, 12}, {40, 2, 7}};
		int sample_rate_idx = HDR_GET_SAMPLE_RATE(hdr);
		unsigned kbps = hdr_bitrate_kbps(hdr) >> (int)(mode != MODE_MONO);
		if (!kbps) /* free-format */
		{
			kbps = 192;
		}

		alloc = g_alloc_L2M1;
		nbands = 27;
		if (kbps < 56) {
			static const L12_subband_alloc_t g_alloc_L2M1_lowrate[] = {{44, 4, 2}, {44, 3, 10}};
			alloc = g_alloc_L2M1_lowrate;
			nbands = sample_rate_idx == 2 ? 12 : 8;
		} else if (kbps >= 96 && sample_rate_idx != 1) {
			nbands = 30;
		}
	}

	sci->total_bands = (uint8_t)nbands;
	sci->stereo_bands = (uint8_t)MINIMP3_MIN(stereo_bands, nbands);

	return alloc;
}

static void L12_read_scalefactors(bs_t* bs, uint8_t* pba, uint8_t* scfcod, int bands, float* scf) {
	static const float g_deq_L12[18 * 3] = {
#define DQ(x) 9.53674316e-07f / x, 7.56931807e-07f / x, 6.00777173e-07f / x
		DQ(3),    DQ(7),    DQ(15),   DQ(31),    DQ(63),    DQ(127),   DQ(255), DQ(511), DQ(1023),
		DQ(2047), DQ(4095), DQ(8191), DQ(16383), DQ(32767), DQ(65535), DQ(3),   DQ(5),   DQ(9)};
	int i, m;
	for (i = 0; i < bands; i++) {
		float s = 0;
		int ba = *pba++;
		int mask = ba ? 4 + ((19 >> scfcod[i]) & 3) : 0;
		for (m = 4; m; m >>= 1) {
			if (mask & m) {
				int b = get_bits(bs, 6);
				s = g_deq_L12[ba * 3 - 6 + b % 3] * (1 << 21 >> b / 3);
			}
			*scf++ = s;
		}
	}
}

static void L12_read_scale_info(const uint8_t* hdr, bs_t* bs, L12_scale_info* sci) {
	static const uint8_t g_bitalloc_code_tab[] = {
		0,  17, 3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 0,  17, 18, 3,  19, 4,  5,
		6,  7,  8,  9,  10, 11, 12, 13, 16, 0,  17, 18, 3,  19, 4,  5,  16, 0,  17, 18, 16, 0,  17,
		18, 19, 4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 0,  17, 18, 3,  19, 4,  5,  6,  7,
		8,  9,  10, 11, 12, 13, 14, 0,  2,  3,  4,  5,  6,  7,  8,  9,  10, 11, 12, 13, 14, 15, 16};
	const L12_subband_alloc_t* subband_alloc = L12_subband_alloc_table(hdr, sci);

	int i, k = 0, ba_bits = 0;
	const uint8_t* ba_code_tab = g_bitalloc_code_tab;

	for (i = 0; i < sci->total_bands; i++) {
		uint8_t ba;
		if (i == k) {
			k += subband_alloc->band_count;
			ba_bits = subband_alloc->code_tab_width;
			ba_code_tab = g_bitalloc_code_tab + subband_alloc->tab_offset;
			subband_alloc++;
		}
		ba = ba_code_tab[get_bits(bs, ba_bits)];
		sci->bitalloc[2 * i] = ba;
		if (i < sci->stereo_bands) {
			ba = ba_code_tab[get_bits(bs, ba_bits)];
		}
		sci->bitalloc[2 * i + 1] = sci->stereo_bands ? ba : 0;
	}

	for (i = 0; i < 2 * sci->total_bands; i++) {
		sci->scfcod[i] = sci->bitalloc[i] ? HDR_IS_LAYER_1(hdr) ? 2 : get_bits(bs, 2) : 6;
	}

	L12_read_scalefactors(bs, sci->bitalloc, sci->scfcod, sci->total_bands * 2, sci->scf);

	for (i = sci->stereo_bands; i < sci->total_bands; i++) {
		sci->bitalloc[2 * i + 1] = 0;
	}
}

static int L12_dequantize_granule(float* grbuf, bs_t* bs, L12_scale_info* sci, int group_size) {
	int i, j, k, choff = 576;
	for (j = 0; j < 4; j++) {
		float* dst = grbuf + group_size * j;
		for (i = 0; i < 2 * sci->total_bands; i++) {
			int ba = sci->bitalloc[i];
			if (ba != 0) {
				if (ba < 17) {
					int half = (1 << (ba - 1)) - 1;
					for (k = 0; k < group_size; k++) {
						dst[k] = (float)((int)get_bits(bs, ba) - half);
					}
				} else {
					unsigned mod = (2 << (ba - 17)) + 1;                /* 3, 5, 9 */
					unsigned code = get_bits(bs, mod + 2 - (mod >> 3)); /* 5, 7, 10 */
					for (k = 0; k < group_size; k++, code /= mod) {
						dst[k] = (float)((int)(code % mod - mod / 2));
					}
				}
			}
			dst += choff;
			choff = 18 - choff;
		}
	}
	return group_size * 4;
}

static void L12_apply_scf_384(L12_scale_info* sci, const float* scf, float* dst) {
	int i, k;
	memcpy(dst + 576 + sci->stereo_bands * 18, dst + sci->stereo_bands * 18,
		   (sci->total_bands - sci->stereo_bands) * 18 * sizeof(float));
	for (i = 0; i < sci->total_bands; i++, dst += 18, scf += 6) {
		for (k = 0; k < 12; k++) {
			dst[k + 0] *= scf[0];
			dst[k + 576] *= scf[3];
		}
	}
}
#endif /* MINIMP3_ONLY_MP3 */

static void L3_read_scalefactors(uint8_t* scf, uint8_t* ist_pos, const uint8_t* scf_size, const uint8_t* scf_count,
								 bs_t* bitbuf, int scfsi) {
	int i, k;
	for (i = 0; i < 4 && scf_count[i]; i++, scfsi *= 2) {
		int cnt = scf_count[i];
		if (scfsi & 8) {
			memcpy(scf, ist_pos, cnt);
		} else {
			int bits = scf_size[i];
			if (!bits) {
				memset(scf, 0, cnt);
				memset(ist_pos, 0, cnt);
			} else {
				int max_scf = (scfsi < 0) ? (1 << bits) - 1 : -1;
				for (k = 0; k < cnt; k++) {
					int s = get_bits(bitbuf, bits);
					ist_pos[k] = (s == max_scf ? -1 : s);
					scf[k] = s;
				}
			}
		}
		ist_pos += cnt;
		scf += cnt;
	}
	scf[0] = scf[1] = scf[2] = 0;
}

static float L3_ldexp_q2(float y, int exp_q2) {
	static const float g_expfrac[4] = {9.31322575e-10f, 7.83145814e-10f, 6.58544508e-10f, 5.53767716e-10f};
	int e;
	do {
		e = MINIMP3_MIN(30 * 4, exp_q2);
		y *= g_expfrac[e & 3] * (1 << 30 >> (e >> 2));
	} while ((exp_q2 -= e) > 0);
	return y;
}

static void L3_decode_scalefactors(const uint8_t* hdr, uint8_t* ist_pos, bs_t* bs, const mp3dec_granule_t* gr, float* scf,
								   int ch) {
	static const uint8_t g_scf_partitions[3][28] = {
		{6, 5, 5, 5, 6, 5, 5, 5, 6, 5, 7, 3, 11, 10, 0, 0, 7, 7, 7, 0, 6, 6, 6, 3, 8, 8, 5, 0},
		{8, 9, 6, 12, 6, 9, 9, 9, 6, 9, 12, 6, 15, 18, 0, 0, 6, 15, 12, 0, 6, 12, 9, 6, 6, 18, 9, 0},
		{9, 9, 6, 12, 9, 9, 9, 9, 9, 9, 12, 6, 18, 18, 0, 0, 12, 12, 12, 0, 12, 9, 9, 6, 15, 12, 9, 0}};
	const uint8_t* scf_partition = g_scf_partitions[!!gr->n_short_sfb + !gr->n_long_sfb];
	uint8_t scf_size[4], iscf[40];
	int i, scf_shift = gr->scalefac_scale + 1, gain_exp, scfsi = gr->scfsi;
	float gain;

	if (HDR_TEST_MPEG1(hdr)) {
		static const uint8_t g_scfc_decode[16] = {0, 1, 2, 3, 12, 5, 6, 7, 9, 10, 11, 13, 14, 15, 18, 19};
		int part = g_scfc_decode[gr->scalefac_compress];
		scf_size[1] = scf_size[0] = (uint8_t)(part >> 2);
		scf_size[3] = scf_size[2] = (uint8_t)(part & 3);
	} else {
		static const uint8_t g_mod[6 * 4] = {5, 5, 4, 4, 5, 5, 4, 1, 4, 3, 1, 1, 5, 6, 6, 1, 4, 4, 4, 1, 4, 3, 1, 1};
		int k, modprod, sfc, ist = HDR_TEST_I_STEREO(hdr) && ch;
		sfc = gr->scalefac_compress >> ist;
		for (k = ist * 3 * 4; sfc >= 0; sfc -= modprod, k += 4) {
			for (modprod = 1, i = 3; i >= 0; i--) {
				scf_size[i] = (uint8_t)(sfc / modprod % g_mod[k + i]);
				modprod *= g_mod[k + i];
			}
		}
		scf_partition += k;
		scfsi = -16;
	}
	L3_read_scalefactors(iscf, ist_pos, scf_size, scf_partition, bs, scfsi);

	if (gr->n_short_sfb) {
		int sh = 3 - scf_shift;
		for (i = 0; i < gr->n_short_sfb; i += 3) {
			iscf[gr->n_long_sfb + i + 0] += gr->subblock_gain[0] << sh;
			iscf[gr->n_long_sfb + i + 1] += gr->subblock_gain[1] << sh;
			iscf[gr->n_long_sfb + i + 2] += gr->subblock_gain[2] << sh;
		}
	} else if (gr->preflag) {
		static const uint8_t g_preamp[10] = {1, 1, 1, 1, 2, 2, 3, 3, 3, 2};
		for (i = 0; i < 10; i++) {
			iscf[11 + i] += g_preamp[i];
		}
	}

	gain_exp = gr->global_gain + BITS_DEQUANTIZER_OUT * 4 - 210 - (HDR_IS_MS_STEREO(hdr) ? 2 : 0);
	gain = L3_ldexp_q2(1 << (MAX_SCFI / 4), MAX_SCFI - gain_exp);
	for (i = 0; i < (int)(gr->n_long_sfb + gr->n_short_sfb); i++) {
		scf[i] = L3_ldexp_q2(gain, iscf[i] << scf_shift);
	}
}

static const float g_pow43[129 + 16] = {
	0,           -1,          -2.519842f,  -4.326749f,  -6.349604f,  -8.549880f,  -10.902724f, -13.390518f, -16.000000f,
	-18.720754f, -21.544347f, -24.463781f, -27.473142f, -30.567351f, -33.741992f, -36.993181f, 0,           1,
	2.519842f,   4.326749f,   6.349604f,   8.549880f,   10.902724f,  13.390518f,  16.000000f,  18.720754f,  21.544347f,
	24.463781f,  27.473142f,  30.567351f,  33.741992f,  36.993181f,  40.317474f,  43.711787f,  47.173345f,  50.699631f,
	54.288352f,  57.937408f,  61.644865f,  65.408941f,  69.227979f,  73.100443f,  77.024898f,  81.000000f,  85.024491f,
	89.097188f,  93.216975f,  97.382800f,  101.593667f, 105.848633f, 110.146801f, 114.487321f, 118.869381f, 123.292209f,
	127.755065f, 132.257246f, 136.798076f, 141.376907f, 145.993119f, 150.646117f, 155.335327f, 160.060199f, 164.820202f,
	169.614826f, 174.443577f, 179.305980f, 184.201575f, 189.129918f, 194.090580f, 199.083145f, 204.107210f, 209.162385f,
	214.248292f, 219.364564f, 224.510845f, 229.686789f, 234.892058f, 240.126328f, 245.389280f, 250.680604f, 256.000000f,
	261.347174f, 266.721841f, 272.123723f, 277.552547f, 283.008049f, 288.489971f, 293.998060f, 299.532071f, 305.091761f,
	310.676898f, 316.287249f, 321.922592f, 327.582707f, 333.267377f, 338.976394f, 344.709550f, 350.466646f, 356.247482f,
	362.051866f, 367.879608f, 373.730522f, 379.604427f, 385.501143f, 391.420496f, 397.362314f, 403.326427f, 409.312672f,
	415.320884f, 421.350905f, 427.402579f, 433.475750f, 439.570269f, 445.685987f, 451.822757f, 457.980436f, 464.158883f,
	470.357960f, 476.577530f, 482.817459f, 489.077615f, 495.357868f, 501.658090f, 507.978156f, 514.317941f, 520.677324f,
	527.056184f, 533.454404f, 539.871867f, 546.308458f, 552.764065f, 559.238575f, 565.731879f, 572.243870f, 578.774440f,
	585.323483f, 591.890898f, 598.476581f, 605.080431f, 611.702349f, 618.342238f, 625.000000f, 631.675540f, 638.368763f,
	645.079578f};

static float L3_pow_43(int x) {
	float frac;
	int sign, mult = 256;

	if (x < 129) {
		return g_pow43[16 + x];
	}

	if (x < 1024) {
		mult = 16;
		x <<= 3;
	}

	sign = 2 * x & 64;
	frac = (float)((x & 63) - sign) / ((x & ~63) + sign);
	return g_pow43[16 + ((x + sign) >> 6)] * (1.f + frac * ((4.f / 3) + frac * (2.f / 9))) * mult;
}

static void L3_huffman(float* dst, bs_t* bs, const mp3dec_granule_t* gr_info, const float* scf, int layer3gr_limit) {
	static const int16_t tabs[] = {
		0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
		0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
		785,   785,   785,   785,   784,   784,   784,   784,   513,   513,   513,   513,   513,   513,   513,   513,
		256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,
		-255,  1313,  1298,  1282,  785,   785,   785,   785,   784,   784,   784,   784,   769,   769,   769,   769,
		256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,
		290,   288,   -255,  1313,  1298,  1282,  769,   769,   769,   769,   529,   529,   529,   529,   529,   529,
		529,   529,   528,   528,   528,   528,   528,   528,   528,   528,   512,   512,   512,   512,   512,   512,
		512,   512,   290,   288,   -253,  -318,  -351,  -367,  785,   785,   785,   785,   784,   784,   784,   784,
		769,   769,   769,   769,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,
		256,   256,   256,   256,   819,   818,   547,   547,   275,   275,   275,   275,   561,   560,   515,   546,
		289,   274,   288,   258,   -254,  -287,  1329,  1299,  1314,  1312,  1057,  1057,  1042,  1042,  1026,  1026,
		784,   784,   784,   784,   529,   529,   529,   529,   529,   529,   529,   529,   769,   769,   769,   769,
		768,   768,   768,   768,   563,   560,   306,   306,   291,   259,   -252,  -413,  -477,  -542,  1298,  -575,
		1041,  1041,  784,   784,   784,   784,   769,   769,   769,   769,   256,   256,   256,   256,   256,   256,
		256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   -383,  -399,  1107,  1092,  1106,  1061,
		849,   849,   789,   789,   1104,  1091,  773,   773,   1076,  1075,  341,   340,   325,   309,   834,   804,
		577,   577,   532,   532,   516,   516,   832,   818,   803,   816,   561,   561,   531,   531,   515,   546,
		289,   289,   288,   258,   -252,  -429,  -493,  -559,  1057,  1057,  1042,  1042,  529,   529,   529,   529,
		529,   529,   529,   529,   784,   784,   784,   784,   769,   769,   769,   769,   512,   512,   512,   512,
		512,   512,   512,   512,   -382,  1077,  -415,  1106,  1061,  1104,  849,   849,   789,   789,   1091,  1076,
		1029,  1075,  834,   834,   597,   581,   340,   340,   339,   324,   804,   833,   532,   532,   832,   772,
		818,   803,   817,   787,   816,   771,   290,   290,   290,   290,   288,   258,   -253,  -349,  -414,  -447,
		-463,  1329,  1299,  -479,  1314,  1312,  1057,  1057,  1042,  1042,  1026,  1026,  785,   785,   785,   785,
		784,   784,   784,   784,   769,   769,   769,   769,   768,   768,   768,   768,   -319,  851,   821,   -335,
		836,   850,   805,   849,   341,   340,   325,   336,   533,   533,   579,   579,   564,   564,   773,   832,
		578,   548,   563,   516,   321,   276,   306,   291,   304,   259,   -251,  -572,  -733,  -830,  -863,  -879,
		1041,  1041,  784,   784,   784,   784,   769,   769,   769,   769,   256,   256,   256,   256,   256,   256,
		256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   -511,  -527,  -543,  1396,  1351,  1381,
		1366,  1395,  1335,  1380,  -559,  1334,  1138,  1138,  1063,  1063,  1350,  1392,  1031,  1031,  1062,  1062,
		1364,  1363,  1120,  1120,  1333,  1348,  881,   881,   881,   881,   375,   374,   359,   373,   343,   358,
		341,   325,   791,   791,   1123,  1122,  -703,  1105,  1045,  -719,  865,   865,   790,   790,   774,   774,
		1104,  1029,  338,   293,   323,   308,   -799,  -815,  833,   788,   772,   818,   803,   816,   322,   292,
		307,   320,   561,   531,   515,   546,   289,   274,   288,   258,   -251,  -525,  -605,  -685,  -765,  -831,
		-846,  1298,  1057,  1057,  1312,  1282,  785,   785,   785,   785,   784,   784,   784,   784,   769,   769,
		769,   769,   512,   512,   512,   512,   512,   512,   512,   512,   1399,  1398,  1383,  1367,  1382,  1396,
		1351,  -511,  1381,  1366,  1139,  1139,  1079,  1079,  1124,  1124,  1364,  1349,  1363,  1333,  882,   882,
		882,   882,   807,   807,   807,   807,   1094,  1094,  1136,  1136,  373,   341,   535,   535,   881,   775,
		867,   822,   774,   -591,  324,   338,   -671,  849,   550,   550,   866,   864,   609,   609,   293,   336,
		534,   534,   789,   835,   773,   -751,  834,   804,   308,   307,   833,   788,   832,   772,   562,   562,
		547,   547,   305,   275,   560,   515,   290,   290,   -252,  -397,  -477,  -557,  -622,  -653,  -719,  -735,
		-750,  1329,  1299,  1314,  1057,  1057,  1042,  1042,  1312,  1282,  1024,  1024,  785,   785,   785,   785,
		784,   784,   784,   784,   769,   769,   769,   769,   -383,  1127,  1141,  1111,  1126,  1140,  1095,  1110,
		869,   869,   883,   883,   1079,  1109,  882,   882,   375,   374,   807,   868,   838,   881,   791,   -463,
		867,   822,   368,   263,   852,   837,   836,   -543,  610,   610,   550,   550,   352,   336,   534,   534,
		865,   774,   851,   821,   850,   805,   593,   533,   579,   564,   773,   832,   578,   578,   548,   548,
		577,   577,   307,   276,   306,   291,   516,   560,   259,   259,   -250,  -2107, -2507, -2764, -2909, -2974,
		-3007, -3023, 1041,  1041,  1040,  1040,  769,   769,   769,   769,   256,   256,   256,   256,   256,   256,
		256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   -767,  -1052, -1213, -1277, -1358, -1405,
		-1469, -1535, -1550, -1582, -1614, -1647, -1662, -1694, -1726, -1759, -1774, -1807, -1822, -1854, -1886, 1565,
		-1919, -1935, -1951, -1967, 1731,  1730,  1580,  1717,  -1983, 1729,  1564,  -1999, 1548,  -2015, -2031, 1715,
		1595,  -2047, 1714,  -2063, 1610,  -2079, 1609,  -2095, 1323,  1323,  1457,  1457,  1307,  1307,  1712,  1547,
		1641,  1700,  1699,  1594,  1685,  1625,  1442,  1442,  1322,  1322,  -780,  -973,  -910,  1279,  1278,  1277,
		1262,  1276,  1261,  1275,  1215,  1260,  1229,  -959,  974,   974,   989,   989,   -943,  735,   478,   478,
		495,   463,   506,   414,   -1039, 1003,  958,   1017,  927,   942,   987,   957,   431,   476,   1272,  1167,
		1228,  -1183, 1256,  -1199, 895,   895,   941,   941,   1242,  1227,  1212,  1135,  1014,  1014,  490,   489,
		503,   487,   910,   1013,  985,   925,   863,   894,   970,   955,   1012,  847,   -1343, 831,   755,   755,
		984,   909,   428,   366,   754,   559,   -1391, 752,   486,   457,   924,   997,   698,   698,   983,   893,
		740,   740,   908,   877,   739,   739,   667,   667,   953,   938,   497,   287,   271,   271,   683,   606,
		590,   712,   726,   574,   302,   302,   738,   736,   481,   286,   526,   725,   605,   711,   636,   724,
		696,   651,   589,   681,   666,   710,   364,   467,   573,   695,   466,   466,   301,   465,   379,   379,
		709,   604,   665,   679,   316,   316,   634,   633,   436,   436,   464,   269,   424,   394,   452,   332,
		438,   363,   347,   408,   393,   448,   331,   422,   362,   407,   392,   421,   346,   406,   391,   376,
		375,   359,   1441,  1306,  -2367, 1290,  -2383, 1337,  -2399, -2415, 1426,  1321,  -2431, 1411,  1336,  -2447,
		-2463, -2479, 1169,  1169,  1049,  1049,  1424,  1289,  1412,  1352,  1319,  -2495, 1154,  1154,  1064,  1064,
		1153,  1153,  416,   390,   360,   404,   403,   389,   344,   374,   373,   343,   358,   372,   327,   357,
		342,   311,   356,   326,   1395,  1394,  1137,  1137,  1047,  1047,  1365,  1392,  1287,  1379,  1334,  1364,
		1349,  1378,  1318,  1363,  792,   792,   792,   792,   1152,  1152,  1032,  1032,  1121,  1121,  1046,  1046,
		1120,  1120,  1030,  1030,  -2895, 1106,  1061,  1104,  849,   849,   789,   789,   1091,  1076,  1029,  1090,
		1060,  1075,  833,   833,   309,   324,   532,   532,   832,   772,   818,   803,   561,   561,   531,   560,
		515,   546,   289,   274,   288,   258,   -250,  -1179, -1579, -1836, -1996, -2124, -2253, -2333, -2413, -2477,
		-2542, -2574, -2607, -2622, -2655, 1314,  1313,  1298,  1312,  1282,  785,   785,   785,   785,   1040,  1040,
		1025,  1025,  768,   768,   768,   768,   -766,  -798,  -830,  -862,  -895,  -911,  -927,  -943,  -959,  -975,
		-991,  -1007, -1023, -1039, -1055, -1070, 1724,  1647,  -1103, -1119, 1631,  1767,  1662,  1738,  1708,  1723,
		-1135, 1780,  1615,  1779,  1599,  1677,  1646,  1778,  1583,  -1151, 1777,  1567,  1737,  1692,  1765,  1722,
		1707,  1630,  1751,  1661,  1764,  1614,  1736,  1676,  1763,  1750,  1645,  1598,  1721,  1691,  1762,  1706,
		1582,  1761,  1566,  -1167, 1749,  1629,  767,   766,   751,   765,   494,   494,   735,   764,   719,   749,
		734,   763,   447,   447,   748,   718,   477,   506,   431,   491,   446,   476,   461,   505,   415,   430,
		475,   445,   504,   399,   460,   489,   414,   503,   383,   474,   429,   459,   502,   502,   746,   752,
		488,   398,   501,   473,   413,   472,   486,   271,   480,   270,   -1439, -1455, 1357,  -1471, -1487, -1503,
		1341,  1325,  -1519, 1489,  1463,  1403,  1309,  -1535, 1372,  1448,  1418,  1476,  1356,  1462,  1387,  -1551,
		1475,  1340,  1447,  1402,  1386,  -1567, 1068,  1068,  1474,  1461,  455,   380,   468,   440,   395,   425,
		410,   454,   364,   467,   466,   464,   453,   269,   409,   448,   268,   432,   1371,  1473,  1432,  1417,
		1308,  1460,  1355,  1446,  1459,  1431,  1083,  1083,  1401,  1416,  1458,  1445,  1067,  1067,  1370,  1457,
		1051,  1051,  1291,  1430,  1385,  1444,  1354,  1415,  1400,  1443,  1082,  1082,  1173,  1113,  1186,  1066,
		1185,  1050,  -1967, 1158,  1128,  1172,  1097,  1171,  1081,  -1983, 1157,  1112,  416,   266,   375,   400,
		1170,  1142,  1127,  1065,  793,   793,   1169,  1033,  1156,  1096,  1141,  1111,  1155,  1080,  1126,  1140,
		898,   898,   808,   808,   897,   897,   792,   792,   1095,  1152,  1032,  1125,  1110,  1139,  1079,  1124,
		882,   807,   838,   881,   853,   791,   -2319, 867,   368,   263,   822,   852,   837,   866,   806,   865,
		-2399, 851,   352,   262,   534,   534,   821,   836,   594,   594,   549,   549,   593,   593,   533,   533,
		848,   773,   579,   579,   564,   578,   548,   563,   276,   276,   577,   576,   306,   291,   516,   560,
		305,   305,   275,   259,   -251,  -892,  -2058, -2620, -2828, -2957, -3023, -3039, 1041,  1041,  1040,  1040,
		769,   769,   769,   769,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,   256,
		256,   256,   256,   256,   -511,  -527,  -543,  -559,  1530,  -575,  -591,  1528,  1527,  1407,  1526,  1391,
		1023,  1023,  1023,  1023,  1525,  1375,  1268,  1268,  1103,  1103,  1087,  1087,  1039,  1039,  1523,  -604,
		815,   815,   815,   815,   510,   495,   509,   479,   508,   463,   507,   447,   431,   505,   415,   399,
		-734,  -782,  1262,  -815,  1259,  1244,  -831,  1258,  1228,  -847,  -863,  1196,  -879,  1253,  987,   987,
		748,   -767,  493,   493,   462,   477,   414,   414,   686,   669,   478,   446,   461,   445,   474,   429,
		487,   458,   412,   471,   1266,  1264,  1009,  1009,  799,   799,   -1019, -1276, -1452, -1581, -1677, -1757,
		-1821, -1886, -1933, -1997, 1257,  1257,  1483,  1468,  1512,  1422,  1497,  1406,  1467,  1496,  1421,  1510,
		1134,  1134,  1225,  1225,  1466,  1451,  1374,  1405,  1252,  1252,  1358,  1480,  1164,  1164,  1251,  1251,
		1238,  1238,  1389,  1465,  -1407, 1054,  1101,  -1423, 1207,  -1439, 830,   830,   1248,  1038,  1237,  1117,
		1223,  1148,  1236,  1208,  411,   426,   395,   410,   379,   269,   1193,  1222,  1132,  1235,  1221,  1116,
		976,   976,   1192,  1162,  1177,  1220,  1131,  1191,  963,   963,   -1647, 961,   780,   -1663, 558,   558,
		994,   993,   437,   408,   393,   407,   829,   978,   813,   797,   947,   -1743, 721,   721,   377,   392,
		844,   950,   828,   890,   706,   706,   812,   859,   796,   960,   948,   843,   934,   874,   571,   571,
		-1919, 690,   555,   689,   421,   346,   539,   539,   944,   779,   918,   873,   932,   842,   903,   888,
		570,   570,   931,   917,   674,   674,   -2575, 1562,  -2591, 1609,  -2607, 1654,  1322,  1322,  1441,  1441,
		1696,  1546,  1683,  1593,  1669,  1624,  1426,  1426,  1321,  1321,  1639,  1680,  1425,  1425,  1305,  1305,
		1545,  1668,  1608,  1623,  1667,  1592,  1638,  1666,  1320,  1320,  1652,  1607,  1409,  1409,  1304,  1304,
		1288,  1288,  1664,  1637,  1395,  1395,  1335,  1335,  1622,  1636,  1394,  1394,  1319,  1319,  1606,  1621,
		1392,  1392,  1137,  1137,  1137,  1137,  345,   390,   360,   375,   404,   373,   1047,  -2751, -2767, -2783,
		1062,  1121,  1046,  -2799, 1077,  -2815, 1106,  1061,  789,   789,   1105,  1104,  263,   355,   310,   340,
		325,   354,   352,   262,   339,   324,   1091,  1076,  1029,  1090,  1060,  1075,  833,   833,   788,   788,
		1088,  1028,  818,   818,   803,   803,   561,   561,   531,   531,   816,   771,   546,   546,   289,   274,
		288,   258,   -253,  -317,  -381,  -446,  -478,  -509,  1279,  1279,  -811,  -1179, -1451, -1756, -1900, -2028,
		-2189, -2253, -2333, -2414, -2445, -2511, -2526, 1313,  1298,  -2559, 1041,  1041,  1040,  1040,  1025,  1025,
		1024,  1024,  1022,  1007,  1021,  991,   1020,  975,   1019,  959,   687,   687,   1018,  1017,  671,   671,
		655,   655,   1016,  1015,  639,   639,   758,   758,   623,   623,   757,   607,   756,   591,   755,   575,
		754,   559,   543,   543,   1009,  783,   -575,  -621,  -685,  -749,  496,   -590,  750,   749,   734,   748,
		974,   989,   1003,  958,   988,   973,   1002,  942,   987,   957,   972,   1001,  926,   986,   941,   971,
		956,   1000,  910,   985,   925,   999,   894,   970,   -1071, -1087, -1102, 1390,  -1135, 1436,  1509,  1451,
		1374,  -1151, 1405,  1358,  1480,  1420,  -1167, 1507,  1494,  1389,  1342,  1465,  1435,  1450,  1326,  1505,
		1310,  1493,  1373,  1479,  1404,  1492,  1464,  1419,  428,   443,   472,   397,   736,   526,   464,   464,
		486,   457,   442,   471,   484,   482,   1357,  1449,  1434,  1478,  1388,  1491,  1341,  1490,  1325,  1489,
		1463,  1403,  1309,  1477,  1372,  1448,  1418,  1433,  1476,  1356,  1462,  1387,  -1439, 1475,  1340,  1447,
		1402,  1474,  1324,  1461,  1371,  1473,  269,   448,   1432,  1417,  1308,  1460,  -1711, 1459,  -1727, 1441,
		1099,  1099,  1446,  1386,  1431,  1401,  -1743, 1289,  1083,  1083,  1160,  1160,  1458,  1445,  1067,  1067,
		1370,  1457,  1307,  1430,  1129,  1129,  1098,  1098,  268,   432,   267,   416,   266,   400,   -1887, 1144,
		1187,  1082,  1173,  1113,  1186,  1066,  1050,  1158,  1128,  1143,  1172,  1097,  1171,  1081,  420,   391,
		1157,  1112,  1170,  1142,  1127,  1065,  1169,  1049,  1156,  1096,  1141,  1111,  1155,  1080,  1126,  1154,
		1064,  1153,  1140,  1095,  1048,  -2159, 1125,  1110,  1137,  -2175, 823,   823,   1139,  1138,  807,   807,
		384,   264,   368,   263,   868,   838,   853,   791,   867,   822,   852,   837,   866,   806,   865,   790,
		-2319, 851,   821,   836,   352,   262,   850,   805,   849,   -2399, 533,   533,   835,   820,   336,   261,
		578,   548,   563,   577,   532,   532,   832,   772,   562,   562,   547,   547,   305,   275,   560,   515,
		290,   290,   288,   258};
	static const uint8_t tab32[] = {130, 162, 193, 209, 44,  28,  76,  140, 9,   9,   9,   9,  9,   9,
									9,   9,   190, 254, 222, 238, 126, 94,  157, 157, 109, 61, 173, 205};
	static const uint8_t tab33[] = {252, 236, 220, 204, 188, 172, 156, 140, 124, 108, 92, 76, 60, 44, 28, 12};
	static const int16_t tabindex[2 * 16] = {0,    32,   64,   98,   0,    132,  180,  218,  292,  364,  426,
											 538,  648,  746,  0,    1126, 1460, 1460, 1460, 1460, 1460, 1460,
											 1460, 1460, 1842, 1842, 1842, 1842, 1842, 1842, 1842, 1842};
	static const uint8_t g_linbits[] = {0, 0, 0, 0, 0, 0, 0,  0,  0, 0, 0, 0, 0, 0, 0,  0,
										1, 2, 3, 4, 6, 8, 10, 13, 4, 5, 6, 7, 8, 9, 11, 13};

#define PEEK_BITS(n) (bs_cache >> (32 - n))
#define FLUSH_BITS(n)                                                                                                  \
	{                                                                                                                  \
		bs_cache <<= (n);                                                                                              \
		bs_sh += (n);                                                                                                  \
	}
#define CHECK_BITS                                                                                                     \
	while (bs_sh >= 0) {                                                                                               \
		bs_cache |= (uint32_t)*bs_next_ptr++ << bs_sh;                                                                 \
		bs_sh -= 8;                                                                                                    \
	}
#define BSPOS ((bs_next_ptr - bs->buf) * 8 - 24 + bs_sh)

	float one = 0.0f;
	int ireg = 0, big_val_cnt = gr_info->big_values;
	const uint8_t* sfb = gr_info->sfbtab;
	const uint8_t* bs_next_ptr = bs->buf + bs->pos / 8;
	uint32_t bs_cache = (((bs_next_ptr[0] * 256u + bs_next_ptr[1]) * 256u + bs_next_ptr[2]) * 256u + bs_next_ptr[3])
						<< (bs->pos & 7);
	int pairs_to_decode, np, bs_sh = (bs->pos & 7) - 8;
	bs_next_ptr += 4;

	while (big_val_cnt > 0) {
		int tab_num = gr_info->table_select[ireg];
		int sfb_cnt = gr_info->region_count[ireg++];
		const int16_t* codebook = tabs + tabindex[tab_num];
		int linbits = g_linbits[tab_num];
		if (linbits) {
			do {
				np = *sfb++ / 2;
				pairs_to_decode = MINIMP3_MIN(big_val_cnt, np);
				one = *scf++;
				do {
					int j, w = 5;
					int leaf = codebook[PEEK_BITS(w)];
					while (leaf < 0) {
						FLUSH_BITS(w);
						w = leaf & 7;
						leaf = codebook[PEEK_BITS(w) - (leaf >> 3)];
					}
					FLUSH_BITS(leaf >> 8);

					for (j = 0; j < 2; j++, dst++, leaf >>= 4) {
						int lsb = leaf & 0x0F;
						if (lsb == 15) {
							lsb += PEEK_BITS(linbits);
							FLUSH_BITS(linbits);
							CHECK_BITS;
							*dst = one * L3_pow_43(lsb) * ((int32_t)bs_cache < 0 ? -1 : 1);
						} else {
							*dst = g_pow43[16 + lsb - 16 * (bs_cache >> 31)] * one;
						}
						FLUSH_BITS(lsb ? 1 : 0);
					}
					CHECK_BITS;
				} while (--pairs_to_decode);
			} while ((big_val_cnt -= np) > 0 && --sfb_cnt >= 0);
		} else {
			do {
				np = *sfb++ / 2;
				pairs_to_decode = MINIMP3_MIN(big_val_cnt, np);
				one = *scf++;
				do {
					int j, w = 5;
					int leaf = codebook[PEEK_BITS(w)];
					while (leaf < 0) {
						FLUSH_BITS(w);
						w = leaf & 7;
						leaf = codebook[PEEK_BITS(w) - (leaf >> 3)];
					}
					FLUSH_BITS(leaf >> 8);

					for (j = 0; j < 2; j++, dst++, leaf >>= 4) {
						int lsb = leaf & 0x0F;
						*dst = g_pow43[16 + lsb - 16 * (bs_cache >> 31)] * one;
						FLUSH_BITS(lsb ? 1 : 0);
					}
					CHECK_BITS;
				} while (--pairs_to_decode);
			} while ((big_val_cnt -= np) > 0 && --sfb_cnt >= 0);
		}
	}

	for (np = 1 - big_val_cnt;; dst += 4) {
		const uint8_t* codebook_count1 = (gr_info->count1_table) ? tab33 : tab32;
		int leaf = codebook_count1[PEEK_BITS(4)];
		if (!(leaf & 8)) {
			leaf = codebook_count1[(leaf >> 3) + (bs_cache << 4 >> (32 - (leaf & 3)))];
		}
		FLUSH_BITS(leaf & 7);
		if (BSPOS > layer3gr_limit) {
			break;
		}
#define RELOAD_SCALEFACTOR                                                                                             \
	if (!--np) {                                                                                                       \
		np = *sfb++ / 2;                                                                                               \
		if (!np)                                                                                                       \
			break;                                                                                                     \
		one = *scf++;                                                                                                  \
	}
#define DEQ_COUNT1(s)                                                                                                  \
	if (leaf & (128 >> s)) {                                                                                           \
		dst[s] = ((int32_t)bs_cache < 0) ? -one : one;                                                                 \
		FLUSH_BITS(1)                                                                                                  \
	}
		RELOAD_SCALEFACTOR;
		DEQ_COUNT1(0);
		DEQ_COUNT1(1);
		RELOAD_SCALEFACTOR;
		DEQ_COUNT1(2);
		DEQ_COUNT1(3);
		CHECK_BITS;
	}

	bs->pos = layer3gr_limit;
}

static void L3_midside_stereo(float* left, int n) {
	int i = 0;
	float* right = left + 576;
#if HAVE_SIMD
	if (have_simd())
		for (; i < n - 3; i += 4) {
			f4 vl = VLD(left + i);
			f4 vr = VLD(right + i);
			VSTORE(left + i, VADD(vl, vr));
			VSTORE(right + i, VSUB(vl, vr));
		}
#endif /* HAVE_SIMD */
	for (; i < n; i++) {
		float a = left[i];
		float b = right[i];
		left[i] = a + b;
		right[i] = a - b;
	}
}

static void L3_intensity_stereo_band(float* left, int n, float kl, float kr) {
	int i;
	for (i = 0; i < n; i++) {
		left[i + 576] = left[i] * kr;
		left[i] = left[i] * kl;
	}
}

static void L3_stereo_top_band(const float* right, const uint8_t* sfb, int nbands, int max_band[3]) {
	int i, k;

	max_band[0] = max_band[1] = max_band[2] = -1;

	for (i = 0; i < nbands; i++) {
		for (k = 0; k < sfb[i]; k += 2) {
			if (right[k] != 0 || right[k + 1] != 0) {
				max_band[i % 3] = i;
				break;
			}
		}
		right += sfb[i];
	}
}

static void L3_stereo_process(float* left, const uint8_t* ist_pos, const uint8_t* sfb, const uint8_t* hdr,
							  int max_band[3], int mpeg2_sh) {
	static const float g_pan[7 * 2] = {0,    1,           0.21132487f, 0.78867513f, 0.36602540f, 0.63397460f, 0.5f,
									   0.5f, 0.63397460f, 0.36602540f, 0.78867513f, 0.21132487f, 1,           0};
	unsigned i, max_pos = HDR_TEST_MPEG1(hdr) ? 7 : 64;

	for (i = 0; sfb[i]; i++) {
		unsigned ipos = ist_pos[i];
		if ((int)i > max_band[i % 3] && ipos < max_pos) {
			float kl, kr, s = HDR_TEST_MS_STEREO(hdr) ? 1.41421356f : 1;
			if (HDR_TEST_MPEG1(hdr)) {
				kl = g_pan[2 * ipos];
				kr = g_pan[2 * ipos + 1];
			} else {
				kl = 1;
				kr = L3_ldexp_q2(1, (ipos + 1) >> 1 << mpeg2_sh);
				if (ipos & 1) {
					kl = kr;
					kr = 1;
				}
			}
			L3_intensity_stereo_band(left, sfb[i], kl * s, kr * s);
		} else if (HDR_TEST_MS_STEREO(hdr)) {
			L3_midside_stereo(left, sfb[i]);
		}
		left += sfb[i];
	}
}

static void L3_intensity_stereo(float* left, uint8_t* ist_pos, const mp3dec_granule_t* gr, const uint8_t* hdr) {
	int max_band[3], n_sfb = gr->n_long_sfb + gr->n_short_sfb;
	int i, max_blocks = gr->n_short_sfb ? 3 : 1;

	L3_stereo_top_band(left + 576, gr->sfbtab, n_sfb, max_band);
	if (gr->n_long_sfb) {
		max_band[0] = max_band[1] = max_band[2] = MINIMP3_MAX(MINIMP3_MAX(max_band[0], max_band[1]), max_band[2]);
	}
	for (i = 0; i < max_blocks; i++) {
		int default_pos = HDR_TEST_MPEG1(hdr) ? 3 : 0;
		int itop = n_sfb - max_blocks + i;
		int prev = itop - max_blocks;
		ist_pos[itop] = max_band[i] >= prev ? default_pos : ist_pos[prev];
	}
	L3_stereo_process(left, ist_pos, gr->sfbtab, hdr, max_band, gr[1].scalefac_compress & 1);
}

static void L3_reorder(float* grbuf, float* scratch, const uint8_t* sfb) {
	int i, len;
	float *src = grbuf, *dst = scratch;

	for (; 0 != (len = *sfb); sfb += 3, src += 2 * len) {
		for (i = 0; i < len; i++, src++) {
			*dst++ = src[0 * len];
			*dst++ = src[1 * len];
			*dst++ = src[2 * len];
		}
	}
	memcpy(grbuf, scratch, (dst - scratch) * sizeof(float));
}

static void L3_antialias(float* grbuf, int nbands) {
	static const float g_aa[2][8] = {
		{0.85749293f, 0.88174200f, 0.94962865f, 0.98331459f, 0.99551782f, 0.99916056f, 0.99989920f, 0.99999316f},
		{0.51449576f, 0.47173197f, 0.31337745f, 0.18191320f, 0.09457419f, 0.04096558f, 0.01419856f, 0.00369997f}};

	for (; nbands > 0; nbands--, grbuf += 18) {
		int i = 0;
#if HAVE_SIMD
		if (have_simd())
			for (; i < 8; i += 4) {
				f4 vu = VLD(grbuf + 18 + i);
				f4 vd = VLD(grbuf + 14 - i);
				f4 vc0 = VLD(g_aa[0] + i);
				f4 vc1 = VLD(g_aa[1] + i);
				vd = VREV(vd);
				VSTORE(grbuf + 18 + i, VSUB(VMUL(vu, vc0), VMUL(vd, vc1)));
				vd = VADD(VMUL(vu, vc1), VMUL(vd, vc0));
				VSTORE(grbuf + 14 - i, VREV(vd));
			}
#endif /* HAVE_SIMD */
#ifndef MINIMP3_ONLY_SIMD
		for (; i < 8; i++) {
			float u = grbuf[18 + i];
			float d = grbuf[17 - i];
			grbuf[18 + i] = u * g_aa[0][i] - d * g_aa[1][i];
			grbuf[17 - i] = u * g_aa[1][i] + d * g_aa[0][i];
		}
#endif /* MINIMP3_ONLY_SIMD */
	}
}

static void L3_decode(const uint8_t* hdr, mp3dec_scratch_t* s, const mp3dec_granule_t* gr_info, int nch) {
	int ch;

	for (ch = 0; ch < nch; ch++) {
		int layer3gr_limit = s->bs.pos + gr_info[ch].part_23_length;
		L3_decode_scalefactors(hdr, s->ist_pos[ch], &s->bs, gr_info + ch, s->scf, ch);
		L3_huffman(s->grbuf[ch], &s->bs, gr_info + ch, s->scf, layer3gr_limit);
	}

	if (HDR_TEST_I_STEREO(hdr)) {
		L3_intensity_stereo(s->grbuf[0], s->ist_pos[1], gr_info, hdr);
	} else if (HDR_IS_MS_STEREO(hdr)) {
		L3_midside_stereo(s->grbuf[0], 576);
	}

	for (ch = 0; ch < nch; ch++, gr_info++) {
		int aa_bands = 31;
		int n_long_bands = (gr_info->mixed_block_flag ? 2 : 0) << (int)(HDR_GET_MY_SAMPLE_RATE(hdr) == 2);

		if (gr_info->n_short_sfb) {
			aa_bands = n_long_bands - 1;
			L3_reorder(s->grbuf[ch] + n_long_bands * 18, s->reorder, gr_info->sfbtab + gr_info->n_long_sfb);
		}

		L3_antialias(s->grbuf[ch], aa_bands);
	}
}

int mp3dec_unpack_layer3(const uint8_t* hdr, const mp3dec_granule_t* granules, const uint8_t* main_data,
						int data_bytes, mp3dec_spectral_frame_t* spectrum) {
	int channels = HDR_IS_MONO(hdr) ? 1 : 2;
	mp3dec_scratch_t scratch;

	memcpy(scratch.maindata, main_data, data_bytes);
	bs_init(&scratch.bs, scratch.maindata, data_bytes);
	for (int igr = 0; igr < (HDR_TEST_MPEG1(hdr) ? 2 : 1); igr++) {
		memset(scratch.grbuf, 0, sizeof(scratch.grbuf));
		L3_decode(hdr, &scratch, granules + igr * channels, channels);
		memcpy((*spectrum)[igr], scratch.grbuf, sizeof(scratch.grbuf));
	}
	return scratch.bs.pos;
}

#endif /* MINIMP3_IMPLEMENTATION && !_MINIMP3_IMPLEMENTATION_GUARD */
