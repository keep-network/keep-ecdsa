pragma solidity 0.5.17;

import "@keep-network/sortition-pools/contracts/api/IStaking.sol";
import "@keep-network/sortition-pools/contracts/SortitionPool.sol";

contract Stake is IStaking {
    function eligibleStake(
        address operator,
        address operatorContract
    ) external view returns (uint256) {
      return 1;
    }
}

contract TestSortitionPool is SortitionPool(new Stake(), 1, 1, msg.sender) {}
