import * as React from 'react'
import { useContext } from 'react'
import {
  BrowserRouter as Router,
  Redirect,
  Switch,
  Route,
  Link,
} from 'react-router-dom'
import { Hello } from './editpage'
import { UserProvider, UserContext } from '../contexts/usercontext'
import LoginPage from './login'

export function App(): JSX.Element {
  return (
    <UserProvider>
      <Router>
        <div>
          <Switch>
            <Route path="/login">
              <LoginPage />
            </Route>
            <PrivateRoute path="/edit/:id">
              <Hello />
            </PrivateRoute>
            <Route path="/">
              <div>Index!</div>
            </Route>
          </Switch>
        </div>
      </Router>
    </UserProvider>
  )
}

type Props = {
  children: React.ReactNode
  path: string
}

// A wrapper for <Route> that redirects to the login
// screen if you're not yet authenticated.
function PrivateRoute({ children, ...rest }: Props) {
  const { userState } = useContext(UserContext)

  return (
    <Route
      {...rest}
      render={({ location }) =>
        userState.isLoggedIn ? (
          children
        ) : (
          <Redirect
            to={{
              pathname: '/login',
              state: { from: location },
            }}
          />
        )
      }
    />
  )
}
