pragma solidity ^0.5.4;

import "../../contracts/ECDSAKeepRewards.sol";

contract ECDSAKeepRewardsStub is ECDSAKeepRewards {
    constructor (
        uint256 _termLength,
        address _keepToken,
        uint256 _minimumKeepsPerInterval,
        address factoryAddress,
        uint256 _initiated,
        uint256[] memory _intervalWeights
    )
        public
        ECDSAKeepRewards(
            _termLength,
            _keepToken,
            _minimumKeepsPerInterval,
            factoryAddress,
            _initiated,
            _intervalWeights
        )
    {}

    function getTotalRewards() public view returns (uint256) {
        return totalRewards;
    }

    function getUnallocatedRewards() public view returns (uint256) {
        return unallocatedRewards;
    }

    function getPaidOutRewards() public view returns (uint256) {
        return paidOutRewards;
    }

    function currentTime() public view returns (uint256) {
        return block.timestamp;
    }

    function findEndpoint(uint256 intervalEndpoint) public view returns (uint256) {
        return _findEndpoint(intervalEndpoint);
    }
    function getEndpoint(uint256 interval) public returns (uint256) {
        return _getEndpoint(interval);
    }
    function keepsInInterval(uint256 interval) public returns (uint256) {
        return _keepsInInterval(interval);
    }
    function keepCountAdjustment(uint256 interval) public returns (uint256) {
        return _keepCountAdjustment(interval);
    }
    function getIntervalWeight(uint256 interval) public view returns (uint256) {
        return _getIntervalWeight(interval);
    }
    function getIntervalCount() public view returns (uint256) {
        return _getIntervalCount();
    }
    function baseAllocation(uint256 interval) public view returns (uint256) {
        return _baseAllocation(interval);
    }
    function adjustedAllocation(uint256 interval) public returns (uint256) {
        return _adjustedAllocation(interval);
    }
    function rewardPerKeep(uint256 interval) public returns (uint256) {
        return _rewardPerKeep(interval);
    }
    function allocateRewards(uint256 interval) public {
        _allocateRewards(interval);
    }
    function getAllocatedRewards(uint256 interval) public view returns (uint256) {
        return _getAllocatedRewards(interval);
    }
    function isAllocated(uint256 interval) public view returns (bool) {
        return _isAllocated(interval);
    }
}
