pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeepVendor.sol";

/// @title Bonded ECDSA Keep Vendor Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepVendorStub is BondedECDSAKeepVendor {
    constructor(
        string memory _version,
        address _implementation,
        bytes memory _data
    ) public BondedECDSAKeepVendor(_version, _implementation, _data) {}

    function upgradeVersion() public view returns (string memory) {
        Implementation memory implementation = getImplementation(
            upgradeImplementationID()
        );

        return implementation.version;
    }

    function upgradeImplementation() public view returns (address) {
        Implementation memory implementation = getImplementation(
            upgradeImplementationID()
        );

        return implementation.implementationContract;
    }

    function upgradeInitialization() public view returns (bytes memory) {
        Implementation memory implementation = getImplementation(
            upgradeImplementationID()
        );

        return implementation.initializationData;
    }
}
