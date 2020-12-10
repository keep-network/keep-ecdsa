import BigNumber from "bignumber.js"

export const decimalPlaces = 2
export const noDecimalPlaces = 0

export const shorten18Decimals = (value) =>
  new BigNumber(value).dividedBy(new BigNumber(1e18)).toFixed(decimalPlaces)
