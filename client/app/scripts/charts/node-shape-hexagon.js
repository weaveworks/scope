import React from 'react';
import classNames from 'classnames';

import { nodeShapePolygon } from '../utils/node-shape-utils';
import {
  getMetricValue,
  getMetricColor,
  getClipPathDefinition,
  renderMetricValue,
} from '../utils/metric-utils';
import {
  NODE_SHAPE_HIGHLIGHT_RADIUS,
  NODE_SHAPE_BORDER_RADIUS,
  NODE_SHAPE_SHADOW_RADIUS,
} from '../constants/styles';

export default class NodeShapeHexagon extends React.Component {
  render() {
    const { id, highlighted, color, metric } = this.props;
    const { height, hasMetric, formattedValue } = getMetricValue(metric);
    const metricStyle = { fill: getMetricColor(metric) };

    const className = classNames('shape', 'shape-hexagon', { metrics: hasMetric });
    const pathProps = r => ({ d: nodeShapePolygon(r, 6) });
    const clipId = `mask-${id}`;

    // if (id === '30a215df22be6253ac71098dc0505b18a6b852236a9e804c104e94a3b01b5c0b;<container>') {
    //   console.log(this.props);
    // }
    return (
      <g className={className}>
        {hasMetric && getClipPathDefinition(clipId, height)}
        {highlighted &&
          <path className="highlighted" {...pathProps(NODE_SHAPE_HIGHLIGHT_RADIUS)} />}
        <path className="border" stroke={color} {...pathProps(NODE_SHAPE_BORDER_RADIUS)} />
        <path className="shadow" {...pathProps(NODE_SHAPE_SHADOW_RADIUS)} />
        {hasMetric && <path
          className="metric-fill"
          clipPath={`url(#${clipId})`}
          style={metricStyle}
          {...pathProps(NODE_SHAPE_SHADOW_RADIUS)}
        />}
        {renderMetricValue(formattedValue, highlighted && hasMetric)}
      </g>
    );
  }
}

// export default function NodeShapeHexagon({ id, highlighted, color, metric }) {
//   const { height, hasMetric, formattedValue } = getMetricValue(metric);
//   const metricStyle = { fill: getMetricColor(metric) };
//
//   const className = classNames('shape', 'shape-hexagon', { metrics: hasMetric });
//   const pathProps = r => ({ d: nodeShapePolygon(r, 6) });
//   const clipId = `mask-${id}`;
//
//   return (
//     <g className={className}>
//       {hasMetric && getClipPathDefinition(clipId, height)}
//       {highlighted &&
//         <path className="highlighted" {...pathProps(NODE_SHAPE_HIGHLIGHT_RADIUS)} />}
//       <path className="border" stroke={color} {...pathProps(NODE_SHAPE_BORDER_RADIUS)} />
//       <path className="shadow" {...pathProps(NODE_SHAPE_SHADOW_RADIUS)} />
//       {hasMetric && <path
//         className="metric-fill"
//         clipPath={`url(#${clipId})`}
//         style={metricStyle}
//         {...pathProps(NODE_SHAPE_SHADOW_RADIUS)}
//       />}
//       {renderMetricValue(formattedValue, highlighted && hasMetric)}
//     </g>
//   );
// }
