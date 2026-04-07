cmake_minimum_required (VERSION 3.11)
project (nox-port)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS ON)

option (USE_ASAN "Enable address sanitizer" OFF)

add_subdirectory (src)
