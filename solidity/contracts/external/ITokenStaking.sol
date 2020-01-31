pragma solidity ^0.5.4;

/// @title Token Staking interface.
/// @notice Interface for interaction with TokenStaking contract.
contract ITokenStaking {
    /// @dev Gets the stake balance of the specified address.
    /// @param _address The address to query the balance of.
    /// @return An uint256 representing the amount staked by the passed address.
    function balanceOf(address _address) public view returns (uint256 balance);
}
