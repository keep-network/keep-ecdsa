#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

WORKDIR=$PWD

printf "${LOG_START}Initializing merkle-distributor submodule...${LOG_END}"

git submodule update --init --recursive --remote --rebase --force

printf "${LOG_START}Installing dependencies...${LOG_END}"

cd "$WORKDIR/include/merkle-distributor"
yarn

cd "$WORKDIR/staker-rewards"
yarn

printf "${LOG_START}Generating merkle output object...${LOG_END}"

REWARDS_INPUT_PATH="staker-rewards/example-rewards-input.json"
if [[ $1 == *"--input"* ]]; then
    v="${1/--/}"
    declare REWARDS_INPUT_PATH="$2"
fi

yarn ts-node generate-merkle-root.ts --input "$WORKDIR/$REWARDS_INPUT_PATH"