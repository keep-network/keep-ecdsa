using OtherBeneficiary as otherBeneficiary

/*
    This is a specification file for smart contract verification with the Certora prover.
    For more information, visit: https://www.certora.com/
*/


/*
    Declaration of methods that are used in the rules.
    envfree indicate that the method is not dependent on the environment (msg.value, msg.sender).
    Methods that are not declared here are assumed to be dependent on env.
*/
methods {
    authorizerOf(address) returns address envfree
    balanceOf(address) returns uint256 envfree
    beneficiaryOf(address) returns address envfree
    bondAmount(address, address, uint256) returns uint256 envfree
    createBond(address, address, uint256, uint256, address)
    deposit(address)
    everDeposited(address) returns uint256 envfree
    getDelegatedAuthority(address) returns address envfree
    init_state() envfree
    ownerOf(address) returns address envfree
    seizeBond(address, uint, uint, address)
    unbondedValue(address) returns uint256 envfree
    withdraw(uint256, address)
}


// A few macros
definition MAXINT() returns uint256 =
        0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;

definition safeAdd(uint256 x, uint256 y) returns bool =
        x + y >= y && x + y >=x && x+y <= MAXINT();



/* A ghost can be thought of as an additional variable in the background. */

// A mapping from bondID to operator. e.g. mapping(uint256 => address).
ghost bondIDToOperator(uint256) returns address;

// A mapping from operator to the sum of all their bonds e.g. mapping(address => uint256).
ghost totalLockedBonds(uint256) returns uint256;


// The total amount of lockedBonds of all operators
ghost allLockedBond() returns uint256
{
    init_state axiom allLockedBond() == 0;
}


// The total amount of unbonded of all operators
ghost allUnbonded() returns uint256
{
  init_state axiom allUnbonded() == 0;
}


/*
    Update bondIdToOperator whenever operatorHolderRefToBondID is changed.
    e.g. after,
    operatorHolderRefToBondID[operatorY][holder][referenceID] = bondID;
    the following is simulated:
    bondIDToOperator[bondID] = operatorY;
*/
hook Sload uint bondID operatorHolderRefToBondID[KEY uint operatorY][KEY uint holder][KEY uint referenceID] STORAGE {
    havoc bondIDToOperator assuming bondIDToOperator@new(bondID) == operatorY &&
            (forall uint otherBondID. otherBondID != bondID => bondIDToOperator@new(otherBondID) == bondIDToOperator@old(otherBondID));
}


/*
    Update totalLockedBonds whenever lockedBonds is changed based on the old value of both lockedBonds and
    totalLockedBonds.
    e.g. after,
    lockedBonds[bondID] = x ;
    the following is simulated:
    totalLockedBonds[bondIDToOperator[bondID]] = new_value;
    ....
*/
hook Sstore lockedBonds[KEY uint bondID] uint value (uint old_value) STORAGE {
    uint256 operatorX = bondIDToOperator(bondID);
    havoc totalLockedBonds assuming totalLockedBonds@new(operatorX) == totalLockedBonds@old(operatorX) + value - old_value &&
            (forall uint256 otherOperator. otherOperator != operatorX =>
                totalLockedBonds@new(otherOperator) == totalLockedBonds@old(otherOperator));
    havoc allLockedBond assuming allLockedBond@new() == allLockedBond@old() + value - old_value;
}

hook Sstore unbondedValue[KEY uint operatorY] uint value (uint old_value) STORAGE {
    havoc allUnbonded assuming allUnbonded@new() == allUnbonded@old() + value - old_value;
}


/**
    @title Valid Operator ✔️
    @notice Zero cannot be an operator.
           beneficiaryOf(o) ≠ 0  ⇔
              ( o ≠ 0 ⋀ ownerOf(o) ≠ 0  ⋀ authorizerOf(o) ≠ 0 )
*/
invariant validOperator(address operator)
        beneficiaryOf(operator) != 0  <=>  ( operator != 0 && ownerOf(operator) != 0 && authorizerOf(operator) != 0 )


