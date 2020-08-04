import * as React from 'react'
import { start } from '../pad'

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
