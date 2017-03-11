import React from 'react';

const frameFill = 'rgba(100, 100, 100, 0.3)';
const frameStroke = 'rgba(255, 255, 255, 1)';
const frameStrokeWidth = 1.5;

export default class NodeResourceBox extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  handleMouseClick() {
    console.log(this.props.meta.toJS());
  }

  render() {
    const { info, color, width, height, x, y, consumption } = this.props;
    const innerHeight = height * consumption;

    return (
      <g className="node-resource-box" onClick={this.handleMouseClick}>
        <title>{info}</title>
        <rect
          className="wrapper"
          fill={frameFill}
          stroke={frameStroke}
          strokeWidth={frameStrokeWidth}
          vectorEffect="non-scaling-stroke"
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
          vectorEffect="non-scaling-stroke"
          height={innerHeight}
          width={width}
          x={x}
          y={y + (height * (1 - consumption))}
        />
      </g>
    );
  }
}
