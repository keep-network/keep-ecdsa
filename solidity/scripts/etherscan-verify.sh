#!/bin/bash

set -e

echo "Verifying contracts on Etherscan..."

npx truffle run verify \
    BondedECDSAKeep \
    BondedECDSAKeepFactory \
    BondedECDSAKeepVendor \
    BondedECDSAKeepVendorImplV1 \
    BondedSortitionPoolFactory \
    Branch \
    ECDSARewards \
    ECDSARewardsDistributor \
    FullyBackedBonding \
    FullyBackedECDSAKeep \
    FullyBackedECDSAKeepFactory \
    FullyBackedSortitionPoolFactory \
    KeepBonding \
    Leaf \
    LPRewardsKEEPETH \
    LPRewardsKEEPTBTC \
    LPRewardsTBTCETH \
    LPRewardsTBTCSaddle \
    LPRewardsTBTCv2Saddle \
    LPRewardsTBTCv2SaddleV2 \
    Migrations \
    Position \
    StackLib \
    TestToken \
    --network $TRUFFLE_NETWORK
