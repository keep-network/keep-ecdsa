pragma solidity ^0.5.4;

import "../../contracts/KeepBonding.sol";

contract KeepBondingStub is KeepBonding {

    constructor(
        address registryAddress,
        address stakingContractAddress
    )
        public
        KeepBonding(
            registryAddress,
            stakingContractAddress
        )
    {}

    function authorizeSortitionPoolContractStub(
        address _operator,
        address _poolAddress
    ) public {
        require(
            stakingContract.authorizerOf(_operator) == _operator,
            "Not authorized"
        );
        authorizedPools[_operator][_poolAddress] = true;
    }
}
