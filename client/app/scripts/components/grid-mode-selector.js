import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { toggleGridMode, toggleResourceView } from '../actions/app-actions';


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

  constructor(props, context) {
    super(props, context);

    this.enableGridMode = this.enableGridMode.bind(this);
    this.disableGridMode = this.disableGridMode.bind(this);
  }

  enableGridMode() {
    return this.props.toggleGridMode(true);
  }

  disableGridMode() {
    return this.props.toggleGridMode(false);
  }

  render() {
    const { gridMode, resourceView } = this.props;

    return (
      <div className="grid-mode-selector">
        <div className="grid-mode-selector-wrapper">
          {Item('fa fa-share-alt', 'Graph', !gridMode, this.disableGridMode)}
          {Item('fa fa-table', 'Table', gridMode, this.enableGridMode)}
        </div>
        <div className="grid-mode-selector-wrapper">
          {Item('fa fa-bar-chart', 'Resource view', resourceView, this.props.toggleResourceView)}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    gridMode: state.get('gridMode'),
    resourceView: state.get('resourceView'),
  };
}

export default connect(
  mapStateToProps,
  { toggleGridMode, toggleResourceView }
)(GridModeSelector);
