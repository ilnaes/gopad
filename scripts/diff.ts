import { Op, State } from './main'

function diff(s1: string, s2: string): Op[] {
  return []
}

function apply(ops: Op[], s: State): State {
  return {
    Body: s.Body,
    Selstart: s.Selstart,
    Selend: s.Selend,
  }
}
