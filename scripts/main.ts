import { App } from './app'

export type Op = {
  Loc: number
  Ch: number
  Add: boolean

  DocId: number
  Uid: number
  Seq: number
}

export type State = {
  Body: string
  Selstart: number
  Selend: number
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

  const worker = new Worker('/static/worker.js')

  // let pad = document.querySelector('#textbox')
}

window.addEventListener('load', main)
