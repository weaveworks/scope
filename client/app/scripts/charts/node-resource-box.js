import React from 'react';
import { connect } from 'react-redux';


class NodeResourceBox extends React.Component {
  defaultRectProps(relativeHeight = 1) {
    const stroke = this.props.contrastMode ? 'black' : 'white';
    const translateY = this.props.height * (1 - relativeHeight);
    return {
      transform: `translate(0, ${translateY})`,
      height: this.props.height * relativeHeight,
      width: this.props.width,
      x: this.props.x,
      y: this.props.y,
      vectorEffect: 'non-scaling-stroke',
      strokeWidth: 1,
      stroke,
    };
  }

  render() {
    const { color, withCapacity, activeMetric } = this.props;
    const { relativeConsumption, info } = activeMetric.toJS();
    const frameFill = 'rgba(150, 150, 150, 0.4)';

    return (
      <g className="node-resource-box">
        <title>{info}</title>
        {withCapacity && <rect className="frame" fill={frameFill} {...this.defaultRectProps()} />}
        <rect className="bar" fill={color} {...this.defaultRectProps(relativeConsumption)} />
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode')
  };
}

export default connect(mapStateToProps)(NodeResourceBox);
