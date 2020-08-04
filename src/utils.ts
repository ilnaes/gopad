import { Op } from './index'

// diff that turns s1 -> s2
export function diff(s1: string, s2: string, uid: number, sess: number): Op[] {
  // trim beginning
  let delta = 0
  for (let i = 0; i < Math.min(s1.length, s2.length); i++) {
    if (s1[i] != s2[i]) {
      s1 = s1.substring(i)
      s2 = s2.substring(i)
      delta = i
      break
    }
  }

  // trim end
  for (let i = 0; i < Math.min(s1.length, s2.length); i++) {
    if (s1[s1.length - 1 - i] != s2[s2.length - 1 - i]) {
      s1 = s1.substring(0, s1.length - i)
      s2 = s2.substring(0, s2.length - i)
      break
    }
  }

  const dp = new Array(s1.length + 1)
  dp[0] = Array.from(Array(s2.length + 1).keys())

  // DP to calculate diff
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

  const res: Op[] = []

  // collect diff into slice
  while (i > 0 || j > 0) {
    if (i == 0) {
      res.push({
        Type: 'Add',
        Loc: delta + i,
        Ch: s2.charCodeAt(j - 1),
        Uid: uid,
        Session: sess,
      })
      j--
    } else if (j == 0) {
      res.push({
        Type: 'Del',
        Loc: delta + i - 1,
        Ch: s1.charCodeAt(i - 1),
        Uid: uid,
        Session: sess,
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
            Type: 'Add',
            Loc: delta + i,
            Ch: s2.charCodeAt(j - 1),
            Uid: uid,
            Session: sess,
          })
          j--
        } else {
          // Delete s1[i-1]
          res.push({
            Type: 'Del',
            Loc: delta + i - 1,
            Ch: s1.charCodeAt(i - 1),
            Uid: uid,
            Session: sess,
          })
          i--
        }
      }
    }
  }

  return res.reverse()
}

// modify o2 to take into account o1 happening first
export function xform(o1: Op[], o2: Op[]): Op[] {
  const res: Op[] = []

  let i = 0
  let j = 0
  let delta = 0

  while (j < o2.length) {
    if (i == o1.length || o1[i].Loc > o2[j].Loc) {
      res.push(o2[j])
      res[res.length - 1].Loc += delta
      j++
    } else if (o1[i].Loc == o2[j].Loc) {
      if (o1[i].Type == 'Del' && o2[j].Type == 'Del') {
        // two deletes turn to NOOP
        o2[j].Type = 'NOOP'
        res.push(o2[j])
        j++
        i++
        delta--
      } else {
        // do Add first
        if (o1[i].Type == 'Add') {
          delta++
          i++
        } else {
          res.push(o2[j])
          res[res.length - 1].Loc += delta
          j++
        }
      }
    } else {
      if (o1[i].Type == 'Add') {
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
  const res: [number, number] = [...pos]
  for (let j = 0; j < ops.length; j++) {
    const op = ops[j]

    if (op.Loc >= pos[1]) {
      break
    }

    if (op.Type == 'Add') {
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
    const op = ops[j]
    if (op.Type == 'NOOP') {
      continue
    }
    res += base.substring(i, op.Loc)
    i = op.Loc

    if (op.Type == 'Add') {
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

export async function sleep(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}
