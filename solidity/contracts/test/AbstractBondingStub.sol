pragma solidity 0.5.17;

import "../../contracts/AbstractBonding.sol";

import "./StakingInfoStub.sol";

import "@keep-network/keep-core/contracts/KeepRegistry.sol";

contract AbstractBondingStub is AbstractBonding {
    StakingInfoStub stakingInfoStub;

    constructor(KeepRegistry _keepRegistry, StakingInfoStub _stakingInfoStub)
        public
        AbstractBonding(_keepRegistry)
    {
        stakingInfoStub = _stakingInfoStub;
    }

    function withdraw(uint256 amount, address operator) public {
        revert("abstract function");
    }

    function withdrawBondExposed(uint256 amount, address operator) public {
        withdrawBond(amount, operator);
    }

    function isAuthorizedForOperator(
        address _operator,
        address _operatorContract
    ) public view returns (bool) {
        return
            stakingInfoStub.isAuthorizedForOperator(
                _operator,
                _operatorContract
            );
    }

    function authorizerOf(address _operator) public view returns (address) {
        return stakingInfoStub.authorizerOf(_operator);
    }

    function beneficiaryOf(address _operator)
        public
        view
        returns (address payable)
    {
        return stakingInfoStub.beneficiaryOf(_operator);
    }
}
