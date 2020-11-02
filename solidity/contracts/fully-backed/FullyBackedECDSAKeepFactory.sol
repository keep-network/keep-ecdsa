/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓         ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "./FullyBackedECDSAKeep.sol";

import "./FullyBackedBonding.sol";
import "../api/IBondedECDSAKeepFactory.sol";
import "../KeepCreator.sol";
import "../GroupSelectionSeed.sol";
import "../CandidatesPools.sol";

import {
    AuthorityDelegator
} from "@keep-network/keep-core/contracts/Authorizations.sol";

import "@keep-network/sortition-pools/contracts/api/IFullyBackedBonding.sol";
import "@keep-network/sortition-pools/contracts/FullyBackedSortitionPoolFactory.sol";
import "@keep-network/sortition-pools/contracts/FullyBackedSortitionPool.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Fully Backed Bonded ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps that are fully backed by ETH.
/// @dev We avoid redeployment of bonded ECDSA keep contract by using the clone factory.
/// Proxy delegates calls to sortition pool and therefore does not affect contract's
/// state. This means that we only need to deploy the bonded ECDSA keep contract
/// once. The factory provides clean state for every new bonded ECDSA keep clone.
contract FullyBackedECDSAKeepFactory is
    IBondedECDSAKeepFactory,
    KeepCreator,
    AuthorityDelegator,
    GroupSelectionSeed,
    CandidatesPools
{
    FullyBackedSortitionPoolFactory sortitionPoolFactory;
    FullyBackedBonding bonding;

    using SafeMath for uint256;

    // Sortition pool is created with a minimum bond of 20 ETH to avoid
    // small operators joining and griefing future selections before the
    // minimum bond is set to the right value by the application.
    //
    // Anyone can create a sortition pool for an application with the default
    // minimum bond value but the application can change this value later, at
    // any point.
    //
    // The minimum bond value is a boundary value for an operator to remain
    // in the sortition pool. Once operator's unbonded value drops below the
    // minimum bond value the operator is removed from the sortition pool.
    // Operator can top-up the unbonded value deposited in bonding contract
    // and re-register to the sortition pool.
    //
    // This property is configured along with `MINIMUM_DELEGATION_DEPOSIT` defined
    // in `FullyBackedBonding` contract. Minimum delegation deposit determines
    // a minimum value of ether that should be transferred to the bonding contract
    // by an owner when delegating to an operator.
    uint256 public constant defaultMinimumBond = 20 ether;

    // Signer candidates in bonded sortition pool are weighted by their eligible
    // stake divided by a constant divisor. The divisor is set to 1 ETH so that
    // all ETHs in available unbonded value matter when calculating operator's
    // eligible weight for signer selection.
    uint256 public constant bondWeightDivisor = 1 ether;

    // Maps a keep to an application for which the keep was created.
    // keep address -> application address
    mapping(address => address) keepApplication;

    // Notification that a new keep has been created.
    event FullyBackedECDSAKeepCreated(
        address indexed keepAddress,
        address[] members,
        address indexed owner,
        address indexed application,
        uint256 honestThreshold
    );

    // Notification when an operator gets banned in a sortition pool for
    // an application.
    event OperatorBanned(address indexed operator, address indexed application);

    constructor(
        address _masterKeepAddress,
        address _sortitionPoolFactoryAddress,
        address _bondingAddress,
        address _randomBeaconAddress
    )
        public
        KeepCreator(_masterKeepAddress)
        GroupSelectionSeed(_randomBeaconAddress)
    {
        sortitionPoolFactory = FullyBackedSortitionPoolFactory(
            _sortitionPoolFactoryAddress
        );

        bonding = FullyBackedBonding(_bondingAddress);
    }

    /// @notice Sets the minimum bondable value required from the operator to
    /// join the sortition pool of the given application. It is up to the
    /// application to specify a reasonable minimum bond for operators trying to
    /// join the pool to prevent griefing by operators joining without enough
    /// bondable value.
    /// @param _minimumBondableValue The minimum bond value the application
    /// requires from a single keep.
    /// @param _groupSize Number of signers in the keep.
    /// @param _honestThreshold Minimum number of honest keep signers.
    function setMinimumBondableValue(
        uint256 _minimumBondableValue,
        uint256 _groupSize,
        uint256 _honestThreshold
    ) external {
        uint256 memberBond = bondPerMember(_minimumBondableValue, _groupSize);
        FullyBackedSortitionPool(getSortitionPool(msg.sender))
            .setMinimumBondableValue(memberBond);
    }

    /// @notice Opens a new ECDSA keep.
    /// @dev Selects a list of signers for the keep based on provided parameters.
    /// A caller of this function is expected to be an application for which
    /// member candidates were registered in a pool.
    /// @param _groupSize Number of signers in the keep.
    /// @param _honestThreshold Minimum number of honest keep signers.
    /// @param _owner Address of the keep owner.
    /// @param _bond Value of ETH bond required from the keep in wei.
    /// @param _stakeLockDuration Stake lock duration in seconds. Ignored by
    /// this implementation.
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond,
        uint256 _stakeLockDuration
    ) external payable nonReentrant returns (address keepAddress) {
        require(_groupSize > 0, "Minimum signing group size is 1");
        require(_groupSize <= 16, "Maximum signing group size is 16");
        require(
            _honestThreshold > 0,
            "Honest threshold must be greater than 0"
        );
        require(
            _honestThreshold <= _groupSize,
            "Honest threshold must be less or equal the group size"
        );

        address application = msg.sender;
        address pool = getSortitionPool(application);

        uint256 memberBond = bondPerMember(_bond, _groupSize);
        require(memberBond > 0, "Bond per member must be greater than zero");

        require(
            msg.value >= openKeepFeeEstimate(),
            "Insufficient payment for opening a new keep"
        );

        address[] memory members = FullyBackedSortitionPool(pool)
            .selectSetGroup(
            _groupSize,
            bytes32(groupSelectionSeed),
            memberBond
        );

        newGroupSelectionSeed();

        // createKeep sets keepOpenedTimestamp value for newly created keep which
        // is required to be set before calling `keep.initialize` function as it
        // is used to determine token staking delegation authority recognition
        // in `__isRecognized` function.
        keepAddress = createKeep();

        FullyBackedECDSAKeep(keepAddress).initialize(
            _owner,
            members,
            _honestThreshold,
            address(bonding),
            address(this)
        );

        for (uint256 i = 0; i < _groupSize; i++) {
            bonding.createBond(
                members[i],
                keepAddress,
                uint256(keepAddress),
                memberBond,
                pool
            );
        }

        keepApplication[keepAddress] = application;

        emit FullyBackedECDSAKeepCreated(
            keepAddress,
            members,
            _owner,
            application,
            _honestThreshold
        );
    }

    /// @notice Verifies if delegates authority recipient is valid address recognized
    /// by the factory for token staking authority delegation.
    /// @param _delegatedAuthorityRecipient Address of the delegated authority
    /// recipient.
    /// @return True if provided address is recognized delegated token staking
    /// authority for this factory contract.
    function __isRecognized(address _delegatedAuthorityRecipient)
        external
        returns (bool)
    {
        return keepOpenedTimestamp[_delegatedAuthorityRecipient] > 0;
    }

    /// @notice Gets a fee estimate for opening a new keep.
    /// @return Uint256 estimate.
    function openKeepFeeEstimate() public view returns (uint256) {
        return newEntryFeeEstimate();
    }

    /// @notice Checks if the factory has the authorization to operate on stake
    /// represented by the provided operator.
    ///
    /// @param _operator operator's address
    /// @return True if the factory has access to the staked token balance of
    /// the provided operator and can slash that stake. False otherwise.
    function isOperatorAuthorized(address _operator)
        public
        view
        returns (bool)
    {
        return bonding.isAuthorizedForOperator(_operator, address(this));
    }

    /// @notice Bans members of a calling keep in a sortition pool associated
    /// with the application for which the keep was created.
    /// @dev The function can be called only by a keep created by this factory.
    function banKeepMembers() public onlyKeep() {
        FullyBackedECDSAKeep keep = FullyBackedECDSAKeep(msg.sender);

        address[] memory members = keep.getMembers();

        address application = keepApplication[address(keep)];

        FullyBackedSortitionPool sortitionPool = FullyBackedSortitionPool(
            getSortitionPool(application)
        );

        for (uint256 i = 0; i < members.length; i++) {
            address operator = members[i];

            sortitionPool.ban(operator);

            emit OperatorBanned(operator, application);
        }
    }

    function newSortitionPool(address _application) internal returns (address) {
        return
            sortitionPoolFactory.createSortitionPool(
                IFullyBackedBonding(address(bonding)),
                defaultMinimumBond,
                bondWeightDivisor
            );
    }

    /// @notice Calculates bond requirement per member performing the necessary
    /// rounding.
    /// @param _keepBond The bond required from a keep.
    /// @param _groupSize Number of signers in the keep.
    /// @return Bond value required from each keep member.
    function bondPerMember(uint256 _keepBond, uint256 _groupSize)
        internal
        pure
        returns (uint256)
    {
        // In Solidity, division rounds towards zero (down) and dividing
        // '_bond' by '_groupSize' can leave a remainder. Even though, a remainder
        // is very small, we want to avoid this from happening and memberBond is
        // rounded up by: `(bond + groupSize - 1 ) / groupSize`
        // Ex. (100 + 3 - 1) / 3 = 34
        return (_keepBond.add(_groupSize).sub(1)).div(_groupSize);
    }

    /// @notice Checks if caller is a keep created by this factory.
    modifier onlyKeep() {
        require(
            keepOpenedTimestamp[msg.sender] > 0,
            "Caller is not a keep created by the factory"
        );
        _;
    }
}
