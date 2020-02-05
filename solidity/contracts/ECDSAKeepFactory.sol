pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./api/IBondedECDSAKeepFactory.sol";
import "./external/ITokenStaking.sol";
import "./utils/AddressArrayUtils.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "@keep-network/sortition-pools/contracts/SortitionPool.sol";
import "@keep-network/sortition-pools/contracts/SortitionPoolFactory.sol";

/// @title ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
contract ECDSAKeepFactory is
    IBondedECDSAKeepFactory // TODO: Rename to BondedECDSAKeepFactory
{
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

    uint256 feeEstimate;
    bytes32 groupSelectionSeed;

    SortitionPoolFactory sortitionPoolFactory;
    ITokenStaking tokenStaking;

    constructor(address _sortitionPoolFactory, address _tokenStaking) public {
        sortitionPoolFactory = SortitionPoolFactory(_sortitionPoolFactory);
        tokenStaking = ITokenStaking(_tokenStaking);
    }

    /// @notice Register caller as a candidate to be selected as keep member
    /// for the provided customer application
    /// @dev If caller is already registered it returns without any changes.
    /// @param _application Customer application address.
    function registerMemberCandidate(address _application) external {
        if (candidatesPools[_application] == address(0)) {
            // This is the first time someone registers as signer for this
            // application so let's create a signer pool for it.
            candidatesPools[_application] = sortitionPoolFactory
                .createSortitionPool();
        }

        SortitionPool candidatesPool = SortitionPool(
            candidatesPools[_application]
        );

        address operator = msg.sender;
        if (!candidatesPool.isOperatorRegistered(operator)) {
            candidatesPool.insertOperator(operator, eligibleStake(operator));
        }
    }

    /// @notice Checks if operator is registered as a candidate for the given
    /// customer application.
    /// @param _operator Operator's address.
    /// @param _application Customer application address.
    /// @return True if operator is already registered in the candidates pool,
    /// else false.
    function isOperatorRegistered(address _operator, address _application)
        public
        view
        returns (bool)
    {
        if (candidatesPools[_application] == address(0)) {
            return false;
        }

        SortitionPool candidatesPool = SortitionPool(
            candidatesPools[_application]
        );

        return candidatesPool.isOperatorRegistered(_operator);
    }

    /// @notice Gets the eligible stake balance of the operator.
    /// @dev Calls Token Staking contract to get eligible stake of the operator
    /// for this contract.
    /// @param _operator Operator's address.
    /// @return Eligible stake balance.
    function eligibleStake(address _operator) public view returns (uint256) {
        return tokenStaking.eligibleStake(_operator, address(this));
    }

    /// @notice Gets a fee estimate for opening a new keep.
    /// @return Uint256 estimate.
    function openKeepFeeEstimate() external returns (uint256) {
        return feeEstimate;
    }

    /// @notice Open a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// A caller of this function is expected to be an application for which
    /// member candidates were registered in a pool.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @param _bond Value of ETH bond required from the keep.
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond
    ) external payable returns (address keepAddress) {
        _bond; // TODO: assign bond for created keep

        address application = msg.sender;
        address pool = candidatesPools[application];
        require(pool != address(0), "No signer pool for this application");

        address[] memory selected = SortitionPool(pool).selectSetGroup(
            _groupSize,
            groupSelectionSeed
        );

        address payable[] memory members = new address payable[](_groupSize);
        for (uint256 i = 0; i < _groupSize; i++) {
            // TODO: for each selected member, validate staking weight and create,
            // bond. If validation failed or bond could not be created, remove
            // operator from pool and try again.
            members[i] = address(uint160(selected[i]));
        }

        ECDSAKeep keep = new ECDSAKeep(_owner, members, _honestThreshold);

        keepAddress = address(keep);

        emit ECDSAKeepCreated(keepAddress, members, _owner, application);

        // TODO: as beacon for new entry and update groupSelectionSeed in callback

    }
}