/**
    @title Valid state of an operator ❌.
    @notice Operators with assets must have an owner, a beneficiary, and an authorizer.

        (unbondedValue(o) + lockedBonds(o)) > 0 ⟹
            ( ownerOf(o) ≠ 0 ⋀ beneficiaryOf(o) ≠ 0 ⋀ authorizerOf(o) ≠ 0 )
*/
rule validState(address operator, uint unbonded, uint totalLocked) {
     env e;
     require unbondedValue(operator) == unbonded;
     require totalLockedBonds(operator) == totalLocked;
     require safeAdd(unbonded, totalLocked);

     uint sum_before = unbonded + totalLocked;
     require safeAdd(sum_before, e.msg.value);

     require sum_before > 0 => beneficiaryOf(operator) != 0 ;
     require e.msg.sender != 0;
     require operator != 0;
     requireInvariant validOperator(operator);

     method f;
     if (f.selector != seizeBond(address, uint256, uint256, address).selector) {
        calldataarg args;
        f(e, args);

     } else {

        address seizeAddress;
        uint refID;
        uint amount;
        address destination;

        uint256 before = bondAmount(operator, e.msg.sender, refID);
        require totalLocked >= before;

        seizeBond(e, seizeAddress, refID, amount, destination);
     }

     uint sum_after = unbondedValue(operator) + totalLockedBonds(operator);
     assert sum_after > 0 =>
                ( ownerOf(operator) != 0 && beneficiaryOf(operator) != 0 );
}


/**
    @title: No bankruptcy of the system (no money lost)  ✔*
    @notice The balance of the system is more than the obligations
    (the total assets deposited in the system and is in either unbounded or in a locked state )

         ethSystemBalance(a) ≥  sum for all operator o. ( unbondedValue(o) + lockedBonds(o) )
*/

invariant noBankruptcy()
    (allUnbonded() + allLockedBond()) <= balanceOf(currentContract) {
    preserved deposit(env e, address operator) {
        require e.msg.sender != currentContract;
        require (allUnbonded() + allLockedBond()) <= balanceOf(currentContract);
    }
    preserved topUp(env e, address _) {
        require e.msg.sender != currentContract;
        require (allUnbonded() + allLockedBond()) <= balanceOf(currentContract);
    }
    preserved delegate(env e, address _1 ,address _2, address _3) {
        require e.msg.sender != currentContract;
        require (allUnbonded() + allLockedBond()) <= balanceOf(currentContract);
    }
}


/**
    @title User cannot gain assets. ✔*
    @notice The total assets of an operator can not be more than all funds ever deposited to it.
            everDeposited(o) ≥ (unbondedValue(o) + lockedBonds(o))
*/
rule assetsLessThanEverDepositedAsRule(address operator, method f, mathint t, address holder, uint256 referenceID ) {
    env e;
    calldataarg args;
    require t == totalLockedBonds(operator);
    require totalLockedBonds(operator) >= bondAmount(operator, e.msg.sender, referenceID);
    require safeAdd(unbondedValue(operator), totalLockedBonds(operator));
    require everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
    require f.selector==init_state().selector => totalLockedBonds(operator)==0;
    uint256 before = bondAmount(operator, e.msg.sender, referenceID);
    f(e, args);
    require (f.selector==freeBond(address,uint256).selector                             ||
             f.selector==seizeBond(address,uint256,uint256,address).selector            ||
             f.selector==createBond(address,address,uint256,uint256,address).selector   ||
             f.selector==reassignBond(address, uint256, address, uint256).selector )    =>
             before != bondAmount(operator, e.msg.sender, referenceID);
    assert safeAdd(unbondedValue(operator), totalLockedBonds(operator)) &&
            everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
   }


/**
    @title Total assets of a user is the sum of assets within the system (either locked or unbounded) and outside the
            system.

    @notice The total assets of an operator within the system is preserved, except on deposits, withdrawal, and seizing.
        { b = unbondedValue(o) + lockedBonds(o) }
            op(u, x)
        { b =  unbondedValue(o) + lockedBonds(o) + x }

        Note that this property holds for every successful operation performed by any user or holder u, possibly sending
         in x amount of Wei (msg.value), except for siezeBond() and withdraw() which are defined differently:
*/
rule totalAssetsPreserved(address operator, address owner, address beneficiary,  method f) {
    env e;
    calldataarg args;
    mathint totalAssetsBefore = totalLockedBonds(operator) + unbondedValue(operator);
    require f.selector != seizeBond(address,uint256,uint256,address).selector &&
             f.selector != withdraw(uint256,address).selector;
    f(e, args);
    mathint totalAssetsAfter = totalLockedBonds(operator) + unbondedValue(operator);
    assert totalAssetsAfter == totalAssetsBefore ||  totalAssetsAfter == totalAssetsBefore + e.msg.value,
    "$f can change the total assets of an operator in an unexpected way";
}


