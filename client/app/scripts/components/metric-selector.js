import React from 'react';
import _ from 'lodash';
import { selectMetric, lockMetric } from '../actions/app-actions';
import classNames from 'classnames';

const METRICS = {
  'CPU': 'process_cpu_usage_percent',
  'Memory': 'process_memory_usage_bytes',
  'Open Files': 'open_files_count'
};

// docker_cpu_total_usage
// docker_memory_usage

function onMouseOver(k) {
  return selectMetric(k);
}

function onMouseClick(k) {
  return lockMetric(k);
}

function onMouseOut(k) {
  console.log('onMouseOut', k);
  selectMetric(k);
}

export default function MetricSelector({selectedMetric, lockedMetric}) {
  return (
    <div
      className="available-metrics"
      onMouseLeave={() => onMouseOut(lockedMetric)}>
      {_.map(METRICS, (key, name) => {
        return (
          <div
            key={key}
            className={classNames('sidebar-item', {
              'locked': (key === lockedMetric),
              'selected': (key === selectedMetric)
            })}
            onMouseOver={() => onMouseOver(key)}
            onClick={() => onMouseClick(key)}>
            {name}
          </div>
        );
      })}
    </div>
  );
}
