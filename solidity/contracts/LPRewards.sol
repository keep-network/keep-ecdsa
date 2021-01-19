/*
   ____            __   __        __   _
  / __/__ __ ___  / /_ / /  ___  / /_ (_)__ __
 _\ \ / // // _ \/ __// _ \/ -_)/ __// / \ \ /
/___/ \_, //_//_/\__//_//_/\__/ \__//_/ /_\_\
     /___/

* Synthetix: Unipool.sol
*
* Docs: https://docs.synthetix.io/
*
*
* MIT License
* ===========
*
* Copyright (c) 2020 Synthetix
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
*/

/// These contracts reward users for adding liquidity to Uniswap https://uniswap.org/
/// These contracts were obtained from Synthetix and added some minor changes.
/// You can find the original contracts here:
/// https://etherscan.io/address/0x48d7f315fedcad332f68aafa017c7c158bc54760#code

pragma solidity 0.5.17;

import "@keep-network/keep-core/contracts/KeepToken.sol";
import "@keep-network/keep-core/contracts/PhasedEscrow.sol";

import "openzeppelin-solidity/contracts/math/Math.sol";
import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "openzeppelin-solidity/contracts/token/ERC20/SafeERC20.sol";

contract IRewardDistributionRecipient is Ownable {
    address rewardDistribution;

    function notifyRewardAmount(uint256 reward) external;

    modifier onlyRewardDistribution() {
        require(
            msg.sender == rewardDistribution,
            "Caller is not reward distribution"
        );
        _;
    }

    function setRewardDistribution(address _rewardDistribution)
        external
        onlyOwner
    {
        rewardDistribution = _rewardDistribution;
    }
}

contract LPTokenWrapper {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;

    IERC20 public wrappedToken; // Pairs: KEEP/ETH, TBTC/ETH, KEEP/TBTC

    constructor(IERC20 _wrappedToken) public {
        wrappedToken = _wrappedToken;
    }

    function totalSupply() public view returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) public view returns (uint256) {
        return _balances[account];
    }

    function stake(uint256 amount) public {
        _totalSupply = _totalSupply.add(amount);
        _balances[msg.sender] = _balances[msg.sender].add(amount);
        wrappedToken.safeTransferFrom(msg.sender, address(this), amount);
    }

    function withdraw(uint256 amount) public {
        _totalSupply = _totalSupply.sub(amount);
        _balances[msg.sender] = _balances[msg.sender].sub(amount);
        wrappedToken.safeTransfer(msg.sender, amount);
    }
}

contract LPRewards is
    LPTokenWrapper,
    IRewardDistributionRecipient,
    IStakingPoolRewards
{
    IERC20 public keepToken;
    uint256 public constant DURATION = 7 days;

    uint256 public periodFinish = 0;
    uint256 public rewardRate = 0;
    uint256 public lastUpdateTime;
    uint256 public rewardPerTokenStored;
    mapping(address => uint256) public userRewardPerTokenPaid;
    mapping(address => uint256) public rewards;

    event RewardAdded(uint256 reward);
    event Staked(address indexed user, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount);
    event RewardPaid(address indexed user, uint256 reward);

    constructor(IERC20 _keepToken, IERC20 _wrappedToken)
        public
        LPTokenWrapper(_wrappedToken)
    {
        keepToken = _keepToken;
    }

    modifier updateReward(address account) {
        rewardPerTokenStored = rewardPerToken();
        lastUpdateTime = lastTimeRewardApplicable();
        if (account != address(0)) {
            rewards[account] = earned(account);
            userRewardPerTokenPaid[account] = rewardPerTokenStored;
        }
        _;
    }

    function exit() external {
        withdraw(balanceOf(msg.sender));
        getReward();
    }

    function notifyRewardAmount(uint256 reward)
        external
        onlyRewardDistribution
        updateReward(address(0))
    {
        keepToken.safeTransferFrom(msg.sender, address(this), reward);

        if (block.timestamp >= periodFinish) {
            rewardRate = reward.div(DURATION);
        } else {
            uint256 remaining = periodFinish.sub(block.timestamp);
            uint256 leftover = remaining.mul(rewardRate);
            rewardRate = reward.add(leftover).div(DURATION);
        }
        lastUpdateTime = block.timestamp;
        periodFinish = block.timestamp.add(DURATION);
        emit RewardAdded(reward);
    }

    function lastTimeRewardApplicable() public view returns (uint256) {
        return Math.min(block.timestamp, periodFinish);
    }

    function rewardPerToken() public view returns (uint256) {
        if (totalSupply() == 0) {
            return rewardPerTokenStored;
        }
        return
            rewardPerTokenStored.add(
                lastTimeRewardApplicable()
                    .sub(lastUpdateTime)
                    .mul(rewardRate)
                    .mul(1e18)
                    .div(totalSupply())
            );
    }

    function earned(address account) public view returns (uint256) {
        return
            balanceOf(account)
                .mul(rewardPerToken().sub(userRewardPerTokenPaid[account]))
                .div(1e18)
                .add(rewards[account]);
    }

    // stake visibility is public as overriding LPTokenWrapper's stake() function
    function stake(uint256 amount) public updateReward(msg.sender) {
        require(amount > 0, "Cannot stake 0");
        super.stake(amount);
        emit Staked(msg.sender, amount);
    }

    function withdraw(uint256 amount) public updateReward(msg.sender) {
        require(amount > 0, "Cannot withdraw 0");
        super.withdraw(amount);
        emit Withdrawn(msg.sender, amount);
    }

    function getReward() public updateReward(msg.sender) {
        uint256 reward = earned(msg.sender);
        if (reward > 0) {
            rewards[msg.sender] = 0;
            keepToken.safeTransfer(msg.sender, reward);
            emit RewardPaid(msg.sender, reward);
        }
    }
}

/// @title KEEP rewards for TBTC-ETH liquidity pool.
contract LPRewardsTBTCETH is LPRewards {
    constructor(KeepToken keepToken, IERC20 tbtcEthUniswapPair)
        public
        LPRewards(keepToken, tbtcEthUniswapPair)
    {}
}

/// @title KEEP rewards for KEEP-ETH liquidity pool.
contract LPRewardsKEEPETH is LPRewards {
    constructor(KeepToken keepToken, IERC20 keepEthUniswapPair)
        public
        LPRewards(keepToken, keepEthUniswapPair)
    {}
}

/// @title KEEP rewards for KEEP-TBTC liquidity pool.
contract LPRewardsKEEPTBTC is LPRewards {
    constructor(KeepToken keepToken, IERC20 keepTbtcUniswapPair)
        public
        LPRewards(keepToken, keepTbtcUniswapPair)
    {}
}

/// @title KEEP rewards for the tBTC Saddle liquidity pool.
contract LPRewardsTBTCSaddle is LPRewards {
    bool public gated = true;

    constructor(KeepToken keepToken, IERC20 tbtcSaddleLPToken)
        public
        LPRewards(keepToken, tbtcSaddleLPToken)
    {}

    // if the pool is gated, disallow tx's that didn't come from msg.sender.
    function stake(uint256 amount) public {
        require(
            // solium-disable-next-line security/no-tx-origin
            !gated || msg.sender == tx.origin,
            "Only Externally Owned Account can stake"
        );
        super.stake(amount);
    }

    function setGated(bool _gated) public onlyOwner {
        gated = _gated;
    }
}
