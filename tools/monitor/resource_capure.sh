#!/bin/bash
vmstat 1 | awk '{print $3" "$4" "$13" "$14}'