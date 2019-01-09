// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
import React from 'react';
import PropTypes from 'prop-types';
import { min as d3Min, max as d3Max, mean as d3Mean } from 'd3-array';
import { isoParse as parseDate } from 'd3-time-format';
import { line, curveLinear } from 'd3-shape';
import { scaleLinear } from 'd3-scale';

import { formatMetricSvg } from '../utils/string-utils';


const HOVER_RADIUS_MULTIPLY = 1.5;
const HOVER_STROKE_MULTIPLY = 5;
const MARGIN = 2;

export default class Sparkline extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.x = scaleLinear();
    this.y = scaleLinear();
    this.line = line()
      .x(d => this.x(d.date))
      .y(d => this.y(d.value));
  }

  initRanges(hasCircle) {
    // adjust scales and leave some room for the circle on the right, upper, and lower edge
    let circleSpace = MARGIN;
    if (hasCircle) {
      circleSpace += Math.ceil(this.props.circleRadius * HOVER_RADIUS_MULTIPLY);
    }

    this.x.range([MARGIN, this.props.width - circleSpace]);
    this.y.range([this.props.height - circleSpace, circleSpace]);
    this.line.curve(curveLinear);
  }

  getGraphData() {
    // data is of shape [{date, value}, ...] and is sorted by date (ASC)
    let { data } = this.props;

    this.initRanges(true);

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

    return {
      data, lastX, lastY, title
    };
  }

  getEmptyGraphData() {
    this.initRanges(false);
    const first = new Date(0);
    const last = new Date(15);
    this.x.domain([first, last]);
    this.y.domain([0, 1]);

    return {
      data: [
        {date: first, value: 0},
        {date: last, value: 0},
      ],
      lastX: this.x(last),
      lastY: this.y(0),
      title: '',
    };
  }

  render() {
    const dash = 5;
    const hasData = this.props.data && this.props.data.length > 0;
    const strokeColor = this.props.hovered && hasData
      ? this.props.hoverColor
      : this.props.strokeColor;
    const strokeWidth = this.props.strokeWidth * (this.props.hovered ? HOVER_STROKE_MULTIPLY : 1);
    const strokeDasharray = hasData ? undefined : `${dash}, ${dash}`;
    const radius = this.props.circleRadius * (this.props.hovered ? HOVER_RADIUS_MULTIPLY : 1);
    const fillOpacity = this.props.hovered ? 1 : 0.6;
    const circleColor = hasData && this.props.hovered ? strokeColor : strokeColor;
    const graph = hasData ? this.getGraphData() : this.getEmptyGraphData();

    return (
      <div title={graph.title}>
        <svg width={this.props.width} height={this.props.height}>
          <path
            className="sparkline"
            fill="none"
            stroke={strokeColor}
            strokeWidth={strokeWidth}
            strokeDasharray={strokeDasharray}
            d={this.line(graph.data)}
          />
          {hasData && <circle
            className="sparkcircle"
            cx={graph.lastX}
            cy={graph.lastY}
            fill={circleColor}
            fillOpacity={fillOpacity}
            stroke="none"
            r={radius}
          />}
        </svg>
      </div>
    );
  }
}

Sparkline.propTypes = {
  circleRadius: PropTypes.number,
  data: PropTypes.arrayOf(PropTypes.object),
  height: PropTypes.number,
  hoverColor: PropTypes.string,
  hovered: PropTypes.bool,
  strokeColor: PropTypes.string,
  strokeWidth: PropTypes.number,
  width: PropTypes.number,
};

Sparkline.defaultProps = {
  circleRadius: 1.75,
  data: [],
  height: 24,
  hoverColor: '#7d7da8',
  hovered: false,
  strokeColor: '#7d7da8',
  strokeWidth: 0.5,
  width: 80,
};
