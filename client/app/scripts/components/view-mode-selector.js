import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import MetricSelector from './metric-selector';
import { setGraphView, setTableView, setResourceView } from '../actions/app-actions';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import { availableMetricsSelector } from '../selectors/node-metric';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';


const Item = (icons, label, isSelected, onClick, isEnabled = true) => {
  const className = classNames('view-mode-selector-action', {
    'view-mode-selector-action-selected': isSelected,
  });
  return (
    <div
      className={className}
      disabled={!isEnabled}
      onClick={isEnabled && onClick}
      title={`View ${label.toLowerCase()}`}>
      <span className={icons} style={{fontSize: 12}} />
      <span className="label">{label}</span>
    </div>
  );
};

class ViewModeSelector extends React.Component {
  componentWillReceiveProps(nextProps) {
    if (nextProps.isResourceViewMode && !nextProps.hasResourceView) {
      nextProps.setGraphView();
    }
  }

  render() {
    const { isGraphViewMode, isTableViewMode, isResourceViewMode, hasResourceView } = this.props;

    return (
      <div className="view-mode-selector">
        <div className="view-mode-selector-wrapper">
          {Item('fa fa-share-alt', 'Graph', isGraphViewMode, this.props.setGraphView)}
          {Item('fa fa-table', 'Table', isTableViewMode, this.props.setTableView)}
          {Item('fa fa-bar-chart', 'Resources', isResourceViewMode, this.props.setResourceView,
            hasResourceView)}
        </div>
        <MetricSelector />
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    isGraphViewMode: isGraphViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
    hasResourceView: !layersTopologyIdsSelector(state).isEmpty(),
    showingMetricsSelector: availableMetricsSelector(state).count() > 0,
  };
}

export default connect(
  mapStateToProps,
  { setGraphView, setTableView, setResourceView }
)(ViewModeSelector);
