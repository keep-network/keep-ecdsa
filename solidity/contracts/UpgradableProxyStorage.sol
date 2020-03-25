pragma solidity ^0.5.4;

/// @title Storage for two-step upgradable proxy contract.
/// @dev This contract can be used to hold implementation details in a two-step
/// upgradeable proxy to hold details of the implementation. In proxy pattern
/// data should be stored in a ways that reduces possibility of collisions between
/// proxy and implementation contracts. In this contract we define variables at
/// fixed positions according to the Unstructured Storage pattern and a mapping
/// which is allocating storage slot for the data in a dynamic way.
contract UpgradableProxyStorage {
    /// @dev Storage slot with the admin of the contract.
    /// This is the keccak-256 hash of "eip1967.proxy.admin" subtracted by 1, and is
    /// validated in the constructor.
    bytes32 internal constant ADMIN_SLOT = 0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103;

    /// @dev Storage slot with the upgrade time delay. Upgrade time delay defines a
    /// period for implementation upgrade.
    /// This is the keccak-256 hash of "network.keep.proxy.upgradetimedelay"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant UPGRADE_TIME_DELAY_SLOT = 0xa0361fd71e28dd9e0781644db9ca971cf5aeafd12b5d0aaaebff68299a2cbbee;

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
    /// This is the keccak-256 hash of "network.keep.proxy.upgradeinitiatedtimestamp"
    /// subtracted by 1, and is validated in the constructor.
    bytes32 internal constant UPGRADE_INIT_TIMESTAMP_SLOT = 0xf4917e72ec7208070be0cec8873a121e5971a5d90b452600ca1524183a489ed5;

    // Structure holding details of an implementation.
    struct Implementation {
        string version;
        address implementationContract;
        bytes initializationData;
    }

    // Mapping from version ID to implementation contract. It is expected that
    // the map key is a keccak256 of the implementation version.
    mapping(uint256 => Implementation) public implementations;

    constructor() public {
        assertSlot(ADMIN_SLOT, "eip1967.proxy.admin");
        assertSlot(
            UPGRADE_TIME_DELAY_SLOT,
            "network.keep.proxy.upgradetimedelay"
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
            "network.keep.proxy.upgradeinitiatedtimestamp"
        );
    }

    function getImplementation(string memory _version)
        internal
        view
        returns (Implementation memory)
    {
        return
            getImplementation(uint256(keccak256(abi.encodePacked(_version))));
    }

    function getImplementation(uint256 _version)
        internal
        view
        returns (Implementation memory)
    {
        return implementations[_version];
    }

    function setImplementation(
        string memory _version,
        address _implementation,
        bytes memory _initializationData
    ) internal {
        uint256 versionInt = uint256(keccak256(abi.encodePacked(_version)));

        implementations[versionInt].version = _version;
        implementations[versionInt].implementationContract = _implementation;
        implementations[versionInt].initializationData = _initializationData;
    }

    /// @notice Asserts correct slot for provided key.
    /// @dev To avoid clashing with implementation's fields the proxy contract
    /// defines its' fields on specific slots. Slot is calculated as hash of a string
    /// subtracted by 1 to reduce chances of a possible attack. For details see
    /// EIP-1967.
    function assertSlot(bytes32 slot, bytes memory key) internal pure {
        assert(slot == bytes32(uint256(keccak256(key)) - 1));
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

    /// @notice The admin slot.
    /// @return The contract owner's address.
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
}
