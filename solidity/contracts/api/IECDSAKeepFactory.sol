pragma solidity ^0.5.4;

/// @title ECDSA Keep Factory
/// @notice Contract reflecting a ECDSA Keep Factory.
contract IECDSAKeepFactory { // TODO: Rename to IBondedECDSAKeepFactory
    /// @notice Open a new ECDSA Keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Opened keep address.
    function openKeep( // TODO: Add `bond` parameter.
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress);
}
