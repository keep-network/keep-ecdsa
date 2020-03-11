pragma solidity ^0.5.4;

import "../../contracts/BondedECDSAKeepVendorImplV1.sol";

/// @title Bonded ECDSA Keep Vendor Implementation Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepVendorImplV1Stub is BondedECDSAKeepVendorImplV1 {
    function getKeepFactory() public view returns (address) {
        return keepFactory;
    }

    function getNewKeepFactory() public view returns (address) {
        return newKeepFactory;
    }

    function getFactoryRegistrationInitiatedTimestamp()
        public
        view
        returns (uint256)
    {
        return factoryRegistrationInitiatedTimestamp;
    }
}
