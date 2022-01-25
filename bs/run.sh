#!/usr/bin/env bash

source "$(pwd)/bs/build.sh"

$GO_BIN run -race . "$DIR_OBSIDIAN"
