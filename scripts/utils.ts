import { Op, State } from './main.js'

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
  return [0, 0]
}

export function applyString(base: string, ops: Op[]): string {
  let res = ''

  let i = 0
  for (let j = 0; j < ops.length; j++) {
    let op = ops[j]
    res += base.substring(i, op.Loc)
    i = op.Loc

    if (op.Add) {
      res += op.Ch
    } else {
      i++
    }
  }

  if (i < base.length) {
    res += base.substring(i)
  }

  return res
}

// taken from https://spin.atomicobject.com/2018/09/10/javascript-concurrency/
export class Mutex {
  private mutex = Promise.resolve()

  lock(): PromiseLike<() => void> {
    let begin: (unlock: () => void) => void = (unlock) => {}

    this.mutex = this.mutex.then(() => {
      return new Promise(begin)
    })

    return new Promise((res) => {
      begin = res
    })
  }

  async dispatch<T>(fn: (() => T) | (() => PromiseLike<T>)): Promise<T> {
    const unlock = await this.lock()
    try {
      return await Promise.resolve(fn())
    } finally {
      unlock()
    }
  }
}

export function sleep(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}