/**
    On withdraw, the total assets within the system of an operator and the balance of the beneficiary is preserved.

    { b = unbondedValue(o) + lockedBonds(o) + beneficiaryOf(o).balance }
        withdraw(x, o)
    { b = unbondedValue(o) + lockedBonds(o) + beneficiaryOf(o).balance }

*/
rule totalAssetsPreservedOnWithdraw(address operator, address owner, address beneficiary, uint256 amount) {
     env e;
     require beneficiaryOf(operator) == beneficiary;
     require beneficiary != currentContract;
     mathint totalAssets = totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(beneficiary);
     withdraw(e, amount, operator);
     assert totalAssets == totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(beneficiary),
     "withdraw can change the total assets of an operator in an unexpected way";
}


/**
    When seizing of a bond, the total assets within the system of an operator and the destination address receiving the
        bond's value is preserved.
    { b = unbondedValue(o) + lockedBonds(o) + destination.balance }
        seizeBond(o, ref, x, destination)
    { b = unbondedValue(o) + lockedBonds(o) + destination.balance }
*/
rule totalAssetsPreservedOnSeizeBond(address operator, address owner, uint256 referenceID,
            uint256 amount, address destination) {
    env e;
    require destination == otherBeneficiary;
    mathint totalAssetsBefore = totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(destination);
    seizeBond(e, operator, referenceID, amount, destination);
    mathint totalAssetsAfter = totalLockedBonds(operator) + unbondedValue(operator) + balanceOf(destination);
    assert  totalAssetsBefore == totalAssetsAfter,
        "seizeBond can change the total assets of operator in an unexpected way";
}


/**
    @title Integrity of withdraw
    @notice Successful withdraw of a value x of an operator o decreases o's unbonded value by x and transfers x to the
            beneficiary of the operator
                 { 	beneficiary = beneficiaryOf(o)  ⋀
                    u = unbondedValue(o)            ⋀
                    b = beneficiary.balanceOf()
                }
                withdraw(x, o)
                {
                      u - x = unbondedValue(o)          ⋀
                      b + x = beneficiary.balanceOf()
                }
*/
rule integrityOfWithdraw(address operator,address beneficiary, uint256 x) {
    env e;
    require beneficiaryOf(operator) == beneficiary;
    require beneficiary != currentContract;
    require safeAdd(unbondedValue(operator),totalLockedBonds(operator));
    require everDeposited(operator) >= unbondedValue(operator) + totalLockedBonds(operator);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(beneficiary) ;
    uint s = balanceOf(currentContract);
    withdraw(e, x, operator);
    assert u - x == unbondedValue(operator) &&  b + x == balanceOf(beneficiary) &&
           s - x == balanceOf(currentContract) && x <= u && x <= everDeposited(operator),
           "withdraw integrity does not hold";
}


/**
    @title Maximal withdraw
    @notice When withdrawing the total assets ever deposited by an operator, that operator's assets are zeroed out.
            { }
            r = withdraw(everDeposited(o), o)
            { r ⟹ ( unbondedValue(o) = 0 ⋀ lockedBonds(o) = 0 ) }
*/
rule maximalWithdraw(address operator) {
    env e;
    require safeAdd(unbondedValue(operator), totalLockedBonds(operator));
    require everDeposited(operator) >= (unbondedValue(operator) + totalLockedBonds(operator));
    uint256 x = everDeposited(operator);
    invoke withdraw(e, x, operator);
    assert !lastReverted => (unbondedValue(operator) == 0 && totalLockedBonds(operator)==0),
            "maximum withdraw does not hold";
}


/**
    @title Additivity of withdraw
    @notice Withdrawing is additive, i.e., withdrawing the sum is identical to the sum of withdrawals.

    (withdraw(x, o) ; withdraw(y, o) ) ~ withdraw(x+y, o)

    Here we expect the effect of withdrawing and x and then withdrawing y to be the same as withdrawing them
    simultaneously. The correctness of this rule on all inputs increases the confidence that the protocol is less
    fragile, e.g., to rounding errors.
*/
rule additiveWithdraw(address operator, address owner, address beneficiary, uint256 x, uint y) {
    env e;
    require ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary;
    require beneficiary != currentContract;
    require x + y < MAXINT();
    require balanceOf(beneficiary) + x + y <= MAXINT();

    storage init_storage = lastStorage; // store the current state
    invoke withdraw(e, x, operator);
    bool successX = !lastReverted;
    invoke withdraw(e, y, operator);
    bool successY = !lastReverted;
    uint256 uScenario1 = unbondedValue(operator);
    uint256 bScenario1 = balanceOf(beneficiary);

    invoke withdraw(e, x + y, operator) at init_storage; // back to the initial state
    bool successXY = !lastReverted;
    uint256 uScenario2 = unbondedValue(operator);
    uint256 bScenario2 = balanceOf(beneficiary);
    assert (successX && successY) <=> successXY, "withdraw is not additive";
    assert successXY => (uScenario1 == uScenario2 && bScenario1 == bScenario2),
        "withdraw is not additive";
}


