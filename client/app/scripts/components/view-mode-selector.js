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
  const className = classNames('view-mode-selector-action', {
    'view-mode-selector-action-selected': isSelected
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

class ViewModeSelector extends React.Component {
  render() {
    const { isGraphViewMode, isTableViewMode, isResourceViewMode } = this.props;

    return (
      <div className="view-mode-selector">
        <div className="view-mode-selector-wrapper">
          {Item('fa fa-share-alt', 'Graph', isGraphViewMode, this.props.setGraphView)}
          {Item('fa fa-table', 'Table', isTableViewMode, this.props.setTableView)}
          {Item('fa fa-bar-chart', 'Resources', isResourceViewMode, this.props.setResourceView)}
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
)(ViewModeSelector);
