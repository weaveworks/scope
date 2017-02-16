import React from 'react';

import { doControl } from '../../actions/app-actions';

export default class NodeDetailsControlButton extends React.PureComponent {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  render() {
    let className = `node-control-button fa ${this.props.control.icon}`;
    if (this.props.pending) {
      className += ' node-control-button-pending';
    }
    return (
      <span className={className} title={this.props.control.human} onClick={this.handleClick} />
    );
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.dispatch(doControl(this.props.nodeId, this.props.control));
  }
}
