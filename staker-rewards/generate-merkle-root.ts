import { program } from 'commander'

import * as fs from 'fs'
import { parseBalanceMap } from './merkle-distributor/src/parse-balance-map'

program
  .version('0.0.0')
  .requiredOption(
    '-i, --input <path>',
    'input JSON file location containing a map of account addresses to string balances'
  )

program.parse(process.argv)

const json = JSON.parse(fs.readFileSync(program.input, { encoding: 'utf8' }))

if (typeof json !== 'object') throw new Error('Invalid JSON')

fs.writeFileSync('./output-merkle-object.json', JSON.stringify(parseBalanceMap(json), null, 2))
