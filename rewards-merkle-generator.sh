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

printf "${LOG_START}Generating merkle root...${LOG_END}"

yarn ts-node scripts/generate-merkle-root.ts --input "$WORKDIR/staker-rewards/staker-rewards.json"