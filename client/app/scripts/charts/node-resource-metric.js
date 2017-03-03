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
    const transform = 'scale(1,-1)';

    return (
      <g className="node-resource-metric" transform={transform} onClick={this.handleMouseClick}>
        <title>{label}</title>
        <rect fill={frameFill} stroke={frameStroke} x={x} y={y} width={width} height={height} />
        <rect fill={color} x={x} y={y} width={width} height={innerHeight} />
      </g>
    );
  }
}
