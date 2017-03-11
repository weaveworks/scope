import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { selectMetric, pinMetric, unpinMetric } from '../actions/app-actions';

class MetricSelectorItem extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  onMouseOver() {
    const k = this.props.metric.get('id');
    this.props.selectMetric(k);
  }

  onMouseClick() {
    const k = this.props.metric.get('id');
    const pinnedMetric = this.props.pinnedMetric;

    if (k !== pinnedMetric) {
      this.props.pinMetric(k);
    } else if (!this.props.alwaysPinned) {
      this.props.unpinMetric(k);
    }
  }

  render() {
    const {metric, selectedMetric, pinnedMetric} = this.props;
    const id = metric.get('id');
    const isPinned = (id === pinnedMetric);
    const isSelected = (id === selectedMetric);
    const className = classNames('metric-selector-action', {
      'metric-selector-action-selected': isSelected
    });

    return (
      <div
        key={id}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}>
        {metric.get('label')}
        {isPinned && <span className="fa fa-thumb-tack" />}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    selectedMetric: state.get('selectedMetric'),
    pinnedMetric: state.get('pinnedMetric')
  };
}

export default connect(
  mapStateToProps,
  { selectMetric, pinMetric, unpinMetric }
)(MetricSelectorItem);
