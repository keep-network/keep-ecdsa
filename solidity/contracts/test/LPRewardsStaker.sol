pragma solidity 0.5.17;

import "../LPRewards.sol";

contract LPRewardsStaker {
    IERC20 public lpToken;
    LPRewards public lpRewards;

    constructor(IERC20 _lpToken, LPRewards _lpRewards) public {
        lpToken = _lpToken;
        lpRewards = _lpRewards;
    }

    function stake(uint256 amount) public {
        lpToken.approve(address(lpRewards), amount);
        lpRewards.stake(amount);
    }
}
