pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "@keep-network/keep-core/contracts/Registry.sol";
import "./api/IBondedECDSAKeepVendor.sol";
import "./utils/AddressArrayUtils.sol";

/// @title Bonded ECDSA Keep Vendor
/// @notice The contract can be used to obtain a new Bonded ECDSA keep.
/// @dev Interacts with ECDSA keep factory to obtain a new instance of the ECDSA
/// keep. Several versions of ECDSA keep factories can be registered for the vendor.
contract BondedECDSAKeepVendorImplV1 is IBondedECDSAKeepVendor, Ownable {
    using AddressArrayUtils for address payable[];

    // Mapping to store new implementation versions that inherit from this contract.
    mapping (string => bool) internal _initialized;

    // Registry contract with a list of approved factories (operator contracts) and upgraders.
    Registry internal registry;

    // List of ECDSA keep factories.
    address payable[] public factories;

    /**
     * @dev Initialize Keep Vendor contract implementation.
     * @param registryAddress Keep registry contract linked to this contract.
     */
    function initialize(
        uint256 registryAddress
    )
        public
    {
        require(!initialized(), "Contract is already initialized.");
        _initialized["BondedECDSAKeepVendorImplV1"] = true;
        registry = Registry(registryAddress);
    }

    /**
     * @dev Checks if this contract is initialized.
     */
    function initialized() public view returns (bool) {
        return _initialized["BondedECDSAKeepVendorImplV1"];
    }

    /// @notice Register new ECDSA keep factory.
    /// @dev Adds a factory address to the list of registered factories. Address
    /// cannot be zero and cannot be already registered.
    /// @param _factory ECDSA keep factory address.
    function registerFactory(address payable _factory) external onlyOwner {
        require(!factories.contains(_factory), "Factory address already registered");

        factories.push(_factory);
    }

    /// @notice Select a recommended ECDSA keep factory from all registered
    /// ECDSA keep factories.
    /// @dev This is a stub implementation returning first factory on the list.
    /// @return Selected ECDSA keep factory address.
    function selectFactory() public view returns (address payable) {
        require(factories.length > 0, "No factories registered");

        // TODO: Implement factory selection mechanism.
        return factories[factories.length - 1];
    }
}
