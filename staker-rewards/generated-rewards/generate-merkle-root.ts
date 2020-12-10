import { parseBalanceMap } from '../merkle-distributor/src/parse-balance-map'
import { program } from 'commander'
import * as fs from 'fs'


program
  .version('0.0.0')
  .requiredOption(
    '-i, --input <path>',
    'input JSON file location containing a map of account addresses to string balances'
  )

program.parse(process.argv)

const output_merkle_objects = './output-merkle-objects.json'

// read existing merkle objects if any
let merkleObjects = {}
if (fs.existsSync(output_merkle_objects)) {
  merkleObjects = JSON.parse(fs.readFileSync(output_merkle_objects, { encoding: 'utf8' }))
}

// new rewards for merkle interval
const json = JSON.parse(fs.readFileSync(program.input, { encoding: 'utf8' }))
if (typeof json !== 'object') throw new Error('Invalid JSON')

const merkleObject = parseBalanceMap(json)
const totalAndClaims = {
  tokenTotal: merkleObject.tokenTotal,
  claims: merkleObject.claims
}
merkleObjects[merkleObject.merkleRoot] = totalAndClaims

fs.writeFileSync(output_merkle_objects, JSON.stringify(merkleObjects, null, 2))
