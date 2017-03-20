import React from 'react';
import { connect } from 'react-redux';

import { selectMetric } from '../actions/app-actions';
import { availableMetricsSelector, pinnedMetricSelector } from '../selectors/node-metric';
import MetricSelectorItem from './metric-selector-item';

class MetricSelector extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onMouseOut() {
    this.props.selectMetric(this.props.pinnedMetric);
  }

  render() {
    const { pinnedMetric, alwaysPinned, availableMetrics } = this.props;
    const shouldPinMetric = alwaysPinned && !pinnedMetric && !availableMetrics.isEmpty();

    return (
      <div className="metric-selector">
        <div className="metric-selector-wrapper" onMouseLeave={this.onMouseOut}>
          {availableMetrics.map(metric => (
            <MetricSelectorItem
              key={metric.get('id')}
              alwaysPinned={alwaysPinned}
              metric={metric}
            />
          ))}
        </div>
        {shouldPinMetric && <span className="metric-selector-message">
          &laquo; Select a metric
        </span>}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    availableMetrics: availableMetricsSelector(state),
    pinnedMetric: pinnedMetricSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { selectMetric }
)(MetricSelector);
