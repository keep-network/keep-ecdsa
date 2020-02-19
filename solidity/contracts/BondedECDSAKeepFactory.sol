pragma solidity ^0.5.4;

import "./BondedECDSAKeep.sol";
import "./KeepBonding.sol";
import "./api/IBondedECDSAKeepFactory.sol";
import "./CloneFactory.sol";

import "@keep-network/sortition-pools/contracts/api/IStaking.sol";
import "@keep-network/sortition-pools/contracts/api/IBonding.sol";
import "@keep-network/sortition-pools/contracts/BondedSortitionPool.sol";
import "@keep-network/sortition-pools/contracts/BondedSortitionPoolFactory.sol";

import "@keep-network/keep-core/contracts/IRandomBeacon.sol";
import "@keep-network/keep-core/contracts/utils/AddressArrayUtils.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

/// @title Bonded ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
/// @dev We avoid redeployment of bonded ECDSA keep contract by using the clone factory.
/// Proxy delegates calls to sortition pool and therefore does not affect contract's
/// state. This means that we only need to deploy the bonded ECDSA keep contract
/// once. The factory provides clean state for every new bonded ECDSA keep clone.
contract BondedECDSAKeepFactory is IBondedECDSAKeepFactory, CloneFactory {
    using AddressArrayUtils for address[];
    using SafeMath for uint256;

    // Notification that a new sortition pool has been created.
    event SortitionPoolCreated(address application, address sortitionPool);

    // Notification that a new keep has been created.
    event BondedECDSAKeepCreated(
        address keepAddress,
        address[] members,
        address owner,
        address application
    );

    // Holds the address of the bonded ECDSA keep contract that will be used as a
    // master contract for cloning.
    address public masterBondedECDSAKeepAddress;

    // Mapping of pools with registered member candidates for each application.
    mapping(address => address) candidatesPools; // application -> candidates pool

    uint256 feeEstimate;
    uint256 public groupSelectionSeed;

    BondedSortitionPoolFactory sortitionPoolFactory;
    address tokenStaking;
    KeepBonding keepBonding;
    IRandomBeacon randomBeacon;

    uint256 public minimumStake = 200000 * 1e18;

    // Sortition pool is created with a minimum bond of 1 to avoid
    // griefing.
    //
    // Anyone can create a sortition pool for an application. If a pool is
    // created with a ridiculously high bond, nobody can join it and
    // updating bond is not possible because trying to select a group
    // with an empty pool reverts.
    //
    // We set the minimum bond value to 1 to prevent from this situation and
    // to allow the pool adjust the minimum bond during the first signer
    // selection.
    uint256 public constant minimumBond = 1;

    // Gas required for a callback from the random beacon. The value specifies
    // gas required to call `setGroupSelectionSeed` function in the worst-case
    // scenario with all the checks and maximum allowed uint256 relay entry as
    // a callback parameter.
    uint256 callbackGas = 41830;

    // Random beacon sends back callback surplus to the requestor. It may also
    // decide to send additional request subsidy fee. What's more, it may happen
    // that the beacon is busy and we'll not refresh group selection seed from
    // the beacon. We accumulate all funds received from the beacon in the
    // subsidy pool and later distribute funds from this pull to selected
    // signers.
    uint256 public subsidyPool;

    constructor(
        address _masterBondedECDSAKeepAddress,
        address _sortitionPoolFactory,
        address _tokenStaking,
        address _keepBonding,
        address _randomBeacon
    ) public {
        masterBondedECDSAKeepAddress = _masterBondedECDSAKeepAddress;
        sortitionPoolFactory = BondedSortitionPoolFactory(
            _sortitionPoolFactory
        );
        tokenStaking = _tokenStaking;
        keepBonding = KeepBonding(_keepBonding);
        randomBeacon = IRandomBeacon(_randomBeacon);
    }

    /// @notice Adds any received funds to the factory subsidy pool.
    function() external payable {
        subsidyPool += msg.value;
    }

    /// @notice Creates new sortition pool for the application.
    /// @dev Emits an event after sortition pool creation.
    /// @param _application Address of the application.
    /// @return Address of the created sortition pool contract.
    function createSortitionPool(address _application)
        external
        returns (address)
    {
        require(
            candidatesPools[_application] == address(0),
            "Sortition pool already exists"
        );

        address sortitionPoolAddress = sortitionPoolFactory.createSortitionPool(
            IStaking(tokenStaking),
            IBonding(address(keepBonding)),
            minimumStake,
            minimumBond
        );

        candidatesPools[_application] = sortitionPoolAddress;

        emit SortitionPoolCreated(_application, sortitionPoolAddress);

        return candidatesPools[_application];
    }

    /// @notice Gets the sortition pool address for the given application.
    /// @dev Reverts if sortition does not exits for the application.
    /// @param _application Address of the application.
    /// @return Address of the sortition pool contract.
    function getSortitionPool(address _application)
        external
        view
        returns (address)
    {
        require(
            candidatesPools[_application] != address(0),
            "No pool found for the application"
        );

        return candidatesPools[_application];
    }

    /// @notice Register caller as a candidate to be selected as keep member
    /// for the provided customer application.
    /// @dev If caller is already registered it returns without any changes.
    /// @param _application Address of the application.
    function registerMemberCandidate(address _application) external {
        require(
            candidatesPools[_application] != address(0),
            "No pool found for the application"
        );

        BondedSortitionPool candidatesPool = BondedSortitionPool(
            candidatesPools[_application]
        );

        address operator = msg.sender;
        if (!candidatesPool.isOperatorInPool(operator)) {
            candidatesPool.joinPool(operator);
        }
    }

    /// @notice Checks if operator is registered as a candidate for the given
    /// customer application.
    /// @param _operator Operator's address.
    /// @param _application Customer application address.
    /// @return True if operator is already registered in the candidates pool,
    /// false otherwise.
    function isOperatorRegistered(address _operator, address _application)
        public
        view
        returns (bool)
    {
        if (candidatesPools[_application] == address(0)) {
            return false;
        }

        BondedSortitionPool candidatesPool = BondedSortitionPool(
            candidatesPools[_application]
        );

        return candidatesPool.isOperatorRegistered(_operator);
    }

    /// @notice Checks if operator's details in the member candidates pool are
    /// up to date for the given application. If not update operator status
    /// function should be called by the one who is monitoring the status.
    /// @param _operator Operator's address.
    /// @param _application Customer application address.
    function isOperatorUpToDate(address _operator, address _application)
        external
        view
        returns (bool)
    {
        BondedSortitionPool candidatesPool = getSortitionPoolForOperator(
            _operator,
            _application
        );

        return candidatesPool.isOperatorUpToDate(_operator);
    }

    /// @notice Invokes update of operator's details in the member candidates pool
    /// for the given application
    /// @param _operator Operator's address.
    /// @param _application Customer application address.
    function updateOperatorStatus(address _operator, address _application)
        external
    {
        BondedSortitionPool candidatesPool = getSortitionPoolForOperator(
            _operator,
            _application
        );

        candidatesPool.updateOperatorStatus(_operator);
    }

    /// @notice Gets bonded sortition pool of specific application for the
    /// operator.
    /// @dev Reverts if the operator is not registered for the application.
    /// @param _operator Operator's address.
    /// @param _application Customer application address.
    /// @return Bonded sortition pool.
    function getSortitionPoolForOperator(
        address _operator,
        address _application
    ) internal view returns (BondedSortitionPool) {
        require(
            isOperatorRegistered(_operator, _application),
            "Operator not registered for the application"
        );

        return BondedSortitionPool(candidatesPools[_application]);
    }

    /// @notice Gets a fee estimate for opening a new keep.
    /// @return Uint256 estimate.
    function openKeepFeeEstimate() public view returns (uint256) {
        return randomBeacon.entryFeeEstimate(callbackGas);
    }

    /// @notice Open a new ECDSA keep.
    /// @dev Selects a list of members for the keep based on provided parameters.
    /// A caller of this function is expected to be an application for which
    /// member candidates were registered in a pool.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @param _bond Value of ETH bond required from the keep (wei).
    /// @return Created keep address.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond
    ) external payable returns (address keepAddress) {
        require(_groupSize <= 16, "Maximum signing group size is 16");

        address application = msg.sender;
        address pool = candidatesPools[application];
        require(pool != address(0), "No signer pool for this application");

        // In Solidity, division rounds towards zero (down) and dividing
        // '_bond' by '_groupSize' can leave a remainder. Even though, a remainder
        // is very small, we want to avoid this from happening and memberBond is
        // rounded up by: `(bond + groupSize - 1 ) / groupSize`
        // Ex. (100 + 3 - 1) / 3 = 34
        uint256 memberBond = (_bond.add(_groupSize).sub(1)).div(_groupSize);
        require(memberBond > 0, "Bond per member must be greater than zero");

        require(
            msg.value >= openKeepFeeEstimate(),
            "Insufficient payment for opening a new keep"
        );

        address[] memory members = BondedSortitionPool(pool).selectSetGroup(
            _groupSize,
            bytes32(groupSelectionSeed),
            memberBond
        );

        newGroupSelectionSeed();

        keepAddress = createClone(masterBondedECDSAKeepAddress);
        BondedECDSAKeep keep = BondedECDSAKeep(keepAddress);
        keep.initialize(
            _owner,
            members,
            _honestThreshold,
            tokenStaking,
            address(keepBonding)
        );

        for (uint256 i = 0; i < _groupSize; i++) {
            keepBonding.createBond(
                members[i],
                keepAddress,
                uint256(keepAddress),
                memberBond,
                pool
            );
        }

        // If subsidy pool is non-empty, distribute the value to signers but
        // never distribute more than the payment for opening a keep.
        uint256 signerSubsidy = subsidyPool < msg.value
            ? subsidyPool
            : msg.value;
        if (signerSubsidy > 0) {
            subsidyPool -= signerSubsidy;
            keep.distributeETHToMembers.value(signerSubsidy)();
        }

        emit BondedECDSAKeepCreated(keepAddress, members, _owner, application);
    }

    /// @notice Updates group selection seed.
    /// @dev The main goal of this function is to request the random beacon to
    /// generate a new random number. The beacon generates the number asynchronously
    /// and will call a callback function when the number is ready. In the meantime
    /// we update current group selection seed to a new value using a hash function.
    /// In case of the random beacon request failure this function won't revert
    /// but add beacon payment to factory's subsidy pool.
    function newGroupSelectionSeed() internal {
        // Calculate new group selection seed based on the current seed.
        // We added address of the factory as a key to calculate value different
        // than sortition pool RNG will, so we don't end up selecting almost
        // identical group.
        groupSelectionSeed = uint256(
            keccak256(abi.encodePacked(groupSelectionSeed, address(this)))
        );

        // Call the random beacon to get a random group selection seed.
        //
        // Limiting forwarded gas to prevent malicious behavior in case the
        // beacon service contract gets compromised. Relay request should not
        // consume more than 360k of gas. We set the limit to 400k to have
        // a safety margin for future updates.
        (bool success, ) = address(randomBeacon).call.gas(400000).value(msg.value)(
            abi.encodeWithSignature(
                "requestRelayEntry(address,string,uint256)",
                address(this),
                "setGroupSelectionSeed(uint256)",
                callbackGas
            )
        );
        if (!success) {
            subsidyPool += msg.value; // beacon is busy
        }
    }

    /// @notice Sets a new group selection seed value.
    /// @dev The function is expected to be called in a callback by the random
    /// beacon.
    /// @param _groupSelectionSeed New value of group selection seed.
    function setGroupSelectionSeed(uint256 _groupSelectionSeed)
        external
        onlyRandomBeacon
    {
        groupSelectionSeed = _groupSelectionSeed;
    }

    /// @notice Checks if the caller is the random beacon.
    /// @dev Throws an error if called by any account other than the random beacon.
    modifier onlyRandomBeacon() {
        require(
            address(randomBeacon) == msg.sender,
            "Caller is not the random beacon"
        );
        _;
    }
}
