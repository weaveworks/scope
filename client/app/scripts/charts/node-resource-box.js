import React from 'react';
import { connect } from 'react-redux';


class NodeResourceBox extends React.Component {
  defaultRectProps(relativeHeight = 1) {
    const { translateX, translateY, scaleX, scaleY } = this.props.transform;
    const innerTranslateY = this.props.height * scaleY * (1 - relativeHeight);
    const stroke = this.props.contrastMode ? 'black' : 'white';
    return {
      transform: `translate(0, ${innerTranslateY})`,
      opacity: this.props.contrastMode ? 1 : 0.85,
      height: this.props.height * scaleY * relativeHeight,
      width: this.props.width * scaleX,
      x: (this.props.x * scaleX) + translateX,
      y: (this.props.y * scaleY) + translateY,
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
