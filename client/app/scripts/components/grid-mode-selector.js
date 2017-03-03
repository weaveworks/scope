import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { setGraphView, setTableView, setResourceView } from '../actions/app-actions';
import {
  isGraphViewModeSelector,
  isTableViewModeSelector,
  isResourceViewModeSelector,
} from '../selectors/topology';

const Item = (icons, label, isSelected, onClick) => {
  const className = classNames('grid-mode-selector-action', {
    'grid-mode-selector-action-selected': isSelected
  });
  return (
    <div
      className={className}
      onClick={onClick} >
      <span className={icons} style={{fontSize: 12}} />
      <span>{label}</span>
    </div>
  );
};

class GridModeSelector extends React.Component {
  render() {
    const { isGraphViewMode, isTableViewMode, isResourceViewMode } = this.props;

    return (
      <div className="grid-mode-selector">
        <div className="grid-mode-selector-wrapper">
          {Item('fa fa-share-alt', 'Graph', isGraphViewMode, this.props.setGraphView)}
          {Item('fa fa-table', 'Table', isTableViewMode, this.props.setTableView)}
        </div>
        <div className="grid-mode-selector-wrapper">
          {Item('fa fa-bar-chart', 'Resource view', isResourceViewMode, this.props.setResourceView)}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    isGraphViewMode: isGraphViewModeSelector(state),
    isTableViewMode: isTableViewModeSelector(state),
    isResourceViewMode: isResourceViewModeSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { setGraphView, setTableView, setResourceView }
)(GridModeSelector);
