// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import { min as d3Min, max as d3Max, mean as d3Mean } from 'd3-array';
import { isoParse as parseDate } from 'd3-time-format';
import { line, curveLinear } from 'd3-shape';
import { scaleLinear } from 'd3-scale';

import { formatMetricSvg } from '../utils/string-utils';
import { round } from '../utils/math-utils';


export default class Sparkline extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.x = scaleLinear();
    this.y = scaleLinear();
    this.line = line()
      .x(d => this.x(d.date))
      .y(d => this.y(d.value));
  }

  getGraphData() {
    // data is of shape [{date, value}, ...] and is sorted by date (ASC)
    let data = this.props.data;

    // Do nothing if no data or data w/o date are passed in.
    if (data === undefined || data.length === 0 || data[0].date === undefined) {
      return <div />;
    }

    // adjust scales
    this.x.range([2, this.props.width - 2]);
    this.y.range([this.props.height - 2, 2]);
    this.line.curve(this.props.curve);

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
    const minValue = this.props.min !== undefined ? this.props.min : d3Min(data, d => d.value);
    const maxValue = this.props.max !== undefined
      ? Math.max(this.props.max, d3Max(data, d => d.value)) : d3Max(data, d => d.value);
    this.y.domain([minValue, maxValue]);

    const lastValue = data[data.length - 1].value;
    const lastX = this.x(lastDate);
    const lastY = this.y(lastValue);
    const min = formatMetricSvg(d3Min(data, d => d.value), this.props);
    const max = formatMetricSvg(d3Max(data, d => d.value), this.props);
    const mean = formatMetricSvg(d3Mean(data, d => d.value), this.props);
    const title = `Last ${round((lastDate - firstDate) / 1000)} seconds, ` +
      `${data.length} samples, min: ${min}, max: ${max}, mean: ${mean}`;

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
          <path
            className="sparkline" fill="none" stroke={this.props.strokeColor}
            strokeWidth={this.props.strokeWidth} d={this.line(data)}
          />
          <circle
            className="sparkcircle" cx={lastX} cy={lastY} fill="#46466a"
            fillOpacity="0.6" stroke="none" r={this.props.circleDiameter}
          />
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
  curve: curveLinear,
  circleDiameter: 1.75
};
