pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";

/// @title Keep Registry
/// @notice Contract handling keeps registry.
/// @dev The keep registry serves the role of the master list and tracks sanctioned
/// vendors. It ensures that only approved contracts are used. A new type of keep
/// can be added without upgradeable registry.
/// TODO: This is a stub contract - needs to be implemented.
contract KeepRegistry is Ownable {
    // Registered keep vendors. Mapping of a keep type to a keep vendor address.
    mapping (string => address) internal keepVendors;

    /// @notice Set a keep vendor contract address for a keep type.
    /// @dev Only contract owner can call this function.
    /// @param _keepType Keep type.
    /// @param _vendorAddress Keep Vendor contract address.
    function setVendor(string memory _keepType, address _vendorAddress) public onlyOwner {
        keepVendors[_keepType] = _vendorAddress;
    }

    /// @notice Get a keep vendor contract address for a keep type.
    /// @param _keepType Keep type.
    /// @return Keep vendor contract address.
    function getVendor(string memory _keepType) public view returns (address) {
        return keepVendors[_keepType];
    }
}
