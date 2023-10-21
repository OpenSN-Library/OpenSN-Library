cmake_minimum_required(VERSION 3.16)
project(netlink_test_userspace C)

set(CMAKE_C_STANDARD 11)

add_executable(netlink_test_userspace
        netlink.c)
