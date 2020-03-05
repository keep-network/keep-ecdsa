pragma solidity ^0.5.4;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract ECDSAKeepRewards {

    IERC20 keepToken;
    IBondedECDSAKeepFactory factory;

    // Total number of keep tokens to distribute.
    uint256 totalRewards;
    // Rewards that haven't been allocated to finished intervals
    uint256 unallocatedRewards;

    // Length of one interval.
    uint256 termLength;
    // Timestamp of first interval beginning.
    uint256 initiated;

    // Minimum number of keep submissions for each interval.
    uint256 minimumSubmissions;

    // Array representing the percentage of unallocated rewards
    // available for each reward interval.
    uint256[] intervalWeights; // percent array
    // Mapping of interval number to tokens allocated for the interval.
    mapping(uint256 => uint256) intervalAllocations;

    // Total number of intervals. (Implicit in intervalWeights)
    // uint256 termCount = intervalWeights.length;

    // mapping of keeps to booleans. True if the keep has been used to calim a reward.
    mapping(address => bool) claimed;

    // Number of submissions for each interval.
    mapping(uint256 => uint256) intervalSubmissions;

    // Array of timestamps marking interval's end.
    uint256[] intervalEndpoints;

    // Mapping of interval to number of keeps created in/before the interval
    mapping(uint256 => uint256) keepsByInterval;

    // Mapping of interval to number of keeps whose rewards have been paid out,
    // or reallocated because the keep closed unhappily
    mapping(uint256 => uint256) intervalKeepsProcessed;

    constructor (
        uint256 _termLength,
        uint256 _totalRewards,
        address _keepToken,
        uint256 _minimumSubmissions,
        address factoryAddress,
        uint256 _initiated,
        uint256[] memory _intervalWeights
    )
    public {
       keepToken = IERC20(_keepToken);
       totalRewards = _totalRewards;
       unallocatedRewards = totalRewards;
       termLength = _termLength;
       initiated = _initiated;
       minimumSubmissions = _minimumSubmissions;
       factory = IBondedECDSAKeepFactory(factoryAddress);
       intervalWeights = _intervalWeights;
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
        uint256 _totalTermRewards = totalRewards * intervalWeights[term] / 100;
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
        uint256 newInterval = findEndpoint(_timestamp);

        uint256 totalSubmissions = intervalEndpointsLength > 0 ?
        newInterval:
        newInterval - intervalEndpoints[intervalEndpointsLength - 1];

        intervalSubmissions[intervalEndpointsLength] = totalSubmissions;
        if (totalSubmissions < minimumSubmissions){
            if(intervalEndpointsLength >= intervalWeights.length){
                return newInterval;
            }
            intervalWeights[intervalEndpointsLength + 1] +=  intervalWeights[intervalEndpointsLength];
            intervalWeights[intervalEndpointsLength] = 0;
        }
        return newInterval;
    }

    /// @notice Checks if a keep is eligible to receive rewards.
    /// @dev Keeps that close dishonorably or early are not eligible for rewards.
    /// @param _keep The keep to check.
    /// @return True if the keep is eligible, false otherwise
    function eligibleForReward(address _keep) public view returns (bool){
        // check that keep closed properly
        return IBondedECDSAKeep(_keep).isClosed();
    }

    function findEndpoint(uint256 intervalEndpoint) public view returns (uint256) {
        require(
            intervalEndpoint <= currentTime(),
            "interval hasn't ended yet"
        );
        uint256 keepCount = factory.getKeepCount();
        // no keeps created yet -> return 0
        if (keepCount == 0) {
            return 0;
        }

        uint256 lb = 0; // lower bound, inclusive
        uint256 timestampLB = factory.getCreationTime(factory.getKeepAtIndex(lb));
        // all keeps created after the interval -> return 0
        if (timestampLB >= intervalEndpoint) {
            return 0;
        }

        uint256 ub = keepCount - 1; // upper bound, inclusive
        uint256 timestampUB = factory.getCreationTime(factory.getKeepAtIndex(ub));
        // all keeps created in or before the interval -> return next keep
        if (timestampUB < intervalEndpoint) {
            return keepCount;
        }

        // The above cases also cover the case
        // where only 1 keep has been created;
        // lb == ub
        // if it was created after the interval, return 0
        // otherwise, return 1

        return _find(lb, timestampLB, ub, timestampUB, intervalEndpoint);
    }

    // Invariants:
    //   lb >= 0, lbTime < target
    //   ub < keepCount, ubTime >= target
    function _find(
        uint256 lb,
        uint256 lbTime,
        uint256 ub,
        uint256 ubTime,
        uint256 target
    ) internal view returns (uint256) {
        uint256 len = ub - lb;
        while (len > 1) {
            // ub >= lb + 2
            // mid > lb
            uint256 mid = lb + (len / 2);
            uint256 midTime = factory.getCreationTime(factory.getKeepAtIndex(mid));

            if (midTime >= target) {
                ub = mid;
                ubTime = midTime;
            } else {
                lb = mid;
                lbTime = midTime;
            }
            len = ub - lb;
        }
        return ub;
    }

   function currentTime() public view returns (uint256) {
       return block.timestamp;
   }

   /// @notice Return the interval number
   /// the provided timestamp falls within.
   /// @dev Reverts if the timestamp is before `initiated`.
   /// @param timestamp The timestamp whose interval is queried.
   /// @return The interval of the timestamp.
   function intervalOf(uint256 timestamp) public view returns (uint256) {
       uint256 _initiated = initiated;
       uint256 _termLength = termLength;

       require(
           timestamp >= _initiated,
           "Timestamp is before the first interval"
       );

       uint256 difference = timestamp - _initiated;
       uint256 interval = difference / _termLength;

       return interval;
   }

   /// @notice Return the timestamp corresponding to the start of the interval.
   function startOf(uint256 interval) public view returns (uint256) {
       return initiated + (interval * termLength);
   }

   /// @notice Return the timestamp corresponding to the end of the interval.
   function endOf(uint256 interval) public view returns (uint256) {
       return startOf(interval + 1);
   }

   /// @notice Return the endpoint index of the interval,
   /// i.e. the number of keeps created in and before the interval.
   /// The interval must have ended;
   /// otherwise the endpoint might still change.
   /// @dev Uses a locally cached result,
   /// and stores the result if it isn't cached yet.
   /// All keeps created before the initiation fall in interval 0.
   /// @param interval The number of the interval.
   /// @return endpoint The number of keeps the factory had created
   /// before the end of the interval.
   function getEndpoint(uint256 interval) public returns (uint256 endpoint) {
       uint256 intervalEnd = endOf(interval);
       require(
           block.timestamp >= intervalEnd,
           "Interval hasn't ended yet"
       );
       // Get the endpoint from local cache;
       // might not be recorded yet
       uint256 maybeEndpoint = keepsByInterval[interval];

       // Either the endpoint is zero
       // (no keeps created by the end of the interval)
       // or the endpoint isn't cached yet
       if (maybeEndpoint == 0) {
           // Check what the real endpoint is
           // if the actual value is 0, this call short-circuits
           // so we don't need to special-case the zero
           uint256 realEndpoint = findEndpoint(endOf(interval));
           // We didn't have the correct value cached,
           // so store it
           if (realEndpoint != 0) {
               keepsByInterval[interval] = realEndpoint;
           }
           endpoint = realEndpoint;
       } else {
           endpoint = maybeEndpoint;
       }
       return endpoint;
   }

   function getPreviousEndpoint(uint256 interval) public returns (uint256) {
       if (interval == 0) {
           return 0;
       } else {
           return getEndpoint(interval - 1);
       }
   }

   function getIntervalWeight(uint256 interval) public view returns (uint256) {
       if (interval < getIntervalCount()) {
           return intervalWeights[interval];
       } else {
           return 100;
       }
   }

   function getIntervalCount() public view returns (uint256) {
       return intervalWeights.length;
   }
}

interface IBondedECDSAKeep {
    function getOwner() external view returns (address);
    function getTimestamp() external view returns (uint256);
    function isClosed() external view returns (bool);
    function isTerminated() external view returns (bool);
    function isActive() external view returns (bool);
}

interface IBondedECDSAKeepFactory {
    function getKeepCount() external view returns (uint256);
    function getKeepAtIndex(uint256 index) external view returns (address);
    function getCreationTime(address _keep) external view returns (uint256);
}
