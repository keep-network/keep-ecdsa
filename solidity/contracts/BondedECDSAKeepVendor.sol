pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Proxy contract for Bonded ECDSA Keep vendor.
contract BondedECDSAKeepVendor {
    using SafeMath for uint256;

    // The contract owner. The value is stored in the first position. The field
    // should not be moved as the first slot is used for the owner field in the
    // implementation contract.
    //
    // There is a collision for this slot between implementation and proxy, but
    // it is desired as we want to use this value in the implementation contract
    // to protect the initialization function.
    //
    // DO NOT MOVE THIS FIELD FROM THE FIRST POSITION.
    address private _owner;

    // Storage position of the address of the current implementation.
    bytes32 private constant implementationPosition = keccak256(
        "network.keep.bondedecdsavendor.proxy.implementation"
    );

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
        require(
            _implementation != address(0),
            "Implementation address can't be zero."
        );

        initializeImplementation(_implementation, _data);

        setImplementation(_implementation);

        setUpgradeTimeDelay(1 days); // TODO: Determine right value for this property.

        _owner = msg.sender;
    }

    /// @notice Delegates call to the current implementation contract.
    function() external payable {
        address _impl = implementation();
        /* solium-disable-next-line */
        assembly {
            let ptr := mload(0x40)
            calldatacopy(ptr, 0, calldatasize)
            let result := delegatecall(gas, _impl, ptr, calldatasize, 0, 0)
            let size := returndatasize
            returndatacopy(ptr, 0, size)

            switch result
                case 0 {
                    revert(ptr, size)
                }
                default {
                    return(ptr, size)
                }
        }
    }

    /// @notice Starts upgrade of the current vendor implementation.
    /// @dev It is the first part of the two-step implementation address update
    /// process. The function emits an event containing the new value and current
    /// block timestamp.
    /// @param _implementation Address of the new vendor implementation contract.
    /// @param _data Delegate call data for implementation initialization.
    function upgradeToAndCall(address _implementation, bytes memory _data)
        public
        onlyOwner
    {
        address currentImplementation = implementation();
        require(
            _implementation != address(0),
            "Implementation address can't be zero."
        );
        require(
            _implementation != currentImplementation,
            "Implementation address must be different from the current one."
        );

        initializeImplementation(_implementation, _data);

        setNewImplementation(_implementation);

        /* solium-disable-next-line security/no-block-members */
        setUpgradeInitiatedTimestamp(block.timestamp);

        /* solium-disable-next-line security/no-block-members */
        emit UpgradeStarted(_implementation, block.timestamp);
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

    /// @notice Gets the address of the current vendor implementation.
    /// @return Address of the current implementation.
    function implementation() public view returns (address _implementation) {
        bytes32 position = implementationPosition;
        /* solium-disable-next-line */
        assembly {
            _implementation := sload(position)
        }
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

    /// @notice Sets the address of the current implementation.
    /// @param _implementation Address representing the new implementation to
    /// be set.
    function setImplementation(address _implementation) internal {
        bytes32 position = implementationPosition;
        /* solium-disable-next-line */
        assembly {
            sstore(position, _implementation)
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

    /// @notice The owner of the contract.
    /// @return The contract owner's address.
    function owner() public view returns (address) {
        return _owner;
    }

    /// @dev Throws if called by any account other than the contract owner.
    modifier onlyOwner() {
        require(msg.sender == owner(), "Caller is not the owner");
        _;
    }
}
