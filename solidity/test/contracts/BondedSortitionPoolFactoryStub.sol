pragma solidity ^0.5.10;

import "./BondedSortitionPoolStub.sol";

contract BondedSortitionPoolFactoryStub {
    function createSortitionPool(
        IStaking stakingContract,
        IBonding bondingContract,
        uint256 minimumStake,
        uint256 initialMinimumBond
    ) public returns (address) {
        return
            address(
                new BondedSortitionPoolStub(
                    stakingContract,
                    bondingContract,
                    minimumStake,
                    initialMinimumBond,
                    msg.sender
                )
            );
    }
}
