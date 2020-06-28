import { App } from './app.js'

export class State {
  body = ''
  selStart = 0
  selEnd = 0
}

export type Op = {
  Loc: number
  Ch: number
  Add: boolean

  Seq: number
  Uid: number
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
  Uid: number
  View: number
  Ops?: Op[][]
}

function main() {
  let split = document.location.pathname.lastIndexOf('/')
  let docId = parseInt(document.location.pathname.slice(split + 1))
  let app = new App(docId)
}

window.addEventListener('load', main)
