pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/PhasedEscrow.sol";

/// @title ECDSABackportRewardsEscrowBeneficiary
/// @notice Trasfer the received tokens to a designated
///         ECDSABackportRewardsEscrowBeneficiary contract.
contract ECDSABackportRewardsEscrowBeneficiary is StakerRewardsBeneficiary {
    constructor(IERC20 _token, IStakerRewards _stakerRewards)
        public
        StakerRewardsBeneficiary(_token, _stakerRewards)
    {}
}

/// @title ECDSARewardsEscrowBeneficiary
/// @notice Transfer the received tokens to a designated
///         ECDSARewardsEscrowBeneficiary contract.
contract ECDSARewardsEscrowBeneficiary is StakerRewardsBeneficiary {
    constructor(IERC20 _token, IStakerRewards _stakerRewards)
        public
        StakerRewardsBeneficiary(_token, _stakerRewards)
    {}
}
