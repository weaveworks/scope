import React from 'react';
import ReactDOM from 'react-dom';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsTableNodeLink extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    clickRelative(this.props.id, this.props.topologyId, this.props.label,
      ReactDOM.findDOMNode(this).getBoundingClientRect());
  }

  render() {
    const titleLines = [`${this.props.label} (${this.props.topologyId})`];
    this.props.metadata.forEach(data => {
      titleLines.push(`${data.label}: ${data.value}`);
    });
    const title = titleLines.join('\n');

    if (this.props.linkable) {
      return (
        <span className="node-details-table-node-link truncate" title={title}
          onClick={this.handleClick}>
          {this.props.label}
        </span>
      );
    }
    return (
      <span className="node-details-table-node truncate" title={title}>
        {this.props.label}
      </span>
    );
  }
}
