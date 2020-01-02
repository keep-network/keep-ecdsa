pragma solidity ^0.5.4;

import "../../contracts/KeepBonding.sol";

/// @title Keep Bond Stub
/// @dev This contract is for testing purposes only.
contract KeepBondStub is KeepBonding {

    /// @notice Get registered locked bonds.
    /// @dev This is a stub implementation to validate bonds mapping.
    /// @return Value assigned in the locked bond mapping.
    function getLockedBonds(address holder, address operator, uint256 ref) public view returns (uint256) {
        bytes memory bondID = abi.encodePacked(operator, holder, ref);
        return lockedBonds[bondID];
    }

    /// @notice Get registered pot for account.
    /// @dev This is a stub implementation to validate bonds mapping.
    /// @return Value assigned in the bond assignments mapping.
    function getBondAssignments(address holder, address operator) public view returns (uint256[] memory) {
        bytes memory assignment = abi.encodePacked(operator, holder);
        return bondAssignments[assignment];
    }
}
