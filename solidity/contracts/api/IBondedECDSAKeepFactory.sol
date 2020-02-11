pragma solidity ^0.5.4;

/// @title Bonded ECDSA Keep Factory
/// @notice Factory for Bonded ECDSA Keeps.
contract IBondedECDSAKeepFactory {
    /// @notice Open a new ECDSA Keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @param _bond Value of ETH bond required from the keep.
    /// @return Address of the opened keep.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond
    ) external payable returns (address keepAddress);

    /// @notice Gets a fee estimate for opening a new keep.
    /// @return Uint256 estimate.
    function openKeepFeeEstimate() public view returns (uint256);
}
