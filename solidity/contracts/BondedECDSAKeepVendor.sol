pragma solidity ^0.5.4;

import "./UpgradableProxyStorage.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@openzeppelin/upgrades/contracts/upgradeability/Proxy.sol";

/// @title Proxy contract for Bonded ECDSA Keep vendor.
contract BondedECDSAKeepVendor is Proxy, UpgradableProxyStorage {
    using SafeMath for uint256;

    /// @dev Storage slot with the admin of the contract.
    /// This is the keccak-256 hash of "eip1967.proxy.admin" subtracted by 1, and is
    /// validated in the constructor.
    bytes32 internal constant ADMIN_SLOT = 0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103;

    /// @dev Storage slot with the upgrade time delay. Upgrade time delay defines a
    /// period for implementation upgrade.
    /// This is the keccak-256 hash of "network.keep.bondedecdsavendor.proxy.upgradeTimeDelay"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant UPGRADE_TIME_DELAY_SLOT = 0x3ca583dafde9ce8bdb41fe825f85984a83b08ecf90ffaccbc4b049e8d8703563;

    /// @dev Storage slot with the ID of the current implementation. The ID is
    /// an unique identifier of the implementation calculated based on the version.
    /// This is the keccak-256 hash of "network.keep.proxy.currentimplementationid"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant CURRENT_IMPL_ID_SLOT = 0x2fdc2b7761cb35b3c6cce7021b2c04230161423bd5b9520fd8a7ae15aa03ac9f;

    /// @dev Storage slot with the new implementation ID.
    /// This is the keccak-256 hash of "network.keep.proxy.upgradeimplementationid"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant UPGRADE_IMPL_ID_SLOT = 0x1836489d800664edd6c29124d8cf32e481082cc2687a6f957a01e0f20a003d40;

    /// @dev Storage slot with the implementation address upgrade initiation.
    /// This is the keccak-256 hash of "network.keep.bondedecdsavendor.proxy.upgradeInitiatedTimestamp"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant UPGRADE_INIT_TIMESTAMP_SLOT = 0x0816e8d9eeb2554df0d0b7edc58e2d957e6ce18adf92c138b50dd78a420bebaf;

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
    ) public {
        assertSlot(ADMIN_SLOT, "eip1967.proxy.admin");
        assertSlot(
            UPGRADE_TIME_DELAY_SLOT,
            "network.keep.bondedecdsavendor.proxy.upgradeTimeDelay"
        );
        assertSlot(
            CURRENT_IMPL_ID_SLOT,
            "network.keep.proxy.currentimplementationid"
        );
        assertSlot(
            UPGRADE_IMPL_ID_SLOT,
            "network.keep.proxy.upgradeimplementationid"
        );
        assertSlot(
            UPGRADE_INIT_TIMESTAMP_SLOT,
            "network.keep.bondedecdsavendor.proxy.upgradeInitiatedTimestamp"
        );

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

    /// @notice Asserts correct slot for provided key.
    /// @dev To avoid clashing with implementation's fields the proxy contract
    /// defines its' fields on specific slots. Slot is calculated as hash of a string
    /// subtracted by 1 to reduce chances of a possible attack. For details see
    /// EIP-1967.
    function assertSlot(bytes32 slot, bytes memory key) internal pure {
        assert(slot == bytes32(uint256(keccak256(key)) - 1));
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

    function currentImplementationID()
        internal
        view
        returns (uint256 _version)
    {
        bytes32 position = CURRENT_IMPL_ID_SLOT;
        /* solium-disable-next-line */
        assembly {
            _version := sload(position)
        }
    }

    function setCurrentImplementationID(uint256 _id) internal {
        bytes32 position = CURRENT_IMPL_ID_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _id)
        }
    }

    function upgradeImplementationID() internal view returns (uint256 _newID) {
        bytes32 position = UPGRADE_IMPL_ID_SLOT;
        /* solium-disable-next-line */
        assembly {
            _newID := sload(position)
        }
    }

    function setUpgradeImplementationID(uint256 _newID) internal {
        bytes32 position = UPGRADE_IMPL_ID_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _newID)
        }
    }

    function upgradeInitiatedTimestamp()
        public
        view
        returns (uint256 _upgradeInitiatedTimestamp)
    {
        bytes32 position = UPGRADE_INIT_TIMESTAMP_SLOT;
        /* solium-disable-next-line */
        assembly {
            _upgradeInitiatedTimestamp := sload(position)
        }
    }

    function setUpgradeInitiatedTimestamp(uint256 _upgradeInitiatedTimestamp)
        internal
    {
        bytes32 position = UPGRADE_INIT_TIMESTAMP_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _upgradeInitiatedTimestamp)
        }
    }

    function upgradeTimeDelay()
        public
        view
        returns (uint256 _upgradeTimeDelay)
    {
        bytes32 position = UPGRADE_TIME_DELAY_SLOT;
        /* solium-disable-next-line */
        assembly {
            _upgradeTimeDelay := sload(position)
        }
    }

    function setUpgradeTimeDelay(uint256 _upgradeTimeDelay) internal {
        bytes32 position = UPGRADE_TIME_DELAY_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _upgradeTimeDelay)
        }
    }

    /// @notice The admin slot.
    /// @return The contract owner's address.
    function admin() public view returns (address adm) {
        bytes32 slot = ADMIN_SLOT;
        /* solium-disable-next-line */
        assembly {
            adm := sload(slot)
        }
    }

    /// @dev Sets the address of the proxy admin.
    /// @param _newAdmin Address of the new proxy admin.
    function setAdmin(address _newAdmin) internal {
        bytes32 slot = ADMIN_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(slot, _newAdmin)
        }
    }

    /// @dev Throws if called by any account other than the contract owner.
    modifier onlyAdmin() {
        require(msg.sender == admin(), "Caller is not the admin");
        _;
    }
}
