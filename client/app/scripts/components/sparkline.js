// Forked from: https://github.com/KyleAMathews/react-sparkline at commit a9d7c5203d8f240938b9f2288287aaf0478df013
const React = require('react');
const d3 = require('d3');

const Sparkline = React.createClass({
  getDefaultProps: function() {
    return {
      width: 100,
      height: 16,
      strokeColor: 'black',
      strokeWidth: '0.5px',
      interpolate: 'basis',
      circleDiameter: 1.5,
      data: [1, 23, 5, 5, 23, 0, 0, 0, 4, 32, 3, 12, 3, 1, 24, 1, 5, 5, 24, 23] // Some semi-random data.
    };
  },

  componentDidMount: function() {
    return this.renderSparkline();
  },

  renderSparkline: function() {
    // If the sparkline has already been rendered, remove it.
    let el = this.getDOMNode();
    while (el.firstChild) {
      el.removeChild(el.firstChild);
    }

    let data = this.props.data.slice();

    // Do nothing if no data is passed in.
    if (data.length === 0) {
      return;
    }

    let x = d3.scale.linear().range([2, this.props.width - 2]);
    let y = d3.scale.linear().range([this.props.height - 2, 2]);

    // react-sparkline allows you to pass in two types of data.
    // Data tied to dates and linear data. We need to change our line and x/y
    // functions depending on the type of data.

    // These are objects with a date key
    let line;
    let lastX;
    let lastY;
    if (data[0].date) {
      // Convert dates into D3 dates
      data.forEach(d => {
        d.date = d3.time.format.iso.parse(d.date);
      });

      line = d3.svg.line().
        interpolate(this.props.interpolate).
        x(d => x(d.date)).
        y(d => y(d.value));

      x.domain([
        this.props.first ? d3.time.format.iso.parse(this.props.first) : d3.min(data, d => d.date),
        this.props.last ? d3.time.format.iso.parse(this.props.last) : d3.max(data, d => d.date)
      ]);

      y.domain([
        this.props.min || d3.min(data, d => d.value),
        this.props.max || d3.max(data, d => d.value)
      ]);

      lastX = x(data[data.length - 1].date);
      lastY = y(data[data.length - 1].value);
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
    }

    let svg = d3.select(this.getDOMNode()).
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
      attr('fill', 'red').
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
