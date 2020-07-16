import { Op } from './main.js'
import { diff, applyString, xform } from './utils.js'

// applies ops to state
// arguments:
// ops - ops from server to be considered
// uid - uid
// base - document known to be on server
// view - view of base
// delta - current outstanding op commit
// curr - applyString(delta, base)
// val - textbox value
function handleMessage(e: MessageEvent) {
  let [ops, uid, base, view, delta, curr, val]: [
    Op[][],
    number,
    string,
    number,
    Op[],
    string,
    string
  ] = e.data

  let delta1 = diff(curr, val, uid)
  let seq1 = -1

  for (let i = 0; i < ops.length; i++) {
    base = applyString(base, ops[i])

    if (ops[i][0].Uid == uid) {
      // found delta
      seq1 = ops[i][0].Seq!
    } else {
      delta = xform(ops[i], delta)
      delta1 = xform(ops[i], delta1)
      //     pos = applyPos(pos, resp.Ops[i])
    }
  }

  if (seq1 == -1) {
    curr = applyString(base, delta)
    //     pos = applyPos(pos, resp.Ops[i])
  } else {
    curr = base
  }

  let val1 = applyString(curr, delta1)
  //     pos = applyPos(pos, resp.Ops[i])

  postMessage([val, seq1, view + ops.length, base, curr, val1, delta, delta1])
}

onmessage = (e) => {
  handleMessage(e)
}
