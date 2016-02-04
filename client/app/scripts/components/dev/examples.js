import React from 'react';
import SimpleChart from '../../charts/simple-chart';
import d3 from 'd3';
import _ from 'lodash';

// Inspired by Lee Byron's test data generator.
function bumpLayer(n) {
  function bump(a) {
    const x = 1 / (0.1 + Math.random());
    const y = 2 * Math.random() - 0.5;
    const z = 10 / (0.1 + Math.random());
    for (let i = 0; i < n; i++) {
      const w = (i / n - y) * z;
      a[i] += x * Math.exp(-w * w);
    }
  }

  const a = [];
  let i;
  for (i = 0; i < n; ++i) a[i] = 0;
  for (i = 0; i < 5; ++i) bump(a);
  return a.map(function(d) { return Math.max(0, d * 100); });
}

const color = d3.scale.category20();

const simpleData = [1, 3, 2, 5, 7, 3, 1, 2, 6, 8, 7];
const simpleData2 = [0, 8, 6, 8, 4, 5, 1, 5, 4, 3, 5];

function generateDates(interval, n) {
  const start = new Date('2016-01-10');
  const end = interval.offset(start, n);
  return interval.range(start, end);
}

function zipValsDates(values, dates) {
  return values.map((v, i) => {
    return {value: v, date: dates[i]};
  });
}

function dataSet(data) {
  return [{
    id: 'instance-01',
    color: 'steelBlue',
    label: 'CPU',
    data: data
  }];
}

function dataSet2(data, data2) {
  return [{
    id: 'instance-01',
    color: 'steelBlue',
    label: 'CPU',
    data: data
  }, {
    id: 'instance-02',
    color: 'green',
    label: 'Mem.',
    data: data2
  }];
}

const DATASET2 = dataSet2(
  zipValsDates(simpleData, generateDates(d3.time.hour, simpleData.length)),
  zipValsDates(simpleData2, generateDates(d3.time.hour, simpleData2.length))
);

const INTERVALS = [
  d3.time.minute,
  d3.time.hour,
  d3.time.day
];

export default class ComponentExamples extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      intervalIndex: 0,
      width: 500,
      height: 200
    };
    this.incrementIntervalIndex = this.incrementIntervalIndex.bind(this);
    this.swapWidthHeight = this.swapWidthHeight.bind(this);
  }

  incrementIntervalIndex(ev) {
    ev.preventDefault();
    const nextIndex = (this.state.intervalIndex + 1) % INTERVALS.length;
    this.setState({intervalIndex: nextIndex});
  }

  getCurrentIntervalData(nSeries = 1) {
    const n = d3.round(Math.random() * 100);
    const interval = INTERVALS[this.state.intervalIndex];
    const times = generateDates(interval, n);
    return _.range(nSeries).map(i => {
      return {
        id: 'zing' + i,
        color: color(i),
        label: 'Z' + i,
        data: zipValsDates(bumpLayer(n), times)
      };
    });
  }

  swapWidthHeight() {
    this.setState({width: this.state.height, height: this.state.width});
  }

  render() {
    return (
      <div className="app-main">
        <h1>Charts</h1>

        <h2>Various time intervals</h2>
        {INTERVALS.map((interval, i) => {
          const times = generateDates(interval, simpleData.length);
          const data = dataSet(zipValsDates(simpleData, times));
          return (
            <SimpleChart key={i} width={500} height={200} data={data} />
          );
        })}

        <h2>Various value ranges</h2>
        {[10, 100000, 1000000].map((n, i) => {
          const times = generateDates(d3.time.hour, simpleData.length);
          const data = dataSet(zipValsDates(simpleData.map(v => v * n), times));
          return (
            <SimpleChart key={i} width={500} height={200} data={data} />
          );
        })}

        <h2>Multiple series</h2>
        <SimpleChart width={500} height={200} data={DATASET2} />

        <h2>Updating data</h2>
        <button onClick={this.incrementIntervalIndex}>Next interval please</button>
        <SimpleChart width={500} height={200} data={this.getCurrentIntervalData(5)} />

        <h2>Tiny</h2>
        <button onClick={this.incrementIntervalIndex}>Next interval please</button>
        <SimpleChart width={300} height={100} data={this.getCurrentIntervalData(5)} />

        <h2>Updating dimensions</h2>
        <button onClick={this.swapWidthHeight}>Swap w/h</button>
        <SimpleChart width={this.state.width} height={this.state.height}
          data={this.getCurrentIntervalData()} />
      </div>
    );
  }
}
