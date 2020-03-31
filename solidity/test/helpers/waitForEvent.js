/**
 *  Resolves when a given event happens or rejects after a timeout.
 *  @param {Event} event Contract event to watch for.
 *  @param {number} timeout Time to wait in milliseconds (default 5000 ms).
 *  @return {Promise} Promise which resolves when the event is seen.
 */
async function waitForEvent(event, timeout = 5000) {
  return new Promise((resolve, reject) => {
    const timeoutSet = setTimeout(() => {
      clearTimeout(timeoutSet)
      return reject(new Error("Timeout waiting for event"))
    }, timeout)

    event.on("data", (result) => {
      clearTimeout(timeoutSet)
      return resolve(result)
    })
  })
}

module.exports = waitForEvent
