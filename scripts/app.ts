import { Req, Res, Op } from './main.js'
import { diff, applyString, applyPos, xform, sleep } from './utils.js'

const PULL_INTERVAL = 1000

export class App {
  docId: number = -1
  uid: number = -1

  pollStart: boolean = false
  commitStart: boolean = false

  base: string = ''
  curr: string = ''
  delta: Op[] = []
  ops: Op[][] = []

  view: number = -1 // view of base
  seq: number = 0
  seenseq: number = -1

  textbox: HTMLTextAreaElement
  ws?: WebSocket
  worker: Worker

  constructor(docId: number) {
    this.docId = docId
    this.uid = Math.floor(Math.random() * 1e18)

    this.worker = new Worker('../static/worker.js', { type: 'module' })
    this.worker.onmessage = (e) => {
      this.handleWorker(e)
    }

    this.textbox = document.querySelector('#textbox') as HTMLTextAreaElement

    this.connect()
  }

  connect() {
    this.ws = new WebSocket('ws://localhost:8080/ws/' + this.docId.toString())

    this.ws.onopen = () => {
      this.ws!.send(this.uid.toString())
      this.poll()
      this.commit()
    }

    this.ws.onmessage = (e) => this.handleResp(e)

    this.ws.onclose = () => {
      this.connect()
    }

    this.ws.onerror = () => {
      this.ws!.close()
    }
  }

  // continuously query the server
  async poll() {
    if (this.pollStart) {
      return
    }

    this.pollStart = true
    while (true) {
      if (this.ws && this.ws.readyState == WebSocket.OPEN) {
        let req: Req = {
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
  async commit() {
    if (this.commitStart) {
      return
    }

    this.commitStart = true
    while (true) {
      if (this.ws && this.ws.readyState == WebSocket.OPEN) {
        if (this.delta.length > 0) {
          this.delta[0].Seq = this.seq

          let req: Req = {
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
  handleWorker(e: MessageEvent) {
    let [val, seq, view, base, curr, val1, delta, delta1]: [
      string,
      number,
      number,
      string,
      string,
      string,
      Op[],
      Op[]
    ] = e.data
    // delta: base -> curr
    // delta1: curr -> val1

    if (this.textbox.value == val && this.view <= view) {
      this.base = base
      this.textbox.value = val1
      // this.textbox.setSelectionRange(pos[0], pos[1])

      this.ops = this.ops.splice(view - this.view)
      this.view = view

      if (
        seq > this.seenseq ||
        (this.delta.length == 0 && delta1.length != 0)
      ) {
        // create a new delta to send
        this.delta = delta1
        this.base = curr
        this.curr = val1

        if (seq > this.seenseq) {
          this.seenseq = seq
          this.seq = seq + 1
          console.log(this.seq)
        }
      } else {
        this.delta = delta
        this.curr = curr
      }
    }
  }

  // sends current state with ops to be applied
  updateState() {
    // let pos: [number, number] = [
    //   this.textbox.selectionStart,
    //   this.textbox.selectionEnd,
    // ]
    this.worker.postMessage([
      this.ops,
      this.uid,
      this.base,
      this.view,
      this.delta,
      this.curr,
      this.textbox.value,
    ])
  }

  handleResp(event: MessageEvent) {
    let resp: Res = JSON.parse(event.data)

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
