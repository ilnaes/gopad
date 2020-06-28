import { Mutex } from './utils.js'
import { Op } from './main.js'

let mu = new Mutex()
let log: Op[][] = []

async function handleMessage(e: MessageEvent) {
  let unlock = await mu.lock()
  log.push(e.data)
  console.log(log)
  unlock()
}

onmessage = (e) => {
  handleMessage(e).then(() => {})
}
