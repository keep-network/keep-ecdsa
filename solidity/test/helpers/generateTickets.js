import packTicket from './packTicket';

export default function generateTickets(randomBeaconValue, stakerValue, stakerWeight) {
    let tickets = [];
    for (let i = 1; i <= stakerWeight; i++) {
      let ticketValueHex = web3.utils.soliditySha3({t: 'uint', v: randomBeaconValue}, {t: 'uint', v: stakerValue}, {t: 'uint', v: i})
      tickets.push(packTicket(ticketValueHex, i, stakerValue));
    }
    return tickets
  }