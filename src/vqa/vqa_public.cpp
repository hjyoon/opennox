#include "vqa_file.h"

extern "C" SDL_Surface* movieSurface;
extern "C" SDL_Surface* g_backbuffer1;
extern "C" Uint32 g_format;

extern "C" char* resolve_case_insensitive_path(const char* path);
extern "C" int PlayMovieCallback(byte* frame, dword cx, dword cy);

extern "C" int PlayMovie(char* filename)
{
#ifdef NOX_E2E_TEST
    return 0;
#endif
    char* resolved = NULL;
    char* path = filename;
    char* backslash = strchr(filename, '\\');
    if (backslash != NULL)
    {
        backslash[0] = '/';
    }
    resolved = resolve_case_insensitive_path(filename);
    if (resolved != NULL)
    {
        path = resolved;
    }
    Cvqa_file file(path);
    file.register_decode(&PlayMovieCallback);
    file.post_open();
    bool valid = file.is_valid();
    if (valid)
    {
        if (movieSurface != NULL)
        {
            SDL_FreeSurface(movieSurface);
            movieSurface = NULL;
        }
        movieSurface = SDL_CreateRGBSurfaceWithFormat(0, file.get_cx(), file.get_cy(), 16, g_format);
        SDL_FillRect(g_backbuffer1, NULL, SDL_MapRGB(g_backbuffer1->format, 0, 0, 0));
        file.extract_both();
        SDL_FreeSurface(movieSurface);
        movieSurface = NULL;
    }
    free(resolved);
    return 0;
}
