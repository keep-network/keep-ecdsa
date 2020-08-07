pragma solidity 0.5.17;

import "../../contracts/AbstractBonding.sol";

contract AbstractBondingStub is AbstractBonding {
    constructor(
        address registryAddress,
        address authorizationsAddress,
        address stakeDelegatableAddress
    )
        public
        AbstractBonding(
            registryAddress,
            authorizationsAddress,
            stakeDelegatableAddress
        )
    {}

    function withdraw(uint256 amount, address operator) public {
        revert("abstract function");
    }

    function withdrawBondExposed(uint256 amount, address operator) public {
        withdrawBond(amount, operator);
    }
}
