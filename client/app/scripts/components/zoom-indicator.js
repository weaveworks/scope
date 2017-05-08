import React from 'react';
import Slider from 'rc-slider';


export default class ZoomIndicator extends React.Component {
  constructor() {
    super();

    this.handleChange = this.handleChange.bind(this);
  }

  handleChange(value) {
    const scale = Math.exp(value) * this.props.minScale;
    this.props.slideAction(scale);
  }

  render() {
    const { minScale, maxScale, scale } = this.props;
    const max = Math.log(maxScale / minScale);
    const value = Math.log(scale / minScale);

    return (
      <div className="zoom-indicator-wrapper">
        <span className="zoom-indicator">
          Zoom <Slider min={0} max={max} value={value} step={0.001} onChange={this.handleChange} />
        </span>
      </div>
    );
  }
}
