pragma solidity ^0.5.4;

/// @title Bonded ECDSA Keep Factory
/// @notice Factory for Bonded ECDSA Keeps.
contract IECDSAKeepFactory { // TODO: Rename to IBondedECDSAKeepFactory
    /// @notice Open a new ECDSA Keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @param _bond value of ETH bond required from the keep.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond
    ) external payable returns (address keepAddress);
}
