import { Op } from './main.js'
import { diff } from './utils.js'

async function handleMessage(e: MessageEvent) {
  let [view, uid, seq, prev, curr] = e.data
  postMessage([diff(prev, curr, seq, uid, view), seq])
}

onmessage = (e) => {
  handleMessage(e)
}
