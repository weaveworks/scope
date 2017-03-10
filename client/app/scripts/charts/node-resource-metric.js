import React from 'react';

const frameFill = 'rgba(100, 100, 100, 0.2)';
const frameStroke = 'rgba(255, 255, 255, 1)';
const frameStrokeWidth = 1.5;

export default class NodeResourceMetric extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  handleMouseClick() {
    console.log(this.props.meta.toJS());
  }

  // renderLabel() {
  //   return (
  //     <foreignObject
  //       className="node-labels-container"
  //       y={0}
  //       x={0}
  //       vectorEffect="non-scaling-stroke"
  //       width={labelWidth}
  //       height="5em">
  //       <div className="node-label-wrapper">
  //         {this.props.label}
  //       </div>
  //     </foreignObject>
  //   );
  // }

  render() {
    const { label, info, color, width, height, x, y, consumption } = this.props;
    const innerHeight = height * consumption;
    const labelX = Math.max(0, x) + 10;
    const labelWidth = Math.max(0, (x + width) - (labelX + 10));
    const labelShown = (labelWidth > 20);

    return (
      <g className="node-resource-metric" onClick={this.handleMouseClick}>
        <title>{info}</title>
        <rect
          className="wrapper"
          fill={frameFill}
          stroke={frameStroke}
          strokeWidth={frameStrokeWidth}
          height={height}
          width={width}
          x={x}
          y={y}
        />
        <rect
          className="bar"
          fill={color}
          stroke={frameStroke}
          strokeWidth={frameStrokeWidth}
          height={innerHeight}
          width={width}
          x={x}
          y={y + (height * (1 - consumption))}
        />
        {labelShown && <foreignObject
          className="node-label-container truncate"
          y={y + 10}
          x={labelX}
          width={labelWidth}
          height="5em">
          <div className="node-label-wrapper truncate">
            {label}
          </div>
        </foreignObject>}
      </g>
    );
  }
}
