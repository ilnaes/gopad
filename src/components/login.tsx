import * as React from 'react'
import { useContext, useState } from 'react'
import { UserContext, UserState } from '../contexts/usercontext'
import jwtDecode from 'jwt-decode'
import { useLocation } from 'react-router-dom'

type LocationState = {
  from: {
    pathname: string
  }
}

type LoginProps = {
  register: boolean
}

export default function LoginPage(props: LoginProps): JSX.Element {
  const loc = useLocation<LocationState>()

  console.log('HERE')

  const { updateUser } = useContext(UserContext)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [response, setResponse] = useState('')

  const { from } = loc.state || { from: { pathname: '/' } }

  const login = async () => {
    const options = {
      method: props.register ? 'PUT' : 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    }

    const endpoint = props.register
      ? 'http://localhost:8080/register'
      : 'http://localhost:8080/login'
    const res = await fetch(endpoint, options)
    if (res.ok) {
      const token = await res.text()
      const uid = (jwtDecode(token) as UserState).uid

      localStorage.setItem('token', token)
      updateUser({ token, uid, isLoggedIn: true })

      location.replace(from.pathname)
    } else {
      setResponse(await res.text())
    }
  }

  const other = props.register ? '/login' : '/register'

  return (
    <div>
      <h2>{props.register ? 'Register' : 'Sign In'}</h2>
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
      <br />
      <a href={other}>{props.register ? 'Login' : 'Register'}</a>
      <br />
      {response}
    </div>
  )
}
