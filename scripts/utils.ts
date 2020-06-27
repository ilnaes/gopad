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
