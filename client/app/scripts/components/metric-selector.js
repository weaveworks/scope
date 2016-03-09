import React from 'react';
import { selectMetric, lockMetric } from '../actions/app-actions';
import classNames from 'classnames';

// docker_cpu_total_usage
// docker_memory_usage

function onMouseOver(k) {
  return selectMetric(k);
}

function onMouseClick(k) {
  return lockMetric(k);
}

function onMouseOut(k) {
  selectMetric(k);
}

export default function MetricSelector({availableCanvasMetrics, selectedMetric, lockedMetric}) {
  return (
    <div
      className="available-metrics"
      onMouseLeave={() => onMouseOut(lockedMetric)}>
      {availableCanvasMetrics.map(({id, label}) => {
        return (
          <div
            key={id}
            className={classNames('sidebar-item', {
              'locked': (id === lockedMetric),
              'selected': (id === selectedMetric)
            })}
            onMouseOver={() => onMouseOver(id)}
            onClick={() => onMouseClick(id)}>
            {label}
          </div>
        );
      })}
    </div>
  );
}
