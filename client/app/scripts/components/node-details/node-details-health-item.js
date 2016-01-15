import React from 'react';

import Sparkline from '../sparkline';
import { formatMetric } from '../../utils/string-utils';

export default (props) => {
  return (
    <div className="node-details-health-item">
    <div className="node-details-health-item-value">{formatMetric(props.item.value, props.item)}</div>
      <div className="node-details-health-item-sparkline">
        <Sparkline data={props.item.samples} min={0} max={props.item.max}
          first={props.item.first} last={props.item.last} interpolate="none" />
      </div>
      <div className="node-details-health-item-label">{props.item.label}</div>
    </div>
  );
};
