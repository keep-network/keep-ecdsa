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
    function setKeepTypeVendor(string memory _keepType, address _vendorAddress) public onlyOwner {
        require(_vendorAddress != address(0), "Vendor address cannot be zero");

        keepVendors[_keepType] = _vendorAddress;
    }

    /// @notice Get a keep vendor contract address for a keep type.
    /// @param _keepType Keep type.
    /// @return Keep vendor contract address.
    function getKeepVendor(string memory _keepType) public view returns (address) {
        return keepVendors[_keepType];
    }

    /// @notice Remove a keep type from the registry.
    /// @dev Only contract owner can call this function.
    /// @param _keepType Keep type.
    function removeKeepType(string memory _keepType) public onlyOwner {
        delete keepVendors[_keepType];
    }
}
