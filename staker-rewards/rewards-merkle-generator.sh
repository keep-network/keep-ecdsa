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
   echo "Usage: $0"\
        "--verify"\
        "--eth-host <eth_host>"\
        "--start-timestamp <start_timestamp>"\
        "--end-timestamp <end_timestamp>"\
        "--start-block <start_block>"\
        "--end-block <end_block>"\
        "--allocation <allocation>"\
        "--tenderly-url <tenderly_url>"\
        "--tenderly-token <tenderly_token>"
   echo -e "\t--verify Optional verify flag just to display the results without generation of a merkle tree"
   echo -e "\t--eth-host Websocket endpoint of the Ethereum node"
   echo -e "\t--start-timestamp Start of the interval passed as UNIX timestamp"
   echo -e "\t--end-timestamp End of the interval passed as UNIX timestamp"
   echo -e "\t--start-block Start block of the interval"
   echo -e "\t--end-block End block of the interval"
   echo -e "\t--allocation Total KEEP rewards distributed within the given interval passed as 18-decimals number"
   echo -e "\t--tenderly-url Optional Tenderly API project URL"
   echo -e "\t--tenderly-token Optional access token for Tenderly API used to fetch transactions from the chain"
   exit 1 # Exit script after printing help
}

if [ "$1" == "-help" ]; then
  help
fi

printf "${LOG_START}Processing input parameters...${LOG_END}"

# Transform long options to short ones
for arg in "$@"; do
  shift
  case "$arg" in
    "--verify")          set -- "$@" "-v" ;;
    "--eth-host")        set -- "$@" "-h" ;;
    "--start-timestamp") set -- "$@" "-s" ;;
    "--end-timestamp")   set -- "$@" "-e" ;;
    "--start-block")     set -- "$@" "-x" ;;
    "--end-block")       set -- "$@" "-y" ;;
    "--allocation")      set -- "$@" "-a" ;;
    "--tenderly-url")    set -- "$@" "-u" ;;
    "--tenderly-token")  set -- "$@" "-t" ;;
    *)                   set -- "$@" "$arg"
  esac
done

# Parse short options
OPTIND=1
while getopts "vh:s:e:x:y:a:u:t:" opt
do
   case "$opt" in
      v ) verify=true ;;
      h ) eth_host="$OPTARG" ;;
      s ) start="$OPTARG" ;;
      e ) end="$OPTARG" ;;
      x ) start_block="$OPTARG" ;;
      y ) end_block="$OPTARG" ;;
      a ) rewards="$OPTARG" ;;
      u ) tenderly_url="$OPTARG" ;;
      t ) tenderly_token="$OPTARG" ;;
      ? ) help ;; # Print help in case parameter is non-existent
   esac
done
shift $(expr $OPTIND - 1) # remove options from positional parameters

#Print help in case required parameters are empty
if [ -z "$eth_host" ] || [ -z "$start" ] || [ -z "$end" ] || [ -z "$start_block" ] || [ -z "$end_block" ] || [ -z "$rewards" ]
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
node --experimental-json-modules rewards.js "$start" "$end" "$start_block" "$end_block" "$rewards"

printf "${LOG_START}Generating merkle output object...${LOG_END}"

cd "$WORKDIR/distributor"
npm i

npm run generate-merkle-root -- --input="$WORKDIR/$STAKER_REWARD"

printf "${LOG_START}Script finished successfully${LOG_END}"