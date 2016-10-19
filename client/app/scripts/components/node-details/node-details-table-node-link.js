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
    const { label, labelMinor, linkable } = this.props;
    const title = !labelMinor ? label : `${label} (${labelMinor})`;

    if (linkable) {
      return (
        <span className="node-details-table-node-link" title={title}
          onClick={this.handleClick}>
          {label}
        </span>
      );
    }
    return (
      <span className="node-details-table-node" title={title}>
        {label}
      </span>
    );
  }
}

export default connect()(NodeDetailsTableNodeLink);
