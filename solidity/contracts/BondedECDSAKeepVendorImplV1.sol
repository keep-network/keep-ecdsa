pragma solidity ^0.5.4;

import "./api/IBondedECDSAKeepVendor.sol";

import "@keep-network/keep-core/contracts/Registry.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Bonded ECDSA Keep Vendor
/// @notice The contract is used to obtain a new Bonded ECDSA keep factory.
contract BondedECDSAKeepVendorImplV1 is IBondedECDSAKeepVendor {
    using SafeMath for uint256;

    // Mapping to store new implementation versions that inherit from
    // this contract.
    mapping(string => bool) internal _initialized;

    // Registry contract with a list of approved factories (operator contracts)
    // and upgraders.
    Registry internal registry;

    // Upgrade time delay defines a waiting period for factory address registration.
    // When new factory registration is initialized it has to wait this period
    // before the change can take effect on the currently registered facotry
    // address.
    uint256 public factoryRegistrationTimeDelay;

    // Address of ECDSA keep factory.
    address payable keepFactory;

    // Address of a new ECDSA keep factory in update process.
    address payable newKeepFactory;

    // Timestamp of the moment when factory registration process has started.
    uint256 factoryRegistrationInitiatedTimestamp;

    event FactoryRegistrationStarted(address factory, uint256 timestamp);
    event FactoryRegistered(address factory);

    /// @notice Initializes Keep Vendor contract implementation.
    /// @param registryAddress Keep registry contract linked to this contract.
    /// @param factory Keep factory contract registered initially in the vendor
    /// contract.
    function initialize(address registryAddress, address payable factory)
        public
    {
        require(registryAddress != address(0), "Incorrect registry address");
        require(factory != address(0), "Incorrect factory address");

        require(!initialized(), "Contract is already initialized.");
        _initialized["BondedECDSAKeepVendorImplV1"] = true;

        registry = Registry(registryAddress);
        keepFactory = factory;

        factoryRegistrationTimeDelay = 1 days; // TODO: Determine right value for this property.
    }

    /// @notice Checks if this contract is initialized.
    function initialized() public view returns (bool) {
        return _initialized["BondedECDSAKeepVendorImplV1"];
    }

    /// @notice Starts new ECDSA keep factory registration.
    /// @dev Registers a new ECDSA keep factory. Address cannot be zero
    /// and cannot be the same that is currently registered. It is a first part
    /// of the two-step factory registration process. The function emits an event
    /// containing the new factory address and current block timestamp.
    /// @param _factory ECDSA keep factory address.
    function registerFactory(address payable _factory)
        external
        onlyOperatorContractUpgrader
    {
        require(_factory != address(0), "Incorrect factory address");

        if (upgradeInitiated()) {
            require(
                newKeepFactory != _factory,
                "Factory upgrade already initiated"
            );
        } else {
            require(keepFactory != _factory, "Factory already registered");
        }

        require(
            registry.isApprovedOperatorContract(_factory),
            "Factory contract is not approved"
        );

        newKeepFactory = _factory;

        /* solium-disable-next-line security/no-block-members */
        uint256 timestamp = block.timestamp;

        factoryRegistrationInitiatedTimestamp = timestamp;

        emit FactoryRegistrationStarted(_factory, timestamp);
    }

    /// @notice Finalizes ECDSA keep factory registration.
    /// @dev It is the second part of the two-step factory registration process.
    /// The function can be called after factory registration time delay period
    /// has passed since the new factory registration. It emits an event
    /// containing the new factory address.
    function completeFactoryRegistration() public {
        require(upgradeInitiated(), "Upgrade not initiated");

        require(
            /* solium-disable-next-line security/no-block-members */
            block.timestamp.sub(factoryRegistrationInitiatedTimestamp) >=
                factoryRegistrationTimeDelay,
            "Timer not elapsed"
        );

        keepFactory = newKeepFactory;
        newKeepFactory = address(0);
        factoryRegistrationInitiatedTimestamp = 0;

        emit FactoryRegistered(keepFactory);
    }

    function upgradeInitiated() internal view returns (bool) {
        return factoryRegistrationInitiatedTimestamp > 0;
    }

    /// @notice Selects the latest ECDSA keep factory.
    /// @return ECDSA keep factory address.
    function selectFactory() public view returns (address payable) {
        require(keepFactory != address(0), "Keep factory is not registered");
        return keepFactory;
    }

    /// @dev Throws if called by any account other than the operator contract
    /// upgrader authorized for this service contract.
    modifier onlyOperatorContractUpgrader() {
        address operatorContractUpgrader = registry.operatorContractUpgraderFor(
            address(this)
        );
        require(
            operatorContractUpgrader == msg.sender,
            "Caller is not operator contract upgrader"
        );
        _;
    }
}
