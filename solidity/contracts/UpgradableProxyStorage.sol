pragma solidity ^0.5.4;

/// @title Storage for upgradable proxy contract used to track implementations.
/// @dev This contract can be used to hold implementation details in a two-step
/// upgradeable proxy to hold details of the implementation. In proxy pattern
/// data should be stored in a ways that reduces possibility of collisions between
/// proxy and implementation contracts. In this contract we use a mapping which
/// is allocating storage slot for the data in a dynamic way.
contract UpgradableProxyStorage {
    // Structure holding details of an implementation.
    struct Implementation {
        string version;
        address implementationContract;
        bytes initializationData;
    }

    // Mapping from version ID to implementation contract. It is expected that
    // the map key is a keccak256 of the implementation version.
    mapping(uint256 => Implementation) public implementations;

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
}
