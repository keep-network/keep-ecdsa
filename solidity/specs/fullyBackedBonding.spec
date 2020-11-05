methods {
    withdraw(uint256,address)
    deposit(address)
    beneficiaryOf(address) returns address envfree
    ownerOf(address) returns address envfree
    unbondedValue(address) returns uint256 envfree
    bondAmount(address, address, uint256) returns uint256 envfree
    balanceOf(address) returns uint256 envfree
    getDelegatedAuthority(address) returns address envfree
}
/**
@title Valid state of a bond
@notice
        unbondedValue[o] + lockedBonds[o] > 0 ⇒
       		( ownerOf(o) ≠ 0 ⋀ beneficiaryOf(o) ≠ 0 )
*/
invariant validState(address operator, address holder, uint256 referenceID) /* status: WIP */
     unbondedValue(operator) + bondAmount(operator, holder, referenceID) > 0 => (ownerOf(operator) != 0 && beneficiaryOf(operator) != 0 )

/**
@title  No cyclic authorization
@notice There is  no cycle in delegating authority, as it can cause denial of service

        ㄱ (delegatedAuthority(a) = b ⋀ delegatedAuthority(b) = c ⋀ delegatedAuthority(c) = a )

*/
invariant noCyclicDelegatedAuthority(address a, address b, address c) /* status: found violation */
     (a!=0 && b!=0 && c!=0 => !(getDelegatedAuthority(a)==b && getDelegatedAuthority(b)==c && getDelegatedAuthority(c)==a)) && getDelegatedAuthority(0) == 0

/**
@title Total assets of bond participants is preserved (one does not lose money)
    Total assets of a user is defined as the sum of assets within the system (either locked or unbonded) and outside the system
        totalAssets(a) ≡ a.balanceOf() + unbondedValue(a) + lockedBonds(a)

    The total assets of an operator, his beneficiary and his owner can only increase. This should hold for every operation in the system by any user
    {  b = totalAssets(operator) + totalAssets(beneficiary) + totalAssets(owner) }
	    op
    {   (ownerOf(operator) = owner ⋀ beneficiaryOf(operator) =  beneficiary ) ⇒
        b ≤ totalAssets(operator) + totalAssets(beneficiary) + totalAssets(owner) }
*/
rule totalAssets(address operator, address owner, address beneficiary, address holder, uint256 referenceID, method f) { /*status : WIP */
    uint256 totalAssetsOwnerBefore = balanceOf(owner) + bondAmount(owner, holder, referenceID) + unbondedValue(owner);
    uint256 totalAssetsOperatorBefore = balanceOf(operator) + bondAmount(operator, holder, referenceID) + unbondedValue(operator);
    uint256 totalAssetsBeneficiaryBefore = balanceOf(beneficiary) + bondAmount(beneficiary, holder, referenceID) + unbondedValue(beneficiary);
    mathint b = totalAssetsOwnerBefore + totalAssetsOperatorBefore + totalAssetsBeneficiaryBefore;
    require totalAssetsOwnerBefore < 1000 && totalAssetsOperatorBefore < 1000 && totalAssetsBeneficiaryBefore < 1000; // XXX
    env e;
    calldataarg args;
    sinvoke f(e, args);
    uint256 totalAssetsOwnerAfter = balanceOf(owner) + bondAmount(owner, holder, referenceID) + unbondedValue(owner); // mathint?
    uint256 totalAssetsOperatorAfter = balanceOf(operator) + bondAmount(operator, holder, referenceID) + unbondedValue(operator);
    uint256 totalAssetsBeneficiaryAfter = balanceOf(beneficiary) + bondAmount(beneficiary, holder, referenceID) + unbondedValue(beneficiary);
    assert ( ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary )  =>
            ( b <= totalAssetsOwnerAfter + totalAssetsOperatorAfter + totalAssetsBeneficiaryAfter );
}


