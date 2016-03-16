import React from 'react';
import classNames from 'classnames';
import { selectMetric, lockMetric, unlockMetric } from '../actions/app-actions';


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


export function label(m) {
  return METRIC_LABELS[m.id];
}


export class MetricSelectorItem extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  onMouseOver() {
    const k = this.props.metric.id;
    selectMetric(k);
  }

  onMouseClick() {
    const k = this.props.metric.id;
    const lockedMetric = this.props.lockedMetric;

    if (k === lockedMetric) {
      unlockMetric(k);
    } else {
      lockMetric(k);
    }
  }

  render() {
    const {metric, selectedMetric, lockedMetric} = this.props;
    const id = metric.id;
    const isLocked = (id === lockedMetric);
    const isSelected = (id === selectedMetric);
    const className = classNames('sidebar-item', {
      locked: isLocked,
      selected: isSelected
    });

    return (
      <div
        key={id}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}>
        {label(metric)}
        {isLocked && <span className="sidebar-item-actions">
          <span className="sidebar-item-action fa fa-thumb-tack"></span>
        </span>}
      </div>
    );
  }
}
