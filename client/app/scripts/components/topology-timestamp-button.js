import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { isPausedSelector } from '../selectors/timeline';
import { TIMELINE_TICK_INTERVAL } from '../constants/timer';


class TopologyTimestampButton extends React.Component {
  componentDidMount() {
    this.timer = setInterval(() => {
      if (!this.props.isPaused) {
        this.forceUpdate();
      }
    }, TIMELINE_TICK_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  renderTimestamp() {
    const { isPaused, updatePausedAt, millisecondsInPast } = this.props;
    const timestamp = isPaused ? updatePausedAt : moment().utc().subtract(millisecondsInPast);

    return (
      <time>{timestamp.format('MMMM Do YYYY, h:mm:ss a')} UTC</time>
    );
  }

  render() {
    const { selected, onClick, millisecondsInPast } = this.props;
    const isCurrent = (millisecondsInPast === 0);

    const className = classNames('button topology-timestamp-button', {
      selected, current: isCurrent
    });

    return (
      <a className={className} onClick={onClick}>
        <span className="topology-timestamp-info">
          {isCurrent ? 'now' : this.renderTimestamp()}
        </span>
        <span className="fa fa-clock-o" />
      </a>
    );
  }
}

function mapStateToProps({ scope }) {
  return {
    isPaused: isPausedSelector(scope),
    updatePausedAt: scope.get('updatePausedAt'),
  };
}

export default connect(mapStateToProps)(TopologyTimestampButton);
