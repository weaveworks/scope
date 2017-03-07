import React from 'react';
import { connect } from 'react-redux';

import { selectMetric } from '../actions/app-actions';
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
    return (
      <div className="metric-selector">
        <div className="metric-selector-wrapper" onMouseLeave={this.onMouseOut}>
          {this.props.availableCanvasMetrics.map(metric => (
            <MetricSelectorItem key={metric.get('id')} metric={metric} />
          ))}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    availableCanvasMetrics: state.get('availableCanvasMetrics'),
    pinnedMetric: state.get('pinnedMetric')
  };
}

export default connect(
  mapStateToProps,
  { selectMetric }
)(MetricSelector);
