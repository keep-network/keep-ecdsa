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
        address _factoryAddress,
        address _operatorContract,
        address _stakingContract
    ) public Rewards(
        _token,
        bondedECDSAKeepFactoryDeployment,
        backportECDSAIntervalWeight,
        backportECDSATermLength,
        minimumECDSAKeepsPerInterval
    ) {
        factory = IBondedECDSAKeepFactory(_factoryAddress);
    }

}
