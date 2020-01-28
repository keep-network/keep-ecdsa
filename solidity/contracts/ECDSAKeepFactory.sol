pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./api/IECDSAKeepFactory.sol";
import "./utils/AddressArrayUtils.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@keep-network/sortition-pools/contracts/SortitionPool.sol";
import "@keep-network/sortition-pools/contracts/SortitionPoolFactory.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
contract ECDSAKeepFactory is IECDSAKeepFactory { // TODO: Rename to BondedECDSAKeepFactory
    using AddressArrayUtils for address payable[];
    using SafeMath for uint256;

    // Notification that a new keep has been created.
    event ECDSAKeepCreated(
        address keepAddress,
        address payable[] members,
        address owner,
        address application
    );

    // Mapping of pools with registered member candidates for each application.
    mapping(address => address) candidatesPools; // application -> candidates pool

    bytes32 groupSelectionSeed;

    SortitionPoolFactory sortitionPoolFactory;

    constructor(address _sortitionPoolFactory) public {
        sortitionPoolFactory = SortitionPoolFactory(_sortitionPoolFactory);
    }

    /// @notice Register caller as a candidate to be selected as keep member
    /// for the provided customer application
    /// @dev If caller is already registered it returns without any changes.
    function registerMemberCandidate(address _application) external {
        if (candidatesPools[_application] == address(0)) {
            // This is the first time someone registers as signer for this
            // application so let's create a signer pool for it.
            candidatesPools[_application] = sortitionPoolFactory.createSortitionPool();
        }

        SortitionPool candidatesPool = SortitionPool(candidatesPools[_application]);
        candidatesPool.insertOperator(msg.sender, 500); // TODO: take weight from staking contract
    }

    /// @notice Open a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// A caller of this function is expected to be an application for which
    /// member candidates were registered in a pool.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner
    ) external payable returns (address keepAddress) {
        address application = msg.sender;
        address pool = candidatesPools[application];
        require(pool != address(0), "No signer pool for this application");

        // TODO: Check if enough operators are registered in the pool.

        address[] memory selected = SortitionPool(pool).selectGroup(
            _groupSize,
            groupSelectionSeed
        );

        uint256 latestSelectionIteration = _groupSize;

        address payable[] memory members = new address payable[](_groupSize);
        for (uint256 i = 0; i < _groupSize; i++) {
            // TODO: This is a temporary solution until client is able to handle
            // multiple members for one operator.
            (members[i], latestSelectionIteration) = ensureMemberUniqueness(
                SortitionPool(pool),
                members,
                address(uint160(selected[i])),
                latestSelectionIteration++
            );

          // TODO: for each selected member, validate staking weight and create,
          // bond. If validation failed or bond could not be created, remove
          // operator from pool and try again.

        }

        ECDSAKeep keep = new ECDSAKeep(_owner, members, _honestThreshold);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress, members, _owner, application);

        // TODO: as beacon for new entry and update groupSelectionSeed in callback

    }

    /// @notice Ensures a new member is unique for the current members array.
    /// @dev If the member is already included in the current members array it
    /// selects a new member in a loop until a unique member is obtained. The pool's
    /// `selectGroup` function ensures that the same set of members is selected
    /// for given group size and seed, hence we need to modify the group size
    /// to get new selection results until we hit a unique one.
    /// @param _pool Sortition Pool to use for group selection.
    /// @param _currentMembers Currently selected unique members.
    /// @param _newMember Address of a member to validate uniqueness.
    /// @param _currentIteration Track iteration for group selection in the sortition
    /// pool.
    /// @return Unique member for the current members array.
    function ensureMemberUniqueness(
        SortitionPool _pool,
        address payable[] memory _currentMembers,
        address payable _newMember,
        uint256 _currentIteration
    ) internal returns (address payable, uint256) {
        if (_currentMembers.contains(_newMember)) {
            address replacement = _pool.selectGroup(
                _currentIteration,
                groupSelectionSeed
            )[_currentIteration - 1];

            return
                ensureMemberUniqueness(
                    _pool,
                    _currentMembers,
                    address(uint160(replacement)),
                    _currentIteration + 1
                );
        } else {
            return (_newMember, _currentIteration);
        }
    }
}
