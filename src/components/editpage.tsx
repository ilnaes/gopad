import * as React from 'react'
import { useEffect, useState, useContext } from 'react'
import { UserContext } from '../contexts/usercontext'
import { start, Pad } from '../pad'
import { useParams } from 'react-router-dom'

interface ParamType {
  id: string
}

export function EditPage(): JSX.Element {
  const { id } = useParams<ParamType>()
  const [pad, setState] = useState({} as Pad)
  const state = useContext(UserContext)

  useEffect(() => {
    const res = start(
      parseInt(id, 10),
      state.userState.uid,
      state.userState.token
    )
    setState(res)
  }, [])

  return (
    <>
      <a
        href="#"
        onClick={() => {
          pad.kill()
          state.logoff()
        }}
      >
        Log off
      </a>
      <textarea id="textbox" rows={45} cols={150} disabled></textarea>
    </>
  )
}
