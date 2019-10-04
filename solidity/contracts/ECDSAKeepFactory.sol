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
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress) {
        address payable[] memory _members = selectECDSAKeepMembers(_groupSize);

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
    ) internal pure returns (address payable[] memory members){
        // TODO: Implement members selection
        _groupSize;

        members = new address payable[](1);

        // For development we use a member address calculated from the following
        // private key: 0x0789df7d07e6947a93576b9ef60b97aed9adb944fb3ff6bae5215fd3ab0ad0dd
        members[0] = 0x1C25f178599d00b3887BF6D9084cf0C6d49a3097;
    }
}
