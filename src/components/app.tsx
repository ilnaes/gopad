import * as React from 'react'
import { BrowserRouter as Router, Switch, Route, Link } from 'react-router-dom'
import { Hello } from './editpage'

export function App(): JSX.Element {
  return (
    <Router>
      <div>
        <Switch>
          <Route path="/edit/:id">
            <Hello />
          </Route>
          <Route path="/index">
            <div>Index!</div>
          </Route>
        </Switch>
      </div>
    </Router>
  )
}
