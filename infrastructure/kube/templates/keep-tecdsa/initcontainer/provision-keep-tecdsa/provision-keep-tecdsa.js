const fs = require('fs');
const toml = require('toml');
const tomlify = require('tomlify-j0.4');
const concat = require('concat-stream');
const Web3 = require('web3');
const HDWalletProvider = require("@truffle/hdwallet-provider");

// ETH host info
const ethRPCUrl = process.env.ETH_RPC_URL
const ethWSUrl = process.env.ETH_WS_URL
const ethNetworkId = process.env.ETH_NETWORK_ID;

// Contract owner info
var contractOwnerAddress = process.env.CONTRACT_OWNER_ETH_ACCOUNT_ADDRESS;
var authorizer = contractOwnerAddress
var purse = contractOwnerAddress

var contractOwnerProvider = new HDWalletProvider(process.env.CONTRACT_OWNER_ETH_ACCOUNT_PRIVATE_KEY, ethRPCUrl);

var operatorKeyFile = process.env.KEEP_TECDSA_ETH_KEYFILE_PATH

// LibP2P network info
const libp2pPeers = [process.env.KEEP_TECDSA_PEERS]
const libp2pPort = Number(process.env.KEEP_TECDSA_PORT)
const libp2pAnnouncedAddresses = [process.env.KEEP_TECDSA_ANNOUNCED_ADDRESSES]

/*
We override transactionConfirmationBlocks and transactionBlockTimeout because they're
25 and 50 blocks respectively at default.  The result of this on small private testnets
is long wait times for scripts to execute.
*/
const web3_options = {
  defaultBlock: 'latest',
  defaultGas: 4712388,
  transactionBlockTimeout: 25,
  transactionConfirmationBlocks: 3,
  transactionPollingTimeout: 480
};

const web3 = new Web3(contractOwnerProvider, null, web3_options);

/*
Each <contract.json> file is sourced directly from the InitContainer.  Files are generated by
Truffle during contract migration and copied to the InitContainer image via Circle.
*/
const bondedECDSAKeepFactory = getWeb3Contract('BondedECDSAKeepFactory');
const keepBondingContract = getWeb3Contract('KeepBonding');
const tokenStakingContract = getWeb3Contract('TokenStaking');
const keepTokenContract = getWeb3Contract('KeepToken');

// Address of the external TBTCSystem contract which should be set for the InitContainer
// execution.
const tbtcSystemContractAddress = process.env.TBTC_SYSTEM_ADDRESS;

// Returns a web3 contract object based on a truffle contract artifact JSON file.
function getWeb3Contract(contractName) {

  const filePath = `/tmp/${contractName}.json`;
  const parsed = JSON.parse(fs.readFileSync(filePath));
  const abi = parsed.abi;
  const address = parsed.networks[ethNetworkId].address;
  return new web3.eth.Contract(abi, address);
}

async function provisionKeepTecdsa() {

  try {

    console.log('###########  Provisioning keep-tecdsa! ###########');

    console.log(`\n<<<<<<<<<<<< Create Sortition Pool for TBTCSystem: ${tbtcSystemContractAddress} >>>>>>>>>>>>`);
    const sortitionPoolContractAddress = await createSortitionPool(tbtcSystemContractAddress);

    console.log(`\n<<<<<<<<<<<< Read operator address from key file >>>>>>>>>>>>`)
    const operatorAddress = readAddressFromKeyFile(operatorKeyFile)

    console.log(`\n<<<<<<<<<<<< Funding Operator Account ${operatorAddress} >>>>>>>>>>>>`)
    await fundOperator(operatorAddress, purse, '10')

    console.log(`\n<<<<<<<<<<<< Deposit to KeepBondingContract ${keepBondingContract.address} >>>>>>>>>>>>`)
    await depositUnbondedValue(operatorAddress, purse, '50')

    console.log(`\n<<<<<<<<<<<< Staking Operator Account ${operatorAddress} >>>>>>>>>>>>`)
    await stakeOperator(operatorAddress, contractOwnerAddress, authorizer)

    console.log(`\n<<<<<<<<<<<< Authorizing Operator Contract ${bondedECDSAKeepFactory.address} >>>>>>>>>>>>`)
    await authorizeOperatorContract(operatorAddress, bondedECDSAKeepFactory.address, authorizer)

    console.log(`\n<<<<<<<<<<<< Authorizing Sortition Pool Contract ${sortitionPoolContractAddress} >>>>>>>>>>>>`)
    await authorizeSortitionPoolContract(operatorAddress, sortitionPoolContractAddress, authorizer)

    console.log('\n<<<<<<<<<<<< Creating keep-tecdsa Config File >>>>>>>>>>>>');
    await createKeepTecdsaConfig();

    console.log("\n########### keep-tecdsa Provisioning Complete! ###########");
    process.exit()
  }
  catch (error) {
    console.error(error.message);
    throw error;
  }
};

function readAddressFromKeyFile(keyFilePath) {
  const keyFile = JSON.parse(fs.readFileSync(keyFilePath, 'utf8'))

  return web3.utils.toHex(keyFile.address)
}

async function fundOperator(operatorAddress, purse, requiredEtherBalance) {
  let requiredBalance = web3.utils.toWei(requiredEtherBalance, 'ether');

  const currentBalance = web3.utils.toBN(await web3.eth.getBalance(operatorAddress))
  if (currentBalance.gte(requiredBalance)) {
    console.log('Operator address is already funded, exiting!');
    return;
  }

  const transferAmount = requiredBalance.sub(currentBalance)

  console.log(`Funding account ${operatorAddress} with ${web3.utils.fromWei(transferAmount)} ether from purse ${purse}`);
  await web3.eth.sendTransaction({ from: purse, to: operatorAddress, value: transferAmount });
  console.log(`Account ${operatorAddress} funded!`);

};

