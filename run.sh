#!/usr/bin/env bash

make clean
make

./notes &

trap 'pkill -P $$' EXIT

./users


