import React from 'react';

import Sparkline from '../sparkline';

export default class NodeDetailsTableRowSparkline extends React.Component {
  render() {
    const row = this.props.row;
    return (
      <div className="node-details-table-row-value">
        <div className="node-details-table-row-value-sparkline"><Sparkline data={row.metric.samples} min={0} max={row.metric.max} first={row.metric.first} last={row.metric.last} interpolate="none" />{row.value_major}</div>
        <div className="node-details-table-row-value-unit">{row.value_minor}</div>
      </div>
    );
  }
}
