import React from 'react';
import ReactDOM from 'react-dom';
import { connect } from 'react-redux';

import { clickRelative } from '../../actions/app-actions';

class NodeDetailsTableNodeLink extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.dispatch(clickRelative(this.props.nodeId, this.props.topologyId,
      this.props.label, ReactDOM.findDOMNode(this).getBoundingClientRect()));
  }

  render() {
    if (this.props.linkable) {
      return (
        <span className="node-details-table-node-link" title={this.props.label}
          onClick={this.handleClick}>
          {this.props.label}
        </span>
      );
    }
    return (
      <span className="node-details-table-node" title={this.props.label}>
        {this.props.label}
      </span>
    );
  }
}

export default connect()(NodeDetailsTableNodeLink);
