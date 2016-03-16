import React from 'react';
import _ from 'lodash';
import { selectMetric } from '../actions/app-actions';
import { MetricSelectorItem, label } from './metric-selector-item';

// const CROSS = '\u274C';
// const MINUS = '\u2212';
// const DOT = '\u2022';
//


export default class MetricSelector extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onMouseOut() {
    selectMetric(this.props.lockedMetric);
  }

  render() {
    const {availableCanvasMetrics} = this.props;

    const items = _.sortBy(availableCanvasMetrics, label).map(metric => (
      <MetricSelectorItem key={metric.id} metric={metric} {...this.props} />
    ));

    return (
      <div
        className="available-metrics"
        onMouseLeave={this.onMouseOut}>
        <div className="sidebar-item">
          METRICS
        </div>
        {items}
      </div>
    );
  }
}

