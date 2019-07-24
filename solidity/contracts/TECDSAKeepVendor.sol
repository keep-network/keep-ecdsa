pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./utils/AddressArrayUtils.sol";
import "./TECDSAKeepFactory.sol";

/// @title TECDSA Keep Vendor
/// @notice The contract can be used to obtain a new TECDSA keep.
/// @dev Interacts with TECDSA keep factory to obtain a new instance of the TECDSA
/// keep. Several versions of TECDSA keep factories can be registered for the vendor.
/// TODO: This is a stub contract - needs to be implemented.
/// TODO: When more keep types are added consider extracting registration and
/// selection to a separate inheritable contract.
contract TECDSAKeepVendor is Ownable {
    using AddressArrayUtils for address[];

    // List of TECDSA keep factories.
    address[] public factories;

    /// @notice Register new TECDSA keep factory.
    /// @dev Adds a factory address to the list of registered factories. Address
    /// cannot be zero and cannot be already registered.
    /// @param _factory TECDSA keep factory address.
    function registerFactory(address _factory) public onlyOwner {
        require(!factories.contains(_factory), "Factory address already registered");

        factories.push(_factory);
    }

    /// @notice Select a recommended TECDSA keep factory from all registered
    /// TECDSA keep factories.
    /// @dev This is a stub implementation returning first factory on the list.
    /// @return Selected TECDSA keep factory address.
    function selectFactory() internal view returns (address) {
        // TODO: Implement factory selection mechanism.
        return factories[factories.length - 1];
    }

    /// @notice Open a new TECDSA keep.
    /// @dev Calls a recommended TECDSA keep factory to open a new keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Opened keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) public payable returns (address keepAddress) {
        address factory = selectFactory();

        return TECDSAKeepFactory(factory).openKeep(
            _groupSize,
            _honestThreshold,
            _owner
        );
    }
}
