import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { isPausedSelector } from '../selectors/time-travel';
import { TIMELINE_TICK_INTERVAL } from '../constants/timer';


class TimeTravelTimestamp extends React.Component {
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
      <time>{timestamp.format()}</time>
    );
  }

  render() {
    const { selected, onClick, millisecondsInPast } = this.props;
    const isCurrent = (millisecondsInPast === 0);

    const className = classNames('button time-travel-timestamp', {
      selected, current: isCurrent
    });

    return (
      <a className={className} onClick={onClick}>
        <span className="time-travel-timestamp-info">
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

export default connect(mapStateToProps)(TimeTravelTimestamp);
