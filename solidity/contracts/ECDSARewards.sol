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
import "./BondedECDSAKeep.sol";

contract ECDSARewards is Rewards {
    // BondedECDSAKeepFactory deployment date, Sep-14-2020 interval started.
    // https://etherscan.io/address/0xA7d9E842EFB252389d613dA88EDa3731512e40bD
    uint256 internal constant ecdsaFirstIntervalStart = 1600087297;

    /// Weights of the 24 reward intervals assigned over
    // 24 * backportECDSATermLength days.
    uint256[] internal intervalWeights = [
        4, 8, 10, 12, 15, 15,
        15, 15, 15, 15, 15, 15,
        15, 15, 15, 15, 15, 15,
        15, 15, 15, 15, 15, 15
    ];

    // Each interval is 30 days long.
    uint256 internal constant backportECDSATermLength = 30 days;

    // TODO: Define..
    uint256 internal constant minimumECDSAKeepsPerInterval = 2;

    BondedECDSAKeepFactory factory;

    constructor(address _token, address payable _factoryAddress)
        public
        Rewards(
            _token,
            ecdsaFirstIntervalStart,
            intervalWeights,
            backportECDSATermLength,
            minimumECDSAKeepsPerInterval
        )
    {
        factory = BondedECDSAKeepFactory(_factoryAddress);
    }

    function _getKeepCount() internal view returns (uint256) {
        return factory.getKeepCount();
    }

    function _getKeepAtIndex(uint256 i) internal view returns (bytes32) {
        return fromAddress(factory.getKeepAtIndex(i));
    }

    function _getCreationTime(bytes32 _keep) internal view returns (uint256) {
        return factory.getKeepOpenedTimestamp(toAddress(_keep));
    }

    function _isClosed(bytes32 _keep) internal view returns (bool) {
        return BondedECDSAKeep(toAddress(_keep)).isClosed();
    }

    function _isTerminated(bytes32 _keep)
        internal
        view
        returns (bool)
    {
        return BondedECDSAKeep(toAddress(_keep)).isTerminated();
    }

    // A keep is recognized if it was opened by this factory.
    function _recognizedByFactory(bytes32 _keep) internal view returns (bool) {
        return factory.getKeepOpenedTimestamp(toAddress(_keep)) != 0;
    }

    function _distributeReward(bytes32 _keep, uint256 amount)
        internal
        isAddress(_keep)
    {
        token.approve(toAddress(_keep), amount);

        BondedECDSAKeep(toAddress(_keep)).distributeERC20Reward(
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
