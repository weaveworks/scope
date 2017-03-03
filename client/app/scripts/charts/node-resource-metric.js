import React from 'react';

export default class NodeResourceMetric extends React.Component {
  render() {
    const { index } = this.props;
    return (
      <g className="node-resource-metric">
        <rect x={index * 100} y="0" width="90" height="300" />
      </g>
    );
  }
}
