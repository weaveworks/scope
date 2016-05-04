import React from 'react';
import { connect } from 'react-redux';

import { doControl } from '../../actions/app-actions';

class NodeDetailsControlButton extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  render() {
    let className = `node-control-button fa ${this.props.control.icon}`;
    if (this.props.pending) {
      className += ' node-control-button-pending';
    }
    // If the control allows raw pipes, let the user choose that option
    // if (this.props.control.can_raw_pipe) {
    //  className += ' node-control-raw';
    // }
    return (
      <span className={className} title={this.props.control.human} onClick={this.handleClick} />
    );
  }

  handleClick(ev) {
    const args = {};
    ev.preventDefault();
    // If the use chose the raw option include it in the request
    // if (this.raw_pipe_chosen) {
    //  args.raw_pipe = 'true';
    // }
    this.props.dispatch(doControl(this.props.nodeId, this.props.control, args));
  }
}

export default connect()(NodeDetailsControlButton);
