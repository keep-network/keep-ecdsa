pragma solidity ^0.5.4;

/// @title ECDSA Keep Vendor
/// @notice Contract reflecting an ECDSA keep vendor.
contract IECDSAKeepVendor {
    /// @notice Open a new ECDSA keep.
    /// @dev Calls a recommended ECDSA keep factory to open a new keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Opened keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress);
}
