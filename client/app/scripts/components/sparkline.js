// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import PropTypes from 'prop-types';
import { min as d3Min, max as d3Max, mean as d3Mean } from 'd3-array';
import { isoParse as parseDate } from 'd3-time-format';
import { line, curveLinear } from 'd3-shape';
import { scaleLinear } from 'd3-scale';

import { formatMetricSvg } from '../utils/string-utils';
import { brightenColor, darkenColor } from '../utils/color-utils';


const HOVER_RADIUS_MULTIPLY = 1.5;
const HOVER_STROKE_MULTIPLY = 5;

export default class Sparkline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.x = scaleLinear();
    this.y = scaleLinear();
    this.line = line()
      .x(d => this.x(d.date))
      .y(d => this.y(d.value));
  }

  initRanges() {
    // adjust scales and leave some room for the circle on the right, upper, and lower edge
    const padding = 2 + Math.ceil(this.props.circleRadius * HOVER_RADIUS_MULTIPLY);
    this.x.range([2, this.props.width - padding]);
    this.y.range([this.props.height - padding, padding]);
    this.line.curve(this.props.curve);
  }

  getGraphData() {
    // data is of shape [{date, value}, ...] and is sorted by date (ASC)
    let data = this.props.data;

    this.initRanges();

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
    const title = `Last ${Math.round((lastDate - firstDate) / 1000)} seconds, ` +
      `${data.length} samples, min: ${min}, max: ${max}, mean: ${mean}`;

    return {title, lastX, lastY, data};
  }

  getEmptyGraphData() {
    this.initRanges();
    const first = new Date(0);
    const last = new Date(15);
    this.x.domain([first, last]);
    this.y.domain([0, 1]);

    return {
      title: '',
      lastX: this.x(last),
      lastY: this.y(0),
      data: [
        {date: first, value: 0},
        {date: last, value: 0},
      ],
    };
  }

  render() {
    let strokeColor = this.props.strokeColor;
    let strokeWidth = this.props.strokeWidth;
    let radius = this.props.circleRadius;
    let fillOpacity = 0.6;
    let circleColor;
    let graph = {};

    if (!this.props.data || this.props.data.length === 0 || this.props.data[0].date === undefined) {
      // no data means just a dead line w/o circle
      graph = this.getEmptyGraphData();
      strokeColor = brightenColor(strokeColor);
      radius = 0;
    } else {
      graph = this.getGraphData();

      if (this.props.hovered) {
        strokeColor = this.props.hoverColor;
        circleColor = strokeColor;
        strokeWidth *= HOVER_STROKE_MULTIPLY;
        radius *= HOVER_RADIUS_MULTIPLY;
        fillOpacity = 1;
      } else {
        circleColor = darkenColor(strokeColor);
      }
    }

    return (
      <div title={graph.title}>
        <svg width={this.props.width} height={this.props.height}>
          <path
            className="sparkline" fill="none" stroke={strokeColor}
            strokeWidth={strokeWidth} d={this.line(graph.data)}
          />
          <circle
            className="sparkcircle" cx={graph.lastX} cy={graph.lastY} fill={circleColor}
            fillOpacity={fillOpacity} stroke="none" r={radius}
          />
        </svg>
      </div>
    );
  }
}

Sparkline.propTypes = {
  data: PropTypes.arrayOf(PropTypes.object)
};

Sparkline.defaultProps = {
  width: 80,
  height: 24,
  strokeColor: '#7d7da8',
  strokeWidth: 0.5,
  hoverColor: '#7d7da8',
  curve: curveLinear,
  circleRadius: 1.75,
  hovered: false,
  data: [],
};
