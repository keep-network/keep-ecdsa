pragma solidity ^0.5.4;

import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
// TODO: For simplification we use import for the other contracts. `ECDSAKeepFactory.sol` 
// and `KeepRegistry.sol` will be kept in different repos in the future so a better
// way of calling another contract from this contract should be introduced.
import "./ECDSAKeepFactory.sol";

/// @title Keep Registry
/// @notice Contract handling keeps registry.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract KeepRegistry is Ownable {
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

    function setECDSAKeepFactory(address _ecdsaKeepFactory) public onlyOwner {
        ecdsaKeepFactory = _ecdsaKeepFactory;
    }

    /// @notice Create a new ECDSA keep.
    /// @dev Calls ECDSA Keep Factory to create a keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @return Created keep address.
    function createECDSAKeep(
        uint256 _groupSize,
        uint256 _honestThreshold
    ) public payable returns (address keep) {
        keep = ECDSAKeepFactory(ecdsaKeepFactory).createNewKeep(
            _groupSize,
            _honestThreshold,
            msg.sender
        );

        keeps.push(Keep(msg.sender, keep, KeepTypes.ECDSA));
    }

    //Hacky function to get Keep address of the latest added Keep
    function getLatestKeepAddress()public view returns (address){
        require(keeps.length != 0, "There are no recent keeps");
        return keeps[keeps.length - 1].keepAddress;
    }
}
