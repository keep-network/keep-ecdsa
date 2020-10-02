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
import "./api/IBondedECDSAKeepFactory.sol";
import "./api/IBondedECDSAKeep.sol";

contract ECDSABackportRewards is Rewards {

    // BondedECDSAKeepFactory deployment date, May-13-2020 interval started.
    // https://etherscan.io/address/0x18758f16988E61Cd4B61E6B930694BD9fB07C22F
    uint256 internal constant bondedECDSAKeepFactoryDeployment = 1589408351;

    // We are going to have one interval, with a weight of 100%.
    uint256[] internal backportECDSAIntervalWeight = [100];
    uint256 internal constant lastInterval = 0;

    // Interval is the difference in time of creation between older and newer 
    // versions of BondedECDSAKeepFactory.
    // Older: https://etherscan.io/address/0x18758f16988E61Cd4B61E6B930694BD9fB07C22F
    // Newer: https://etherscan.io/address/0xA7d9E842EFB252389d613dA88EDa3731512e40bD
    uint256 internal constant backportECDSATermLength = 123 days; // 10678946 sec

    // There were 41 Keeps created by BondedECDSAKeepFactory : 0x18758f16988E61Cd4B61E6B930694BD9fB07C22F
    // The last Keep was opened on May-17-2020
    // https://etherscan.io/address/0x45A3cACA2F2a78A53607618651C86111c9720AA5
    uint256 internal constant numberOfCreatedECDSAKeeps = 41;

    // We allocate ecdsa backport rewards to all 41 keeps.
    uint256 internal constant minimumECDSAKeepsPerInterval = numberOfCreatedECDSAKeeps;

    IBondedECDSAKeepFactory factory;

    constructor (
        address _token,
        address _factoryAddress
    ) public Rewards(
        _token,
        bondedECDSAKeepFactoryDeployment,
        backportECDSAIntervalWeight,
        backportECDSATermLength,
        minimumECDSAKeepsPerInterval
    ) {
        factory = IBondedECDSAKeepFactory(_factoryAddress);
    }

    /// @notice Sends the reward for a keep to the group signing member beneficiaries.
    /// @param groupIndex Index of the keep to receive a reward.
    function receiveReward(uint256 groupIndex) public {
        receiveReward(bytes32(groupIndex));
    }

    function _getKeepCount() internal view returns (uint256) {
        // Between May 17 2020 - Sep 14 2020 there were 41 keeps opened.
        return numberOfCreatedECDSAKeeps;
    }

    function _getKeepAtIndex(uint256 i) internal view returns (bytes32) {
        return bytes32(i);
    }

    function _getCreationTime(bytes32) internal view returns (uint256) {
        // Assign each keep to the starting timestamp of its interval.
        return startOf(0);
    }

    function _isClosed(bytes32) internal view returns (bool) {
        // All keeps created between May 17 2020 - Sep 14 2020 are considered closed.
        return true;
    }

    function _isTerminated(bytes32 groupIndexBytes) internal view returns (bool) {
        return false;
    }

    // A keep is recognized if its index is at most `last eligible keep`.
    function _recognizedByFactory(bytes32 groupIndexBytes) internal view returns (bool) {
        return numberOfCreatedECDSAKeeps > uint256(groupIndexBytes);
    }

    function _distributeReward(bytes32 _keep, uint256 amount) internal isAddress(_keep) {
        token.approve(toAddress(_keep), amount);

        IBondedECDSAKeep(toAddress(_keep)).distributeERC20Reward(address(token), amount);
    }

    function validAddressBytes(bytes32 keepBytes) internal pure returns (bool) {
        return fromAddress(toAddress(keepBytes)) == keepBytes;
    }

    function toAddress(bytes32 keepBytes) internal pure returns (address) {
        return address(bytes20(keepBytes));
    }

    function fromAddress(address keepAddress) internal pure returns (bytes32) {
        return bytes32(bytes20(keepAddress));
    }

    modifier isAddress(bytes32 _keep) {
        require(validAddressBytes(_keep), "Invalid keep address");
        _;
    }
}
