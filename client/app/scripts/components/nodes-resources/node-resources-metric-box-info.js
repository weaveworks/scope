import React from 'react';

import { getHumanizedMetricInfo } from '../../utils/metric-utils';


const HEIGHT = '45px';

export default class NodeResourcesMetricBoxInfo extends React.Component {
  render() {
    const { node, width, x, y } = this.props;
    const humanizedMetricInfo = getHumanizedMetricInfo(node.get('activeMetric'));

    return (
      <foreignObject className="node-resource-info" x={x} y={y} width={width} height={HEIGHT}>
        <span className="wrapper label truncate">{node.get('label')}</span>
        <span className="wrapper consumption truncate">{humanizedMetricInfo}</span>
      </foreignObject>
    );
  }
}
