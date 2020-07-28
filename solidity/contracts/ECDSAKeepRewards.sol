pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "openzeppelin-solidity/contracts/token/ERC20/SafeERC20.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "keep-network/keep-core/contracts/Rewards.sol";

/// @title ECDSAKeepRewards
/// @dev A contract for distributing KEEP token rewards to ECDSA keeps.
/// When a reward contract is created,
/// the creator defines a reward schedule
/// consisting of one or more reward intervals and their interval weights,
/// the length of reward intervals,
/// and the quota of how many keeps must be created in an interval
/// for the full reward for that interval to be paid out.
///
/// The amount of KEEP to be distributed is determined by funding the contract,
/// and additional KEEP can be added at any time.
/// The reward contract is funded with `approveAndCall` with no extra data,
/// but it also collects any KEEP mistakenly sent to it in any other way.
///
/// An interval is defined by the timestamps [startOf, endOf);
/// a keep created at the time `startOf(i)` belongs to interval `i`
/// and one created at `endOf(i)` belongs to `i+1`.
///
/// When an interval is over,
/// it will be allocated a percentage of the remaining unallocated rewards
/// based on its weight,
/// and adjusted by the number of keeps created in the interval
/// if the quota is not met.
/// The adjustment for not meeting the keep quota is a percentage
/// that equals the percentage of the quota that was met;
/// if the number of keeps created is 80% of the quota
/// then 80% of the base reward will be allocated for the interval.
///
/// Any unallocated rewards will stay in the unallocated rewards pool,
/// to be allocated for future intervals.
/// Intervals past the initially defined schedule have a weight of 100%,
/// meaning that all remaining unallocated rewards
/// will be allocated to the interval.
///
/// ECDSA keeps created by the defined `factory` can receive rewards
/// once the interval they were created in is over,
/// and the keep has closed happily.
/// There is no time limit to receiving rewards,
/// nor is there need to wait for all keeps from the interval to close.
/// Calling `receiveReward` automatically allocates the rewards
/// for the interval the specified keep was created in
/// and all previous intervals.
///
/// If a keep is terminated,
/// that fact can be reported to the reward contract.
/// Reporting a terminated keep returns its allocated reward
/// to the pool of unallocated rewards.
contract ECDSAKeepRewards is Rewards {
    uint256 constant MAX_UINT20 = 1 << 20 - 1;
    IBondedECDSAKeepFactory factory;

    constructor (
        uint256 _termLength,
        address _token,
        uint256 _minimumKeepsPerInterval,
        address factoryAddress,
        uint256 _firstIntervalStart,
        uint256[] memory _intervalWeights
    ) public Rewards(
        _termLength,
        _token,
        _minimumKeepsPerInterval,
        _firstIntervalStart,
        _intervalWeights
    ) {
       factory = IBondedECDSAKeepFactory(factoryAddress);
    }

   function _getKeepCount() internal view returns (uint256) {
       return factory.getKeepCount();
   }

   function _getKeepAtIndex(uint256 index) internal view returns (bytes32) {
       return bytes32(factory.getKeepAtIndex(index));
   }

   function _getCreationTime(bytes32 _keep) isAddress(_keep) internal view returns (uint256) {
       return factory.getCreationTime(address(bytes20(_keep)));
   }

   function _isClosed(bytes32 _keep) isAddress(_keep) internal view returns (bool) {
       return IBondedECDSAKeep(address(bytes20(_keep))).isClosed();
   }

   function _isTerminated(bytes32 _keep) isAddress(_keep) internal view returns (bool) {
       return IBondedECDSAKeep(address(bytes20(_keep))).isTerminated();
   }

   function _recognizedByFactory(bytes32 _keep) isAddress(_keep) internal view returns (bool) {
       return factory.getCreationTime(address(bytes20(_keep))) != 0;
   }

   function _distributeReward(bytes32 _keep, uint256 amount) isAddress(_keep) internal {
       token.approve(address(_keep), amount);
       IBondedECDSAKeep(address(_keep)).distributeERC20Reward(
           address(token),
           amount
       );
   }

   modifier isAddress(bytes32 _keep) {
       require(
           _keep == (_keep & bytes32(MAX_UINT20)),
           "Invalid keep address"
       );
       _;
   }
}

interface IBondedECDSAKeep {
    function getOwner() external view returns (address);
    function getTimestamp() external view returns (uint256);
    function isClosed() external view returns (bool);
    function isTerminated() external view returns (bool);
    function isActive() external view returns (bool);
    function distributeERC20Reward(address _erc20, uint256 amount) external;
}

interface IBondedECDSAKeepFactory {
    function getKeepCount() external view returns (uint256);
    function getKeepAtIndex(uint256 index) external view returns (address);
    function getCreationTime(address _keep) external view returns (uint256);
}
