import React from 'react';

import { formatMetricSvg } from '../../utils/string-utils';


export default class NodeResourcesMetricBoxInfo extends React.Component {
  humanizedMetricInfo() {
    const metric = this.props.activeMetric.toJS();
    const showExtendedInfo = metric.withCapacity && metric.format !== 'percent';
    const totalCapacity = formatMetricSvg(metric.totalCapacity, metric);
    const absoluteConsumption = formatMetricSvg(metric.absoluteConsumption, metric);
    const relativeConsumption = formatMetricSvg(100.0 * metric.relativeConsumption,
      { format: 'percent' });

    return (
      <span>
        <strong>
          {showExtendedInfo ? relativeConsumption : absoluteConsumption}
        </strong> consumed
        {showExtendedInfo && <i>{' - '}
          ({absoluteConsumption} / <strong>{totalCapacity}</strong>)
        </i>}
      </span>
    );
  }

  render() {
    const { width, x, y } = this.props;
    return (
      <foreignObject x={x} y={y} width={width} height="45px">
        <div className="node-resources-metric-box-info">
          <span className="wrapper label truncate">{this.props.label}</span>
          <span className="wrapper consumption truncate">{this.humanizedMetricInfo()}</span>
        </div>
      </foreignObject>
    );
  }
}
