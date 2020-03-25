pragma solidity ^0.5.4;

import "./UpgradableProxyStorage.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@openzeppelin/upgrades/contracts/upgradeability/Proxy.sol";

/// @title Proxy contract for Bonded ECDSA Keep vendor.
contract BondedECDSAKeepVendor is Proxy, UpgradableProxyStorage {
    using SafeMath for uint256;

    event UpgradeStarted(
        string version,
        address implementation,
        bytes initialization,
        uint256 timestamp
    );
    event UpgradeCompleted(string version, address implementation);

    constructor(
        string memory _version,
        address _implementation,
        bytes memory _data
    ) public UpgradableProxyStorage() {
        require(
            _implementation != address(0),
            "Implementation address can't be zero."
        );

        setCurrentImplementationID(
            uint256(keccak256(abi.encodePacked(_version)))
        );

        setImplementation(_version, _implementation, _data);

        if (_data.length > 0) {
            initializeImplementation(_implementation, _data);
        }

        setUpgradeTimeDelay(1 days);

        setAdmin(msg.sender);
    }

    /// @notice Starts upgrade of the current vendor implementation.
    /// @dev It is the first part of the two-step implementation address update
    /// process. The function emits an event containing the new value and current
    /// block timestamp.
    /// @param _newVersion Version of the new vendor implementation contract.
    /// @param _newImplementation Address of the new vendor implementation contract.
    /// @param _data Delegate call data for implementation initialization.
    function upgradeToAndCall(
        string memory _newVersion,
        address _newImplementation,
        bytes memory _data
    ) public onlyAdmin {
        require(
            _newImplementation != address(0),
            "Implementation address can't be zero."
        );
        require(
            bytes(_newVersion).length > 0,
            "Version can't be empty string."
        );

        uint256 newVersionID = uint256(
            keccak256(abi.encodePacked(_newVersion))
        );

        require(
            newVersionID != currentImplementationID(),
            "Implementation version must be different from the current one."
        );

        // Check if the new version is already registered, but in case the upgrade
        // for given version is already in progress let to update it's details.
        if (newVersionID != upgradeImplementationID()) {
            Implementation memory implementation = getImplementation(
                newVersionID
            );
            require(
                bytes(implementation.version).length == 0,
                "Implementation version has already been registered before."
            );
        }

        setUpgradeImplementationID(newVersionID);

        setImplementation(_newVersion, _newImplementation, _data);

        /* solium-disable-next-line security/no-block-members */
        setUpgradeInitiatedTimestamp(block.timestamp);

        emit UpgradeStarted(
            _newVersion,
            _newImplementation,
            _data,
            /* solium-disable-next-line security/no-block-members */
            block.timestamp
        );
    }

    /// @notice Finalizes implementation address upgrade.
    /// @dev It is the second part of the two-step implementation address update
    /// process. The function emits an event containing the new implementation
    /// address. It can be called after upgrade time delay period has passed since
    /// upgrade initiation. If the upgrade initialization data had been stored
    /// the function will call the new implementation contract to initialize.
    function completeUpgrade() public onlyAdmin {
        require(upgradeInitiatedTimestamp() > 0, "Upgrade not initiated");

        require(
            /* solium-disable-next-line security/no-block-members */
            block.timestamp.sub(upgradeInitiatedTimestamp()) >=
                upgradeTimeDelay(),
            "Timer not elapsed"
        );

        setCurrentImplementationID(upgradeImplementationID());

        Implementation memory implementation = getImplementation(
            upgradeImplementationID()
        );

        if (implementation.initializationData.length > 0) {
            initializeImplementation(
                implementation.implementationContract,
                implementation.initializationData
            );
        }

        setUpgradeImplementationID(0);
        setUpgradeInitiatedTimestamp(0);

        emit UpgradeCompleted(
            implementation.version,
            implementation.implementationContract
        );
    }

    /// @notice Initializes implementation contract.
    /// @dev Delegates a call to the implementation with provided data. It is
    /// expected that data contains details of function to be called.
    /// @param _implementation Address of the new vendor implementation contract.
    /// @param _data Delegate call data for implementation initialization.
    function initializeImplementation(
        address _implementation,
        bytes memory _data
    ) internal {
        (bool success, bytes memory returnData) = _implementation.delegatecall(
            _data
        );

        require(success, string(returnData));
    }

    /// @notice Gets the address of the current vendor implementation address.
    /// @return Address of the current implementation.
    function implementation() public view returns (address) {
        return _implementation();
    }

    /// @dev Returns the current implementation address. Implements function
    /// from `Proxy` contract.
    /// @return Address of the current implementation
    function _implementation() internal view returns (address) {
        Implementation memory implementation = getImplementation(
            currentImplementationID()
        );

        return implementation.implementationContract;
    }

    /// @dev Returns the current implementation version.
    /// @return Version of the current implementation
    function version() public view returns (string memory) {
        Implementation memory implementation = getImplementation(
            currentImplementationID()
        );

        return implementation.version;
    }

    /// @dev Throws if called by any account other than the contract owner.
    modifier onlyAdmin() {
        require(msg.sender == admin(), "Caller is not the admin");
        _;
    }
}
