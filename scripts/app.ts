import { Req, Res, Op } from './main.js'
import { applyString, xform, sleep } from './utils.js'

const PULL_INTERVAL = 200

export class App {
  docId: number = -1
  uid: number = -1

  view: number = -1
  seq: number = 0
  discardPoint: number = 0
  commitPoint: number = 0
  opsPoint: number = 0

  base: string = ''
  prev: string = ''
  ops: Op[][] = []

  textbox: HTMLTextAreaElement
  ws: WebSocket
  worker: Worker

  constructor(docId: number) {
    this.docId = docId
    this.uid = Math.floor(Math.random() * 1e18)

    this.worker = new Worker('../static/worker.js', { type: 'module' })
    this.worker.onmessage = (e) => {
      this.handleWorker(e)
    }

    this.textbox = document.querySelector('#textbox') as HTMLTextAreaElement
    this.textbox.addEventListener('input', (e) => this.handleEvent())

    this.ws = new WebSocket('ws://localhost:8080/ws/' + docId.toString())
    this.ws.addEventListener('open', () => {
      this.ws.send(this.uid.toString())
      this.poll()
      this.commit()
    })
    this.ws.addEventListener('message', (e) => this.handleResp(e))
  }

  async commit() {
    while (true) {
      if (this.ops.length > 0) {
        let req: Req = {
          IsQuery: false,
          View: this.view,
          DocId: this.docId,
          Uid: this.uid,
          Ops: this.ops,
        }
        this.ws.send(JSON.stringify(req))
      }

      await sleep(PULL_INTERVAL)
    }
  }

  // enqueues an Op sequence from Worker
  async handleWorker(e: MessageEvent) {
    let [ops, seq]: [Op[], number] = e.data
    while (true) {
      // make sure to push in order
      if (this.opsPoint == seq) {
        this.ops.push(ops)
        this.opsPoint++
        break
      }

      await sleep(100)
    }
  }

  // when textbox changes ask Worker to compute diff
  async handleEvent() {
    this.worker.postMessage([this.uid, this.seq, this.prev, this.textbox.value])
    this.prev = this.textbox.value
    this.seq++
  }

  // continuously query the server
  async poll() {
    while (true) {
      let req: Req = {
        IsQuery: true,
        DocId: this.docId,
        Uid: this.uid,
        View: this.view,
      }
      this.ws.send(JSON.stringify(req))

      await sleep(PULL_INTERVAL)
    }
  }

  async handleResp(event: MessageEvent) {
    let resp: Res = JSON.parse(event.data)
    console.log(resp)

    if (this.view == -1 && resp.Type != 'DocRes') {
      return
    }

    switch (resp.Type) {
      case 'DocRes': {
        if (resp.View > this.view) {
          this.view = resp.View
          this.seq = resp.Seq
          this.base = resp.Body
          this.prev = resp.Body

          // TODO: diff and xform
          this.textbox.disabled = false
          this.textbox.value = resp.Body
        }
        break
      }
      case 'OpsRes': {
        console.log(resp.Ops)
        if (this.view < resp.View) {
          break
        }

        let pos: [number, number] = [
          this.textbox.selectionStart,
          this.textbox.selectionEnd,
        ]

        if (this.view > resp.View + resp.Ops.length) {
          // remove ops that have been seen
          for (let i = resp.Ops.length - 1; i >= 0; i--) {
            if (resp.Ops[i][0].Uid == this.uid) {
              this.ops = this.ops.splice(
                resp.Ops[i][0].Seq - (this.seq - this.ops.length)
              )
              break
            }
          }

          // apply to base and xform log
          for (let i = this.view - resp.View; i < resp.Ops.length; i++) {
            this.base = applyString(this.base, resp.Ops[i])

            if (resp.Ops[i][0].Uid != this.uid) {
              for (let j = 0; j < this.ops.length; j++) {
                this.ops[j] = xform(resp.Ops[i], this.ops[j])
              }
            }
          }

          // update textbox
          this.textbox.value = this.base
          for (let i = 0; i < this.ops.length; i++) {
            // TODO:
          }
        }
        break
      }
    }
  }
}
