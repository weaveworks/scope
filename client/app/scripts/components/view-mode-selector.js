import React from 'react';
import { connect } from 'react-redux';

import ViewModeButton from './view-mode-button';
import MetricSelector from './metric-selector';
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

  render() {
    return (
      <div className="view-mode-selector">
        <div className="view-mode-selector-wrapper">
          <ViewModeButton
            label="Graph"
            icons="fa fa-share-alt"
            viewMode={GRAPH_VIEW_MODE}
            onClick={this.props.setGraphView}
          />
          <ViewModeButton
            label="Table"
            icons="fa fa-table"
            viewMode={TABLE_VIEW_MODE}
            onClick={this.props.setTableView}
          />
          <ViewModeButton
            label="Resources"
            icons="fa fa-bar-chart"
            viewMode={RESOURCE_VIEW_MODE}
            onClick={this.props.setResourceView}
            disabled={!this.props.hasResourceView}
          />
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
