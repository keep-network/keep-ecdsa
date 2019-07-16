pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating ECDSA keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeepFactory {
    // List of keeps.
    ECDSAKeep[] keeps;

    // Notification that a new keep has been created.
    event ECDSAKeepCreated(
        address keepAddress
    );

    /// @notice Open a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Created keep.
    function openKeep(
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
        // TODO: Implement
        _groupSize;

        members = new address[](1);
        members[0] = 0xE1d6c440DC87476242F313aA1179196fAE89B93e;
    }
}
