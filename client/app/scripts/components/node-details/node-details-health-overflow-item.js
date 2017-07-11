import React from 'react';

import { formatMetric } from '../../utils/string-utils';

function NodeDetailsHealthOverflowItem(props) {
  return (
    <div className="node-details-health-overflow-item">
      <div className="node-details-health-overflow-item-value">
        {props.value !== undefined && formatMetric(props.value, props)}
      </div>
      <div className="node-details-health-overflow-item-label truncate">{props.label}</div>
    </div>
  );
}

export default NodeDetailsHealthOverflowItem;
