import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { hoverMetric, pinMetric, unpinMetric } from '../actions/app-actions';
import { selectedMetricTypeSelector } from '../selectors/node-metric';
import { trackMixpanelEvent } from '../utils/tracking-utils';


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
      trackMixpanelEvent('scope.metric.resource.pin', {
        metricType,
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
      this.props.pinMetric(metricType);
    } else {
      trackMixpanelEvent('scope.metric.resource.unpin', {
        metricType,
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
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
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
    pinnedMetricType: state.get('pinnedMetricType'),
    selectedMetricType: selectedMetricTypeSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { hoverMetric, pinMetric, unpinMetric }
)(MetricSelectorItem);
