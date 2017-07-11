import React from 'react';

import NodeDetailsHealthOverflowItem from './node-details-health-overflow-item';

export default class NodeDetailsHealthOverflow extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.handleClick();
  }

  render() {
    const items = this.props.items.slice(0, 4);

    return (
      <div className="node-details-health-overflow" onClick={this.handleClick} title="Expand metrics">
        {items.map(item => <NodeDetailsHealthOverflowItem key={item.id} {...item} />)}
      </div>
    );
  }
}
