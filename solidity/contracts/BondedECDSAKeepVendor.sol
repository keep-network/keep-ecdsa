pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";

/// @title Proxy contract for Bonded ECDSA Keep vendor.
contract BondedECDSAKeepVendor is Ownable {
    // Storage position of the address of the current implementation
    bytes32 private constant implementationPosition = keccak256(
        "network.keep.bondedecdsavendor.proxy.implementation"
    );

    event Upgraded(address implementation);

    constructor(address _implementation) public {
        require(_implementation != address(0), "Implementation address can't be zero.");
        setImplementation(_implementation);
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
            case 0 { revert(ptr, size) }
            default { return(ptr, size) }
        }
    }

    /// @notice Upgrades the current vendor implementation.
    /// @param _implementation Address of the new vendor implementation contract.
    function upgradeTo(address _implementation) public onlyOwner {
        address currentImplementation = implementation();
        require(
            _implementation != address(0),
            "Implementation address can't be zero."
        );
        require(
            _implementation != currentImplementation,
            "Implementation address must be different from the current one."
        );
        setImplementation(_implementation);
        emit Upgraded(_implementation);
    }
}