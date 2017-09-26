import React from 'react';
import { connect } from 'react-redux';

import { trackAnalyticsEvent } from '../../utils/tracking-utils';
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
    return (
      <span className={className} title={this.props.control.human} onClick={this.handleClick} />
    );
  }

  handleClick(ev) {
    ev.preventDefault();
    const { id, human } = this.props.control;
    trackAnalyticsEvent('scope.node.control.click', { id, title: human });
    this.props.dispatch(doControl(this.props.nodeId, this.props.control));
  }
}

// Using this instead of PureComponent because of props.dispatch
export default connect()(NodeDetailsControlButton);
