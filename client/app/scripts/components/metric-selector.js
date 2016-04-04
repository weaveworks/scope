import React from 'react';
import { selectMetric } from '../actions/app-actions';
import { MetricSelectorItem } from './metric-selector-item';


export default class MetricSelector extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onMouseOut() {
    selectMetric(this.props.pinnedMetric);
  }

  render() {
    const {availableCanvasMetrics} = this.props;

    const items = availableCanvasMetrics.map(metric => (
      <MetricSelectorItem key={metric.get('id')} metric={metric} {...this.props} />
    ));

    return (
      <div
        className="metric-selector">
        <div className="metric-selector-wrapper" onMouseLeave={this.onMouseOut}>
          {items}
        </div>
      </div>
    );
  }
}

