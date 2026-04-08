#include <ctype.h>
#include <dirent.h>
#include <fcntl.h>
#include <fenv.h>
#include <glob.h>
#include <limits.h>
#include <stdio.h>
#include <sys/ioctl.h>
#include <sys/time.h>
#include <unistd.h>
#include <wctype.h>
#include "proto.h"

#include <SDL2/SDL_stdinc.h>

#ifdef __EMSCRIPTEN__
#include <emscripten/emscripten.h>
#include <emscripten/html5.h>
#endif

enum
{
    HANDLE_FILE,
    HANDLE_PROCESS,
    HANDLE_THREAD,
    HANDLE_MUTEX,
};

struct _REGKEY
{
    char *path;
};

extern const char *progname;
DWORD last_error;
DWORD last_socket_error;
void *handles[1024];

enum
{
    COMPAT_DS_OK = 0,
    COMPAT_DSERR_GENERIC = 0x88780000,
    COMPAT_DSERR_BUFFERLOST = 0x88780096,
};

enum
{
    COMPAT_DSBSTATUS_PLAYING = 0x00000001,
    COMPAT_DSBSTATUS_BUFFERLOST = 0x00000002,
    COMPAT_DSBSTATUS_LOOPING = 0x00000004,
};

struct compat_mm_timer
{
    int active;
    MMRESULT id;
    SDL_TimerID timer;
    LPTIMECALLBACK callback;
    DWORD user;
};

struct compat_directsound
{
    IDirectSound iface;
    ULONG refcount;
    HWND window;
    DWORD cooperative_level;
};

struct compat_directsound_buffer
{
    IDirectSoundBuffer iface;
    ULONG refcount;
    DWORD flags;
    DWORD buffer_bytes;
    DWORD status;
    DWORD play_cursor;
    Uint32 play_start_ticks;
    DWORD play_start_cursor;
    WAVEFORMATEX format;
    BYTE *data;
};

static MMRESULT g_next_timer_id = 1;
static struct compat_mm_timer g_mm_timers[32];

static HRESULT WINAPI compat_ds_query_interface(IDirectSound *iface, const IID *riid, void **ppvObject);
static ULONG WINAPI compat_ds_add_ref(IDirectSound *iface);
static ULONG WINAPI compat_ds_release(IDirectSound *iface);
static HRESULT WINAPI compat_ds_create_sound_buffer(IDirectSound *iface, LPCDSBUFFERDESC desc, IDirectSoundBuffer **buffer, void *outer);
static HRESULT WINAPI compat_ds_get_caps(IDirectSound *iface, DSCAPS *caps);
static HRESULT WINAPI compat_ds_duplicate_sound_buffer(IDirectSound *iface, IDirectSoundBuffer *src, IDirectSoundBuffer **dst);
static HRESULT WINAPI compat_ds_set_cooperative_level(IDirectSound *iface, HWND hwnd, DWORD level);

static HRESULT WINAPI compat_dsb_query_interface(IDirectSoundBuffer *iface, const IID *riid, void **ppvObject);
static ULONG WINAPI compat_dsb_add_ref(IDirectSoundBuffer *iface);
static ULONG WINAPI compat_dsb_release(IDirectSoundBuffer *iface);
static HRESULT WINAPI compat_dsb_get_caps(IDirectSoundBuffer *iface, void *caps);
static HRESULT WINAPI compat_dsb_get_current_position(IDirectSoundBuffer *iface, DWORD *play_cursor, DWORD *write_cursor);
static HRESULT WINAPI compat_dsb_get_format(IDirectSoundBuffer *iface, WAVEFORMATEX *format, DWORD size, DWORD *written);
static HRESULT WINAPI compat_dsb_get_volume(IDirectSoundBuffer *iface, LONG *volume);
static HRESULT WINAPI compat_dsb_get_pan(IDirectSoundBuffer *iface, LONG *pan);
static HRESULT WINAPI compat_dsb_get_frequency(IDirectSoundBuffer *iface, DWORD *frequency);
static HRESULT WINAPI compat_dsb_get_status(IDirectSoundBuffer *iface, DWORD *status);
static HRESULT WINAPI compat_dsb_initialize(IDirectSoundBuffer *iface, IDirectSound *directsound, LPCDSBUFFERDESC desc);
static HRESULT WINAPI compat_dsb_lock(IDirectSoundBuffer *iface, DWORD offset, DWORD bytes, void **ptr1, DWORD *bytes1, void **ptr2, DWORD *bytes2, DWORD flags);
static HRESULT WINAPI compat_dsb_play(IDirectSoundBuffer *iface, DWORD reserved1, DWORD reserved2, DWORD flags);
static HRESULT WINAPI compat_dsb_set_current_position(IDirectSoundBuffer *iface, DWORD position);
static HRESULT WINAPI compat_dsb_set_format(IDirectSoundBuffer *iface, const WAVEFORMATEX *format);
static HRESULT WINAPI compat_dsb_set_volume(IDirectSoundBuffer *iface, LONG volume);
static HRESULT WINAPI compat_dsb_set_pan(IDirectSoundBuffer *iface, LONG pan);
static HRESULT WINAPI compat_dsb_set_frequency(IDirectSoundBuffer *iface, DWORD frequency);
static HRESULT WINAPI compat_dsb_stop(IDirectSoundBuffer *iface);
static HRESULT WINAPI compat_dsb_unlock(IDirectSoundBuffer *iface, void *ptr1, DWORD bytes1, void *ptr2, DWORD bytes2);
static HRESULT WINAPI compat_dsb_restore(IDirectSoundBuffer *iface);

static IDirectSoundVtbl g_compat_directsound_vtbl = {
    compat_ds_query_interface,
    compat_ds_add_ref,
    compat_ds_release,
    compat_ds_create_sound_buffer,
    compat_ds_get_caps,
    compat_ds_duplicate_sound_buffer,
    compat_ds_set_cooperative_level,
};

static IDirectSoundBufferVtbl g_compat_directsound_buffer_vtbl = {
    compat_dsb_query_interface,
    compat_dsb_add_ref,
    compat_dsb_release,
    compat_dsb_get_caps,
    compat_dsb_get_current_position,
    compat_dsb_get_format,
    compat_dsb_get_volume,
    compat_dsb_get_pan,
    compat_dsb_get_frequency,
    compat_dsb_get_status,
    compat_dsb_initialize,
    compat_dsb_lock,
    compat_dsb_play,
    compat_dsb_set_current_position,
    compat_dsb_set_format,
    compat_dsb_set_volume,
    compat_dsb_set_pan,
    compat_dsb_set_frequency,
    compat_dsb_stop,
    compat_dsb_unlock,
    compat_dsb_restore,
};