/**
@title Integrity of withdraw
@notice successful withdraw of a value x of an operator o decreases o's unbonded value by x and transfers x to the beneficiary of the operator
             { 	beneficiary = beneficiaryOf(o)  ⋀
                u = unbondedValue(o)  ⋀
                b = beneficiary.balanceOf()
            }
            withdraw(x, o)
            {
                  u - x = unbondedValue(o)  ⋀
                  b + x = beneficiary.balanceOf()
            }
*/
//todo - add also check regarding balanceOf(currentContract)
rule integrityOfWithdraw(address operator, address owner, address beneficiary, uint256 x) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(beneficiary) ;
    env e; //no constraint on the e.msg.sender
    sinvoke withdraw(e, x, operator);
    assert ( u - x == unbondedValue(operator) &&  b + x == balanceOf(beneficiary) );
}
/**
@title Additivity of withdraw
@notice Withdrawing is additive, i.e., it can be performed either all at once or in two steps

( withdraw(x, o) ; withdraw(y, o) ) ~ withdraw(x+y, o)

Here we expect the effect of withdrawing and x and then withdrawing y to be the same as withdrawing them simultaneously.
The correctness of this rule on all inputs increases the confidence that the protocol is less fragile, e.g., to rounding errors.
*/
rule additiveWithdraw(address operator, address owner, address beneficiary, uint256 x, uint y) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    storage init_storage = lastStorage; //store the current state
    env eScenario1;
    sinvoke withdraw(eScenario1, x, operator);
    sinvoke withdraw(eScenario1, y, operator);
    uint256 uScenario1 = unbondedValue(operator);
    uint256 bScenario1 = balanceOf(beneficiary);

    env eScenario2;
    sinvoke withdraw(eScenario2, x + y, operator) at init_storage; //back to the initial state
    uint256 uScenario2 = unbondedValue(operator);
    uint256 bScenario2 = balanceOf(beneficiary);
    assert (uScenario1 == uScenario2 && bScenario1 == bScenario2);
}

/**
@title No front running on withdraw
@notice If one can withdraw x amount from the unbounded amount of operator o1, then same one should be able to withdraw after
another user has performed an operation as operator o2.

withdraw(x, o1)   ~unbondedValue(o), beneficiary.balanceOf()   ( op_o2 ; withdraw(x, o1) )

Here we compare two arbitrary executions of the program: one with a single withdrawal,
and another in which another operation is performed by another operator.
We require that the extra operation did not affect the value unbondedValue(o) and beneficiary.balanceOf().
*/

// status: WIP need to limit the f to a different pair of operator and beneficiary
rule noFrontRunningOnWithdraw(address operator, address owner, address beneficiary, uint256 x, method f) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    storage init_storage = lastStorage;
    env eWithdraw;
    sinvoke withdraw(eWithdraw, x, operator) ;
    uint256 uScenario1 = unbondedValue(operator);
    uint256 bScenario1 = balanceOf(beneficiary);

    env eF;
    calldataarg args;
    require (eF.msg.sender != owner && eF.msg.sender != operator && eF.msg.sender != beneficiary  && eF.msg.sender != currentContract);
    sinvoke f(eF,args) at init_storage;
    invoke withdraw(eWithdraw, x, operator);
    uint256 uScenario2 = unbondedValue(operator);
    uint256 bScenario2 = balanceOf(beneficiary);
    assert !lastReverted && (uScenario1 == uScenario2 && bScenario1 == bScenario2);
}

/**
@title Deposit and withdraw are inverse functions
@notice Withdraw is the inverse function of deposit with respect to the system balance and the unbounded value of operator
    {  u = unbondedValue(o)  ⋀
       b = ethSystemBalance()
    }
    ( deposit(x, o) ; withdraw(x, o) )
    {  u = unbondedValue(o)  ⋀
       b = ethSystemBalance()
    }
*/
//todo - change to invoke withdraw
rule inverseOfDepositAndWithdraw(address operator, address owner, address beneficiary, uint256 x ) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(currentContract) ;
    env eDeposit;
    env eWithdraw;
    require eDeposit.msg.value == x;
    require (eDeposit.msg.sender != currentContract);
    sinvoke deposit(eDeposit, operator);
    sinvoke withdraw(eWithdraw, x, operator);
    assert ( u  == unbondedValue(operator) && b == balanceOf(currentContract));
}



//rules for learning the code
rule changeToBalanceOf(address o,  method f) {
    uint256 totalAssetsBefore = balanceOf(o) ;
    env e;
    calldataarg args;
    sinvoke f(e, args);
    uint256 totalAssetsAfter = balanceOf(o) ;
    assert (totalAssetsAfter ==totalAssetsBefore );
}

//todo - certora issue with computing bondAmount
rule changeToBondAmount(address operator, address owner,  method f, uint256 totalAssetsBefore, uint256 totalAssetsAfter) {
    address holder;
    uint256 referenceID;
    require totalAssetsBefore == bondAmount(owner, holder, referenceID) ;
    env e;
    calldataarg args;
    sinvoke f(e, args);
    require totalAssetsAfter == bondAmount(owner, holder, referenceID);
    assert (totalAssetsBefore  == totalAssetsAfter );
}

rule changeToUnbondedValue(address o, address owner,  method f) {
    uint256 totalAssetsBefore = unbondedValue(o);
    env e;
    calldataarg args;
    sinvoke f(e, args);
    uint256 totalAssetsAfter = unbondedValue(o);
    assert (totalAssetsAfter  == totalAssetsBefore);
}
