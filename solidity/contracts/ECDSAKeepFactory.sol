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
    event ECDSAKeepRequested(
        address keepAddress,        // formed keep contract address
        uint256[] keepMembers,      // list of keep members identifiers
        uint256 dishonestThreshold  // maximum number of dishonest keep members
    );

    /// @notice Build a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// @param _groupSize Number of members in the keep.
    /// @param _dishonestThreshold Maximum number of dishonest keep members.
    /// @param _owner Owner of the keep.
    /// @return Built keep.
    function buildNewKeep(
        uint256 _groupSize,
        uint256 _dishonestThreshold,
        address _owner
    ) public payable returns (address keepAddress) {
        uint256[] memory _keepMembers = selectECDSAKeepMembers(_groupSize);

        ECDSAKeep keep = new ECDSAKeep(
            _owner,
            _keepMembers,
            _dishonestThreshold       
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit ECDSAKeepRequested(keepAddress, _keepMembers, _dishonestThreshold);
    }

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations gets only one member with ID `1`.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members IDs.
    function selectECDSAKeepMembers(
        uint256 _groupSize      
    ) internal pure returns (uint256[] memory keepMembers){
        _groupSize;

        keepMembers = new uint256[](1);
        keepMembers[0] = 1;

        // TODO: Currently it assumes members are identified by ID, we should
        // consider changing it to an account address or other unique identfier. 
    }
}
