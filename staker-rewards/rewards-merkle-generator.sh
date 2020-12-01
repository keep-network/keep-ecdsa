#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

WORKDIR=$PWD

printf "${LOG_START}Initializing merkle-distributor submodule...${LOG_END}"

git submodule update --init --recursive --remote --rebase --force

printf "${LOG_START}Installing dependencies...${LOG_END}"

cd "$WORKDIR/merkle-distributor"
npm i

cd "$WORKDIR"
npm i

printf "${LOG_START}Generating merkle output object...${LOG_END}"

REWARDS_INPUT_PATH="example-rewards-input.json"
if [[ $1 == *"--input"* ]]; then
    v="${1/--/}"
    declare REWARDS_INPUT_PATH="$2"
fi

npm run generate-merkle-root -- --input="$WORKDIR/$REWARDS_INPUT_PATH"