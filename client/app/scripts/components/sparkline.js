// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
const React = require('react');
const ReactDOM = require('react-dom');
const d3 = require('d3');

const Sparkline = React.createClass({
  getDefaultProps: function() {
    return {
      width: 100,
      height: 16,
      strokeColor: '#7d7da8',
      strokeWidth: '0.5px',
      interpolate: 'basis',
      circleDiameter: 1.75,
      data: [1, 23, 5, 5, 23, 0, 0, 0, 4, 32, 3, 12, 3, 1, 24, 1, 5, 5, 24, 23] // Some semi-random data.
    };
  },

  componentDidMount: function() {
    return this.renderSparkline();
  },

  renderSparkline: function() {
    // If the sparkline has already been rendered, remove it.
    const el = ReactDOM.findDOMNode(this);
    while (el.firstChild) {
      el.removeChild(el.firstChild);
    }

    const data = this.props.data.slice();

    // Do nothing if no data is passed in.
    if (data.length === 0) {
      return;
    }

    const x = d3.scale.linear().range([2, this.props.width - 2]);
    const y = d3.scale.linear().range([this.props.height - 2, 2]);

    // react-sparkline allows you to pass in two types of data.
    // Data tied to dates and linear data. We need to change our line and x/y
    // functions depending on the type of data.

    // These are objects with a date key
    let line;
    let lastX;
    let lastY;
    let title;
    if (data[0].date) {
      // Convert dates into D3 dates
      data.forEach(d => {
        d.date = d3.time.format.iso.parse(d.date);
      });

      line = d3.svg.line().
        interpolate(this.props.interpolate).
        x(d => x(d.date)).
        y(d => y(d.value));

      const first = this.props.first ? d3.time.format.iso.parse(this.props.first) : d3.min(data, d => d.date);
      const last = this.props.last ? d3.time.format.iso.parse(this.props.last) : d3.max(data, d => d.date);
      x.domain([first, last]);

      y.domain([
        this.props.min || d3.min(data, d => d.value),
        this.props.max || d3.max(data, d => d.value)
      ]);

      lastX = x(data[data.length - 1].date);
      lastY = y(data[data.length - 1].value);
      title = 'Last ' + d3.round((last - first) / 1000) + ' seconds, ' + data.length + ' samples, min: ' + d3.round(d3.min(data, d => d.value), 2) + ', max: ' + d3.round(d3.max(data, d => d.value), 2) + ', mean: ' + d3.round(d3.mean(data, d => d.value), 2);
    } else {
      line = d3.svg.line().
        interpolate(this.props.interpolate).
        x((d, i) => x(i)).
        y(d => y(d));

      x.domain([
        this.props.first || 0,
        this.props.last || data.length
      ]);

      y.domain([
        this.props.min || d3.min(data),
        this.props.max || d3.max(data)
      ]);

      lastX = x(data.length - 1);
      lastY = y(data[data.length - 1]);
      title = data.length + ' samples, min: ' + d3.round(d3.min(data), 2) + ', max: ' + d3.round(d3.max(data), 2) + ', mean: ' + d3.round(d3.mean(data), 2);
    }

    d3.select(ReactDOM.findDOMNode(this)).attr('title', title);

    const svg = d3.select(ReactDOM.findDOMNode(this)).
      append('svg').
      attr('width', this.props.width).
      attr('height', this.props.height).
      append('g');

    svg.append('path').
      datum(data).
      attr('class', 'sparkline').
      style('fill', 'none').
      style('stroke', this.props.strokeColor).
      style('stroke-width', this.props.strokeWidth).
      attr('d', line);

    svg.append('circle').
      attr('class', 'sparkcircle').
      attr('cx', lastX).
      attr('cy', lastY).
      attr('fill', '#46466a').
      attr('fill-opacity', 0.6).
      attr('stroke', 'none').
      attr('r', this.props.circleDiameter);
  },

  render: function() {
    return (
      <div/>
    );
  },

  componentDidUpdate: function() {
    return this.renderSparkline();
  }
});

module.exports = Sparkline;
