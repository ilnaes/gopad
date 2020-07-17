import { Op } from './main.js'
import { diff, applyPos, applyString, xform } from './utils.js'

export type WorkerRet = [
  string,
  number,
  number,
  string,
  string,
  string,
  Op[],
  Op[],
  [number, number]
]

export type WorkerArg = [
  Op[][],
  number,
  string,
  number,
  Op[],
  string,
  string,
  [number, number]
]

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
  let [ops, uid, base, view, delta, curr, val, pos]: WorkerArg = e.data

  let delta1 = diff(base, val, uid)
  let seq1 = -1
  let val1 = ''

  for (let i = 0; i < ops.length; i++) {
    if (ops[i][0].Uid == uid) {
      // found delta so change delta1 to not
      // incorporate it
      seq1 = ops[i][0].Seq!

      val1 = applyString(base, delta1)
      base = applyString(base, ops[i])
      delta1 = diff(base, val1, uid)
    } else {
      delta = xform(ops[i], delta)
      delta1 = xform(ops[i], delta1)

      base = applyString(base, ops[i])
      pos = applyPos(pos, ops[i])
    }
  }

  if (seq1 == -1) {
    curr = applyString(base, delta)
  } else {
    curr = base
  }

  val1 = applyString(base, delta1)

  if (seq1 == -1) {
    delta1 = diff(curr, val1, uid)
  }

  postMessage([
    val,
    seq1,
    view + ops.length,
    base,
    curr,
    val1,
    delta,
    delta1,
    pos,
  ] as WorkerRet)
}

onmessage = (e) => {
  handleMessage(e)
}
