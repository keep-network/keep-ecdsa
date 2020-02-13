pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./api/IBondedECDSAKeepVendor.sol";

/// @title Bonded ECDSA Keep Vendor.
/// @notice The contract can be used to obtain a new Bonded ECDSA keep.
/// @dev Interacts with ECDSA keep factory to obtain a new instance of the ECDSA
/// keep. The latest version of ECDSA keep factory can be registered for a vendor.
contract BondedECDSAKeepVendorImplV1 is IBondedECDSAKeepVendor, Ownable {

    // Address of ECDSA keep factory.
    address payable keepFactory;

    /// @notice Register a new ECDSA keep factory.
    /// @dev Registers a new ECDSA keep factory. Address cannot be zero
    /// and replaces the old one if it was registered.
    /// @param _factory ECDSA keep factory address.
    function registerFactory(address payable _factory) external onlyOwner {
        require(_factory != address(0), "Incorrect factory address");

        keepFactory = _factory;
    }

    /// @notice Select the latest ECDSA keep factory.
    /// @dev This is a stub implementation returning the latest factory.
    /// @return ECDSA keep factory address.
    function selectFactory() public view returns (address payable) {
        require(keepFactory != address(0), "Keep factory is not registered");

        return keepFactory;
    }
}