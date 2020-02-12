pragma solidity ^0.5.4;

import "../../contracts/KeepBonding.sol";

/// @title Keep Bonding Stub
/// @dev This contract is for testing purposes only.
contract KeepBondingStub is KeepBonding {
    constructor(
        address registryAddress,
        address stakingContractAddress
    ) KeepBonding(registryAddress, stakingContractAddress) public {}

    /// @notice Get registered locked bonds.
    /// @dev This is a stub implementation to validate bonds mapping.
    /// @return Value assigned in the locked bond mapping.
    function getLockedBonds(address holder, address operator, uint256 referenceID) public view returns (uint256) {
        bytes32 bondID = keccak256(abi.encodePacked(operator, holder, referenceID));
        return lockedBonds[bondID];
    }
}
