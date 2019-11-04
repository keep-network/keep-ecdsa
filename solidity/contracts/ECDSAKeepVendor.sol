pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./api/IECDSAKeepVendor.sol";
import "./utils/AddressArrayUtils.sol";
import "./ECDSAKeepFactory.sol";

/// @title ECDSA Keep Vendor
/// @notice The contract can be used to obtain a new ECDSA keep.
/// @dev Interacts with ECDSA keep factory to obtain a new instance of the ECDSA
/// keep. Several versions of ECDSA keep factories can be registered for the vendor.
/// TODO: This is a stub contract - needs to be implemented.
/// TODO: When more keep types are added consider extracting registration and
/// selection to a separate inheritable contract.
contract ECDSAKeepVendor is IECDSAKeepVendor, Ownable {
    using AddressArrayUtils for address payable[];

    // List of ECDSA keep factories.
    address payable[] public factories;

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
    function selectFactory() internal view returns (address payable) {
        require(factories.length > 0, "No factories registered");

        // TODO: Implement factory selection mechanism.
        return factories[factories.length - 1];
    }

    /// @notice Open a new ECDSA keep.
    /// @dev Calls a recommended ECDSA keep factory to open a new keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Opened keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress) {
        address factory = selectFactory();

        keepAddress = ECDSAKeepFactory(factory).openKeep.value(msg.value)(
            _groupSize,
            _honestThreshold,
            _owner
        );
    }
}
