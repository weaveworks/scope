import React from 'react';

import Sparkline from '../sparkline';
import { formatMetric } from '../../utils/string-utils';

function NodeDetailsHealthItem(props) {
  return (
    <div className="node-details-health-item">
      {props.value !== undefined && <div className="node-details-health-item-value">{formatMetric(props.value, props)}</div>}
      {props.samples && <div className="node-details-health-item-sparkline">
        <Sparkline
          data={props.samples} max={props.max} format={props.format}
          first={props.first} last={props.last} strokeWidth={props.strokeWidth}
          strokeColor={props.strokeColor} />
      </div>}
      <div className="node-details-health-item-label" style={{ color: props.labelColor }}>
        {props.label}
      </div>
    </div>
  );
}

export default NodeDetailsHealthItem;
