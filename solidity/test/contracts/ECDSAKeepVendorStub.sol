pragma solidity ^0.5.4;

import "../../contracts/ECDSAKeepVendor.sol";

/// @title ECDSA Keep Vendor Stub
/// @dev This contract is for testing purposes only.
contract ECDSAKeepVendorStub is ECDSAKeepVendor {

    /// @notice Get registered ECDSA keep factories.
    /// @dev This is a stub implementation to validate the factories list.
    /// @return List of registered ECDSA keep factory addresses.
    function getFactories() public view returns (address[] memory) {
        return factories;
    }

    /// @notice Select a recommended ECDSA keep factory from all registered
    /// ECDSA keep factories.
    /// @dev This is a stub implementation to expose the function for testing.
    /// @return Selected ECDSA keep factory address.
    function selectFactoryPublic() public view returns (address) {
        return selectFactory();
    }
}
