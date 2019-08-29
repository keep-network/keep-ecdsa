/**
 *  Resolves when a given event happens or rejects after a timeout. 
 *  @param event Contract event to watch for.
 *  @param timeout Time to wait in milliseconds (default 5000 ms).
 *  @return Promise which resolves when the event is seen.
 */
async function waitForEvent(event, timeout = 5000) {
    return new Promise((resolve, reject) => {
        let timeoutSet = setTimeout(
            () => {
                clearTimeout(timeoutSet)
                return reject(new Error('Timeout waiting for event'))
            },
            timeout
        )

        event.on('data', result => {
            clearTimeout(timeoutSet)
            return resolve(result)
        })
    })
}

module.exports = waitForEvent
