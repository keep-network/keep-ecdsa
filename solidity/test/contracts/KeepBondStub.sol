pragma solidity ^0.5.4;

import "../../contracts/KeepBond.sol";

/// @title Keep Bond Stub
/// @dev This contract is for testing purposes only.
contract KeepBondStub is KeepBond {

    /// @notice Get registered pot for account.
    /// @dev This is a stub implementation to validate the pot mapping.
    /// @return Value assigned for the account in the pot mapping.
    function getPot(address account) public view returns (uint) {
        return pot[account];
    }
}
