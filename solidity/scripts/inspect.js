const BondedECDSAKeepFactory = artifacts.require("BondedECDSAKeepFactory.sol")
const BondedECDSAKeep = artifacts.require("BondedECDSAKeep.sol")

module.exports = async function () {
    try {
        const factory = await BondedECDSAKeepFactory.at(
          "0x9EcCf03dFBDa6A5E50d7aBA14e0c60c2F6c575E6"
        )

        const keepCount = await factory.getKeepCount()
        console.log(`created keeps count: ${keepCount}`)

        const allOperators = new Set()
        const goodOperators = new Set()

        for (i = 0; i < keepCount; i++) {
            const keepAddress = await callWithRetry(() => factory.getKeepAtIndex(i))
            const keep = await BondedECDSAKeep.at(keepAddress) 
            const keepPublicKey = await callWithRetry(() => keep.publicKey())
            const members = await callWithRetry(() => keep.getMembers())
            const isActive = await callWithRetry(() => keep.isActive())
            const bond = await callWithRetry(()=> keep.checkBondAmount())

            console.log(`keep address: ${keepAddress}`)
            console.log(`keep index:   ${i}`)
            console.log(`pubkey:       ${keepPublicKey}`)
            console.log(`members:      ${members}`)
            console.log(`isActive:     ${isActive}`)
            console.log(`bond [wei]:   ${bond}`)
            console.log(`bond [eth]:   ${web3.utils.fromWei(bond)}`)

            members.forEach((member) => allOperators.add(member))
            if (keepPublicKey) {
                members.forEach((member) => goodOperators.add(member))
            }

            console.log(``)
        }

        // all operators who are members of keeps
        console.log(`all operators = ${new Array(...allOperators).join(', ')}`)
        console.log(``)

        // if the operator is a member of at least one keep which generated
        // a public key, it's here
        console.log(`good operators = ${new Array(...goodOperators).join(', ')}`)
        console.log(``)

        // if the operator is a member of at least one keep and that operator 
        // is NOT a member of at least one keep which successfully generated
        // a public key, this operator is here        
        let potentiallyBadOperators = new Set(allOperators)
        for (let goodOperator of goodOperators) {
            potentiallyBadOperators.delete(goodOperator)
        }
        console.log(`potentially bad operators = ${new Array(...potentiallyBadOperators).join(', ')}`)

        process.exit()
    } catch (error) {
      console.log(error)
    }
}

async function callWithRetry(fn) {
    try {
        return await fn()
    } catch (error) {
        console.log(`Error ${error} occurred; retrying...`)
        return await fn()
    }
}