import React from 'react';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsTableNodeLink extends React.PureComponent {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
    this.saveNodeRef = this.saveNodeRef.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.dispatch(clickRelative(
      this.props.nodeId,
      this.props.topologyId,
      this.props.label,
      this.node.getBoundingClientRect()
    ));
  }

  saveNodeRef(ref) {
    this.node = ref;
  }

  render() {
    const { label, labelMinor, linkable } = this.props;
    const title = !labelMinor ? label : `${label} (${labelMinor})`;

    if (linkable) {
      return (
        <span
          className="node-details-table-node-link" title={title}
          ref={this.saveNodeRef} onClick={this.handleClick}>
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
