pragma solidity 0.5.17;

import "../../contracts/BondedECDSAKeepFactory.sol";

/// @title Bonded ECDSA Keep Factory Stub for vendor testing
/// @dev This contract is for testing purposes only.
contract BondedECDSAKeepFactoryVendorStub is BondedECDSAKeepFactory {
    constructor(
        address masterBondedECDSAKeepAddress,
        address sortitionPoolFactory,
        address tokenStaking,
        address keepBonding,
        address randomBeacon
    )
        public
        BondedECDSAKeepFactory(
            masterBondedECDSAKeepAddress,
            sortitionPoolFactory,
            tokenStaking,
            keepBonding,
            randomBeacon
        )
    {}

    // @dev Returns calculated keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond
    ) public payable returns (address) {
        _groupSize;
        _honestThreshold;
        _owner;
        _bond;

        return calculateKeepAddress();
    }

    /// @dev Calculates an address for a keep based on the address of the factory.
    /// We need it to have predictable addresses for factories verification.
    function calculateKeepAddress() public view returns (address) {
        uint256 factoryAddressInt = uint256(address(this));
        return address(factoryAddressInt % 1000000000000);
    }
}
