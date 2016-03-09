import React from 'react';
import { selectMetric, lockMetric, unlockMetric } from '../actions/app-actions';
import classNames from 'classnames';

const CROSS = '\u274C';
// const MINUS = '\u2212';
// const DOT = '\u2022';

// docker_cpu_total_usage
// docker_memory_usage

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

export default function MetricSelector({availableCanvasMetrics, selectedMetric, lockedMetric}) {
  return (
    <div
      className="available-metrics"
      onMouseLeave={() => onMouseOut(lockedMetric)}>
      {availableCanvasMetrics.map(({id, label}) => {
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
            {label}
            {isLocked && <span className="sidebar-item-actions">
              <span className="sidebar-item-action">
                {CROSS}
              </span>
            </span>}
          </div>
        );
      })}
    </div>
  );
}
