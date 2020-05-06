pragma solidity 0.5.17;

import "./api/IBondedECDSAKeepFactory.sol";

import "./api/IBondedECDSAKeepSubFactory.sol";

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/math/Math.sol";

contract BondedECDSAKeepLoadBalancerFactory is
    IBondedECDSAKeepFactory
{
    using SafeMath for uint256;

    IBondedECDSAKeepSubFactory factoryA;
    IBondedECDSAKeepSubFactory factoryB;

    mapping(address => uint256) applicationRequestCounter;

    address public factoryUpgrader;

    constructor() public {
        factoryUpgrader = msg.sender;
    };

    function addFactory(address factoryAddress) public {
        require(
            msg.sender == factoryUpgrader,
            "Only callable by factory upgrader"
        );
        require(
            address(factoryB) == address(0),
            "Both factories already set"
        );
        if (address(factoryA) == address(0)) {
            factoryA = IBondedECDSAKeepSubFactory(factoryAddress);
        } else {
            factoryB = IBondedECDSAKeepSubFactory(factoryAddress);
        }
    }

    /// @notice Open a new ECDSA Keep.
    /// @param _groupSize Number of members in the keep.
    /// @param _honestThreshold Minimum number of honest keep members.
    /// @param _owner Address of the keep owner.
    /// @param _bond Value of ETH bond required from the keep.
    /// @param _stakeLockDuration Stake lock duration in seconds.
    /// @return Address of the opened keep.
    function openKeep(
        uint256 _groupSize,
        uint256 _honestThreshold,
        address _owner,
        uint256 _bond,
        uint256 _stakeLockDuration
    ) external payable returns (address keepAddress) {
        require(
            address(factoryA) != address(0),
            "At least one factory must be set"
        );

        address application = msg.sender;

        IBondedECDSAKeepSubFactory factory = _chooseFactory(application);

        return factory.indirectOpenKeep(
            application,
            _groupSize,
            _honestThreshold,
            _owner,
            _bond,
            _stakeLockDuration
        );
    }

    /// @notice Gets a fee estimate for opening a new keep.
    /// @return Uint256 estimate.
    function openKeepFeeEstimate() external view returns (uint256) {
        if (address(factoryB) == address(0)) {
            return factoryA.openKeepFeeEstimate();
        } else {
            return Math.max(
                factoryA.openKeepFeeEstimate(),
                factoryB.openKeepFeeEstimate()
            );
        }
    }

    function _chooseFactory(address application)
        internal returns (IBondedECDSAKeepSubFactory factory) {
        uint256 counter = applicationRequestCounter[application];
        applicationRequestCounter[application] = counter + 1;

        uint256 weightA = factoryA.getSortitionPoolWeight(application);
        uint256 weightB = factoryB.getSortitionPoolWeight(application);

        uint256 seed = uint256(
            keccak256(abi.encodePacked(application, counter))
        );

        if (_choiceFunction(seed, weightA, weightB)) {
            factory = factoryA;
        } else {
            factory = factoryB;
        }
    }

    function _choiceFunction(uint256 seed, uint256 weightA, uint256 weightB)
        internal returns (bool chooseA) {
        /// TODO
    }
}
