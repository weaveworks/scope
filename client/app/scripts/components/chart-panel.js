import debug from 'debug';
import React from 'react';
import d3 from 'd3';
import { clickCloseMetrics, selectMetric } from '../actions/app-actions';
import { getNodeColor } from '../utils/color-utils';
import FlexChart from '../charts/flex-chart';

const log = debug('scope:chart');

function generateDates(interval, n) {
  const start = new Date('2016-01-10');
  const end = interval.offset(start, n);
  return interval.range(start, end);
}

function zipValsDates(values, dates) {
  return values.map((v, i) => ({value: v, date: dates[i]}));
}

function toDataSet(data) {
  return [{
    id: 'instance-01',
    color: 'steelBlue',
    label: 'CPU',
    data
  }];
}

const simpleData = [1, 3, 2, 5, 7, 3, 1, 2, 6, 8, 7];
const times = generateDates(d3.time.hour, simpleData.length);
const data = toDataSet(zipValsDates(simpleData, times));
log(data);

function getAvailableMetrics(node) {
  const d = node && node.details || {};
  return d.metrics || [];
}

function onChangeSelectedMetric(nodeId, metricId) {
  log('hi');
  selectMetric(nodeId, metricId);
}

export default function ChartPanel({metricQueries, metricData, details}) {
  const queryId = metricQueries.keySeq().first();
  const query = metricQueries.get(queryId);

  const metricId = query.get('metricId');
  const metric = metricData.get(queryId, {});

  const nodeId = query.get('nodeId');
  const node = details.get(nodeId);
  const d = node && node.details || {};
  const onChange = ev => onChangeSelectedMetric(nodeId, ev.target.value);

  const dataSet = {
    id: metricId,
    color: 'steelBlue',
    label: metric.label,
    min: metric.min,
    max: metric.max,
    data: metric.samples && metric.samples.map(({date, value}) => (
      {date: new Date(date), value}
    ))
  };

  return (
    <div className="chart-panel">
      <div className="chart-panel-tools">
        <span
          className="terminal-header-tools-icon fa fa-close"
          onClick={clickCloseMetrics} />
      </div>

      <div className="chart-panel-title">
        <h1 style={{color: getNodeColor(d.rank, d.label)}}>
          {d.label} - {metric.label}
        </h1>
        <select
          value={metricId}
          onChange={onChange}>
          {getAvailableMetrics(node).map(m => (
            <option key={m.id} value={m.id}>{m.label}</option>
          ))}
        </select>
      </div>

      <div className="chart-panel-charts">
        {metric.samples && <FlexChart data={[dataSet]} />}
      </div>
    </div>
  );
}
