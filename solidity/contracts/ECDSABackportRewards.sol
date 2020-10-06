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
import "./BondedECDSAKeepFactory.sol";
import "./api/IBondedECDSAKeep.sol";

contract ECDSABackportRewards is Rewards {
    // BondedECDSAKeepFactory deployment date, May-13-2020 interval started.
    // https://etherscan.io/address/0x18758f16988E61Cd4B61E6B930694BD9fB07C22F
    uint256 internal constant bondedECDSAKeepFactoryDeployment = 1589408351;

    // We are going to have one interval, with a weight of 100%.
    uint256[] internal backportECDSAIntervalWeight = [100];

    // Interval is the difference in time of creation between older and newer
    // versions of BondedECDSAKeepFactory.
    // Older: https://etherscan.io/address/0x18758f16988E61Cd4B61E6B930694BD9fB07C22F
    // Newer: https://etherscan.io/address/0xA7d9E842EFB252389d613dA88EDa3731512e40bD
    uint256 internal constant backportECDSATermLength = 123 days; // 10678946 sec

    uint256 internal constant minimumECDSAKeepsPerInterval = 40;

    BondedECDSAKeepFactory factory;

    constructor(address _token, address payable _factoryAddress)
        public
        Rewards(
            _token,
            bondedECDSAKeepFactoryDeployment,
            backportECDSAIntervalWeight,
            backportECDSATermLength,
            minimumECDSAKeepsPerInterval
        )
    {
        factory = BondedECDSAKeepFactory(_factoryAddress);
    }

    function _getKeepCount() internal view returns (uint256) {
        // Between May 13 2020 - Sep 14 2020 there were 41 keeps opened.
        // The last Keep was opened on May-17-2020
        // https://etherscan.io/address/0x45A3cACA2F2a78A53607618651C86111c9720AA5
        return factory.getKeepCount();
    }

    function _getKeepAtIndex(uint256 index) internal view returns (bytes32) {
        return fromAddress(factory.getKeepAtIndex(index));
    }

    function _getCreationTime(bytes32) internal view returns (uint256) {
        // Assign each keep to the starting timestamp of its interval.
        return startOf(0);
    }

    function _isClosed(bytes32) internal view returns (bool) {
        // Even though we still have some of the keeps opened, all the keeps
        // created between May 13 2020 - Sep 14 2020 are considered closed. Because
        // of the deposits pause https://tbtc.network/news/2020-05-21-details-of-the-tbtc-deposit-pause-on-may-18-2020/
        // closing all the keeps is not easily achievable. However, we do not want
        // to block rewards distribution for good stakers caused by the incident on May 18th.
        return true;
    }

    function _isTerminated(bytes32 groupIndexBytes)
        internal
        view
        returns (bool)
    {
        return false;
    }

    // A keep is recognized if it was created by this factory between May 13 2020 - Sep 14 2020
    function _recognizedByFactory(bytes32 _keep) internal view returns (bool) {
        return factory.getKeepOpenedTimestamp(toAddress(_keep)) != 0;
    }

    function _distributeReward(bytes32 _keep, uint256 amount)
        internal
        isAddress(_keep)
    {
        token.approve(toAddress(_keep), amount);

        IBondedECDSAKeep(toAddress(_keep)).distributeERC20Reward(
            address(token),
            amount
        );
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
