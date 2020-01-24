pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./api/IECDSAKeepFactory.sol";
import "./utils/AddressArrayUtils.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
contract ECDSAKeepFactory is IECDSAKeepFactory { // TODO: Rename to BondedECDSAKeepFactory
    using AddressArrayUtils for address payable[];
    using SafeMath for uint256;

    // List of keeps.
    ECDSAKeep[] keeps;

    // Tickets submitted by member candidates during the signing group selection
    // execution and accepted by the protocol for consideration.
    uint64[] tickets;

    // Map simulates a sorted linked list of ticket values by their indexes.
    // key -> value represent indices from the tickets[] array.
    // 'key' index holds an index of a ticket and 'value' holds an index
    // of the next ticket. Tickets are sorted by their value in
    // descending order starting from the tail.
    // Ex. tickets = [151, 42, 175, 7]
    // tail: 2 because tickets[2] = 175
    // previousTicketIndex[0] -> 1
    // previousTicketIndex[1] -> 3
    // previousTicketIndex[2] -> 0
    // previousTicketIndex[3] -> 3 note: index that holds a lowest
    // value points to itself because there is no `nil` in Solidity.
    // Traversing from tail: [2]->[0]->[1]->[3] result in 175->151->42->7
    mapping(uint256 => uint256) previousTicketIndex;

    // Pseudorandom seed value used as an input for the pool selection.
    // TODO: call random beacon for a new seed. This is hardcoded for now.
    uint256 seed = 31415926535897932384626433832795028841971693993751058209749445923078164062862;

    // Tail represents an index of a ticket in a tickets[] array which holds
    // the highest ticket value. It is a tail of the linked list defined by
    // `previousTicketIndex`.
    uint256 tail;

    // Number of block at which the pool selection started and from which
    // ticket submissions are accepted.
    uint256 ticketSubmissionStartBlock = block.number;

    // Timeout in blocks after which the ticket submission is finished.
    uint256 ticketSubmissionTimeout = 12;

    // Ticket's size pool.
    uint256 public poolSize = 60;

    // List of candidates to be selected as keep members. Once the candidate is
    // registered it remains on the list forever.
    // TODO: It's a temporary solution until we implement proper candidate
    // registration and member selection.
    address payable[] memberCandidates;

    // Information about ticket submitters.
    mapping(uint256 => address payable) candidates;

    // Notification that a new keep has been created.
    event ECDSAKeepCreated(
        address keepAddress,
        address payable[] members
    );

    /// @notice Register caller as a candidate to be selected as keep member.
    /// @dev If caller is already registered it returns without any changes.
    /// TODO: This is a simplified solution until we have proper registration
    /// and pool selection.
    function registerMemberCandidate() external {
        if (!memberCandidates.contains(msg.sender)) {
            memberCandidates.push(msg.sender);
        }
    }

    /**
     * @dev Submits ticket to request to participate in a keep.
     * @param ticket Bytes representation of a ticket that holds the following:
     * - ticketValue: first 8 bytes of a result of keccak256 cryptography hash
     *   function on the combination of the pool selection seed, staker-specific
     *   value (address) and virtual staker index.
     * - stakerValue: a staker-specific value which is the address of the staker.
     * - virtualStakerIndex: 4-bytes number within a range of 1 to staker's weight;
     *   has to be unique for all tickets submitted by the given staker for the
     *   current candidate pool selection.
     */
    function submitTicket(bytes32 ticket) public {
        uint64 ticketValue;
        uint160 stakerValue;
        uint32 virtualStakerIndex;

        bytes memory ticketBytes = abi.encodePacked(ticket);
        /* solium-disable-next-line */
        assembly {
            // ticket value is 8 bytes long
            ticketValue := mload(add(ticketBytes, 8))
            // staker value is 20 bytes long
            stakerValue := mload(add(ticketBytes, 28))
            // virtual staker index is 4 bytes long
            virtualStakerIndex := mload(add(ticketBytes, 32))
        }

        uint256 stakingWeight = 10000; // TODO: hardcoded, need to implement getting the right value.

        if (block.number > ticketSubmissionStartBlock.add(ticketSubmissionTimeout)) {
            revert("Ticket submission is over");
        }

        if (candidates[ticketValue] != address(0)) {
            revert("Duplicate ticket");
        }

        if (isTicketValid(
            ticketValue,
            stakerValue,
            virtualStakerIndex,
            stakingWeight,
            seed
        )) {
            addTicket(ticketValue);
        } else {
            revert("Invalid ticket");
        }
    }

    function isTicketValid(
        uint64 ticketValue,
        uint256 stakerValue,
        uint256 virtualStakerIndex,
        uint256 stakingWeight,
        uint256 selectionSeed
    ) internal view returns(bool) {
        uint64 ticketValueExpected;
        bytes memory ticketBytes = abi.encodePacked(
            keccak256(
                abi.encodePacked(
                    selectionSeed,
                    stakerValue,
                    virtualStakerIndex
                )
            )
        );
        // use first 8 bytes to compare ticket values
        /* solium-disable-next-line */
        assembly {
            ticketValueExpected := mload(add(ticketBytes, 8))
        }

        bool isVirtualStakerIndexValid = virtualStakerIndex > 0 && virtualStakerIndex <= stakingWeight;
        bool isStakerValueValid = stakerValue == uint256(msg.sender);
        bool isTicketValueValid = ticketValue == ticketValueExpected;

        return isVirtualStakerIndexValid && isStakerValueValid && isTicketValueValid;
    }

     /**
     * @dev Adds a new, verified ticket. Ticket is accepted when it is lower
     * than the currently highest ticket or when the number of tickets is still
     * below the pool size.
     */
    function addTicket(uint64 newTicketValue) internal {
        uint256[] memory ordered = getTicketValueOrderedIndices();

        // any ticket goes when the tickets array size is lower than the pool size
        if (tickets.length < poolSize) {
            // no tickets
            if (tickets.length == 0) {
                tickets.push(newTicketValue);
            // higher than the current highest
            } else if (newTicketValue > tickets[tail]) {
                tickets.push(newTicketValue);
                uint256 oldTail = tail;
                tail = tickets.length-1;
                previousTicketIndex[tail] = oldTail;
            // lower than the current lowest
            } else if (newTicketValue < tickets[ordered[0]]) {
                tickets.push(newTicketValue);
                // last element points to itself
                previousTicketIndex[tickets.length - 1] = tickets.length - 1;
                // previous lowest ticket points to the new lowest
                previousTicketIndex[ordered[0]] = tickets.length - 1;
            // higher than the lowest ticket value and lower than the highest ticket value
            } else {
                tickets.push(newTicketValue);
                uint256 j = findReplacementIndex(newTicketValue, ordered);
                previousTicketIndex[tickets.length - 1] = previousTicketIndex[j];
                previousTicketIndex[j] = tickets.length - 1;
            }
            candidates[newTicketValue] = msg.sender;
        } else if (newTicketValue < tickets[tail]) {
            uint256 ticketToRemove = tickets[tail];
            // new ticket is lower than currently lowest
            if (newTicketValue < tickets[ordered[0]]) {
                // replacing highest ticket with the new lowest
                tickets[tail] = newTicketValue;
                uint256 newTail = previousTicketIndex[tail];
                previousTicketIndex[ordered[0]] = tail;
                previousTicketIndex[tail] = tail;
                tail = newTail;
            } else { // new ticket is between lowest and highest
                uint256 j = findReplacementIndex(newTicketValue, ordered);
                tickets[tail] = newTicketValue;
                // do not change the order if a new ticket is still highest
                if (j != tail) {
                    uint newTail = previousTicketIndex[tail];
                    previousTicketIndex[tail] = previousTicketIndex[j];
                    previousTicketIndex[j] = tail;
                    tail = newTail;
                }
            }
            // we are replacing tickets so we also need to replace information
            // about the submitter
            delete candidates[ticketToRemove];
            candidates[newTicketValue] = msg.sender;
        }
    }

    /**
     * @dev Use binary search to find an index for a new ticket in the tickets[] array
     */
    function findReplacementIndex(
        uint64 newTicketValue,
        uint256[] memory ordered
    ) internal view returns (uint256) {
        uint256 lo = 0;
        uint256 hi = ordered.length - 1;
        uint256 mid = 0;
        while (lo <= hi) {
            mid = (lo + hi) >> 1;
            if (newTicketValue < tickets[ordered[mid]]) {
                hi = mid - 1;
            } else if (newTicketValue > tickets[ordered[mid]]) {
                lo = mid + 1;
            } else {
                return ordered[mid];
            }
        }

        return ordered[lo];
    }

    /**
     * @dev Creates an array of ticket indexes based on their values in the ascending order:
     *
     * ordered[n-1] = tail
     * ordered[n-2] = previousTicketIndex[tail]
     * ordered[n-3] = previousTicketIndex[ordered[n-2]]
     */
    function getTicketValueOrderedIndices() internal view returns (uint256[] memory) {
        uint256[] memory ordered = new uint256[](tickets.length);
        if (ordered.length > 0) {
            ordered[tickets.length-1] = tail;
            if (ordered.length > 1) {
                for (uint256 i = tickets.length - 1; i > 0; i--) {
                    ordered[i-1] = previousTicketIndex[ordered[i]];
                }
            }
        }

        return ordered;
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
        address payable[] memory _members = selectECDSAKeepMembers(_groupSize);

        ECDSAKeep keep = new ECDSAKeep(
            _owner,
            _members,
            _honestThreshold
        );
        keeps.push(keep);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress, _members);
    }


    // TODO: Selection of ECDSA Keep members will be rewritten.

    /// @notice Runs member selection for an ECDSA keep.
    /// @dev Stub implementations generates a group with only one member. Member
    /// is randomly selected from registered member candidates.
    /// @param _groupSize Number of members to be selected.
    /// @return List of selected members addresses.
    function selectECDSAKeepMembers(
        uint256 _groupSize
    ) internal view returns (address payable[] memory members){
        require(memberCandidates.length > 0, 'keep member candidates list is empty');

        _groupSize;

        members = new address payable[](1);

        // TODO: Use the random beacon for randomness.
        uint memberIndex = uint256(keccak256(abi.encodePacked(block.timestamp)))
            % memberCandidates.length;

        members[0] = memberCandidates[memberIndex];
    }
}
