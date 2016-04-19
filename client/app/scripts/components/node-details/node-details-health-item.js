import React from 'react';

import Sparkline from '../sparkline';
import { formatMetric } from '../../utils/string-utils';

function NodeDetailsHealthItem(props) {
  return (
    <div className="node-details-health-item">
    <div className="node-details-health-item-value">{formatMetric(props.value, props)}</div>
      <div className="node-details-health-item-sparkline">
        <Sparkline data={props.samples} max={props.max} format={props.format}
          first={props.first} last={props.last} />
      </div>
      <div className="node-details-health-item-label">{props.label}</div>
    </div>
  );
}

export default NodeDetailsHealthItem;
