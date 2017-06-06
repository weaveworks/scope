import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';


const TIMESTAMP_TICK_INTERVAL = 500;

class TopologyTimestampButton extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.state = this.getFreshState();
  }

  componentDidMount() {
    this.timer = setInterval(() => {
      if (!this.props.paused) {
        this.setState(this.getFreshState());
      }
    }, TIMESTAMP_TICK_INTERVAL);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  getFreshState() {
    const { updatePausedAt, offset } = this.props;

    let timestamp = updatePausedAt;
    let showingCurrentState = false;

    if (!updatePausedAt) {
      timestamp = moment().utc();
      showingCurrentState = true;

      if (offset >= 1000) {
        timestamp = timestamp.subtract(offset);
        showingCurrentState = false;
      }
    }
    return { timestamp, showingCurrentState };
  }

  renderTimestamp() {
    return (
      <time>{this.state.timestamp.format('MMMM Do YYYY, h:mm:ss a')} UTC</time>
    );
  }

  render() {
    const { selected, onClick } = this.props;
    const { showingCurrentState } = this.state;
    const className = classNames('button topology-timestamp-button', {
      selected, current: showingCurrentState,
    });

    return (
      <a className={className} onClick={onClick}>
        <span className="topology-timestamp-info">
          {showingCurrentState ? 'now' : this.renderTimestamp()}
        </span>
        <span className="fa fa-clock-o" />
      </a>
    );
  }
}

function mapStateToProps(state) {
  return {
    updatePausedAt: state.get('updatePausedAt'),
  };
}

export default connect(mapStateToProps)(TopologyTimestampButton);
