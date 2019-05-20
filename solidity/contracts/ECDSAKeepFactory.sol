pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";

/// @title ECDSA Keep Factory
/// @notice Contract handling ECDSA keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeepFactory {
    // List of keeps.
    ECDSAKeep[] keeps;

    // Notification that a new keep has been formed. It contains details of the
    // keep.
    event ECDSAKeepCreated(
        address keepAddress        // formed keep contract address
    );

    /// @notice Build a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Owner of the keep.
    /// @return Built keep.
    function createNewKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) public payable returns (address keepAddress) {
        address[] memory _members = selectECDSAKeepMembers(_groupSize);

        ECDSAKeep keep = new ECDSAKeep(
            _owner,
            _members,
            _honestThreshold       
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress);
    }

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations gets only one member with a fixed address.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectECDSAKeepMembers(
        uint256 _groupSize      
    ) internal pure returns (address[] memory members){
        _groupSize;

        members = new address[](1);
        members[0] = 0xE1d6c440DC87476242F313aA1179196fAE89B93e;

        // TODO: Currently it assumes members are identified by ID, we should
        // consider changing it to an account address or other unique identfier. 
    }
}
