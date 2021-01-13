/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓         ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "./api/IBondedECDSAKeepVendor.sol";

import "@keep-network/keep-core/contracts/KeepRegistry.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Bonded ECDSA Keep Vendor
/// @notice The contract is used to obtain a new Bonded ECDSA keep factory.
contract BondedECDSAKeepVendorImplV1 is IBondedECDSAKeepVendor {
    using SafeMath for uint256;

    // Mapping to store new implementation versions that inherit from
    // this contract.
    mapping(string => bool) internal _initialized;

    // KeepRegistry contract with a list of approved factories (operator contracts)
    // and upgraders.
    KeepRegistry internal registry;

    // Upgrade time delay defines a waiting period for factory address upgrade.
    // When new factory upgrade is initialized it has to wait this period
    // before the change can take effect on the currently registered facotry
    // address.
    uint256 public factoryUpgradeTimeDelay;

    // Address of ECDSA keep factory.
    address payable keepFactory;

    // Address of a new ECDSA keep factory in update process.
    address payable newKeepFactory;

    // Timestamp of the moment when factory upgrade process has started.
    uint256 factoryUpgradeInitiatedTimestamp;

    event FactoryUpgradeStarted(address factory, uint256 timestamp);
    event FactoryUpgradeCompleted(address factory);

    constructor() public {
        // Mark as already initialized to block the direct usage of the contract
        // after deployment. The contract is intended to be called via proxy, so
        // setting this value will not affect the clones usage.
        _initialized["BondedECDSAKeepVendorImplV1"] = true;
    }

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

        registry = KeepRegistry(registryAddress);
        keepFactory = factory;

        factoryUpgradeTimeDelay = 1 days;
    }

    /// @notice Checks if this contract is initialized.
    function initialized() public view returns (bool) {
        return _initialized["BondedECDSAKeepVendorImplV1"];
    }

    /// @notice Starts new ECDSA keep factory upgrade.
    /// @dev Registers a new ECDSA keep factory. Address cannot be zero
    /// and cannot be the same that is currently registered. It is a first part
    /// of the two-step factory upgrade process. The function emits an event
    /// containing the new factory address and current block timestamp.
    /// @param _factory ECDSA keep factory address.
    function upgradeFactory(address payable _factory)
        public
        onlyOperatorContractUpgrader
    {
        require(_factory != address(0), "Incorrect factory address");

        if (isUpgradeInitiated()) {
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

        factoryUpgradeInitiatedTimestamp = timestamp;

        emit FactoryUpgradeStarted(_factory, timestamp);
    }

    /// @notice Finalizes ECDSA keep factory upgrade.
    /// @dev It is the second part of the two-step factory upgrade process.
    /// The function can be called after factory upgrade time delay period
    /// has passed since the new factory upgrade. It emits an event
    /// containing the new factory address.
    function completeFactoryUpgrade() public {
        require(isUpgradeInitiated(), "Upgrade not initiated");

        require(
            /* solium-disable-next-line security/no-block-members */
            block.timestamp.sub(factoryUpgradeInitiatedTimestamp) >=
                factoryUpgradeTimeDelay,
            "Timer not elapsed"
        );

        keepFactory = newKeepFactory;
        newKeepFactory = address(0);
        factoryUpgradeInitiatedTimestamp = 0;

        emit FactoryUpgradeCompleted(keepFactory);
    }

    /// @notice Selects the latest ECDSA keep factory.
    /// @return ECDSA keep factory address.
    function selectFactory() public view returns (address payable) {
        require(keepFactory != address(0), "Keep factory is not registered");
        return keepFactory;
    }

    function isUpgradeInitiated() internal view returns (bool) {
        return factoryUpgradeInitiatedTimestamp > 0;
    }

    /// @dev Throws if called by any account other than the operator contract
    /// upgrader authorized for this service contract.
    modifier onlyOperatorContractUpgrader() {
        require(
            address(registry) != address(0),
            "KeepRegistry address is not registered"
        );

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
