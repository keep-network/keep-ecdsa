pragma solidity ^0.5.4;

import "./TECDSAKeep.sol";

/// @title TECDSA Keep Factory
/// @notice Contract creating TECDSA keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract TECDSAKeepFactory {
    // List of keeps.
    TECDSAKeep[] keeps;

    // Notification that a new keep has been created.
    event TECDSAKeepCreated(
        address keepAddress
    );

    /// @notice Open a new TECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) public payable returns (address keepAddress) {
        address[] memory _members = selectMembers(_groupSize);

        TECDSAKeep keep = new TECDSAKeep(
            _owner,
            _members,
            _honestThreshold
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit TECDSAKeepCreated(keepAddress);
    }

    /// @notice Runs member selection for an TECDSA keep.
    /// @dev Stub implementations gets only one member with a fixed address.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectMembers(
        uint256 _groupSize
    ) internal pure returns (address[] memory members){
        // TODO: Implement
        _groupSize;

        members = new address[](1);
        members[0] = 0xE1d6c440DC87476242F313aA1179196fAE89B93e;
    }
}
