import React from 'react';
import _ from 'lodash';
import { selectMetric, lockMetric, unlockMetric } from '../actions/app-actions';
import classNames from 'classnames';

// const CROSS = '\u274C';
// const MINUS = '\u2212';
// const DOT = '\u2022';
//

const METRIC_LABELS = {
  docker_cpu_total_usage: 'Container CPU',
  docker_memory_usage: 'Container Memory',
  host_cpu_usage_percent: 'Host CPU',
  host_mem_usage_bytes: 'Host Memory',
  load1: 'Host Load 1',
  load15: 'Host Load 15',
  load5: 'Host Load 5',
  open_files_count: 'Process Open files',
  process_cpu_usage_percent: 'Process CPU',
  process_memory_usage_bytes: 'Process Memory'
};

function onMouseOver(k) {
  selectMetric(k);
}

function onMouseClick(k, lockedMetric) {
  if (k === lockedMetric) {
    unlockMetric(k);
  } else {
    lockMetric(k);
  }
}

function onMouseOut(k) {
  selectMetric(k);
}

function label(m) {
  return METRIC_LABELS[m.id];
}

export default function MetricSelector({availableCanvasMetrics, selectedMetric, lockedMetric}) {
  return (
    <div
      className="available-metrics"
      onMouseLeave={() => onMouseOut(lockedMetric)}>
      <div className="sidebar-item">
        METRICS
      </div>
      {_.sortBy(availableCanvasMetrics, label).map(m => {
        const id = m.id;
        const isLocked = (id === lockedMetric);
        const isSelected = (id === selectedMetric);
        const className = classNames('sidebar-item', {
          'locked': isLocked,
          'selected': isSelected
        });

        return (
          <div
            key={id}
            className={className}
            onMouseOver={() => onMouseOver(id)}
            onClick={() => onMouseClick(id, lockedMetric)}>
            {label(m)}
            {isLocked && <span className="sidebar-item-actions">
              <span className="sidebar-item-action fa fa-thumb-tack"></span>
            </span>}
          </div>
        );
      })}
    </div>
  );
}
