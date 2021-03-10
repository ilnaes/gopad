import * as React from 'react'
import { useState } from 'react'
import jwtDecode from 'jwt-decode'

export type UserState = {
  token: string
  uid: string
  isLoggedIn: boolean
}

type UserProps = {
  children: React.ReactNode
}

type UserContext_t = {
  userState: UserState
  updateUser: (state: UserState) => void
  logoff: () => void
}

const token = localStorage.getItem('token')
const uid = token && (jwtDecode(token) as UserState).uid

const init: UserState = {
  token: token,
  uid: uid,
  isLoggedIn: uid != null,
}

export const UserContext = React.createContext({} as UserContext_t)

export const UserProvider = (props: UserProps): JSX.Element => {
  const [userState, updateUser] = useState(init)
  const logoff = () => {
    updateUser({ isLoggedIn: false } as UserState)
    localStorage.clear()
  }

  return (
    <UserContext.Provider value={{ userState, updateUser, logoff }}>
      {props.children}
    </UserContext.Provider>
  )
}
