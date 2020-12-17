#!/bin/bash

set -e

LOG_START='\n\e[1;36m' # new line + bold + color
LOG_END='\n\e[0m' # new line + reset color

WORKDIR=$PWD

# default file for calculated staker reward allocation
STAKER_REWARD="distributor/staker-reward-allocation.json"

NODE_CURRENT_VER="$(node --version)"
NODE_REQUIRED_VER="v14.3.0"

if [ "$(printf '%s\n' "$NODE_REQUIRED_VER" "$NODE_CURRENT_VER" | sort -V | head -n1)" != "$NODE_REQUIRED_VER" ]; 
then
      echo "Required node version must be at least ${NODE_REQUIRED_VER}" 
      exit 1
fi

help()
{
   echo ""
   echo "Usage: $0 -v -w <eth_host> -s <start_interval> -e <end_interval> -r <rewards> -u <tenderly_url> -t <tenderly_token>"
   echo -e "\t-v Optional verify flag just to display the results without generation of a merkle tree"
   echo -e "\t-w Websocket endpoint of the Ethereum node"
   echo -e "\t-s Start of the interval passed as UNIX timestamp"
   echo -e "\t-e End of the interval passed as UNIX timestamp"
   echo -e "\t-r Total KEEP rewards distributed within the given interval passed as 18-decimals number"
   echo -e "\t-u Optional Tenderly API project URL"
   echo -e "\t-t Optional access token for Tenderly API used to fetch transactions from the chain"
   exit 1 # Exit script after printing help
}

if [ "$1" == "-help" ]; then
  help
fi

printf "${LOG_START}Processing input parameters...${LOG_END}"

while getopts "vw:s:e:r:u:t:" opt
do
   case "$opt" in
      v ) verify=true ;;
      w ) eth_host="$OPTARG" ;;
      s ) start="$OPTARG" ;;
      e ) end="$OPTARG" ;;
      r ) rewards="$OPTARG" ;;
      u ) tenderly_url="$OPTARG" ;;
      t ) tenderly_token="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done

#Print help in case required parameters are empty
if [ -z "$eth_host" ] || [ -z "$start" ] || [ -z "$end" ] || [ -z "$rewards" ]
then
   echo "Some or all of the required parameters are empty";
   help
fi

printf "${LOG_START}Initializing merkle-distributor submodule...${LOG_END}"

git submodule update --init --recursive --remote --rebase --force

printf "${LOG_START}Installing dependencies for merkle-distributor...${LOG_END}"

cd "$WORKDIR/merkle-distributor"
npm i

printf "${LOG_START}Installing dependencies for staker-rewards...${LOG_END}"

cd "$WORKDIR"
npm i

printf "${LOG_START}Calculating staker rewards...${LOG_END}"

export OUTPUT_MODE="text"
if [ "$verify" == true ]; then
   OUTPUT_MODE=""
fi

ETH_HOSTNAME="$eth_host" \
TENDERLY_PROJECT_URL="$tenderly_url" \
TENDERLY_ACCESS_TOKEN="$tenderly_token" \
REWARDS_PATH="$WORKDIR/$STAKER_REWARD" \
node --experimental-json-modules rewards.js "$start" "$end" "$rewards"

printf "${LOG_START}Generating merkle output object...${LOG_END}"

cd "$WORKDIR/distributor"
npm i

npm run generate-merkle-root -- --input="$WORKDIR/$STAKER_REWARD"

printf "${LOG_START}Script finished successfully${LOG_END}"