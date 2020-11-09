/**
▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓▌        ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

                           Trust math, not hardware.
*/

pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/Rewards.sol";
import "@keep-network/keep-core/contracts/libraries/BinarySearch.sol";
import "./BondedECDSAKeepFactory.sol";
import "./BondedECDSAKeep.sol";

/// @title KEEP ECDSA Signer Subsidy Rewards for the Sep 2020 release.
/// @notice Contract distributes KEEP rewards to signers that were part of
/// the keeps which were created by the BondedECDSAKeepFactory contract.
///
/// The amount of KEEP to be distributed is determined by funding the contract,
/// and additional KEEP can be added at any time.
///
/// When an interval is over, it will be allocated a percentage of the remaining
/// unallocated rewards based on its weight, and adjusted by the number of keeps
/// created in the interval if the quota is not met.
///
/// The adjustment for not meeting the keep quota is a percentage that equals
/// the percentage of the quota that was met; if the number of keeps created is
/// 80% of the quota then 80% of the base reward will be allocated for the
/// interval.
///
/// Any unallocated rewards will stay in the unallocated rewards pool,
/// to be allocated for future intervals. Intervals past the initially defined
/// schedule have a weight of the last scheduled interval.
///
/// Keeps can receive rewards once the interval they were created in is over,
/// and the keep has been marked as closed.
/// There is no time limit to receiving rewards, nor is there need to wait for
/// all keeps from the interval to be marked as closed.
/// Calling `receiveReward` automatically allocates the rewards for the interval
/// the specified keep was created in and all previous intervals.
///
/// If a keep is terminated, that fact can be reported to the reward contract.
/// Reporting a terminated keep returns its allocated reward to the pool of
/// unallocated rewards.
contract DoubleFactoryECDSARewards is RewardPayouts {
    // The amount of tokens each individual beneficiary address
    // can receive in a single interval is capped to 3M tokens.
    uint256 public beneficiaryRewardCap = 3000000 * 10**18;

    // BondedECDSAKeepFactory deployment date, Sep-14-2020 interval started.
    // https://etherscan.io/address/0xA7d9E842EFB252389d613dA88EDa3731512e40bD
    uint256 internal constant ecdsaFirstIntervalStart = 1600041600;

    /// Weights of the 24 reward intervals assigned over
    // 24 * termLength days.
    uint256[] internal intervalWeights = [
        4,
        8,
        10,
        12,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15,
        15
    ];

    // Each interval is 30 days long.
    uint256 internal constant termLength = 30 days;

    uint256 internal constant minimumECDSAKeepsPerInterval = 1000;

    // The total amount of rewards allocated to the given beneficiary address,
    // in the given interval.
    // `allocatedRewards[beneficiary][interval] -> amount`
    mapping(address => mapping(uint256 => uint256)) internal allocatedRewards;
    // The amount of interval rewards withdrawn to the given beneficiary.
    mapping(address => mapping(uint256 => uint256)) internal withdrawnRewards;

    BondedECDSAKeepFactory internal factoryA;
    BondedECDSAKeepFactory internal factoryB;
    TokenStaking internal tokenStaking;

    constructor(
        address _token,
        address payable _factoryAddressA,
        address payable _factoryAddressB,
        address _tokenStakingAddress
    )
        public
        RewardPayouts(
            _token,
            ecdsaFirstIntervalStart,
            intervalWeights,
            termLength,
            minimumECDSAKeepsPerInterval
        )
    {
        factoryA = BondedECDSAKeepFactory(_factoryAddressA);
        factoryB = BondedECDSAKeepFactory(_factoryAddressB);
        tokenStaking = TokenStaking(_tokenStakingAddress);
    }

    /// @notice Get the amount of rewards allocated
    /// for the specified operator's beneficiary in the specified interval.
    /// @param interval The interval
    /// @param operator The operator
    /// @return The amount allocated
    function getAllocatedRewards(uint256 interval, address operator)
        external
        view
        returns (uint256)
    {
        address beneficiary = tokenStaking.beneficiaryOf(operator);
        return allocatedRewards[beneficiary][interval];
    }

    /// @notice Get the amount of rewards already withdrawn
    /// for the specified operator's beneficiary in the specified interval.
    /// @param interval The interval
    /// @param operator The operator
    /// @return The amount already withdrawn
    function getWithdrawnRewards(uint256 interval, address operator)
        external
        view
        returns (uint256)
    {
        address beneficiary = tokenStaking.beneficiaryOf(operator);
        return withdrawnRewards[beneficiary][interval];
    }

    /// @notice Get the amount of rewards withdrawable
    /// for the specified operator's beneficiary in the specified interval.
    /// @param interval The interval
    /// @param operator The operator
    /// @return The amount withdrawable
    function getWithdrawableRewards(uint256 interval, address operator)
        external
        view
        returns (uint256)
    {
        address beneficiary = tokenStaking.beneficiaryOf(operator);
        uint256 allocated = allocatedRewards[beneficiary][interval];
        uint256 withdrawn = withdrawnRewards[beneficiary][interval];
        return allocated.sub(withdrawn);
    }

    /// @notice Withdraw all available rewards for the given interval.
    /// The rewards will be paid to the beneficiary of the specified operator.
    /// @param interval The interval
    /// @param operator The operator
    function withdrawRewards(uint256 interval, address operator) external {
        address beneficiary = tokenStaking.beneficiaryOf(operator);

        uint256 allocated = allocatedRewards[beneficiary][interval];
        uint256 alreadyWithdrawn = withdrawnRewards[beneficiary][interval];

        require(allocated > alreadyWithdrawn, "No rewards to withdraw");

        uint256 withdrawableRewards = allocated.sub(alreadyWithdrawn);

        withdrawnRewards[beneficiary][interval] = allocated;

        token.safeTransfer(beneficiary, withdrawableRewards);
    }

    function _getValueA(uint256 index) internal view returns (uint256) {
        address keepAddress = factoryA.getKeepAtIndex(index);
        return factoryA.getKeepOpenedTimestamp(keepAddress);
    }

    function _getValueB(uint256 index) internal view returns (uint256) {
        address keepAddress = factoryB.getKeepAtIndex(index);
        return factoryB.getKeepOpenedTimestamp(keepAddress);
    }

    function _findEndpoint(uint256 intervalEndpoint) internal view returns (uint256) {
        require(
            intervalEndpoint <= block.timestamp,
            "interval hasn't ended yet"
        );
        uint256 endpointA = BinarySearch.find(
            _getValueA,
            _getKeepCountA(),
            intervalEndpoint
        );
        uint256 endpointB = BinarySearch.find(
            _getValueB,
            _getKeepCountB(),
            intervalEndpoint
        );
        return endpointA.add(endpointB);
    }

    function _getKeepCountA() internal view returns (uint256) {
        return factoryA.getKeepCount();
    }

    function _getKeepCountB() internal view returns (uint256) {
        return factoryB.getKeepCount();
    }

    function _getKeepCount() internal view returns (uint256) {
        return _getKeepCountA().add(_getKeepCountB());
    }

    function _getKeepAtIndex(uint256 i) internal view returns (bytes32) {
        return fromAddress(factory.getKeepAtIndex(i));
    }

    function _getCreationTime(bytes32 _keep)
        internal
        view
        isAddress(_keep)
        returns (uint256)
    {
        uint256 timeAtA = factoryA.getKeepOpenedTimestamp(toAddress(_keep));
        uint256 timeAtB = factoryB.getKeepOpenedTimestamp(toAddress(_keep));
        return timeAtA != 0 ? timeAtA : timeAtB;
    }

    function _isClosed(bytes32 _keep)
        internal
        view
        isAddress(_keep)
        returns (bool)
    {
        return BondedECDSAKeep(toAddress(_keep)).isClosed();
    }

    function _isTerminated(bytes32 _keep)
        internal
        view
        isAddress(_keep)
        returns (bool)
    {
        return BondedECDSAKeep(toAddress(_keep)).isTerminated();
    }

    // A keep is recognized if it was opened by this factory.
    function _recognizedByFactory(bytes32 _keep)
        internal
        view
        isAddress(_keep)
        returns (bool)
    {
        bool recognizedByA = factoryA.getKeepOpenedTimestamp(toAddress(_keep)) != 0;
        bool recognizedByB = factoryB.getKeepOpenedTimestamp(toAddress(_keep)) != 0;
        return recognizedByA || recognizedByB;
    }

    /// @notice Get the members of the specified keep, and distribute the reward
    /// amount between them. The reward isn't paid out immediately,
    /// but is instead kept in the reward contract until each operator
    /// individually requests to withdraw the rewards.
    function _distributeReward(bytes32 _keep, uint256 amount)
        internal
        isAddress(_keep)
    {
        address[] memory members = BondedECDSAKeep(toAddress(_keep))
            .getMembers();
        uint256 interval = intervalOf(_getCreationTime(_keep));

        uint256 memberCount = members.length;
        uint256 dividend = amount.div(memberCount);
        uint256 remainder = amount.mod(memberCount);

        uint256[] memory allocations = new uint256[](memberCount);

        for (uint256 i = 0; i < memberCount - 1; i++) {
            allocations[i] = dividend;
        }
        allocations[memberCount - 1] = dividend.add(remainder);

        for (uint256 i = 0; i < memberCount; i++) {
            address beneficiary = tokenStaking.beneficiaryOf(members[i]);
            uint256 addedAllocation = allocations[i];
            uint256 prevAllocated = allocatedRewards[beneficiary][interval];
            uint256 newAllocation = prevAllocated.add(addedAllocation);
            if (newAllocation > beneficiaryRewardCap) {
                uint256 deallocatedAmount = newAllocation.sub(
                    beneficiaryRewardCap
                );
                newAllocation = beneficiaryRewardCap;
                deallocate(deallocatedAmount);
            }
            allocatedRewards[beneficiary][interval] = newAllocation;
        }
    }

    function toAddress(bytes32 keepBytes) internal pure returns (address) {
        return address(bytes20(keepBytes));
    }

    function fromAddress(address keepAddress) internal pure returns (bytes32) {
        return bytes32(bytes20(keepAddress));
    }

    function validAddressBytes(bytes32 keepBytes) internal pure returns (bool) {
        return fromAddress(toAddress(keepBytes)) == keepBytes;
    }

    modifier isAddress(bytes32 _keep) {
        require(validAddressBytes(_keep), "Invalid keep address");
        _;
    }
}
