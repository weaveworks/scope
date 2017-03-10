import React from 'react';


export default class NodeResourceBox extends React.Component {
  render() {
    const { label, width, x, y } = this.props;

    return (
      <g className="node-resource-label">
        <foreignObject
          className="node-label-container truncate"
          y={y}
          x={x}
          width={width}
          height="5em">
          <div className="label-wrapper truncate">
            {label}
          </div>
        </foreignObject>
      </g>
    );
  }
}
