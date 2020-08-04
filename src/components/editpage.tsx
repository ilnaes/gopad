import * as React from 'react'
import { start } from '../app'

export class Hello extends React.Component {
  componentDidMount(): void {
    this.setState(start())
  }

  render(): React.ReactNode {
    return (
      <>
        <textarea id="textbox" rows={45} cols={150} disabled></textarea>
      </>
    )
  }
}
