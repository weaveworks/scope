import React from 'react';
import ReactDOM from 'react-dom';
import { connect } from 'react-redux';

import { doControl } from '../../actions/app-actions';

class NodeDetailsControlButton extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  componentDidMount() {
    if (!window.isNormal() && this.props.control.human === 'Exec shell') {
      ReactDOM.findDOMNode(this).click();
    }
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

export default connect()(NodeDetailsControlButton);
