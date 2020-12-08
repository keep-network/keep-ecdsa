#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

WORKDIR=$PWD

printf "${LOG_START}Initializing merkle-distributor submodule...${LOG_END}"

git submodule update --init --recursive --remote --rebase --force

printf "${LOG_START}Installing dependencies for merkle-distributor...${LOG_END}"

cd "$WORKDIR/merkle-distributor"
npm i

printf "${LOG_START}Installing dependencies for staker-rewards...${LOG_END}"

cd "$WORKDIR"
npm i

printf "${LOG_START}Processing input parameters...${LOG_END}"

help()
{
   echo ""
   echo "Usage: $0 -h <eth_host> -s <start_interval> -e <end_interval> -r <rewards>"
   echo -e "\t-h Websocket endpoint of the Ethereum node"
   echo -e "\t-s Start of the interval passed as UNIX timestamp"
   echo -e "\t-e End of the interval passed as UNIX timestamp"
   echo -e "\t-r Total KEEP rewards distributed within the given interval passed as 18-decimals number"
   exit 1 # Exit script after printing help
}

while getopts "h:s:e:r:" opt
do
   case "$opt" in
      h ) eth_host="$OPTARG" ;;
      s ) start="$OPTARG" ;;
      e ) end="$OPTARG" ;;
      r ) rewards="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done

#Print help in case parameters are empty
if [ -z "$eth_host" ] || [ -z "$start" ] || [ -z "$end" ] || [ -z "$rewards" ]
then
   echo "Some or all of the parameters are empty";
   help
fi

printf "${LOG_START}Calculating staker rewards...${LOG_END}"

ETH_HOSTNAME="$eth_host" node --experimental-json-modules rewards.js "$start" "$end" "$rewards"

printf "${LOG_START}Generating merkle output object...${LOG_END}"

# default file name
REWARDS_INPUT_PATH="rewards-input.json"

npm run generate-merkle-root -- --input="$WORKDIR/$REWARDS_INPUT_PATH"

printf "${LOG_START}Script finished successfully${LOG_END}"