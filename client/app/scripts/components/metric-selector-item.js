import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { hoverMetric, pinMetric, unpinMetric } from '../actions/app-actions';
import { selectedMetricTypeSelector } from '../selectors/node-metric';
import { trackAnalyticsEvent } from '../utils/tracking-utils';


class MetricSelectorItem extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  trackEvent(eventName) {
    trackAnalyticsEvent(eventName, {
      layout: this.props.topologyViewMode,
      metricType: this.props.metric.get('label'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
      topologyId: this.props.currentTopology.get('id'),
    });
  }

  onMouseOver() {
    const metricType = this.props.metric.get('label');
    this.props.hoverMetric(metricType);
  }

  onMouseClick() {
    const metricType = this.props.metric.get('label');
    const { pinnedMetricType } = this.props;

    if (metricType !== pinnedMetricType) {
      this.trackEvent('scope.metric.selector.pin.click');
      this.props.pinMetric(metricType);
    } else {
      this.trackEvent('scope.metric.selector.unpin.click');
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
        {isPinned && <i className="fa fa-thumbtack" />}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    currentTopology: state.get('currentTopology'),
    pinnedMetricType: state.get('pinnedMetricType'),
    selectedMetricType: selectedMetricTypeSelector(state),
    topologyViewMode: state.get('topologyViewMode'),
  };
}

export default connect(
  mapStateToProps,
  { hoverMetric, pinMetric, unpinMetric }
)(MetricSelectorItem);
