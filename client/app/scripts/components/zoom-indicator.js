import React from 'react';
import Slider from 'rc-slider';


export default class ZoomIndicator extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleChange = this.handleChange.bind(this);
    this.handleZoomOut = this.handleZoomOut.bind(this);
    this.handleZoomIn = this.handleZoomIn.bind(this);
  }

  handleChange(value) {
    this.props.slideAction(Math.exp(value) * this.props.minScale);
  }

  handleZoomOut() {
    const newValue = Math.max(0, this.value - (0.05 * this.max));
    this.props.slideAction(Math.exp(newValue) * this.props.minScale);
  }

  handleZoomIn() {
    const newValue = Math.min(this.max, this.value + (0.05 * this.max));
    this.props.slideAction(Math.exp(newValue) * this.props.minScale);
  }

  render() {
    const { minScale, maxScale, scale } = this.props;
    const max = Math.log(maxScale / minScale);
    const value = Math.log(scale / minScale);
    this.value = value;
    this.max = max;

    return (
      <div className="zoom-indicator-wrapper">
        <span className="zoom-indicator">
          <a className="zoom-in" onClick={this.handleZoomIn}><span className="fa fa-plus" /></a>
          <Slider max={max} value={value} vertical step={0.001} onChange={this.handleChange} />
          <a className="zoom-out" onClick={this.handleZoomOut}><span className="fa fa-minus" /></a>
        </span>
      </div>
    );
  }
}
