import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';


const TIMESTAMP_TICK_INTERVAL = 500;

class TopologyTimestampInfo extends React.PureComponent {
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
    const { updatePausedAt, websocketQueryPastAt, websocketQueryPastRequestMadeAt } = this.props;

    let timestamp = updatePausedAt;
    let showingCurrentState = false;

    if (!updatePausedAt) {
      timestamp = moment().utc();
      showingCurrentState = true;

      if (websocketQueryPastAt) {
        const offset = moment(websocketQueryPastRequestMadeAt).diff(moment(websocketQueryPastAt));
        timestamp = timestamp.subtract(offset);
        showingCurrentState = false;
      }
    }
    return { timestamp, showingCurrentState };
  }

  renderTimestamp() {
    return (
      <time>{this.state.timestamp.format('MMMM Do YYYY, h:mm:ss a')}</time>
    );
  }

  render() {
    const { showingCurrentState } = this.state;

    return (
      <span className="topology-timestamp-info">
        {showingCurrentState ? 'now' : this.renderTimestamp()}
      </span>
    );
  }
}

function mapStateToProps(state) {
  return {
    updatePausedAt: state.get('updatePausedAt'),
    websocketQueryPastAt: state.get('websocketQueryPastAt'),
    websocketQueryPastRequestMadeAt: state.get('websocketQueryPastRequestMadeAt'),
  };
}

export default connect(mapStateToProps)(TopologyTimestampInfo);
