import { Op } from './main.js'
import { diff } from './utils.js'

async function handleMessage(e: MessageEvent) {
  let [uid, seq, prev, curr] = e.data
  postMessage([diff(prev, curr, seq, uid), seq])
}

onmessage = (e) => {
  handleMessage(e)
}