static DWORD compat_dsb_update_position(struct compat_directsound_buffer *buffer)
{
    Uint32 elapsed;
    Uint64 advanced;
    DWORD position;

    if (!(buffer->status & COMPAT_DSBSTATUS_PLAYING))
        return buffer->play_cursor;
    if (!buffer->buffer_bytes || !buffer->format.nAvgBytesPerSec)
        return buffer->play_cursor;

    elapsed = SDL_GetTicks() - buffer->play_start_ticks;
    advanced = (Uint64)elapsed * buffer->format.nAvgBytesPerSec / 1000;
    position = buffer->play_start_cursor + (DWORD)advanced;

    if (buffer->status & COMPAT_DSBSTATUS_LOOPING)
        position %= buffer->buffer_bytes;
    else if (position >= buffer->buffer_bytes)
    {
        position = buffer->buffer_bytes ? buffer->buffer_bytes - 1 : 0;
        buffer->status &= ~COMPAT_DSBSTATUS_PLAYING;
    }

    buffer->play_cursor = position;
    return position;
}

static Uint32 compat_mm_timer_callback(Uint32 interval, void *param)
{
    struct compat_mm_timer *timer = param;

    if (!timer->active || !timer->callback)
        return 0;

    timer->callback(timer->id, 0, timer->user, 0, 0);
    return interval;
}

static HRESULT WINAPI compat_ds_query_interface(IDirectSound *iface, const IID *riid, void **ppvObject)
{
    if (!ppvObject)
        return COMPAT_DSERR_GENERIC;
    *ppvObject = iface;
    compat_ds_add_ref(iface);
    return COMPAT_DS_OK;
}

static ULONG WINAPI compat_ds_add_ref(IDirectSound *iface)
{
    struct compat_directsound *sound = (struct compat_directsound *)iface;
    return ++sound->refcount;
}

static ULONG WINAPI compat_ds_release(IDirectSound *iface)
{
    struct compat_directsound *sound = (struct compat_directsound *)iface;

    if (!sound->refcount)
        return 0;
    if (--sound->refcount)
        return sound->refcount;

    free(sound);
    return 0;
}

