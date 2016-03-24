import React from 'react';
import classNames from 'classnames';
import { selectMetric, lockMetric, unlockMetric } from '../actions/app-actions';


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
    const className = classNames('metric-selector-action', {
      'metric-selector-action-locked': isLocked,
      'metric-selector-action-selected': isSelected
    });

    return (
      <div
        key={id}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}>
        {metric.label}
        {isLocked && <span className="fa fa-thumb-tack"></span>}
      </div>
    );
  }
}