async function depositUnbondedValue(operatorAddress, purse, etherToDeposit) {
  let transferAmount = web3.utils.toWei(etherToDeposit, 'ether');

  await keepBondingContract.methods.deposit(
    operatorAddress,
  ).send({ value: transferAmount, from: purse })

  console.log(`deposited ${etherToDeposit} ETH bonding value for operatorAddress ${operatorAddress}`)
}


async function isStaked(operatorAddress) {

  console.log('Checking if operator address is staked:');
  let stakedAmount = await tokenStakingContract.methods.balanceOf(operatorAddress).call();
  return stakedAmount != 0;
};

async function stakeOperator(operatorAddress, contractOwnerAddress, authorizer) {
  let staked = await isStaked(operatorAddress);

  /*
  We need to stake only in cases where an operator account is not already staked.  If the account
  is staked, or the client type is relay-requester we need to exit staking, albeit for different
  reasons.  In the case where the account is already staked, additional staking will fail.
  Clients of type relay-requester don't need to be staked to submit a request, they're acting more
  as a consumer of the network, rather than an operator.
  */
  if (staked === true) {
    console.log('Operator account already staked, exiting!');
    return;
  } else {
    console.log(`Staking 2000000 KEEP tokens on operator account ${operatorAddress}`);
  };

  let delegation = Buffer.concat([
    Buffer.from(web3.utils.hexToBytes(contractOwnerAddress)),
    Buffer.from(web3.utils.hexToBytes(operatorAddress)),
    Buffer.from(web3.utils.hexToBytes(authorizer))
  ])

  await keepTokenContract.methods.approveAndCall(
    tokenStakingContract.address,
    formatAmount(20000000, 18),
    delegation).send({ from: contractOwnerAddress })

  console.log(`Staked!`);
};

async function authorizeOperatorContract(operatorAddress, operatorContractAddress, authorizer) {
  console.log(`Authorizing Operator Contract ${operatorContractAddress} for operator account ${operatorAddress}`);

  await tokenStakingContract.methods.authorizeOperatorContract(operatorAddress, operatorContractAddress)
    .send({ from: authorizer })

  console.log(`Authorized!`);
};

async function authorizeSortitionPoolContract(operatorAddress, sortitionPoolContractAddress, authorizer) {

  console.log(`Authorizing Sortition Pool Contract ${sortitionPoolContractAddress} for operator account ${operatorAddress}`);

  await keepBondingContract.methods.authorizeSortitionPoolContract(operatorAddress, sortitionPoolContractAddress)
    .send({ from: authorizer })

  console.log(`Authorized!`);
};

async function createSortitionPool(applicationAddress) {
  const ADDRESS_ZERO = "0x0000000000000000000000000000000000000000"

  let sortitionPoolContractAddress

  const create = async () => {
    await bondedECDSAKeepFactory.methods.createSortitionPool(applicationAddress).send({ from: contractOwnerAddress });

    console.log(`created sortition pool for application: [${applicationAddress}]`);

    return await bondedECDSAKeepFactory.methods.getSortitionPool(applicationAddress).call();
  }

  sortitionPoolContractAddress = await bondedECDSAKeepFactory.methods.getSortitionPool(applicationAddress).call();

  if (!sortitionPoolContractAddress || sortitionPoolContractAddress == ADDRESS_ZERO) {
    console.log("sortition pool does not exists yet")
    sortitionPoolContractAddress = await create()
  } else {
    console.log(`sortition pool already exists for application: [${applicationAddress}]`)
  }

  console.log(`sortition pool contract address: ${sortitionPoolContractAddress}`);
  return sortitionPoolContractAddress
};

async function createKeepTecdsaConfig() {
  let parsedConfigFile = toml.parse(fs.readFileSync('/tmp/keep-tecdsa-config-template.toml', 'utf8'));

  parsedConfigFile.ethereum.URL = ethWSUrl;

  parsedConfigFile.ethereum.account.KeyFile = operatorKeyFile

  parsedConfigFile.ethereum.ContractAddresses.BondedECDSAKeepFactory = bondedECDSAKeepFactory.address;

  parsedConfigFile.SanctionedApplications.Addresses = [tbtcSystemContractAddress]

  parsedConfigFile.LibP2P.Peers = libp2pPeers
  parsedConfigFile.LibP2P.Port = libp2pPort
  parsedConfigFile.LibP2P.AnnouncedAddresses = libp2pAnnouncedAddresses

  parsedConfigFile.Storage.DataDir = process.env.KEEP_DATA_DIR

  /*
  tomlify.toToml() writes our Seed/Port values as a float.  The added precision renders our config
  file unreadable by the keep-client as it interprets 3919.0 as a string when it expects an int.
  Here we format the default rendering to write the config file with Seed/Port values as needed.
  */
  let formattedConfigFile = tomlify.toToml(parsedConfigFile, {
    space: 2,
    replace: (key, value) => { return (key == 'Port') ? value.toFixed(0) : false }
  })

  fs.writeFileSync('/mnt/keep-tecdsa/config/keep-tecdsa-config.toml', formattedConfigFile)
  console.log('keep-tecdsa config written to /mnt/keep-tecdsa/config/keep-tecdsa-config.toml');
};

/*
\heimdall aliens numbers.  Really though, the approveAndCall function expects numbers
in a particular format, this function facilitates that.
*/
function formatAmount(amount, decimals) {
  return web3.utils.toHex(web3.utils.toBN(amount).mul(web3.utils.toBN(10).pow(web3.utils.toBN(decimals))))
};

provisionKeepTecdsa().catch(error => {
  console.error(error);
  process.exit(1);
});
