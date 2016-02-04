import React from 'react';
import ReactDOM from 'react-dom';
import d3 from 'd3';

import SimpleChart from './simple-chart';
import { INTERVAL_SECS } from '../../constants/timer';
import { now, requestRange, processValues } from '../../utils/prometheus-utils';

const color = d3.scale.category10();
// set colors for http codes for colors from https://github.com/mbostock/d3/wiki/Ordinal-Scales#categorical-colors
color('301');
color('404');
color('200');
color('500');
color('400');

function toSeriesSet(series) {
  return Object.keys(series).map((code, i) => {
    return {
      id: 'ewq' + i,
      color: color(code),
      label: code,
      data: series[code]
    };
  }).reverse();
}

export default class PrometheusChart extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      series: {},
      width: 0
    };
    this.chartTimer = null;
    this.getData = this.getData.bind(this);
    this.receiveData = this.receiveData.bind(this);
  }

  receiveData(json) {
    const result = json.data.result;
    const series = {};
    result.forEach(obj => {
      series[obj.metric.code] = obj.values.map(processValues);
    });
    // sync chart request to other requests by clocking into a slot
    const timeToNextSlot = (now() + INTERVAL_SECS) * 1000 - new Date;
    this.chartTimer = setTimeout(this.getData, timeToNextSlot);
    if (this.mounted) {
      const container = ReactDOM.findDOMNode(this).parentElement;
      const width = container.clientWidth || 100;
      this.setState({series, width});
    }
  }

  getData() {
    const end = now();
    const start = end - 300;
    requestRange(this.props.spec, start, end, this.receiveData);
  }

  componentDidMount() {
    this.mounted = true;
    this.getData();
  }

  componentWillUnmount() {
    this.mounted = false;
    clearTimeout(this.chartTimer);
  }

  render() {
    const data = toSeriesSet(this.state.series);
    return <SimpleChart data={data} label={this.props.label} height="120" width={this.state.width} />;
  }
}