static HRESULT WINAPI compat_ds_create_sound_buffer(IDirectSound *iface, LPCDSBUFFERDESC desc, IDirectSoundBuffer **buffer, void *outer)
{
    struct compat_directsound_buffer *created;

    (void)iface;
    (void)outer;
    if (!buffer)
        return COMPAT_DSERR_GENERIC;

    created = calloc(1, sizeof(*created));
    if (!created)
        return COMPAT_DSERR_GENERIC;

    created->iface.lpVtbl = &g_compat_directsound_buffer_vtbl;
    created->refcount = 1;
    if (desc)
    {
        created->flags = desc->dwFlags;
        created->buffer_bytes = desc->dwBufferBytes;
        if (desc->lpwfxFormat)
            created->format = *desc->lpwfxFormat;
    }
    if (created->buffer_bytes)
        created->data = calloc(1, created->buffer_bytes);

    *buffer = &created->iface;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_ds_get_caps(IDirectSound *iface, DSCAPS *caps)
{
    (void)iface;
    if (!caps)
        return COMPAT_DSERR_GENERIC;
    memset(caps, 0, sizeof(*caps));
    caps->dwSize = sizeof(*caps);
    caps->dwFlags = 0x20;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_ds_duplicate_sound_buffer(IDirectSound *iface, IDirectSoundBuffer *src, IDirectSoundBuffer **dst)
{
    (void)iface;
    (void)src;
    (void)dst;
    return COMPAT_DSERR_GENERIC;
}

static HRESULT WINAPI compat_ds_set_cooperative_level(IDirectSound *iface, HWND hwnd, DWORD level)
{
    struct compat_directsound *sound = (struct compat_directsound *)iface;
    sound->window = hwnd;
    sound->cooperative_level = level;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_query_interface(IDirectSoundBuffer *iface, const IID *riid, void **ppvObject)
{
    if (!ppvObject)
        return COMPAT_DSERR_GENERIC;
    *ppvObject = iface;
    compat_dsb_add_ref(iface);
    return COMPAT_DS_OK;
}

static ULONG WINAPI compat_dsb_add_ref(IDirectSoundBuffer *iface)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    return ++buffer->refcount;
}

static ULONG WINAPI compat_dsb_release(IDirectSoundBuffer *iface)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;

    if (!buffer->refcount)
        return 0;
    if (--buffer->refcount)
        return buffer->refcount;

    free(buffer->data);
    free(buffer);
    return 0;
}

static HRESULT WINAPI compat_dsb_get_caps(IDirectSoundBuffer *iface, void *caps)
{
    (void)iface;
    (void)caps;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_current_position(IDirectSoundBuffer *iface, DWORD *play_cursor, DWORD *write_cursor)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    DWORD position = compat_dsb_update_position(buffer);

    if (buffer->status & COMPAT_DSBSTATUS_BUFFERLOST)
        return COMPAT_DSERR_BUFFERLOST;
    if (play_cursor)
        *play_cursor = position;
    if (write_cursor)
        *write_cursor = position;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_format(IDirectSoundBuffer *iface, WAVEFORMATEX *format, DWORD size, DWORD *written)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;

    if (written)
        *written = sizeof(buffer->format);
    if (!format)
        return COMPAT_DS_OK;
    if (size < sizeof(buffer->format))
        return COMPAT_DSERR_GENERIC;
    *format = buffer->format;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_volume(IDirectSoundBuffer *iface, LONG *volume)
{
    (void)iface;
    if (volume)
        *volume = 0;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_pan(IDirectSoundBuffer *iface, LONG *pan)
{
    (void)iface;
    if (pan)
        *pan = 0;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_frequency(IDirectSoundBuffer *iface, DWORD *frequency)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    if (frequency)
        *frequency = buffer->format.nSamplesPerSec;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_get_status(IDirectSoundBuffer *iface, DWORD *status)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    compat_dsb_update_position(buffer);
    if (status)
        *status = buffer->status;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_initialize(IDirectSoundBuffer *iface, IDirectSound *directsound, LPCDSBUFFERDESC desc)
{
    (void)iface;
    (void)directsound;
    (void)desc;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_lock(IDirectSoundBuffer *iface, DWORD offset, DWORD bytes, void **ptr1, DWORD *bytes1, void **ptr2, DWORD *bytes2, DWORD flags)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    DWORD available;

    (void)flags;
    if (!buffer->data || !ptr1 || !bytes1 || offset > buffer->buffer_bytes)
        return COMPAT_DSERR_GENERIC;

    available = buffer->buffer_bytes - offset;
    if (bytes > available)
        bytes = available;
    *ptr1 = buffer->data + offset;
    *bytes1 = bytes;
    if (ptr2)
        *ptr2 = NULL;
    if (bytes2)
        *bytes2 = 0;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_play(IDirectSoundBuffer *iface, DWORD reserved1, DWORD reserved2, DWORD flags)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;

    (void)reserved1;
    (void)reserved2;
    buffer->play_start_cursor = buffer->play_cursor;
    buffer->play_start_ticks = SDL_GetTicks();
    buffer->status = COMPAT_DSBSTATUS_PLAYING;
    if (flags)
        buffer->status |= COMPAT_DSBSTATUS_LOOPING;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_set_current_position(IDirectSoundBuffer *iface, DWORD position)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;

    if (buffer->buffer_bytes)
        position %= buffer->buffer_bytes;
    else
        position = 0;
    buffer->play_cursor = position;
    buffer->play_start_cursor = position;
    buffer->play_start_ticks = SDL_GetTicks();
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_set_format(IDirectSoundBuffer *iface, const WAVEFORMATEX *format)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    if (!format)
        return COMPAT_DSERR_GENERIC;
    buffer->format = *format;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_set_volume(IDirectSoundBuffer *iface, LONG volume)
{
    (void)iface;
    (void)volume;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_set_pan(IDirectSoundBuffer *iface, LONG pan)
{
    (void)iface;
    (void)pan;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_set_frequency(IDirectSoundBuffer *iface, DWORD frequency)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    buffer->format.nSamplesPerSec = frequency;
    if (buffer->format.nBlockAlign)
        buffer->format.nAvgBytesPerSec = frequency * buffer->format.nBlockAlign;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_stop(IDirectSoundBuffer *iface)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    compat_dsb_update_position(buffer);
    buffer->status &= ~(COMPAT_DSBSTATUS_PLAYING | COMPAT_DSBSTATUS_LOOPING);
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_unlock(IDirectSoundBuffer *iface, void *ptr1, DWORD bytes1, void *ptr2, DWORD bytes2)
{
    (void)iface;
    (void)ptr1;
    (void)bytes1;
    (void)ptr2;
    (void)bytes2;
    return COMPAT_DS_OK;
}

static HRESULT WINAPI compat_dsb_restore(IDirectSoundBuffer *iface)
{
    struct compat_directsound_buffer *buffer = (struct compat_directsound_buffer *)iface;
    buffer->status &= ~COMPAT_DSBSTATUS_BUFFERLOST;
    return COMPAT_DS_OK;
}

static char *append_path_component(const char *base, const char *component)
{
    size_t base_len = strlen(base);
    size_t comp_len = strlen(component);
    int needs_slash = base_len > 0 && strcmp(base, "/") != 0;
    char *result = malloc(base_len + needs_slash + comp_len + 1);

    if (!result)
        return NULL;

    memcpy(result, base, base_len);
    if (needs_slash)
        result[base_len++] = '/';
    memcpy(result + base_len, component, comp_len + 1);
    return result;
}

static char *find_case_insensitive_component(const char *dirpath, const char *component)
{
    DIR *dir;
    struct dirent *entry;
    char *match = NULL;

    dir = opendir(dirpath);
    if (!dir)
        return NULL;

    while ((entry = readdir(dir)) != NULL)
    {
        if (strcasecmp(entry->d_name, component) == 0)
        {
            match = malloc(strlen(entry->d_name) + 1);
            if (match)
                strcpy(match, entry->d_name);
            break;
        }
    }

    closedir(dir);
    return match;
}

char *resolve_case_insensitive_path(const char *path)
{
    char *input;
    char *cursor;
    char *component;
    char *resolved;
    int absolute;

    if (!path || !path[0])
        return NULL;

    input = malloc(strlen(path) + 1);
    if (!input)
        return NULL;
    strcpy(input, path);

    absolute = path[0] == '/';
    resolved = malloc(absolute ? 2 : 2);
    if (!resolved)
    {
        free(input);
        return NULL;
    }
    strcpy(resolved, absolute ? "/" : ".");

    cursor = input;
    while (*cursor)
    {
        while (*cursor == '/')
            cursor++;
        if (!*cursor)
            break;

        component = cursor;
        while (*cursor && *cursor != '/')
            cursor++;
        if (*cursor)
            *cursor++ = '\0';

        if (strcmp(component, ".") == 0)
            continue;

        if (strcmp(component, "..") == 0)
        {
            free(resolved);
            free(input);
            return NULL;
        }

        {
            char *match = find_case_insensitive_component(resolved, component);
            char *next;

            if (!match)
            {
                free(resolved);
                free(input);
                return NULL;
            }

            next = append_path_component(resolved, match);
            free(match);
            free(resolved);
            if (!next)
            {
                free(input);
                return NULL;
            }
            resolved = next;
        }
    }

    free(input);
    if (!absolute && strncmp(resolved, "./", 2) == 0)
    {
        char *trimmed = malloc(strlen(resolved + 2) + 1);
        if (!trimmed)
        {
            free(resolved);
            return NULL;
        }
        strcpy(trimmed, resolved + 2);
        free(resolved);
        resolved = trimmed;
    }
    return resolved;
}

static inline HANDLE new_handle(unsigned int type, void *data)
{
    unsigned int i;
    for (i = 0; i < 1024; i++)
    {
        if (!handles[i])
        {
            handles[i] = data;
            return (HANDLE)((type << 16) | i);
        }
    }
    return (HANDLE)-1;
}

static inline void *lookup_handle(unsigned int type, HANDLE h)
{
    if (type != ((DWORD)h >> 16))
        return NULL;
    if ((WORD)h >= 1024)
        return NULL;
    return handles[(WORD)h];
}

// Debug functions
VOID WINAPI DebugBreak()
{
#ifdef __EMSCRIPTEN__
#else
    asm volatile ("int3");
#endif
}

VOID WINAPI OutputDebugStringA(LPCSTR lpOutputString)
{
    fprintf(stderr, "%s", lpOutputString);
}

// Memory functions
BOOL WINAPI HeapDestroy(HANDLE hHeap)
{
    DebugBreak();
}

// CRT functions
unsigned int _control87(unsigned int new_, unsigned int mask)
{
    if (new_ == 0x300 && mask == 0x300)
        fesetround(FE_TOWARDZERO);
    else
        DebugBreak();
}

unsigned int _controlfp(unsigned int new_, unsigned int mask)
{
    if (new_ == 0x300 && mask == 0x300)
        fesetround(FE_TOWARDZERO);
    else
        DebugBreak();
}

uintptr_t _beginthread(void( __cdecl *start_address )( void * ), unsigned int stack_size, void *arglist)
{
    pthread_t handle;

#ifdef __EMSCRIPTEN__
    fprintf(stderr, "%s: unsupported\n");
    while (1) {}
#endif

    if (pthread_create(&handle, NULL, start_address, arglist))
        return (uintptr_t)-1;

    return new_handle(HANDLE_THREAD, handle);
}

char *_strrev(char *str)
{
    char *begin, *end;

    if (!str[0])
        return str;

    begin = str;
    end = str + strlen(str) - 1;

    while (begin < end)
    {
        char tmp = *end;
        *end = *begin;
        *begin = tmp;
    }

    return str;
}

char *_itoa(int val, char *s, int radix)
{
    return SDL_itoa(val, s, radix);
}

char *_utoa(unsigned int val, char *s, int radix)
{
    return SDL_uitoa(val, s, radix);
}

wchar_t *_itow(int val, wchar_t *s, int radix)
{
    char tmp[32];
    unsigned int i;

    _itoa(val, tmp, radix);
    for (i = 0; tmp[i]; i++)
        s[i] = tmp[i];
    s[i] = 0;

    return s;
}

void _makepath(char *path, const char *drive, const char *dir, const char *fname, const char *ext)
{
    sprintf(path, "%s%s%s%s%s%s%s",
            drive,
            drive && drive[0] && (!dir || dir[0] != '\\') ? "\\" : "",
            dir,
            dir && dir[0] && dir[strlen(dir) - 1] != '\\' ? "\\" : "",
            fname,
            ext && ext[0] && ext[0] != '.' ? "." : "",
            ext);
    //dprintf("%s: (\"%s\", \"%s\", \"%s\", \"%s\") = \"%s\"", __FUNCTION__, drive, dir, fname, ext, path);
}

void _splitpath(const char *path, char *drive, char *dir, char *fname, char *ext)
{
    const char *tmp;

    if (isalpha(path[0]) && path[1] == ':')
    {
        if (drive)
        {
            drive[0] = path[0];
            drive[1] = path[1];
            drive[2] = 0;
        }
        path += 2;
    }
    else if (drive)
    {
        drive[0] = 0;
    }

    tmp = strrchr(path, '\\');
    if (tmp)
    {
        if (dir)
        {
            memcpy(dir, path, tmp + 1 - path);
            dir[tmp + 1 - path] = 0;
        }
        path = tmp + 1;
    }
    else if (dir)
    {
        dir[0] = 0;
    }

    tmp = strrchr(path, '.');
    if (tmp)
    {
        if (fname)
        {
            memcpy(fname, path, tmp - path);
            fname[tmp - path] = 0;
        }
        path = tmp;
    }
    else if (fname)
    {
        fname[0] = 0;
    }

    if (ext)
    {
        strcpy(ext, path);
    }

    //dprintf("%s: \"%s\" = (\"%s\", \"%s\", \"%s\", \"%s\")", __FUNCTION__, path, drive, dir, fname, ext);
}

// Misc functions
BOOL WINAPI CloseHandle(HANDLE hObject)
{
    switch ((DWORD)hObject >> 16)
    {
    case 0:
        close((WORD)hObject);
        break;
    case HANDLE_THREAD:
        handles[(WORD)hObject] = NULL;
        break;
    case HANDLE_MUTEX:
        {
            pthread_mutex_t *m = lookup_handle(HANDLE_MUTEX, hObject);
            pthread_mutex_destroy(m);
            free(m);
        }
        handles[(WORD)hObject] = NULL;
        break;
    default:
        DebugBreak();
        return FALSE;
    }

    return TRUE;
}

DWORD WINAPI GetModuleFileNameA(HMODULE hModule, LPSTR lpFileName, DWORD nSize)
{
    unsigned int i, size = strlen(progname);
    DWORD ret;

    if (hModule != NULL)
        DebugBreak();

    if (size < nSize)
    {
        strcpy(lpFileName, progname);
        ret = size;
    }
    else if (nSize)
    {
        memcpy(lpFileName, progname, nSize);
        lpFileName[nSize-1] = 0;
        ret = nSize;
    }

    for (i = 0; lpFileName[i]; i++)
        if (lpFileName[i] == '/')
            lpFileName[i] = '\\';

    return ret;
}

DWORD WINAPI GetLastError()
{
    return last_error;
}

BOOL WINAPI GetVersionExA(LPOSVERSIONINFOA lpVersionInformation)
{
    lpVersionInformation->dwOSVersionInfoSize = sizeof(OSVERSIONINFOA);
    lpVersionInformation->dwMajorVersion = 5;
    lpVersionInformation->dwMinorVersion = 1;
    return TRUE;
}

LONG InterlockedExchange(volatile LONG *Target, LONG Value)
{
    return __sync_lock_test_and_set(Target, Value);
}

LONG InterlockedDecrement(volatile LONG *Addend)
{
    return __sync_fetch_and_sub(Addend, 1);
}

LONG InterlockedIncrement(volatile LONG *Addend)
{
    return __sync_fetch_and_add(Addend, 1);
}

int WINAPI MessageBoxA(HWND hWnd, LPCSTR lpText, LPCSTR lpCaption, UINT uType)
{
    DebugBreak();
}

int WINAPI MulDiv(int nNumber, int nNumerator, int nDenominator)
{
    DebugBreak();
}

HINSTANCE WINAPI ShellExecuteA(HWND hwnd, LPCSTR lpOperation, LPCSTR lpFile, LPCSTR lpParameters, LPCSTR lpDirectory, INT nShowCmd)
{
    DebugBreak();
}

int WINAPI WideCharToMultiByte(UINT CodePage, DWORD dwFlags, LPCWSTR lpWideCharStr, int cchWideChar, LPSTR lpMultiByteStr, int cbMultiByte, LPCCH lpDefaultChar, LPBOOL lpUsedDefaultChar)
{
    DebugBreak();
}

// Socket functions
int WINAPI WSAStartup(WORD wVersionRequested, struct WSAData *lpWSAData)
{
    return 0;
}

int WINAPI WSACleanup()
{
    return 0;
}

SOCKET WINAPI socket(int domain, int type, int protocol)
#undef socket
{
#ifdef __EMSCRIPTEN__
    static int fd = 1024;
    return fd++;
#else
    return socket(domain, type, protocol);
#endif
}

int WINAPI closesocket(SOCKET s)
{
    close(s);
    return 0;
}

char *WINAPI inet_ntoa(struct in_addr compat_addr)
#undef in_addr
#undef inet_ntoa
{
    struct in_addr addr;

    addr.s_addr = compat_addr.S_un.S_addr;
    return inet_ntoa(addr);
}

int WINAPI setsockopt(SOCKET s, int level, int opt, const void *value, unsigned int len)
#undef setsockopt
{
#ifdef __EMSCRIPTEN__
    return 0;
#else
    return setsockopt(s, level, opt, value, len);
#endif
}

int WINAPI ioctlsocket(SOCKET s, long cmd, unsigned long *argp)
{
    int ret;

    switch (cmd)
    {
    case 0x4004667f: // FIONREAD
#ifdef __EMSCRIPTEN__
        *argp = EM_ASM_INT({
            return network.available($0);
        }, s);
        ret = 0;
#else
        ret = ioctl(s, FIONREAD, argp);
#endif
        break;
    default:
        DebugBreak();
        ret = -1;
        break;
    }

    return ret;
}

#ifdef __EMSCRIPTEN__
void build_server_info(void *arg)
{
    static char oldbuf[256];
    static int oldlen;
    char buf[256];
    int dummy[3] = { 0, 0, 0 };
    int length, ready;

    ready = EM_ASM_INT({
        return network.isready();
    });

    if (!ready) {
        oldlen = 0;
        return;
    }
    
    length = sub_554040(dummy, sizeof(buf), buf);
    if (oldlen != length || memcmp(buf, oldbuf, length) != 0)
    {
        memcpy(oldbuf, buf, length);
        oldlen = length;

        EM_ASM_({
            network.registerServer($1 > 0 ? Module.HEAPU8.slice($0, $0 + $1) : null);
        }, oldbuf, oldlen);
    }
}
#endif

int WINAPI bind(int sockfd, const struct sockaddr *addr, unsigned int addrlen)
#undef bind
{
    int ret;
#ifdef __EMSCRIPTEN__
    static long updater = -1;

    EM_ASM_({
        network.bind($0);
    }, sockfd);

    if (updater == -1)
        updater = emscripten_set_interval(build_server_info, 1000, NULL);

    ret = 0;
#else
    ret = bind(sockfd, addr, addrlen);
#endif
    if (ret >= 0)
        return ret;

    switch (errno)
    {
    case EADDRINUSE:
        last_socket_error = 10048u;
        break;
    default:
        DebugBreak();
        break;
    }

    return -1;
}

int WINAPI recvfrom(int sockfd, void *buffer, unsigned int length, int flags, struct sockaddr *addr, unsigned int *addrlen)
#undef recvfrom
{
#ifdef __EMSCRIPTEN__
    int ret;
    struct sockaddr_in *addr_in = addr;
    ret = EM_ASM_INT(({
        const [ remote, port, data ] = network.recvfrom($4);
        if (remote === null)
            return -1;
        const length = Math.min(data.length, $1);
        Module.HEAPU8.set(new Uint8Array(data, 0, length), $0);
        if ($2) {
            Module.HEAPU32[$2 >> 2] = remote;
        }
        if ($3) {
            Module.HEAPU8[$3] = port >> 8;
            Module.HEAPU8[$3 + 1] = port >> 0;
        }
        return length;
    }), buffer, length, addr_in ? &addr_in->sin_addr : NULL, addr_in ? &addr_in->sin_port : NULL, sockfd);
    if (addr_in)
        addr_in->sin_family = AF_INET;
    return ret;
#else
    return recvfrom(sockfd, buffer, length, flags, addr, addrlen);
#endif
}

int WINAPI sendto(int sockfd, void *buffer, unsigned int length, int flags, const struct sockaddr *addr, unsigned int addrlen)
#undef sendto
{
#ifdef __EMSCRIPTEN__
    struct sockaddr_in *addr_in = addr;
    unsigned int dest = addr_in->sin_addr.s_addr;
    unsigned char *p = buffer;

    // broadcast packet
    if (dest == 0xffffffff)
    {
        if (p[2] == 12)
        {
            EM_ASM_({
                network.isready() && network.listServers($0, $1)
            }, *(DWORD *)(p + 8), sockfd);
        }
        return length;
    }

    return EM_ASM_INT({
        network.sendto($3, $2, $4, Module.HEAPU8.slice($0, $0 + $1));
        return $1;
    }, buffer, length, dest, sockfd, ntohs(addr_in->sin_port));
#else
    return sendto(sockfd, buffer, length, flags, addr, addrlen);
#endif
}

int WINAPI WSAGetLastError()
{
    return last_socket_error;
}

// Time functions
// compatGetDateFormatA(Locale=2048, dwFlags=1, lpDate=0x1708c6a4, lpFormat=0x00000000, lpDateStr="nox.str:Warrior", cchDate=256) at compat.c:1001
int WINAPI GetDateFormatA(LCID Locale, DWORD dwFlags, const SYSTEMTIME *lpDate, LPCSTR lpFormat, LPSTR lpDateStr, int cchDate)
{
    if (Locale != 0x800 || dwFlags != 1 || lpFormat)
        DebugBreak();

    // default locale, short date (M/d/yy)
    return snprintf(lpDateStr, cchDate, "%d/%d/%02d", lpDate->wMonth, lpDate->wDay, lpDate->wYear % 100);
}

int WINAPI GetTimeFormatA(LCID Locale, DWORD dwFlags, const SYSTEMTIME *lpTime, LPCSTR lpFormat, LPSTR lpTimeStr, int cchTime)
{
    if (Locale != 0x800 || dwFlags != 14 || lpFormat)
        DebugBreak();

    // default locale, 24 hour, no time marker, no seconds
    return snprintf(lpTimeStr, cchTime, "%02d:%02d", lpTime->wHour, lpTime->wMinute);
}

BOOL WINAPI SystemTimeToFileTime(const SYSTEMTIME *lpSystemTime, LPFILETIME lpFileTime)
{
    QWORD t;
    struct tm tm;

    tm.tm_sec = lpSystemTime->wSecond;
    tm.tm_min = lpSystemTime->wMinute;
    tm.tm_hour = lpSystemTime->wHour;
    tm.tm_mday = lpSystemTime->wDay;
    tm.tm_mon = lpSystemTime->wMonth;
    tm.tm_year = lpSystemTime->wYear - 1900;
    tm.tm_isdst = -1;
    tm.tm_zone = NULL;
    tm.tm_gmtoff = 0;

    t = mktime(&tm);
    t = (t * 1000 + lpSystemTime->wMilliseconds) * 10000;
    lpFileTime->dwLowDateTime = t & 0xffffffff;
    lpFileTime->dwHighDateTime = t >> 32;
}

LONG WINAPI CompareFileTime(const FILETIME *lpFileTime1, const FILETIME *lpFileTime2)
{
    if (lpFileTime1->dwHighDateTime != lpFileTime2->dwHighDateTime)
        return (LONG)lpFileTime1->dwHighDateTime - (LONG)lpFileTime2->dwHighDateTime;
    if (lpFileTime1->dwLowDateTime != lpFileTime2->dwLowDateTime)
        return (LONG)lpFileTime1->dwLowDateTime - (LONG)lpFileTime2->dwLowDateTime;
}

VOID WINAPI GetLocalTime(LPSYSTEMTIME lpSystemTime)
{
    time_t t;
    struct tm *tm;

    time(&t);
    tm = localtime(&t);

    lpSystemTime->wYear = tm->tm_year + 1900;
    lpSystemTime->wMonth = tm->tm_mon;
    lpSystemTime->wDayOfWeek = tm->tm_wday;
    lpSystemTime->wDay = tm->tm_mday;
    lpSystemTime->wHour = tm->tm_hour;
    lpSystemTime->wMinute = tm->tm_min;
    lpSystemTime->wSecond = tm->tm_sec;
    lpSystemTime->wMilliseconds = 0;
}

BOOL WINAPI QueryPerformanceCounter(LARGE_INTEGER *lpPerformanceCount)
{
    struct timeval tv;

    gettimeofday(&tv, NULL);
    lpPerformanceCount->QuadPart = (QWORD)tv.tv_sec * 1000000 + tv.tv_usec;
    return TRUE;
}

BOOL WINAPI QueryPerformanceFrequency(LARGE_INTEGER *lpFrequency)
{
    lpFrequency->QuadPart = 1000000;
    return TRUE;
}

VOID WINAPI Sleep(DWORD dwMilliseconds)
{
#ifndef __EMSCRIPTEN__
    SDL_Delay(dwMilliseconds);
#endif
}

DWORD WINAPI timeGetTime()
{
    return SDL_GetTicks();
}

MMRESULT WINAPI timeBeginPeriod(UINT uPeriod)
{
    (void)uPeriod;
    return 0;
}

MMRESULT WINAPI timeEndPeriod(UINT uPeriod)
{
    (void)uPeriod;
    return 0;
}

MMRESULT WINAPI timeSetEvent(UINT uDelay, UINT uResolution, LPTIMECALLBACK lpTimeProc, DWORD dwUser, UINT fuEvent)
{
    struct compat_mm_timer *timer = NULL;
    size_t i;

    (void)uResolution;
    (void)fuEvent;
    if (!lpTimeProc || !uDelay)
        return 0;

    for (i = 0; i < SDL_arraysize(g_mm_timers); ++i)
    {
        if (!g_mm_timers[i].active)
        {
            timer = &g_mm_timers[i];
            break;
        }
    }
    if (!timer)
        return 0;

    memset(timer, 0, sizeof(*timer));
    timer->id = g_next_timer_id++;
    if (!g_next_timer_id)
        g_next_timer_id = 1;
    timer->callback = lpTimeProc;
    timer->user = dwUser;
    timer->active = 1;
    timer->timer = SDL_AddTimer(uDelay, compat_mm_timer_callback, timer);
    if (!timer->timer)
    {
        memset(timer, 0, sizeof(*timer));
        return 0;
    }
    return timer->id;
}

MMRESULT WINAPI timeKillEvent(MMRESULT uTimerID)
{
    size_t i;

    for (i = 0; i < SDL_arraysize(g_mm_timers); ++i)
    {
        if (g_mm_timers[i].active && g_mm_timers[i].id == uTimerID)
        {
            g_mm_timers[i].active = 0;
            if (g_mm_timers[i].timer)
                SDL_RemoveTimer(g_mm_timers[i].timer);
            memset(&g_mm_timers[i], 0, sizeof(g_mm_timers[i]));
            return 0;
        }
    }
    return 1;
}

BOOL WINAPI PeekMessageA(LPMSG lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax, UINT wRemoveMsg)
{
    (void)lpMsg;
    (void)hWnd;
    (void)wMsgFilterMin;
    (void)wMsgFilterMax;
    (void)wRemoveMsg;
    return FALSE;
}

BOOL WINAPI GetMessageA(LPMSG lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax)
{
    (void)lpMsg;
    (void)hWnd;
    (void)wMsgFilterMin;
    (void)wMsgFilterMax;
    return FALSE;
}

BOOL WINAPI TranslateMessage(const MSG *lpMsg)
{
    (void)lpMsg;
    return TRUE;
}

LRESULT WINAPI DispatchMessageA(const MSG *lpMsg)
{
    (void)lpMsg;
    return 0;
}

DWORD WINAPI GetTickCount()
{
    return timeGetTime();
}

int WINAPI ShowCursor(BOOL bShow)
{
    SDL_ShowCursor(bShow ? SDL_ENABLE : SDL_DISABLE);
    return 0;
}

HDC WINAPI GetWindowDC(HWND hWnd)
{
    (void)hWnd;
    return 0;
}

int WINAPI ReleaseDC(HWND hWnd, HDC hDC)
{
    (void)hWnd;
    (void)hDC;
    return 1;
}

BOOL WINAPI TextOutA(HDC hdc, int x, int y, LPCSTR lpString, int c)
{
    (void)hdc;
    (void)x;
    (void)y;
    (void)lpString;
    (void)c;
    return TRUE;
}

HRESULT WINAPI DirectSoundCreate(const GUID *lpGuidDevice, LPDIRECTSOUND *ppDS, void *pUnkOuter)
{
    struct compat_directsound *sound;

    (void)lpGuidDevice;
    (void)pUnkOuter;
    if (!ppDS)
        return COMPAT_DSERR_GENERIC;

    sound = calloc(1, sizeof(*sound));
    if (!sound)
        return COMPAT_DSERR_GENERIC;

    sound->iface.lpVtbl = &g_compat_directsound_vtbl;
    sound->refcount = 1;
    *ppDS = &sound->iface;
    return COMPAT_DS_OK;
}

// File functions
char *dos_to_unix(const char *path)
{
    int i, len = strlen(path);
    char *str = malloc(len + 1);

    if (path[0] == 'C' && path[1] == ':')
        path += 2;
    
    for (i = 0; path[i]; i++)
    {
        if (path[i] == '\\')
            str[i] = '/';
        else
            str[i] = path[i];
    }
    str[i] = 0;

    return str;
}


struct _FIND_FILE
{
    size_t idx;
    glob_t globbuf;
};

static void fill_find_data(const char *path, LPWIN32_FIND_DATAA lpFindFileData)
{
    struct stat st;
    const char *filename = strrchr(path, '/');
    filename = filename == NULL ? path : filename + 1;

    stat(path, &st);

    memset(lpFindFileData, 0, sizeof(*lpFindFileData));
    lpFindFileData->dwFileAttributes = S_ISDIR(st.st_mode) ? FILE_ATTRIBUTE_DIRECTORY : FILE_ATTRIBUTE_NORMAL;
    lpFindFileData->nFileSizeHigh = st.st_size >> 32;
    lpFindFileData->nFileSizeLow = st.st_size;
    strcpy(lpFindFileData->cFileName, filename);

    // dprintf("%s: attr=0x%X, filename=%s", __FUNCTION__, lpFindFileData->dwFileAttributes, lpFindFileData->cFileName);
}

HANDLE WINAPI FindFirstFileA(LPCSTR lpFileName, LPWIN32_FIND_DATAA lpFindFileData)
{
    char *converted = dos_to_unix(lpFileName);
    int len = strlen(converted);
    struct _FIND_FILE *ff = calloc(sizeof(*ff), 1);

    // dprintf("%s: converted=%s", __FUNCTION__, converted);

    // glob is close enough to the semantics
    if (len >= 2 && converted[len - 2] == '.' && converted[len - 1] == '*')
        converted[len - 2] = 0;
    if (glob(converted, GLOB_NOESCAPE, NULL, &ff->globbuf))
    {
        free(ff);
        return (HANDLE)-1;
    }

    fill_find_data(ff->globbuf.gl_pathv[ff->idx++], lpFindFileData);
    return (HANDLE)ff;
}

BOOL WINAPI FindNextFileA(HANDLE hFindFile, LPWIN32_FIND_DATAA lpFindFileData)
{
    struct _FIND_FILE *ff = (struct _FIND_FILE *)hFindFile;

    if (ff->idx >= ff->globbuf.gl_pathc)
    {
        last_error = ERROR_NO_MORE_FILES;
        return FALSE;
    }

    fill_find_data(ff->globbuf.gl_pathv[ff->idx++], lpFindFileData);
    return TRUE;
}

BOOL WINAPI FindClose(HANDLE hFindFile)
{
    struct _FIND_FILE *ff = (struct _FIND_FILE *)hFindFile;

    globfree(&ff->globbuf);
    free(ff);
    return TRUE;
}

HANDLE WINAPI CreateFileA(LPCSTR lpFileName, DWORD dwDesiredAccess, DWORD dwShareMode, LPSECURITY_ATTRIBUTES lpSecurityAttributes, DWORD dwCreationDisposition, DWORD dwFlagsAndAttributes, HANDLE hTemplateFile)
{
    char *converted = dos_to_unix(lpFileName);
    int fd, flags;

    switch (dwDesiredAccess)
    {
    case GENERIC_READ:
        flags = O_RDONLY;
        break;
    case GENERIC_WRITE:
        flags = O_WRONLY;
        break;
    case GENERIC_READ|GENERIC_WRITE:
        flags = O_RDWR;
        break;
    default:
        DebugBreak();
    }

    switch (dwCreationDisposition)
    {
    case CREATE_NEW:
        flags |= O_CREAT | O_EXCL;
        break;
    case CREATE_ALWAYS:
        flags |= O_CREAT | O_TRUNC;
        break;
    default:
    case OPEN_EXISTING:
        break;
    case OPEN_ALWAYS:
        flags |= O_CREAT;
        break;
    case TRUNCATE_EXISTING:
        flags |= O_TRUNC;
        break;
    }
    
    fd = open(converted, flags, 0666);
    //printf("%s: %s = %08x\n", __FUNCTION__, converted, fd);
    free(converted);
    return (HANDLE)fd;
}

BOOL WINAPI ReadFile(HANDLE hFile, LPVOID lpBuffer, DWORD nNumberOfBytesToRead, LPDWORD lpNumberOfBytesRead, LPOVERLAPPED lpOverlapped)
{
    int fd = (int)hFile;
    int ret;

    if (lpOverlapped)
        DebugBreak();

    ret = read(fd, lpBuffer, nNumberOfBytesToRead);
    if (ret >= 0)
    {
        *lpNumberOfBytesRead = ret;
        last_error = 0;
        return TRUE;
    }
    else
    {
        *lpNumberOfBytesRead = 0;
        last_error = 2; // FIXME
        return FALSE;
    }
}

DWORD WINAPI SetFilePointer(HANDLE hFile, LONG lDistanceToMove, PLONG lpDistanceToMoveHigh, DWORD dwMoveMethod)
{
    int fd = (int)hFile;
    int whence;
    off_t offset;

    if (lpDistanceToMoveHigh && *lpDistanceToMoveHigh)
        DebugBreak();

    switch (dwMoveMethod)
    {
    case FILE_BEGIN:
        whence = SEEK_SET;
        break;
    case FILE_CURRENT:
        whence = SEEK_CUR;
        break;
    case FILE_END:
        whence = SEEK_END;
        break;
    default:
        DebugBreak();
    }

    if (lpDistanceToMoveHigh)
        *lpDistanceToMoveHigh = 0;

    offset = lseek(fd, lDistanceToMove, whence);
    if (offset != (off_t)-1)
        last_error = 0;
    else
        last_error = 2; // FIXME
    return offset;
}

BOOL WINAPI CopyFileA(LPCSTR lpExistingFileName, LPCSTR lpNewFileName, BOOL bFailIfExists)
{
    char buf[1024];
    int rfd = _open(lpExistingFileName, O_RDONLY);
    int wfd;

    if (rfd < 0)
        return FALSE;
    
    wfd = _open(lpNewFileName, bFailIfExists ? O_WRONLY | O_CREAT | O_EXCL : O_WRONLY | O_CREAT | O_TRUNC, 0666);
    if (wfd < 0)
    {
error:
        close(rfd);
        return FALSE;
    }

    while (1)
    {
        int ret = read(rfd, buf, sizeof(buf));
        if (ret <= 0)
            break;
        if (write(wfd, buf, ret) != ret)
            goto error;
    }

    close(rfd);
    close(wfd);
    return TRUE;
}

BOOL WINAPI DeleteFileA(LPCSTR lpFileName)
{
    return _unlink(lpFileName) == 0;
}

BOOL WINAPI MoveFileA(LPCSTR lpExistingFileName, LPCSTR lpNewFileName)
{
    char *converted = dos_to_unix(lpExistingFileName);
    free(converted);
    dprintf("%s\n", __FUNCTION__);
    DebugBreak();
}

BOOL WINAPI CreateDirectoryA(LPCSTR lpPathName, LPSECURITY_ATTRIBUTES lpSecurityAttributes)
{
    return _mkdir(lpPathName) == 0;
}

BOOL WINAPI RemoveDirectoryA(LPCSTR lpPathName)
{
    int result;
    char *converted = dos_to_unix(lpPathName);
    result = rmdir(converted);
    free(converted);
    return result == 0;
}

DWORD WINAPI GetCurrentDirectoryA(DWORD nBufferLength, LPSTR lpBuffer)
{
    if (_getcwd(lpBuffer, nBufferLength))
        return strlen(lpBuffer);
    else
        return 0;
}

BOOL WINAPI SetCurrentDirectoryA(LPCSTR lpPathName)
{
    int result;
    char *converted = dos_to_unix(lpPathName);
    result = chdir(converted);
    free(converted);
    return result == 0;
}

int _open(const char *filename, int oflag, ...)
{
    va_list ap;
    int fd;
    char *converted = dos_to_unix(filename);

    va_start(ap, oflag);
    fd = open(converted, oflag, va_arg(ap, int));
    va_end(ap);

    free(converted);
    return fd;
}

int _chmod(const char *filename, int mode)
{
    int result;
    char *converted = dos_to_unix(filename);
    result = chmod(converted, mode);
    free(converted);
    return result;
}

int _access(const char *filename, int mode)
{
    int result;
    char *converted = dos_to_unix(filename);
    result = access(converted, mode);
    if (result)
    {
        char *fallback = resolve_case_insensitive_path(converted);
        if (fallback)
        {
            result = access(fallback, mode);
            free(fallback);
        }
    }
    free(converted);
    return result;
}

long _lseek(int fd, long offset, int origin)
{
    return lseek(fd, offset, origin);
}

long _filelength(int fd)
{
    struct stat st;

    if (fstat(fd, &st) != 0)
        return -1;
    return st.st_size;
}

int _stat(const char *path, struct _stat *buffer)
{
    int result;
    char *converted = dos_to_unix(path);
    struct stat st;

    result = stat(converted, &st);
    if (result)
    {
        char *fallback = resolve_case_insensitive_path(converted);
        if (fallback)
        {
            result = stat(fallback, &st);
            free(fallback);
        }
    }
    free(converted);

    if (result)
        return result;

    buffer->st_dev = st.st_dev;
    buffer->st_ino = st.st_ino;
    buffer->st_mode = st.st_mode;
    buffer->st_nlink = st.st_nlink;
    buffer->st_uid = st.st_uid;
    buffer->st_gid = st.st_gid;
    buffer->st_rdev = st.st_rdev;
    buffer->st_size = st.st_size;
#ifdef __APPLE__
    buffer->st_mtime = st.st_mtimespec.tv_sec;
    buffer->st_atime = st.st_atimespec.tv_sec;
    buffer->st_ctime = st.st_ctimespec.tv_sec;
#else
    buffer->st_mtime = st.st_mtim.tv_sec;
    buffer->st_atime = st.st_atim.tv_sec;
    buffer->st_ctime = st.st_ctim.tv_sec;
#endif
    return result;
}

int _mkdir(const char *path)
{
    int result;
    char *converted = dos_to_unix(path);
    result = mkdir(converted, 0777);
    free(converted);
    return result;
}

int _unlink(const char *filename)
{
    int result;
    char *converted = dos_to_unix(filename);
    result = unlink(converted);
    free(converted);
    return result;
}

char *_getcwd(char *buffer, int maxlen)
{
    int i;

    if (maxlen < 2)
        return NULL;

    strcpy(buffer, "C:");
    if (!getcwd(buffer + 2, maxlen - 2))
        return NULL;
    
    for (i = 0; buffer[i]; i++)
        if (buffer[i] == '/')
            buffer[i] = '\\';

    return buffer;
}

FILE *fopen(const char *path, const char *mode)
#undef fopen
{
    FILE *result;
    char *converted = dos_to_unix(path);
    result = fopen(converted, mode);
    if (!result && mode && mode[0] == 'r')
    {
        char *fallback = resolve_case_insensitive_path(converted);
        if (fallback)
        {
            result = fopen(fallback, mode);
            free(fallback);
        }
    }
    //printf("%s: %s = %08x\n", __FUNCTION__, converted, (int)result);
    free(converted);
    return result;
}

char *fgets(char *str, int size, FILE *stream)
#undef fgets
{
    char *result;
    
    result = fgets(str, size, stream);
    if (result)
    {
        // XXX hack for text-mode line conversion
        size = strlen(result);
        if (size >= 2 && result[size - 1] == '\n' && result[size - 2] == '\r')
        {
            result[size - 2] = '\n';
            result[size - 1] = '\0';
        }
    }

    return result;
}

// Registry functions
LSTATUS WINAPI RegCreateKeyExA(HKEY hKey, LPCSTR lpSubKey, DWORD Reserved, LPSTR lpClass, DWORD dwOptions, REGSAM samDesired, const LPSECURITY_ATTRIBUTES lpSecurityAttributes, PHKEY phkResult, LPDWORD lpdwDisposition)
{
    DebugBreak();
}

LSTATUS WINAPI RegOpenKeyExA(HKEY hKey, LPCSTR lpSubKey, DWORD ulOptions, REGSAM samDesired, PHKEY phkResult)
{
    HKEY hkResult;
    const char *root;

    if (hKey == HKEY_LOCAL_MACHINE)
        root = "HKEY_LOCAL_MACHINE";
    else
        root = hKey->path;

    hkResult = calloc(sizeof(*hkResult), 1);
    hkResult->path = malloc(strlen(root) + strlen(lpSubKey) + 2);
    sprintf(hkResult->path, "%s\\%s", root, lpSubKey);
    *phkResult = hkResult;
    return 0;
}

LSTATUS WINAPI RegQueryValueExA(HKEY hKey, LPCSTR lpValueName, LPDWORD lpReserved, LPDWORD lpType, LPBYTE lpData, LPDWORD lpcbData)
{
    static const char sku[] = "1";
    const char *value = NULL;
    DWORD type = 1;
    DWORD required_size = 0;

    dprintf("%s: key=\"%s\", value=\"%s\"", __FUNCTION__, hKey->path, lpValueName);

    if (strcmp(hKey->path, "HKEY_LOCAL_MACHINE\\SOFTWARE\\Westwood\\Nox") != 0)
        return 3;

    if (strcmp(lpValueName, "Serial") == 0)
    {
        value = "0000000000000000000000";
        required_size = sizeof("0000000000000000000000");
    }
    else if (strcmp(lpValueName, "SKU") == 0)
    {
        value = sku;
        required_size = sizeof(sku);
    }
    else
    {
        return 3;
    }

    if (lpType)
        *lpType = type;

    if (!lpcbData)
        return lpData ? 87 : 0;

    if (!lpData)
    {
        *lpcbData = required_size;
        return 0;
    }

    if (*lpcbData < required_size)
    {
        *lpcbData = required_size;
        return 234;
    }

    if (strcmp(lpValueName, "Serial") == 0)
    {
        DWORD i;
        for (i = 0; i < required_size - 1; i++)
        {
            lpData[i] = (rand() % 10) + '0';
        }
        lpData[required_size - 1] = '\0';
    }
    else
    {
        memcpy(lpData, value, required_size);
    }

    *lpcbData = required_size;
    return 0;
}

LSTATUS WINAPI RegSetValueExA(HKEY hKey, LPCSTR lpValueName, DWORD Reserved, DWORD dwType, const BYTE *lpData, DWORD cbData)
{
    DebugBreak();
}

LSTATUS WINAPI RegCloseKey(HKEY hKey)
{
    free(hKey->path);
    free(hKey);
}

// Synchronization functions
VOID WINAPI InitializeCriticalSection(LPCRITICAL_SECTION lpCriticalSection)
{
    lpCriticalSection->opaque = SDL_CreateMutex();
}

VOID WINAPI DeleteCriticalSection(LPCRITICAL_SECTION lpCriticalSection)
{
    SDL_DestroyMutex(lpCriticalSection->opaque);
}

VOID WINAPI EnterCriticalSection(LPCRITICAL_SECTION lpCriticalSection)
{
    SDL_LockMutex(lpCriticalSection->opaque);
}

VOID WINAPI LeaveCriticalSection(LPCRITICAL_SECTION lpCriticalSection)
{
    SDL_UnlockMutex(lpCriticalSection->opaque);
}

HANDLE WINAPI CreateMutexA(LPSECURITY_ATTRIBUTES lpSecurityAttributes, BOOL bInitialOwner, LPCSTR lpName)
{
    pthread_mutex_t *m = malloc(sizeof(pthread_mutex_t));
    pthread_mutexattr_t attr;

    pthread_mutexattr_init(&attr);
    pthread_mutexattr_settype(&attr, PTHREAD_MUTEX_RECURSIVE);
    pthread_mutex_init(m, &attr);
    pthread_mutexattr_destroy(&attr);

    if (bInitialOwner)
        pthread_mutex_lock(m);

    return new_handle(HANDLE_MUTEX, m);
}

BOOL WINAPI ReleaseMutex(HANDLE hMutex)
{
    pthread_mutex_t *m = lookup_handle(HANDLE_MUTEX, hMutex);
    if (m == NULL)
        return FALSE;
    pthread_mutex_unlock(m);
    return TRUE;
}

BOOL WINAPI SetEvent(HANDLE hEvent)
{
    DebugBreak();
}

DWORD WINAPI WaitForSingleObject(HANDLE hHandle, DWORD dwMilliseconds)
{
    DWORD result = (DWORD)-1;
    struct timespec tv;

    clock_gettime(CLOCK_REALTIME, &tv);
    tv.tv_sec += dwMilliseconds / 1000;
    tv.tv_nsec += (dwMilliseconds % 1000) * 1000000;
    while (tv.tv_nsec >= 1000000000)
    {
        tv.tv_sec++;
        tv.tv_nsec -= 1000000000;
    }

    switch ((DWORD)hHandle >> 16)
    {
    case HANDLE_THREAD:
        {
            pthread_t thr = lookup_handle(HANDLE_THREAD, hHandle);
            result = 0;
            if (dwMilliseconds == INFINITE)
                pthread_join(thr, NULL);
#if defined(__APPLE__) || defined(__EMSCRIPTEN__)
            else
                pthread_join(thr, NULL);
#else
            else if (pthread_timedjoin_np(thr, NULL, &tv) == ETIMEDOUT)
                result = 0x102;
#endif
        }
        break;
    case HANDLE_MUTEX:
        {
            pthread_mutex_t *m = lookup_handle(HANDLE_MUTEX, hHandle);
            result = 0;
            if (dwMilliseconds == INFINITE)
                pthread_mutex_lock(m);
#if defined(__APPLE__) || defined(__EMSCRIPTEN__)
            else
                pthread_mutex_lock(m);
#else
            else if (pthread_mutex_timedlock(m, &tv) == ETIMEDOUT)
                result = 0x102;
#endif
        }
        break;
    default:
        DebugBreak();
        break;
    }

    return result;
}
