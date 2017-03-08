import React from 'react';

const frameFill = 'rgba(100, 100, 100, 0.2)';
const frameStroke = 'rgba(100, 100, 100, 0.5)';

export default class NodeResourceMetric extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  handleMouseClick() {
    console.log(this.props.meta.toJS());
  }

  render() {
    const { label, color, width, height, x, y, consumption } = this.props;
    const innerHeight = height * consumption;
    const transform = `translate(${x},${y})`;

    return (
      <g className="node-resource-metric" onClick={this.handleMouseClick} transform={transform}>
        <title>{label}</title>
        <rect
          className="wrapper"
          fill={frameFill}
          stroke={frameStroke}
          strokeWidth="1"
          vectorEffect="non-scaling-stroke"
          height={height}
          width={width}
        />
        <rect
          className="bar"
          fill={color}
          stroke={frameStroke}
          strokeWidth="1"
          vectorEffect="non-scaling-stroke"
          y={height - innerHeight}
          height={innerHeight}
          width={width}
        />
      </g>
    );
  }
}
