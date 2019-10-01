pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating ECDSA keeps.
/// @dev TODO: This is a stub contract - needs to be implemented.
contract ECDSAKeepFactory {
    // List of keeps.
    ECDSAKeep[] keeps;

    // List of candidates to be selected as keep members.
    // TODO: Remove stale members.
    address[] candidates;

    // Notification that a new keep has been created.
    event ECDSAKeepCreated(
        address keepAddress,
        address[] members
    );

    /// @notice Register candidate to be selected as keep member.
    /// @dev TODO: This is a simplified solution until we have proper registration
    /// and group selection.
    function registerCandidate() external {
        candidates.push(msg.sender);
    }

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
        address[] memory _members = selectECDSAKeepMembers(_groupSize);

        ECDSAKeep keep = new ECDSAKeep(
            _owner,
            _members,
            _honestThreshold
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress, _members);
    }

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations generates a group with only one member. Member
    /// is randomly selected from registered candidates.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectECDSAKeepMembers(
        uint256 _groupSize
    ) internal view returns (address[] memory members){
        require(candidates.length > 0, 'candidates list is empty');

        // TODO: Handle groups with more than one member.
        _groupSize;

        members = new address[](1);

        // TODO: Implement with better randomness source.
        uint randomNumber = uint(blockhash(block.number));
        uint randomIndex = randomNumber % (candidates.length + 1);

        members[0] = candidates[randomIndex];
    }
}
