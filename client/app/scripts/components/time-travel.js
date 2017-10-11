import React from 'react';
import { connect } from 'react-redux';

import TimeTravelComponent from './time-travel-component';


class TimeTravel extends React.Component {
  render() {
    const { visible, timestamp } = this.props;

    return (
      <TimeTravelComponent
        visible={visible}
        timestamp={timestamp}
      />
    );
  }
}

function mapStateToProps(state) {
  return {
    visible: state.get('showingTimeTravel'),
    // topologyViewMode: state.get('topologyViewMode'),
    // currentTopology: state.get('currentTopology'),
    timestamp: state.get('pausedAt'),
  };
}

export default connect(mapStateToProps)(TimeTravel);
