import { Op } from './main.js'
import { diff, applyPos, applyString, xform } from './utils.js'

export type WorkerRet = {
  val: string
  seq: number | undefined
  view: number
  base: string
  curr: string
  val1: string
  delta: Op[]
  delta1: Op[]
  pos: [number, number]
}

export type WorkerArg = {
  ops: Op[][]
  uid: number
  base: string
  view: number
  delta: Op[]
  curr: string
  val: string
  pos: [number, number]
}

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
  const args: WorkerArg = e.data

  let delta1 = diff(args.base, args.val, args.uid)
  let seq1 = undefined

  for (let i = 0; i < args.ops.length; i++) {
    if (args.ops[i][0].Uid == args.uid) {
      // found delta so change delta1 to not
      // incorporate it
      if (args.ops[i][0].Seq !== undefined) {
        seq1 = args.ops[i][0].Seq
      }

      const val1 = applyString(args.base, delta1)
      args.base = applyString(args.base, args.ops[i])
      delta1 = diff(args.base, val1, args.uid)
    } else {
      args.delta = xform(args.ops[i], args.delta)
      delta1 = xform(args.ops[i], delta1)

      args.base = applyString(args.base, args.ops[i])
      args.pos = applyPos(args.pos, args.ops[i])
    }
  }

  if (seq1 == undefined) {
    args.curr = applyString(args.base, args.delta)
  } else {
    args.curr = args.base
  }

  const val1 = applyString(args.base, delta1)

  if (seq1 == undefined) {
    delta1 = diff(args.curr, val1, args.uid)
  }

  postMessage({
    val: args.val,
    seq: seq1,
    view: args.view + args.ops.length,
    base: args.base,
    curr: args.curr,
    val1: val1,
    delta: args.delta,
    delta1: delta1,
    pos: args.pos,
  } as WorkerRet)
}

onmessage = (e) => {
  handleMessage(e)
}
