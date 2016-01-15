import React from 'react';

import { formatMetric } from '../../utils/string-utils';

export default class NodeDetailsHealthOverflowItem extends React.Component {
  render() {
    return (
      <div className="node-details-health-overflow-item">
      <div className="node-details-health-overflow-item-value">{formatMetric(this.props.item.value, this.props.item)}</div>
        <div className="node-details-health-overflow-item-label truncate">{this.props.item.label}</div>
      </div>
    );
  }
}
