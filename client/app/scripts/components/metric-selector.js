import React from 'react';
import { connect } from 'react-redux';

import { unhoverMetric } from '../actions/app-actions';
import { availableMetricsSelector } from '../selectors/node-metric';
import MetricSelectorItem from './metric-selector-item';

class MetricSelector extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onMouseOut() {
    this.props.unhoverMetric();
  }

  render() {
    const { availableMetrics } = this.props;
    const hasMetrics = !availableMetrics.isEmpty();

    return (
      <div className="metric-selector">
        {hasMetrics &&
          <div className="metric-selector-wrapper" onMouseLeave={this.onMouseOut}>
            {availableMetrics.map(metric => (
              <MetricSelectorItem
                key={metric.get('id')}
                metric={metric}
              />
            ))}
          </div>
        }
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
  { unhoverMetric }
)(MetricSelector);
