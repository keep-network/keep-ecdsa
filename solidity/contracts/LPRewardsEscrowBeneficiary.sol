pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/PhasedEscrow.sol";

/// @title LPRewardsEscrowBeneficiary
/// @notice Transfer the received tokens from PhasedEscrow to a designated
///         LPRewards contract.
contract LPRewardsEscrowBeneficiary is StakerRewardsBeneficiary {
    constructor(IERC20 _token, IStakerRewards _lpRewardsContract)
        public
        StakerRewardsBeneficiary(_token, _lpRewardsContract)
    {}
}