/**
    @title No front running on withdraw
    @notice If an operator can withdraw x amount from the unbounded amount of operator o1, then they should be able to
        withdraw that amount after another user has performed an operation as operator o2.

        {o1 != o2}    r1 = withdraw(x, o1)   ∼(r1=r2)   (op(o2,y) ; r2 = withdraw(x, o1) )

    Here we compare two arbitrary executions of the program: one with a single withdrawal and another in which a
    different operator performs a Keep operation. We require that if one withdrawal succeeds, so will the other and vice
    versa.
*/
rule noFrontRunningOnWithdraw(address operator, address owner, address beneficiary, address otherOperator, uint256 x,
        method f) {
    env eF;
    calldataarg args;
    uint256 referenceID;

    require ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary;
    require otherOperator != operator;

    require safeAdd(unbondedValue(operator), unbondedValue(otherOperator)) &&
            allUnbonded()  >= unbondedValue(operator) + unbondedValue(otherOperator);

    require safeAdd(totalLockedBonds(operator), totalLockedBonds(otherOperator)) &&
            allLockedBond()  >= totalLockedBonds(operator) + totalLockedBonds(otherOperator);

    require safeAdd(allUnbonded(), allLockedBond()) &&
            balanceOf(currentContract) >= allUnbonded() + allLockedBond();

    require totalLockedBonds(otherOperator) >= bondAmount(otherOperator, eF.msg.sender, referenceID);

    uint256 allUserAssets = allUnbonded() + allLockedBond();
    require safeAdd(balanceOf(currentContract), allUserAssets);
    uint256 allAssets = balanceOf(currentContract) + allUserAssets;
    require safeAdd(balanceOf(beneficiary), allAssets);
    require beneficiary != currentContract;

    storage init_storage = lastStorage;
    env eWithdraw;
    withdraw(eWithdraw, x, operator);
    bool succSceanrio1 = !lastReverted;

    // f should only change other operator and if changed totalLocked than the specific bondID used in the require
    uint256 u = unbondedValue(otherOperator) at init_storage;
    uint256 l = totalLockedBonds(otherOperator);
    uint256 bondBefore = bondAmount(otherOperator, eF.msg.sender, referenceID);
    f(eF, args);
    require u != unbondedValue(otherOperator) || l != totalLockedBonds(otherOperator);
    require f.selector==seizeBond(address,uint256,uint256,address).selector =>
                bondBefore != bondAmount(otherOperator, eF.msg.sender, referenceID);
    invoke withdraw(eWithdraw, x, operator);
    bool succSceanrio2 = !lastReverted;
    assert succSceanrio2;
}


/**
    @title Deposit and withdraw are inverse functions
    @notice Withdraw is the inverse function of deposit with respect to the system balance and the unbounded value of an
            operator.

            { u = unbondedValue(o) ⋀ b = ethSystemBalance() }
            ( deposit(x, o) ; r = withdraw(x, o) )
            { r ⋀ u = unbondedValue(o) ⋀ b = ethSystemBalance() }

            In addition, withdraw is only possible under the limitations:
                (i) An allowed user with no msg.value
                (ii) After the delegation lock period passed
                (iii) The Beneficiary can accept the transfer
*/
rule inverseOfDepositAndWithdraw(address operator, address owner, address beneficiary, uint256 x) {
    require (ownerOf(operator) == owner && beneficiaryOf(operator) == beneficiary);
    require (beneficiary != currentContract);
    uint256 u = unbondedValue(operator);
    uint256 b = balanceOf(currentContract);
    // successful deposit
    env eDeposit;
    require eDeposit.msg.value == x;
    require (eDeposit.msg.sender != currentContract);
    deposit(eDeposit, operator);
    env eWithdraw;
    require eWithdraw.msg.sender == operator || eWithdraw.msg.sender == owner;
    require eWithdraw.msg.value == 0;
    require _hasDelegationLockPassed(eWithdraw, operator);
    require safeAdd(balanceOf(beneficiary), x);
    invoke withdraw(eWithdraw, x, operator);
    bool success = !lastReverted;
    // verify that succeeded and the the value is as expected
    assert success &&  u  == unbondedValue(operator) && b == balanceOf(currentContract);
}


