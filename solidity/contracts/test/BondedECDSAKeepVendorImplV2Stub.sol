pragma solidity 0.5.17;

/// @title Bonded ECDSA Keep Vendor Implementation V2 Stub
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepVendorImplV2Stub {
    mapping(string => bool) internal _initialized;

    function version() public view returns (string memory) {
        return "V2";
    }

    function initialize(bool shouldFail) public {
        require(!initialized(), "Contract is already initialized.");
        _initialized["BondedECDSAKeepVendorImplV2"] = true;

        require(!shouldFail, "Initialization failed");
    }

    function initialized() public view returns (bool) {
        return _initialized["BondedECDSAKeepVendorImplV2"];
    }
}
