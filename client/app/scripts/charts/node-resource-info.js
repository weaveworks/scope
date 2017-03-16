import React from 'react';

import { getHumanizedMetricInfo } from '../utils/metric-utils';

export default class NodeResourceInfo extends React.Component {
  render() {
    const { node, width, x, y } = this.props;
    const humanizedMetricInfo = getHumanizedMetricInfo(node.get('activeMetric').toJS());

    return (
      <foreignObject className="node-resource-info" x={x} y={y} width={width} height="45px">
        <span className="wrapper label truncate">{node.get('label')}</span>
        <span className="wrapper consumption truncate">{humanizedMetricInfo}</span>
      </foreignObject>
    );
  }
}
