pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@openzeppelin/upgrades/contracts/upgradeability/Proxy.sol";

/// @title Proxy contract for Bonded ECDSA Keep vendor.
contract BondedECDSAKeepVendor is Proxy {
    using SafeMath for uint256;

    /// @dev Storage slot with the admin of the contract.
    /// This is the keccak-256 hash of "eip1967.proxy.admin" subtracted by 1, and is
    /// validated in the constructor.
    bytes32 internal constant ADMIN_SLOT = 0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103;

    /// @dev Storage slot with the address of the current implementation.
    /// This is the keccak-256 hash of "eip1967.proxy.implementation" subtracted by 1, and is
    /// validated in the constructor.
    bytes32 internal constant IMPLEMENTATION_SLOT = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc;

    // Storage position of the upgrade time delay. Upgrade time delay defines a
    // period for implementation upgrade.
    bytes32 private constant upgradeTimeDelayPosition = keccak256(
        "network.keep.bondedecdsavendor.proxy.upgradeTimeDelay"
    );

    // Storage position of the new implementation address.
    bytes32 private constant newImplementationPosition = keccak256(
        "network.keep.bondedecdsavendor.proxy.newImplementation"
    );

    // Storage position of the implementation address upgrade initiation.
    bytes32 private constant upgradeInitiatedTimestampPosition = keccak256(
        "network.keep.bondedecdsavendor.proxy.upgradeInitiatedTimestamp"
    );

    event UpgradeStarted(address implementation, uint256 timestamp);
    event Upgraded(address implementation);

    constructor(address _implementation, bytes memory _data) public {
        assert(
            IMPLEMENTATION_SLOT ==
                bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1)
        );

        assert(
            ADMIN_SLOT == bytes32(uint256(keccak256("eip1967.proxy.admin")) - 1)
        );

        require(
            _implementation != address(0),
            "Implementation address can't be zero."
        );

        initializeImplementation(_implementation, _data);

        setImplementation(_implementation);

        setUpgradeTimeDelay(1 days); // TODO: Determine right value for this property.

        setAdmin(msg.sender);
    }

    /// @notice Starts upgrade of the current vendor implementation.
    /// @dev It is the first part of the two-step implementation address update
    /// process. The function emits an event containing the new value and current
    /// block timestamp.
    /// @param _newImplementation Address of the new vendor implementation contract.
    /// @param _data Delegate call data for implementation initialization.
    function upgradeToAndCall(address _newImplementation, bytes memory _data)
        public
        onlyAdmin
    {
        address currentImplementation = _implementation();
        require(
            _newImplementation != address(0),
            "Implementation address can't be zero."
        );
        require(
            _newImplementation != currentImplementation,
            "Implementation address must be different from the current one."
        );

        initializeImplementation(_newImplementation, _data);

        setNewImplementation(_newImplementation);

        /* solium-disable-next-line security/no-block-members */
        setUpgradeInitiatedTimestamp(block.timestamp);

        /* solium-disable-next-line security/no-block-members */
        emit UpgradeStarted(_newImplementation, block.timestamp);
    }

    /// @notice Finalizes implementation address upgrade.
    /// @dev It is the second part of the two-step implementation address update
    /// process. The function emits an event containing the new implementation
    /// address. It can be called after upgrade time delay period has passed since
    /// upgrade initiation.
    function completeUpgrade() public {
        require(upgradeInitiatedTimestamp() > 0, "Upgrade not initiated");

        require(
            /* solium-disable-next-line security/no-block-members */
            block.timestamp.sub(upgradeInitiatedTimestamp()) >=
                upgradeTimeDelay(),
            "Timer not elapsed"
        );

        address newImplementation = newImplementation();

        setImplementation(newImplementation);
        setUpgradeInitiatedTimestamp(0);

        emit Upgraded(newImplementation);
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
        if (_data.length > 0) {
            (bool success, bytes memory returnData) = _implementation
                .delegatecall(_data);

            require(success, string(returnData));
        }
    }

    /// @notice Gets the address of the current vendor implementation.
    /// @return Address of the current implementation.
    function implementation() public view returns (address) {
        return _implementation();
    }

    /// @dev Returns the current implementation. Implements function from `Proxy`
    /// contract.
    /// @return Address of the current implementation
    function _implementation() internal view returns (address impl) {
        bytes32 slot = IMPLEMENTATION_SLOT;
        /* solium-disable-next-line */
        assembly {
            impl := sload(slot)
        }
    }

    /// @notice Sets the address of the current implementation.
    /// @param _implementation Address representing the new implementation to
    /// be set.
    function setImplementation(address _implementation) internal {
        bytes32 slot = IMPLEMENTATION_SLOT;
        /* solium-disable-next-line */
        assembly {
            sstore(slot, _implementation)
        }
    }

    function upgradeTimeDelay()
        public
        view
        returns (uint256 _upgradeTimeDelay)
    {
        bytes32 position = upgradeTimeDelayPosition;
        /* solium-disable-next-line */
        assembly {
            _upgradeTimeDelay := sload(position)
        }
    }

    function setUpgradeTimeDelay(uint256 _upgradeTimeDelay) internal {
        bytes32 position = upgradeTimeDelayPosition;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _upgradeTimeDelay)
        }
    }

    function newImplementation()
        public
        view
        returns (address _newImplementation)
    {
        bytes32 position = newImplementationPosition;
        /* solium-disable-next-line */
        assembly {
            _newImplementation := sload(position)
        }
    }

    function setNewImplementation(address _newImplementation) internal {
        bytes32 position = newImplementationPosition;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _newImplementation)
        }
    }

    function upgradeInitiatedTimestamp()
        public
        view
        returns (uint256 _upgradeInitiatedTimestamp)
    {
        bytes32 position = upgradeInitiatedTimestampPosition;
        /* solium-disable-next-line */
        assembly {
            _upgradeInitiatedTimestamp := sload(position)
        }
    }

    function setUpgradeInitiatedTimestamp(uint256 _upgradeInitiatedTimestamp)
        internal
    {
        bytes32 position = upgradeInitiatedTimestampPosition;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _upgradeInitiatedTimestamp)
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
