import * as React from 'react'
import { useEffect, useState, useContext } from 'react'
import { UserContext } from '../contexts/usercontext'
import { start } from '../pad'
import { useParams } from 'react-router-dom'

export function EditPage(): JSX.Element {
  const { id } = useParams()
  const [_, setState] = useState({})
  const state = useContext(UserContext)

  useEffect(() => {
    setState(start(parseInt(id, 10), state.userState.uid))
  }, [])

  return (
    <>
      <textarea id="textbox" rows={45} cols={150} disabled></textarea>
    </>
  )
}
