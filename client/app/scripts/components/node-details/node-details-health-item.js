import React from 'react';

import AnimatedSparkline from '../animated-sparkline';
import { formatMetric } from '../../utils/string-utils';

export default (props) => {
  return (
    <div className="node-details-health-item">
    <div className="node-details-health-item-value">{formatMetric(props.item.value, props.item)}</div>
      <div className="node-details-health-item-sparkline">
        <AnimatedSparkline data={props.item.samples} max={props.item.max}
          first={props.item.first} last={props.item.last} />
      </div>
      <div className="node-details-health-item-label">{props.item.label}</div>
    </div>
  );
};