/**
    @title Valid change to BalanceOf ✔️
    @notice On Keep operation op by user s sending x amount, the ETH balance of a user (operator or not) changes as
            follows:

            { b = o.balance }
            op(s,x)
            {
                b = o.balance                                               ∨
               ( b + x = o.balance ∧ isDepositOp(op) )                      ∨
               ( b ≤ o.balance ∧ (isSeizeBond(op) || isWithdraw(op)) )
            }

    where isDepositOp is true for any of the deposit operations (deposit, delegate or topup).
*/
rule validChangeToBalanceOf(address o, method f) {
    env e;
    calldataarg args;
    require (o != currentContract);
    uint256 before = balanceOf(o);
    f(e, args);
    uint256 after = balanceOf(o);
    assert before == after ||
        // depositing case
        ( (f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
           f.selector==delegate(address, address, address).selector  ) &&
           after == before - e.msg.value ) ||
        // withdrawing case
           ( ( f.selector==seizeBond(address,uint256,uint256,address).selector ||
                f.selector==withdraw(uint256,address).selector) &&
           after >= before );
}

/**
    @title Valid change to unboundedValue ✔️
    @notice On Keep operation op by user s sending x amount, the unbounded value of operator o changes as follows:

        { u = unbondedValue(o) }
        op(s, x)
        {
            u = unbondedValue(o)                                              ∨
          ( u + x = unbondedValue(o) ∧ isDepositOp(op) )                    ∨
          ( l ≥ unbondedValue(o) ∧ isWithdraw(op) ∧ (s=o ∨ u=owner(s) )    ∨
          ( l ≥ unbondedValue(o) ∧ isCreateBond(op) )
        }
*/
rule validChangeToTotalBondAmount(address o, method f) {
    uint256 lockedBondsBefore = totalLockedBonds(o) ;
    uint256 unbondedBefore = unbondedValue(o);
    env e;
    calldataarg args;
    f(e, args);
    uint256 lockedBondsAfter = totalLockedBonds(o);
    assert lockedBondsBefore == lockedBondsAfter ||
           // case of unlocking
           ( (f.selector==freeBond(address,uint256).selector ||
              f.selector==seizeBond(address,uint256,uint256,address).selector ) &&
              lockedBondsAfter <= lockedBondsBefore )
           ||
           // case of locking
           ( f.selector==createBond(address,address,uint256,uint256,address).selector &&
             lockedBondsAfter >= lockedBondsBefore && lockedBondsAfter <= lockedBondsBefore + unbondedBefore )
           ;
}

/**
    @title Valid change to unboundedValue ✔️
    @notice On Keep operation op by user s sending x amount, the unbounded value of operator o changes as follows:

            { u = unbondedValue(o) }
            op(s, x)
            {
                u = unbondedValue(o)                                          ∨
              ( u + x = unbondedValue(o) ∧ isDepositOp(op) )                  ∨
              ( u ≥ unbondedValue(o) ∧ isWithdraw(op) ∧ (s=o ∨ u=owner(s) )  ∨
              ( u ≥ unbondedValue(o) ∧ isCreateBond(op) )
            }
*/
rule validChangeToUnbondedValue(address o, address owner,  method f) {
    uint256 unbondedBefore = unbondedValue(o);
    env e;
    calldataarg args;
    f(e, args);
    uint256 unbondedAfter = unbondedValue(o);
    assert  unbondedAfter  == unbondedBefore ||
        //cases for deposit
        (( f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
           f.selector==delegate(address, address, address).selector  ) &&
           unbondedAfter == unbondedBefore + e.msg.value ) ||
        //cases for withdraw
        ( f.selector==withdraw(uint256,address).selector && (e.msg.sender==o || e.msg.sender == ownerOf(o)) &&
          unbondedAfter <= unbondedBefore) ||
        // cases for holder free
        ( f.selector==freeBond(address,uint256).selector &&  unbondedAfter >= unbondedBefore ) ||
        // case for holder locking
        ( f.selector==createBond(address,address,uint256,uint256,address).selector &&
           unbondedAfter <= unbondedBefore );
}

/**
    @title Valid change to everDeposited ✔️
    On Keep operation op by user s sending x amount, the unbounded value of operator o changes as follows:
            { e = everDeposited(o) }
            op(s,x)
            {
                e = everDeposited(o)                            ∨
              ( e + x = everDeposited(o) ∧ isDepositOp(op) )
            }
*/
rule validChangeToEverDeposited(address o,  method f) {
    env e;
    calldataarg args;
    uint256 before = everDeposited(o);
    f(e, args);
    uint256 after = everDeposited(o);
    assert after == before ||
           ( ( f.selector==topUp(address).selector || f.selector==deposit(address).selector ||
               f.selector==delegate(address, address, address).selector ) &&
               after == before + e.msg.value );
}
