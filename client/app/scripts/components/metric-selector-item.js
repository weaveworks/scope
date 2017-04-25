import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { hoverMetric, pinMetric, unpinMetric } from '../actions/app-actions';
import { selectedMetricTypeSelector } from '../selectors/node-metric';


class MetricSelectorItem extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  onMouseOver() {
    const metricType = this.props.metric.get('label');
    this.props.hoverMetric(metricType);
  }

  onMouseClick() {
    const metricType = this.props.metric.get('label');
    const pinnedMetricType = this.props.pinnedMetricType;

    if (metricType !== pinnedMetricType) {
      this.props.pinMetric(metricType);
    } else {
      this.props.unpinMetric();
    }
  }

  render() {
    const { metric, selectedMetricType, pinnedMetricType } = this.props;
    const type = metric.get('label');
    const isPinned = (type === pinnedMetricType);
    const isSelected = (type === selectedMetricType);
    const className = classNames('metric-selector-action', {
      'metric-selector-action-selected': isSelected
    });

    return (
      <div
        key={type}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}>
        {type}
        {isPinned && <span className="fa fa-thumb-tack" />}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    selectedMetricType: selectedMetricTypeSelector(state),
    pinnedMetricType: state.get('pinnedMetricType'),
  };
}

export default connect(
  mapStateToProps,
  { hoverMetric, pinMetric, unpinMetric }
)(MetricSelectorItem);
