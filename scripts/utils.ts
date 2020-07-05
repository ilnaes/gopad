import { Op, State } from './main.js'

// diff that turns s1 -> s2
export function diff(
  s1: string,
  s2: string,
  seq: number,
  uid: number,
  view: number
): Op[] {
  let dp = new Array(s1.length + 1)
  dp[0] = Array.from(Array(s2.length + 1).keys())

  // DP to calculate diff
  // TODO: trim beginning and end

  for (let i = 1; i < s1.length + 1; i++) {
    dp[i] = new Array(s2.length + 1)
    dp[i][0] = i

    for (let j = 1; j < s2.length + 1; j++) {
      dp[i][j] = Math.min(dp[i][j - 1], dp[i - 1][j]) + 1

      if (s1[i - 1] == s2[j - 1] && dp[i - 1][j - 1] < dp[i][j]) {
        dp[i][j] = dp[i - 1][j - 1]
      }
    }
  }

  let i = s1.length
  let j = s2.length

  let res: Op[] = []

  // collect diff into slice
  while (i > 0 || j > 0) {
    if (i == 0) {
      res.push({
        Add: true,
        Loc: i,
        Ch: s2.charCodeAt(j - 1),
        Seq: seq,
        Uid: uid,
        View: view,
      })
      j--
    } else if (j == 0) {
      res.push({
        Add: false,
        Loc: i - 1,
        Ch: s1.charCodeAt(i - 1),
        Seq: seq,
        Uid: uid,
        View: view,
      })
      i--
    } else {
      if (s1[i - 1] == s2[j - 1] && dp[i][j] == dp[i - 1][j - 1]) {
        i--
        j--
      } else {
        if (dp[i][j] == dp[i][j - 1] + 1) {
          // Add s2[j-1]
          res.push({
            Add: true,
            Loc: i,
            Ch: s2.charCodeAt(j - 1),
            Seq: seq,
            Uid: uid,
            View: view,
          })
          j--
        } else {
          // Delete s1[i-1]
          res.push({
            Add: false,
            Loc: i - 1,
            Ch: s1.charCodeAt(i - 1),
            Seq: seq,
            Uid: uid,
            View: view,
          })
          i--
        }
      }
    }
  }

  return res.reverse()
}

export function xform(o1: Op[], o2: Op[]): Op[] {
  let res: Op[] = []

  let i = 0
  let j = 0
  let delta = 0

  while (j < o2.length) {
    if (i == o1.length || o1[i].Loc > o2[j].Loc) {
      res.push(o2[j])
      res[res.length - 1].Loc += delta
      j++
    } else if (o1[i].Loc == o2[j].Loc) {
      if (!o1[i].Add && !o2[j].Add) {
        // two deletes so skip
        j++
        i++
        delta--
      } else {
        // do Add first
        if (o1[i].Add) {
          delta++
          i++
        } else {
          res.push(o2[j])
          res[res.length - 1].Loc += delta
          j++
        }
      }
    } else {
      if (o1[i].Add) {
        delta++
      } else {
        delta--
      }
      i++
    }
  }

  return res
}

export function applyPos(pos: [number, number], ops: Op[]): [number, number] {
  let res: [number, number] = [...pos]
  for (let j = 0; j < ops.length; j++) {
    let op = ops[j]

    if (op.Loc >= pos[1]) {
      break
    }

    if (op.Add) {
      if (pos[0] >= op.Loc) {
        res[0] += 1
      }
      res[1] += 1
    } else {
      if (pos[0] > op.Loc) {
        res[0] -= 1
      }
      res[1] -= 1
    }
  }

  return res
}

export function applyString(base: string, ops: Op[]): string {
  let res = ''

  let i = 0
  for (let j = 0; j < ops.length; j++) {
    let op = ops[j]
    res += base.substring(i, op.Loc)
    i = op.Loc

    if (op.Add) {
      res += String.fromCharCode(op.Ch)
    } else {
      i++
    }
  }

  if (i < base.length) {
    res += base.substring(i)
  }

  return res
}

export function sleep(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}
