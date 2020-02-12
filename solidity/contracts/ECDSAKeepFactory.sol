pragma solidity ^0.5.4;

import "./ECDSAKeep.sol";
import "./KeepBonding.sol";
import "./api/IBondedECDSAKeepFactory.sol";
import "./utils/AddressPayableArrayUtils.sol";

import "@keep-network/sortition-pools/contracts/BondedSortitionPool.sol";
import "@keep-network/sortition-pools/contracts/BondedSortitionPoolFactory.sol";
import "@keep-network/sortition-pools/contracts/api/IStaking.sol";
import "@keep-network/sortition-pools/contracts/api/IBonding.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";

// TODO: This is an interface which is expected to be defined in keep-core and imported
// to use for the integration.
interface IRandomBeacon {
    function entryFeeEstimate(uint256 callbackGas)
        external
        view
        returns (uint256);

    function requestRelayEntry(
        address callbackContract,
        string calldata callbackMethod,
        uint256 callbackGas
    ) external payable returns (uint256);
}

/// @title ECDSA Keep Factory
/// @notice Contract creating bonded ECDSA keeps.
contract ECDSAKeepFactory is
    IBondedECDSAKeepFactory // TODO: Rename to BondedECDSAKeepFactory
{
    using AddressPayableArrayUtils for address payable[];
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
    uint256 groupSelectionSeed;

    BondedSortitionPoolFactory sortitionPoolFactory;
    address tokenStaking;
    KeepBonding keepBonding;
    IRandomBeacon randomBeacon;

    uint256 minimumStake = 1; // TODO: Take from setter
    uint256 minimumBond = 1; // TODO: Take from setter

    // Gas required for a callback from the random beacon. The value specifies
    // gas required to call `setGroupSelectionSeed` function in the worst-case
    // scenario with all the checks and maximum allowed uint256 relay entry as
    // a callback parameter.
    uint256 callbackGas = 41830;

    constructor(
        address _sortitionPoolFactory,
        address _tokenStaking,
        address _keepBonding,
        address _randomBeacon
    ) public {
        sortitionPoolFactory = BondedSortitionPoolFactory(
            _sortitionPoolFactory
        );
        tokenStaking = _tokenStaking;
        keepBonding = KeepBonding(_keepBonding);
        randomBeacon = IRandomBeacon(_randomBeacon);
    }

    // Fallback function to receive ether from the beacon.
    // TODO: Implement proper surplus handling.
    function() external payable {}

    /// @notice Register caller as a candidate to be selected as keep member
    /// for the provided customer application.
    /// @dev If caller is already registered it returns without any changes.
    function registerMemberCandidate(address _application) external {
        if (candidatesPools[_application] == address(0)) {
            // This is the first time someone registers as signer for this
            // application so let's create a signer pool for it.
            candidatesPools[_application] = sortitionPoolFactory
                .createSortitionPool(
                IStaking(tokenStaking),
                IBonding(address(keepBonding)),
                minimumStake,
                minimumBond
            );
        }
        BondedSortitionPool candidatesPool = BondedSortitionPool(
            candidatesPools[_application]
        );

        address operator = msg.sender;
        if (!candidatesPool.isOperatorInPool(operator)) {
            candidatesPool.joinPool(operator);
        }
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
        address application = msg.sender;
        address pool = candidatesPools[application];
        require(pool != address(0), "No signer pool for this application");

        // TODO: The remainder will not be bonded. What should we do with it?
        uint256 memberBond = _bond.div(_groupSize);
        require(memberBond > 0, "Bond per member must be greater than zero");

        require(
            msg.value >= openKeepFeeEstimate(),
            "Insufficient payment for opening a new keep"
        );

        address[] memory selected = BondedSortitionPool(pool).selectSetGroup(
            _groupSize,
            bytes32(groupSelectionSeed),
            memberBond
        );

        newGroupSelectionSeed();

        address payable[] memory members = new address payable[](_groupSize);
        for (uint256 i = 0; i < _groupSize; i++) {
            // TODO: Modify ECDSAKeep to not keep members as payable and do the
            // required casting in distributeERC20ToMembers and distributeETHToMembers.
            members[i] = address(uint160(selected[i]));
        }

        ECDSAKeep keep = new ECDSAKeep(_owner, members, _honestThreshold);

        keepAddress = address(keep);

        for (uint256 i = 0; i < _groupSize; i++) {
            keepBonding.createBond(
                members[i],
                keepAddress,
                uint256(keepAddress),
                memberBond
            );
        }

        emit ECDSAKeepCreated(keepAddress, members, _owner, application);
    }

    /// @notice Updates group selection seed.
    /// @dev The main goal of this function is to request the random beacon to
    /// generate a new random number. The beacon generates the number asynchronously
    /// and will call a callback function when the number is ready. In the meantime
    /// we update current group selection seed to a new value using a hash function.
    /// In case of the random beacon request failure this function won't revert
    /// but .....// TODO: Update when we decide what to do
    function newGroupSelectionSeed() internal {
        // Calculate new group selection seed based on the current seed.
        // We added address of the factory as a key to calculate value different
        // than sortition pool RNG will, so we don't end up selecting almost
        // identical group.
        groupSelectionSeed = uint256(
            keccak256(abi.encodePacked(groupSelectionSeed, address(this)))
        );

        // Call the random beacon to get a random group selection seed.
        // TODO: Replace with try/catch after we upgrade to solidity >= 0.6.0
        (bool success, bytes memory returnData) = address(randomBeacon)
            .call
            .value(msg.value)(
            abi.encodeWithSelector(
                randomBeacon.requestRelayEntry.selector,
                address(this),
                "setGroupSelectionSeed(uint256)",
                callbackGas
            )
        );
        if (!success) {
            // TODO: What should we do in case of `requestRelayEntry` failure?
            // Forward `msg.value` to the keep members?
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
