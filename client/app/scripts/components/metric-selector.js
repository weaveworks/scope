import React from 'react';
import _ from 'lodash';
import { selectMetric } from '../actions/app-actions';
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

export default function MetricSelector({selectedMetric}) {
  return (
    <div className="available-metrics">
      {_.map(METRICS, (key, name) => {
        return (
          <div
            key={key}
            className={classNames('sidebar-item', {
              'selected': (key === selectedMetric)
            })}
            onMouseOver={() => onMouseOver(key)}>
            {name}
          </div>
        );
      })}
    </div>
  );
}
