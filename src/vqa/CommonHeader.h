#pragma once
#include <SDL2/SDL.h>

#include <array>

#include <cassert>
#include <climits>
#include <cstdint>
#include <list>
#include <memory>
#include <map>
#include <set>
#include <string>
#include <vector>

typedef unsigned char byte;
typedef unsigned long dword;

using std::array;
using std::list;
using std::map;
using std::set;
using std::string;
using std::to_string;
using std::vector;
using std::ifstream;
using std::ofstream;
using std::fstream;
using std::ios;

// Expose a few functions in the platform.c for C++.
extern "C" unsigned int nox_platform_get_ticks();
extern "C" void nox_platform_sleep(unsigned int ms);
