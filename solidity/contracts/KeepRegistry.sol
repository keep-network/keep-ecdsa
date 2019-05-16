pragma solidity ^0.5.4;

// TODO: For simplification we use import for the other contracts. `ECDSAKeepFactory.sol` 
// and `KeepRegistry.sol` will be kept in different repos in the future so a better
// way of calling another contract from this contract should be introduced.
import "./ECDSAKeepFactory.sol";

/// @title Keep Registry
/// @notice Contract handling keeps registry.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract KeepRegistry {
    // Enumeration of supported keeps types.
    enum KeepTypes {ECDSA, BondedECDSA}

    // Structure holding keep details.
    struct Keep {
        address owner;          // owner of the keep
        address keepAddress;    // address of the keep contract
        KeepTypes keepType;     // type of the keep
    }

    // Factory handling ECDSA keeps.
    address internal ecdsaKeepFactory;

    // List of created keeps.
    Keep[] keeps;

    constructor(address _ecdsaKeepFactory) public {
        require(_ecdsaKeepFactory != address(0), "Implementation address can't be zero.");
        setECDSAKeepFactory(_ecdsaKeepFactory);
    }

    function setECDSAKeepFactory(address _ecdsaKeepFactory) internal {
        ecdsaKeepFactory = _ecdsaKeepFactory;
    }

    /// @notice Build a new ECDSA keep.
    /// @dev Calls ECDSA Keep Factory to build a keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _dishonestThreshold Maximum number of dishonest keep members.
    /// @return Built keep address.
    function buildECDSAKeep(
        uint256 _groupSize,
        uint256 _dishonestThreshold
    ) public returns (address keep) {
        keep = address(ECDSAKeepFactory(ecdsaKeepFactory).buildNewKeep(
            _groupSize,
            _dishonestThreshold,
            msg.sender
        ));

        keeps.push(Keep(msg.sender, keep, KeepTypes.ECDSA));

        return keep;
    }
}
