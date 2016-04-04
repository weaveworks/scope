import React from 'react';
import classNames from 'classnames';
import { selectMetric, pinMetric, unpinMetric } from '../actions/app-actions';


export class MetricSelectorItem extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  onMouseOver() {
    const k = this.props.metric.get('id');
    selectMetric(k);
  }

  onMouseClick() {
    const k = this.props.metric.get('id');
    const pinnedMetric = this.props.pinnedMetric;

    if (k === pinnedMetric) {
      unpinMetric(k);
    } else {
      pinMetric(k);
    }
  }

  render() {
    const {metric, selectedMetric, pinnedMetric} = this.props;
    const id = metric.get('id');
    const isPinned = (id === pinnedMetric);
    const isSelected = (id === selectedMetric);
    const className = classNames('metric-selector-action', {
      'metric-selector-action-selected': isSelected
    });

    return (
      <div
        key={id}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}>
        {metric.get('label')}
        {isPinned && <span className="fa fa-thumb-tack"></span>}
      </div>
    );
  }
}
