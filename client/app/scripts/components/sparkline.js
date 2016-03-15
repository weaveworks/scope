// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import d3 from 'd3';

const parseDate = d3.time.format.iso.parse;

export default class Sparkline extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.x = d3.scale.linear();
    this.y = d3.scale.linear();
    this.line = d3.svg.line()
      .x(d => this.x(d.date))
      .y(d => this.y(d.value));
  }

  getGraphData() {
    // data is of shape [{date, value}, ...] and is sorted by date (ASC)
    let data = this.props.data;

    // Do nothing if no data or data w/o date are passed in.
    if (data.length === 0 || data[0].date === undefined) {
      return <div />;
    }

    // adjust scales
    this.x.range([2, this.props.width - 2]);
    this.y.range([this.props.height - 2, 2]);
    this.line.interpolate(this.props.interpolate);

    // Convert dates into D3 dates
    data = data.map(d => ({
      date: parseDate(d.date),
      value: d.value
    }));

    // determine date range
    let firstDate = this.props.first ? parseDate(this.props.first) : data[0].date;
    let lastDate = this.props.last ? parseDate(this.props.last) : data[data.length - 1].date;
    // if last prop is after last value, we need to add that difference as
    // padding before first value to right-align sparkline
    const skip = lastDate - data[data.length - 1].date;
    if (skip > 0) {
      firstDate -= skip;
      lastDate -= skip;
    }
    this.x.domain([firstDate, lastDate]);

    // determine value range
    const minValue = this.props.min !== undefined ? this.props.min : d3.min(data, d => d.value);
    const maxValue = this.props.max !== undefined
      ? Math.max(this.props.max, d3.max(data, d => d.value)) : d3.max(data, d => d.value);
    this.y.domain([minValue, maxValue]);

    const lastValue = data[data.length - 1].value;
    const lastX = this.x(lastDate);
    const lastY = this.y(lastValue);
    const title = `Last ${d3.round((lastDate - firstDate) / 1000)} seconds, ` +
      `${data.length} samples, min: ${d3.round(d3.min(data, d => d.value), 2)}` +
      `, max: ${d3.round(d3.max(data, d => d.value), 2)}` +
      `, mean: ${d3.round(d3.mean(data, d => d.value), 2)}`;

    return {title, lastX, lastY, data};
  }

  render() {
    // Do nothing if no data or data w/o date are passed in.
    if (this.props.data.length === 0 || this.props.data[0].date === undefined) {
      return <div />;
    }

    const {lastX, lastY, title, data} = this.getGraphData();

    return (
      <div title={title}>
        <svg width={this.props.width} height={this.props.height}>
          <path className="sparkline" fill="none" stroke={this.props.strokeColor}
            strokeWidth={this.props.strokeWidth} ref="path" d={this.line(data)} />
          <circle className="sparkcircle" cx={lastX} cy={lastY} fill="#46466a"
            fillOpacity="0.6" stroke="none" r={this.props.circleDiameter} />
        </svg>
      </div>
    );
  }

}

Sparkline.propTypes = {
  data: React.PropTypes.array.isRequired
};

Sparkline.defaultProps = {
  width: 80,
  height: 24,
  strokeColor: '#7d7da8',
  strokeWidth: '0.5px',
  interpolate: 'none',
  circleDiameter: 1.75
};
