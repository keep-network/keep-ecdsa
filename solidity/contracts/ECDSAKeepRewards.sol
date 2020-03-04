pragma solidity ^0.5.4;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract ECDSAKeepRewards {

    IERC20 keepToken;
    IBondedECDSAKeepFactory factory;

    // Total number of keep tokens to distribute.
    uint256 totalRewards;
    // Length of one interval.
    uint256 termLength;
    // Timestamp of first interval beginning.
    uint256 initiated;
    // Minimum number of keep submissions for each interval.
    uint256 minimumSubmissions;
    // Array representing the percentage of total rewards available for each term.
    uint8[] InitialTermWeights = [8, 33, 21, 14, 9, 5, 3, 2, 2, 1, 1, 1]; // percent array

    // Total number of intervals.
    uint256 termCount = InitialTermWeights.length;

    // mapping of keeps to booleans. True if the keep has been used to calim a reward.
    mapping(address => bool) claimed;

    // Number of submissions for each interval.
    mapping(uint256 => uint256) intervalSubmissions;

    // Array of timestamps marking interval's end.
    uint256[] intervalEndpoints;

    constructor (
        uint256 _termLength,
        uint256 _totalRewards,
        address _keepToken,
        uint256 _minimumSubmissions,
        address factoryAddress
    )
    public {
       keepToken = IERC20(_keepToken);
       totalRewards = _totalRewards;
       termLength = _termLength;
       initiated = block.timestamp;
       minimumSubmissions = _minimumSubmissions;
       factory = IBondedECDSAKeepFactory(factoryAddress);
    }

    /// @notice Sends the reward for a keep to the keep owner.
    /// @param _keepAddress ECDSA keep factory address.
    function receiveReward(address _keepAddress) public {
        require(eligibleForReward(_keepAddress));
        require(!claimed[_keepAddress],"Reward already claimed.");
        claimed[_keepAddress] = true;

        IBondedECDSAKeep _keep =  IBondedECDSAKeep(_keepAddress);
        uint256 timestampOpened = _keep.getTimestamp();
        uint256 interval = findInterval(timestampOpened);
        uint256 intervalReward = termReward(interval);
        keepToken.transfer(_keep.getOwner(), intervalReward);
    }


    /// @notice Get the rewards interval a given timestamp falls unnder.
    /// @param _timestamp The timestamp to check.
    /// @return The associated interval.
    function findInterval(uint256 _timestamp) public returns (uint256){
        // provide index/rewards interval and validate on-chain?
        // if interval exists, return it. else updateInterval()
        return updateInterval(_timestamp);
    }

    /// @notice Get the reward dividend for each keep for a given reward interval.
    /// @param term The term to check.
    /// @return The reward dividend.
    function termReward(uint256 term) public view returns (uint256){
        uint256 _totalTermRewards = totalRewards * InitialTermWeights[term] / 100;
        return _totalTermRewards / intervalSubmissions[term];
    }

    /// @notice Updates the latest interval.
    /// @dev Interval should only be updated if the _timestamp provided
    ///      does not belong to a pre-existing interval.
    /// @param _timestamp The timestamp to update with.
    /// @return the new interval.
    function updateInterval(uint256 _timestamp) internal returns (uint256){
        require(
            block.timestamp - initiated >= termLength * intervalEndpoints.length + termLength,
            "not due for new interval"
        );
        uint256 intervalEndpointsLength = intervalEndpoints.length;
        uint256 newInterval = intervalEndpointsLength > 0 ?
        find(0, factory.getKeepCount(), _timestamp):
        find(intervalEndpoints[intervalEndpointsLength - 1], factory.getKeepCount(), _timestamp);

        uint256 totalSubmissions = intervalEndpointsLength > 0 ?
        newInterval:
        newInterval - intervalEndpoints[intervalEndpointsLength - 1];

        intervalSubmissions[intervalEndpointsLength] = totalSubmissions;
        if (totalSubmissions < minimumSubmissions){
            if(intervalEndpointsLength >= InitialTermWeights.length){
                return newInterval;
            }            
            InitialTermWeights[intervalEndpointsLength + 1] +=  InitialTermWeights[intervalEndpointsLength];
            InitialTermWeights[intervalEndpointsLength] = 0;
        }
        return newInterval;
    }

    /// @notice Checks if a keep is eligible to receive rewards.
    /// @dev Keeps that close dishonorably or early are not eligible for rewards.
    /// @param _keep The keep to check.
    /// @return True if the keep is eligible, false otherwise
    function eligibleForReward(address _keep) public view returns (bool){
        // check that keep closed properly
        return true;
    }

    /// @notice Find the index of the keep with the largest timestamp smaller than
    ///         _target from the `factory`
    /// @dev   This is a binary search.
    /// @param start Lower bound index for array traversal.
    /// @param end Upper bound index for array traversal.
    /// @param target Target timestamp to check against.
    /// @return The ndex of the keep with the largest timestamp smaller than _target
    function find(uint256 start, uint256 end, uint256 target) public view returns (uint256) {
        uint256 _len;
        uint256 _start = start;
        uint256 _end = end;
        uint256 _mid;
        uint256 timestamp;
        uint256 timestampNext;

        while (_start <= _end){
            _len = _end - _start;
            _mid = _start + _len / 2;
            timestamp = IBondedECDSAKeep(factory.getKeepAtIndex(_mid)).getTimestamp();
            timestampNext = IBondedECDSAKeep(factory.getKeepAtIndex(_mid + 1)).getTimestamp(); // check bound
            if(timestamp <= target && timestampNext > target){
                return _mid;
            }
            else if(timestamp > target){
                _end = _mid - 1;
            }
            else{
                _start = _mid + 1;
            }
        }
        revert("could not find target");
   }
}

interface IBondedECDSAKeep {
    function getOwner() external view returns (address);
    function getTimestamp() external view returns (uint256);
}

interface IBondedECDSAKeepFactory {
    function getKeepCount() external view returns (uint256);
    function getKeepAtIndex(uint256 index) external view returns (address);
    function getCreationTime(address _keep) external view returns (uint256);
}