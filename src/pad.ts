import { sleep } from './utils'
import type { WorkerArg, WorkerRet } from './worker'
import DiffWorker from 'worker-loader!./worker'

const PULL_INTERVAL = 1000

export class State {
  body = ''
  selStart = 0
  selEnd = 0
}

export type Op = {
  Loc: number
  Ch: number
  Type: string
  Seq?: number

  Uid: string
  Session: number
}

export type Res = {
  Type: string
  Body: string
  View: number
  Seq: number
  Ops: Op[][]
}

export type Req = {
  IsQuery: boolean
  DocId: number
  Uid: string
  View: number

  Seq: number
  Ops?: Op[]
}

export function start(docId: number, uid: string, token: string): Pad {
  const textarea: HTMLTextAreaElement = document.getElementsByTagName(
    'textarea'
  )[0]
  textarea.onkeydown = function (e) {
    if (e.key == 'Tab') {
      e.preventDefault()
      const s = textarea.selectionStart
      textarea.value =
        textarea.value.substring(0, textarea.selectionStart) +
        '\t' +
        textarea.value.substring(textarea.selectionEnd)
      textarea.selectionEnd = s + 1
    }
  }
  return new Pad(docId, uid, token)
}

export class Pad {
  alive = true
  docId: number
  uid: string
  session: number
  token: string

  pollStart = false
  commitStart = false

  base = ''
  curr = ''
  delta: Op[] = []
  ops: Op[][] = []

  view = -1 // view of base
  seq = 0
  seenseq = -1

  textbox: HTMLTextAreaElement
  ws?: WebSocket
  worker: Worker

  constructor(docId: number, uid: string, token: string) {
    this.docId = docId
    this.session = Math.floor(Math.random() * 1e18)
    this.uid = uid

    this.token = token

    this.worker = new DiffWorker()
    this.worker.onmessage = (e) => {
      this.handleWorker(e)
    }

    this.textbox = document.querySelector('#textbox') as HTMLTextAreaElement

    this.connect()
  }

  kill(): void {
    this.alive = false
    this.ws.close()
  }

  connect(): void {
    this.ws = new WebSocket('ws://localhost:8080/ws/' + this.docId.toString())

    this.ws.onopen = () => {
      if (this.ws !== undefined) {
        this.ws.send(this.token)
        this.poll()
        this.commit()
      }
    }

    this.ws.onmessage = (e) => this.handleResp(e)

    this.ws.onclose = () => {
      if (this.alive) {
        this.connect()
      }
    }

    this.ws.onerror = () => {
      if (this.ws !== undefined) {
        this.ws.close()
      }
    }
  }

  // continuously query the server
  async poll(): Promise<void> {
    if (this.pollStart) {
      return
    }

    this.pollStart = true
    while (this.alive) {
      if (this.ws && this.ws.readyState == WebSocket.OPEN) {
        const req: Req = {
          IsQuery: true,
          DocId: this.docId,
          Uid: this.uid,
          View: this.view,
          Seq: 0,
        }
        try {
          this.ws.send(JSON.stringify(req))
        } catch (_) {
          this.pollStart = false
          break
        }
      }

      await sleep(PULL_INTERVAL)
    }
  }

  // push changes to server
  async commit(): Promise<void> {
    if (this.commitStart) {
      return
    }

    this.commitStart = true
    while (this.alive) {
      if (this.ws && this.ws.readyState == WebSocket.OPEN) {
        if (this.delta.length > 0) {
          this.delta[0].Seq = this.seq

          const req: Req = {
            IsQuery: false,
            View: this.view,
            DocId: this.docId,
            Uid: this.uid,

            Seq: this.seq,
            Ops: this.delta,
          }
          try {
            this.ws.send(JSON.stringify(req))
          } catch (_) {
            this.commitStart = false
            break
          }
        }
      }

      this.updateState()
      await sleep(PULL_INTERVAL)
    }
  }

  // processes the changes to state from a Worker
  handleWorker(e: MessageEvent): void {
    const {
      val,
      seq,
      view,
      base,
      curr,
      val1,
      delta,
      delta1,
      pos,
      found,
    } = e.data as WorkerRet
    // delta: base -> curr
    // delta1: curr -> val1

    if (this.textbox.value == val && this.view <= view) {
      this.base = base
      this.textbox.value = val1
      this.textbox.setSelectionRange(pos[0], pos[1])

      this.ops = this.ops.splice(view - this.view)
      this.view = view

      if (seq !== undefined && seq > this.seenseq) {
        this.seenseq = seq
        this.seq = seq + 1
        console.log(this.seq)
      }

      if (found || (this.delta.length == 0 && delta1.length != 0)) {
        // create a new delta to send
        this.delta = delta1
        this.base = curr
        this.curr = val1
      } else {
        this.delta = delta
        this.curr = curr
      }
    }
  }

  // sends current state with ops to be applied
  updateState(): void {
    const arg: WorkerArg = {
      ops: this.ops,
      uid: this.uid,
      session: this.session,
      base: this.base,
      view: this.view,
      delta: this.delta,
      curr: this.curr,
      val: this.textbox.value,
      pos: [this.textbox.selectionStart, this.textbox.selectionEnd],
    }
    this.worker.postMessage(arg)
  }

  handleResp(event: MessageEvent): void {
    const resp: Res = JSON.parse(event.data)

    if (this.view == -1 && resp.Type != 'DocRes') {
      return
    }

    switch (resp.Type) {
      case 'DocRes': {
        // we are starting from new
        this.view = resp.View
        this.seq = resp.Seq
        this.base = resp.Body

        this.textbox.disabled = false
        this.textbox.value = resp.Body
        this.curr = resp.Body
        break
      }
      case 'OpsRes': {
        if (this.view + this.ops.length < resp.View) {
          break
        }

        if (this.view + this.ops.length < resp.View + resp.Ops.length) {
          // enqueue all new ops
          for (
            let i = this.view + this.ops.length - resp.View;
            i < resp.Ops.length;
            i++
          ) {
            this.ops.push(resp.Ops[i])
          }
        }
        break
      }
    }
  }
}
