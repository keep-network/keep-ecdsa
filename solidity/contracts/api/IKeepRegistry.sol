pragma solidity ^0.5.4;

/// @title Keep Registry
/// @notice Contract reflecting a keeps registry.
contract IKeepRegistry {
    /// @notice Get a keep vendor contract address for a keep type.
    /// @param _keepType Keep type.
    /// @return Keep vendor contract address.
    function getVendor(string calldata _keepType) external view returns (address);
}
