#!/usr/bin/env bash

make all

./bin/notes &

trap 'pkill -P $$' EXIT

NOTES_HOST="127.0.0.1" ./bin/users
