import { App } from './app'

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

  Uid: number
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
  Uid: number
  View: number

  Seq: number
  Ops?: Op[]
}

function main() {
  const split = document.location.pathname.lastIndexOf('/')
  const docId = parseInt(document.location.pathname.slice(split + 1))

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
  new App(docId)
}

window.addEventListener('load', main)
