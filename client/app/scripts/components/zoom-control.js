import React from 'react';
import Slider from 'rc-slider';
import { scaleLog } from 'd3-scale';


const SLIDER_STEP = 0.001;
const CLICK_STEP = 0.05;

// Returns a log-scale that maps zoom factors to slider values.
const getSliderScale = ({ minScale, maxScale }) => (
  scaleLog()
    // Zoom limits may vary between different views.
    .domain([minScale, maxScale])
    // Taking the unit range for the slider ensures consistency
    // of the zoom button steps across different zoom domains.
    .range([0, 1])
    // This makes sure the input values are always clamped into the valid domain/range.
    .clamp(true)
);

export default class ZoomControl extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleChange = this.handleChange.bind(this);
    this.handleZoomOut = this.handleZoomOut.bind(this);
    this.handleZoomIn = this.handleZoomIn.bind(this);
    this.getSliderValue = this.getSliderValue.bind(this);
    this.toZoomScale = this.toZoomScale.bind(this);
  }

  handleChange(sliderValue) {
    this.props.zoomAction(this.toZoomScale(sliderValue));
  }

  handleZoomOut() {
    this.props.zoomAction(this.toZoomScale(this.getSliderValue() - CLICK_STEP));
  }

  handleZoomIn() {
    this.props.zoomAction(this.toZoomScale(this.getSliderValue() + CLICK_STEP));
  }

  getSliderValue() {
    const toSliderValue = getSliderScale(this.props);
    return toSliderValue(this.props.scale);
  }

  toZoomScale(sliderValue) {
    const toSliderValue = getSliderScale(this.props);
    return toSliderValue.invert(sliderValue);
  }

  render() {
    const value = this.getSliderValue();

    return (
      <div className="zoom-control">
        <button className="zoom-in" onClick={this.handleZoomIn}>
          <span className="fa fa-plus" />
        </button>
        <Slider value={value} max={1} step={SLIDER_STEP} vertical onChange={this.handleChange} />
        <button className="zoom-out" onClick={this.handleZoomOut}>
          <span className="fa fa-minus" />
        </button>
      </div>
    );
  }
}
