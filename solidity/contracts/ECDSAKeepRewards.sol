pragma solidity ^0.5.4;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract ECDSAKeepRewards {

    IERC20 keepToken;
    IBondedECDSAKeepFactory factory;
    uint256 totalRewards;
    uint256 termLength;
    uint256 initiated;
    uint256 minimumSubmissions;
    uint8[] InitialTermWeights = [8, 33, 21, 14, 9, 5, 3, 2, 2, 1, 1, 1]; // percent array
    uint256 termCount = InitialTermWeights.length;

    mapping(address => bool) claimed;
    mapping(uint256 => uint256) intervalSubmissions;
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

    function findInterval(uint256 _timestamp) public returns (uint256){
        // provide index/rewards interval and validate on-chain?
        // if interval exists, return it. else updateInterval()
        return updateInterval(_timestamp);
    }

    function termReward(uint256 term) public view returns (uint256){
        uint256 _totalTermRewards = totalRewards * InitialTermWeights[term] / 100;
        return _totalTermRewards / intervalSubmissions[term];
    }
    function getTermSubmissionCount(uint256 term) internal returns (uint256){
        require(intervalEndpoints.length > term, "interval for given term not over yet");
        return intervalSubmissions[term];
    }

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

    function eligibleForReward(address _keep) public view returns (bool){
        // check that keep closed properly
        return true;
    }

    function findEndpoint(uint256 intervalEndpoint) public view returns (uint256) {
        require(
            intervalEndpoint <= currentTime(),
            "interval hasn't ended yet"
        );
        uint256 start = 0;
        uint256 end = factory.getKeepCount();
        if (end == 0) {
            return 0;
        }
        return find(start, end, intervalEndpoint);
    }

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
                return _mid + 1;
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
   function tt(uint256 ind) public view returns (uint256) {
    return factory.getKeepCount();
}
   function currentTime() public view returns (uint256) {
       return block.timestamp;
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
