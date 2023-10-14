cmake_minimum_required(VERSION 3.26)
project(c C)

set(CMAKE_C_STANDARD 99)

add_executable(c main.c
        common.h
        chunk.h
        chunk.c
        memory.c
        memory.h
        debug.c
        debug.h
        value.c
        value.h)
