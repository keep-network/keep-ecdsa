pragma solidity 0.5.17;

import "../../contracts/ECDSAKeepRewards.sol";

contract ECDSAKeepRewardsStub is ECDSAKeepRewards {
    constructor (
        uint256 _termLength,
        address _keepToken,
        uint256 _minimumKeepsPerInterval,
        address factoryAddress,
        uint256 _initiated,
        uint256[] memory _intervalWeights
    ) public ECDSAKeepRewards(
            _termLength,
            _keepToken,
            _minimumKeepsPerInterval,
            factoryAddress,
            _initiated,
            _intervalWeights
    ) {}

    // function eligibleForRewardA(address keep) public view returns (bool) {
    //     return eligibleForReward(fromAddress(keep));
    // }

    // function eligibleButTerminatedA(address keep) public view returns (bool) {
    //     return eligibleButTerminated(fromAddress(keep));
    // }

    // function receiveRewardA(address keep) public {
    //     return receiveReward(fromAddress(keep));
    // }

    // function reportTerminationA(address keep) public {
    //     return reportTermination(fromAddress(keep));
    // }

    function getUnallocatedRewards() public view returns (uint256) {
        return unallocatedRewards;
    }

    function _toAddress(bytes32 b) public pure returns (address) {
        return toAddress(b);
    }

    function _fromAddress(address a) public pure returns (bytes32) {
        return fromAddress(a);
    }

    function testRoundtrip(bytes32 b) public pure returns (bool) {
        return validAddressBytes(b);
    }
}
