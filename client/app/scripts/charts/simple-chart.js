/* eslint react/jsx-no-bind: "off", no-return-assign: "off" */

import React from 'react';
import ReactDOM from 'react-dom';
import d3 from 'd3';
import _ from 'lodash';
import debug from 'debug';

import { formatMetric } from '../utils/string-utils';

const log = debug('scope:chart');
const MARGINS = {top: 0, bottom: 24, left: 48, right: 72};

const customTimeFormat = d3.time.format.multi([
  ['.%L', d => d.getMilliseconds()],
  ['%H:%M:%S', d => d.getSeconds()],
  ['%H:%M', d => d.getMinutes()],
  ['%H:%M', d => d.getHours()],
  ['%a %d', d => d.getDay() && d.getDate() !== 1],
  ['%b %d', d => d.getDate() !== 1],
  ['%B', d => d.getMonth()],
  ['%Y', () => true]
]);

const customNumberFormatter = formatMetric;

export default class SimpleChart extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      selectedTime: null,
    };
    this.selectTime = this.selectTime.bind(this);
  }

  componentDidUpdate() {
    this.renderAxis();
  }

  componentDidMount() {
    this.renderAxis();
  }

  selectTime(date) {
    log('selectTime', date);
    this.setState({selectedTime: date});
  }

  renderAxis() {
    const xAxisNode = ReactDOM.findDOMNode(this.xAxisRef);
    if (!xAxisNode) {
      return;
    }
    d3.select(xAxisNode).call(this.xAxis)
      .selectAll('text')
      .attr('x', 4)
      .style('text-anchor', 'start');

    const yAxisNode = ReactDOM.findDOMNode(this.yAxisRef);
    d3.select(yAxisNode).call(this.yAxis);
  }

  renderLegend(x, y, w, h, indexedSeries) {
    const offsetX = w + 16;
    const offsetY = 8;
    return (
      <g className="legend">
        {Object.keys(indexedSeries).map((seriesId, i) => {
          const series = indexedSeries[seriesId];
          return (
            <g className="legend-item" key={seriesId}>
              <rect style={{fill: series.color}}
                x={offsetX} y={offsetY + 20 * i} width="12" height="2" />
              <text style={{fill: series.color}}
                x={offsetX + 16} y={offsetY + 20 * i} dy="0.42em">{series.label}</text>
            </g>
          );
        })}
      </g>
    );
  }

  renderSelected(x, y, w, h, selectedTime, selectedData, indexedSeries) {
    const selectedX = x(selectedTime);
    const selectedFlagX = selectedX + 100 < w ? selectedX : (w - 100);
    const color = seriesId => indexedSeries[seriesId].color;
    const label = seriesId => indexedSeries[seriesId].label;

    return (
      <g className="selected">
        <g transform={`translate(${selectedFlagX}, 0)`}>
          {Object.keys(selectedData).map((k, i) => {
            const v = selectedData[k];
            const text = `${formatMetric(v)} (${label(k)})`;
            return (
              <g key={k}>
                <rect style={{fill: color(k)}} x="0" y={20 * i} width="100" height="20" />
                <text x="4" y={20 * i + 14}>{text}</text>
              </g>
            );
          })}

          <rect className="timeFlag" x="0" y={h} width="100" height="24" />
          <text className="time" x="4" y={h + 16}>{customTimeFormat(selectedTime)}</text>
        </g>

        {Object.keys(selectedData).map(k => {
          const v = selectedData[k];
          const selectedY = y(v);
          return (
            <circle key={k} style={{fill: color(k)}}
              transform={`translate(${selectedX}, 0)`} cx="0" cy={selectedY}
              r="4" />
          );
        })}
      </g>
    );
  }

  renderHoverPaths(x, y, w, h, data) {
    const voronoi = d3.geom.voronoi()
      .x(d => x(d.date))
      .y(0)
      // .y(d => y(d.value))
      .clipExtent([[0, 0], [w, h]]);
    const toPath = d => `M${d.join('L')}Z`;
    const hoverPaths = voronoi(data).map((v, i) => (
      {date: data[i].date, path: toPath(v)}
    ));

    return (
      hoverPaths.map(({date, path}, i) => (
        <path onMouseOver={() => this.selectTime(date)}
          className="voro" key={i} d={path} />
      ))
    );
  }

  render() {
    const {width: outerWidth, height: outerHeight, data: dataSet} = this.props;
    const w = Math.max(outerWidth - MARGINS.left - MARGINS.right, 0);
    const h = outerHeight - MARGINS.top - MARGINS.bottom;
    if (!dataSet || !dataSet.length) {
      return <div>Loading...</div>;
    }
    const ds0 = dataSet[0];

    const indexedSeries = _.fromPairs(dataSet.map(s => [s.id, s]));
    const toKeyValue = (s) => s.data.map(d => [d.date, {[s.id]: d.value}]);
    const indexedData = _.merge.apply(_, dataSet.map(s => _.fromPairs(toKeyValue(s))));
    const allTimes = Object.keys(indexedData).map(dateString => new Date(dateString));
    const allValues = _.flatten(_.values(indexedData).map(_.values));

    const x = d3.time.scale()
      .rangeRound([0, w])
      .domain(d3.extent(allTimes));

    this.xScale = x;

    this.xAxis = d3.svg.axis()
      .ticks(4)
      .scale(x)
      .tickPadding(7)
      .tickSize(-h)
      .tickFormat(customTimeFormat);

    const maxY = ds0.max || d3.max(allValues) || 0;
    const y = d3.scale.linear()
      .rangeRound([h, 0])
      .domain([0, maxY])
      .nice();

    this.yAxis = d3.svg.axis()
      .ticks(3)
      .scale(y)
      .tickPadding(4)
      .orient('left')
      .tickSize(-2 * w)
      .tickFormat(customNumberFormatter);

    const line = d3.svg.line()
      // .interpolate('monotone')
      .x(d => x(d.date))
      .y(d => y(d.value));

    const style = {width: outerWidth, height: outerHeight};

    const selectedTime = this.state.selectedTime;
    const selectedData = indexedData[selectedTime];

    return (
      <div style={style} className="chart"
        onMouseLeave={() => this.selectTime(null)}>
        <svg style={style}>
          <g transform={`translate(${MARGINS.left}, ${MARGINS.top})`}>

            <rect className="background" width={w} height={h} />

            <g className="x axis" transform={`translate(0, ${h})`}
              ref={(ref) => this.xAxisRef = ref} />

            <g className="y axis" ref={(ref) => this.yAxisRef = ref} />

            {dataSet.map((s, i) => (
              <path key={i} style={{stroke: indexedSeries[s.id].color}}
                className="series" d={line(s.data)} />
            ))}

            {selectedData && this.renderSelected(x, y, w, h, selectedTime,
                                                 selectedData, indexedSeries)}

            {this.renderHoverPaths(x, y, w, h, ds0.data)}

            {this.renderLegend(x, y, w, h, indexedSeries)}

            <text className="label" y={y(0) - 4}>{this.props.label}</text>
          </g>
        </svg>
      </div>
    );
  }
}
