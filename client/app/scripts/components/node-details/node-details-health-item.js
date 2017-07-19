import React from 'react';

import Sparkline from '../sparkline';
import { formatMetric } from '../../utils/string-utils';

function NodeDetailsHealthItem(props) {
  return (
    <div className="node-details-health-item">
      {!props.valueEmpty && <div className="node-details-health-item-value">{formatMetric(props.value, props)}</div>}
      <div className="node-details-health-item-sparkline">
        <Sparkline
          data={props.samples} max={props.max} format={props.format}
          first={props.first} last={props.last} hoverColor={props.metricColor}
          hovered={props.samples && props.hovered}
        />
      </div>
      <div className="node-details-health-item-label" style={{ color: props.labelColor }}>
        {props.label}
      </div>
    </div>
  );
}

export default NodeDetailsHealthItem;
