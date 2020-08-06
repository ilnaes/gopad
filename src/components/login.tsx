import * as React from 'react'
import { useContext, useState } from 'react'
import { UserContext, UserState } from '../contexts/usercontext'
import * as jwtDecode from 'jwt-decode'
import { useHistory, useLocation } from 'react-router-dom'

type LocationState = {
  from: {
    pathname: string
  }
}

export default function LoginPage(): JSX.Element {
  const history = useHistory()
  const loc = useLocation<LocationState>()

  const { updateUser } = useContext(UserContext)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  const { from } = loc.state || { from: { pathname: '/' } }

  const login = async () => {
    const options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    }

    const res = await fetch('http://localhost:8080/login', options)
    if (res.ok) {
      const token = await res.text()
      const uid = (jwtDecode(token) as UserState).uid

      localStorage.setItem('token', token)
      updateUser({ token, uid, isLoggedIn: true })

      location.replace(from.pathname)
    } else {
      console.log(res)
    }
  }

  return (
    <div>
      <h2>Sign In</h2>
      <input
        type="text"
        placeholder="Username"
        onChange={(e) => {
          setUsername(e.target.value)
        }}
      ></input>
      <br />
      <input
        type="password"
        placeholder="Password"
        onChange={(e) => {
          setPassword(e.target.value)
        }}
      ></input>
      <br />
      <button onClick={login}>Log in</button>
    </div>
  )
}
