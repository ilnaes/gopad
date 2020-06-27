import { Req, Res, Op } from './main'
import { sleep, Mutex } from './utils'

const PULL_INTERVAL = 1000

export class App {
  docId: number = -1
  uid: number = -1
  view: number = -1
  seq: number = 0
  discardPoint: number = 0
  commitPoint: number = 0
  body: string = ''
  ops: Op[] = []

  mu: Mutex
  textbox: HTMLTextAreaElement
  ws: WebSocket

  constructor(docId: number) {
    this.docId = docId
    this.uid = Math.floor(Math.random() * 1e19)

    this.mu = new Mutex()
    this.textbox = document.querySelector('#textbox') as HTMLTextAreaElement
    this.ws = new WebSocket('ws://localhost:8080/ws/' + docId.toString())

    this.ws.addEventListener('open', () => {
      this.ws.send(this.uid.toString())
      this.poll().then(() => {})
    })
    this.ws.addEventListener('message', this.handle)
  }

  // continuously query the server
  async poll() {
    while (true) {
      let unlock = await this.mu.lock()
      let req: Req = {
        IsQuery: true,
        DocId: this.docId,
        Uid: this.uid,
        View: this.view,
      }
      unlock()

      this.ws.send(JSON.stringify(req))
      await sleep(PULL_INTERVAL)
    }
  }

  async handle(event: MessageEvent) {
    let resp: Res = JSON.parse(event.data)

    let unlock = await this.mu.lock()

    if (this.view == -1 && resp.Type != 'DocRes') {
      unlock()
      return
    }

    switch (resp.Type) {
      case 'DocRes': {
        this.view = resp.View
        this.seq = resp.Seq
        this.body = resp.Body
      }
      case 'OpsRes': {
      }
      case 'Ack': {
      }
    }

    unlock()
  }
}
