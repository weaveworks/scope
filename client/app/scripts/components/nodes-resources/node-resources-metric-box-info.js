import React from 'react';

import { formatMetricSvg } from '../../utils/string-utils';


export default class NodeResourcesMetricBoxInfo extends React.Component {
  humanizedMetricInfo() {
    const metricSummary = this.props.metricSummary.toJS();
    const showExtendedInfo = metricSummary.showCapacity && metricSummary.format !== 'percent';
    const totalCapacity = formatMetricSvg(metricSummary.totalCapacity, metricSummary);
    const absoluteConsumption = formatMetricSvg(metricSummary.absoluteConsumption, metricSummary);
    const relativeConsumption = formatMetricSvg(100.0 * metricSummary.relativeConsumption,
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
