pragma solidity 0.5.17;

import "../../contracts/ECDSARewards.sol";

/// @title ECDSA Rewards Stub for ecdsa rewards testing
/// @dev This contract is for testing purposes only.
contract ECDSARewardsStub is ECDSARewards {

    uint256 internal tokenDecimalMultiplier = 10**18;

    constructor(
        address _token,
        address payable _factoryAddress,
        address _tokenStakingAddress
    )
        public
        ECDSARewards(
            _token,
            _factoryAddress,
            _tokenStakingAddress
        )
    {}

    function setBeneficiaryRewardCap(uint256 _beneficiaryRewardCap) public {
        beneficiaryRewardCap = _beneficiaryRewardCap.mul(tokenDecimalMultiplier);
    }
}
