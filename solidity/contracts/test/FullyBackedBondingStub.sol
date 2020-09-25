pragma solidity 0.5.17;

import "../fully-backed/FullyBackedBonding.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Fully Backed Bonding Stub
/// @dev This contract is for testing purposes only.
contract FullyBackedBondingStub is FullyBackedBonding {
    // using SafeMath for uint256;
    // // address public delegatedAuthority;
    // bool slashingShouldFail;
    constructor(KeepRegistry _keepRegistry, uint256 _initializationPeriod)
        public
        FullyBackedBonding(_keepRegistry, _initializationPeriod)
    {}

    function setBeneficiary(address _operator, address payable _beneficiary)
        public
    {
        operators[_operator].beneficiary = _beneficiary;
    }
}
