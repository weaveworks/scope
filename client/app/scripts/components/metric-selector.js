import React from 'react';
import { connect } from 'react-redux';

import { selectMetric } from '../actions/app-actions';
import { availableMetricsSelector } from '../selectors/node-metric';
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
    const { alwaysPinned, availableMetrics } = this.props;

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
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    availableMetrics: availableMetricsSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { selectMetric }
)(MetricSelector);
