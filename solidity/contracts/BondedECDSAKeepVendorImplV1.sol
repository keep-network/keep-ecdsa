pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "@keep-network/keep-core/contracts/Registry.sol";
import "./api/IBondedECDSAKeepVendor.sol";

/// @title Bonded ECDSA Keep Vendor
/// @notice The contract is used to obtain a new Bonded ECDSA keep.
/// @dev Interacts with ECDSA keep factory to obtain a new instance of a bonded
/// ECDSA keep. The latest version of ECDSA keep factory can be registered for
/// a vendor.
contract BondedECDSAKeepVendorImplV1 is IBondedECDSAKeepVendor, Ownable {

    // Mapping to store new implementation versions that inherit from this contract.
    mapping (string => bool) internal _initialized;

    // Registry contract with a list of approved factories (operator contracts) and upgraders.
    Registry internal registry;

    // Address of ECDSA keep factory.
    address payable keepFactory;

    /// @dev Throws if called by any account other than the operator contract upgrader authorized for this service contract.
    modifier onlyOperatorContractUpgrader() {
        address operatorContractUpgrader = registry.operatorContractUpgraderFor(address(this));
        require(operatorContractUpgrader == msg.sender, "Caller is not operator contract upgrader");
        _;
    }

    /// @dev Initialize Keep Vendor contract implementation.
    /// @param registryAddress Keep registry contract linked to this contract.
    function initialize(
        address registryAddress
    )
        public
    {
        require(!initialized(), "Contract is already initialized.");
        _initialized["BondedECDSAKeepVendorImplV1"] = true;
        registry = Registry(registryAddress);
    }

    /// @dev Checks if this contract is initialized.
    function initialized() public view returns (bool) {
        return _initialized["BondedECDSAKeepVendorImplV1"];
    }

    /// @notice Register a new ECDSA keep factory.
    /// @dev Registers a new ECDSA keep factory. Address cannot be zero
    /// and replaces the old one if it was registered.
    /// @param _factory ECDSA keep factory address.
    function registerFactory(address payable _factory) external onlyOperatorContractUpgrader {
        require(_factory != address(0), "Incorrect factory address");
        require(
            registry.isApprovedOperatorContract(_factory),
            "Factory contract is not approved"
        );
        keepFactory = _factory;
    }

    /// @notice Select the latest ECDSA keep factory.
    /// @return ECDSA keep factory address.
    function selectFactory() public view returns (address payable) {
        require(keepFactory != address(0), "Keep factory is not registered");
        return keepFactory;
    }
}
