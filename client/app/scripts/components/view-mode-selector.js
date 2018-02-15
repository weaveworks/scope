import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import MetricSelector from './metric-selector';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { setGraphView, setTableView, setResourceView } from '../actions/app-actions';
import { availableMetricsSelector } from '../selectors/node-metric';
import {
  isResourceViewModeSelector,
  resourceViewAvailableSelector,
} from '../selectors/topology';
import {
  GRAPH_VIEW_MODE,
  TABLE_VIEW_MODE,
  RESOURCE_VIEW_MODE,
} from '../constants/naming';


class ViewModeSelector extends React.Component {
  componentWillReceiveProps(nextProps) {
    if (nextProps.isResourceViewMode && !nextProps.hasResourceView) {
      nextProps.setGraphView();
    }
  }

  renderItem(icons, label, viewMode, setViewModeAction, isEnabled = true) {
    if (label === 'Table') console.log('render table view action');
    const isSelected = (this.props.topologyViewMode === viewMode);
    const className = classNames(`view-mode-selector-action view-${label}-action`, {
      'view-mode-selector-action-selected': isSelected,
    });
    const onClick = () => {
      trackAnalyticsEvent('scope.layout.selector.click', {
        layout: viewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
      setViewModeAction();
    };

    return (
      <div
        className={className}
        disabled={!isEnabled}
        onClick={isEnabled ? onClick : undefined}
        title={`View ${label.toLowerCase()}`}>
        <span className={icons} style={{ fontSize: 12 }} />
        <span className="label">{label}</span>
      </div>
    );
  }

  render() {
    const { hasResourceView } = this.props;

    return (
      <div className="view-mode-selector">
        <div className="view-mode-selector-wrapper">
          {this.renderItem('fa fa-share-alt', 'Graph', GRAPH_VIEW_MODE, this.props.setGraphView)}
          {this.renderItem('fa fa-table', 'Table', TABLE_VIEW_MODE, this.props.setTableView)}
          {this.renderItem(
'fa fa-bar-chart', 'Resources', RESOURCE_VIEW_MODE,
            this.props.setResourceView, hasResourceView
)}
        </div>
        <MetricSelector />
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    isResourceViewMode: isResourceViewModeSelector(state),
    hasResourceView: resourceViewAvailableSelector(state),
    showingMetricsSelector: availableMetricsSelector(state).count() > 0,
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
  };
}

export default connect(
  mapStateToProps,
  { setGraphView, setTableView, setResourceView }
)(ViewModeSelector);
