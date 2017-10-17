import React from 'react';


export default class NodeResourcesMetricBoxInfo extends React.Component {
  humanizedMetricInfo() {
    const {
      humanizedTotalCapacity, humanizedAbsoluteConsumption,
      humanizedRelativeConsumption, showCapacity, format
    } = this.props.metricSummary.toJS();
    const showExtendedInfo = showCapacity && format !== 'percent';

    return (
      <span>
        <strong>
          {showExtendedInfo ? humanizedRelativeConsumption : humanizedAbsoluteConsumption}
        </strong> used
        {showExtendedInfo &&
          <i>
            {' - '}({humanizedAbsoluteConsumption} / <strong>{humanizedTotalCapacity}</strong>)
          </i>
        }
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
