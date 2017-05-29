import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';

import EditableTime from './editable-time';
import { getUpdateBufferSize } from '../utils/update-buffer-utils';
import {
  clickPauseUpdate,
  clickResumeUpdate,
} from '../actions/app-actions';


class TimelineControl extends React.PureComponent {
  render() {
    // pause button
    const isPaused = this.props.updatePausedAt !== null;
    const updateCount = getUpdateBufferSize();
    const hasUpdates = updateCount > 0;
    const pauseTitle = isPaused ?
      `Paused ${moment(this.props.updatePausedAt).fromNow()}` :
      'Pause updates (freezes the nodes in their current layout)';
    const pauseAction = isPaused ? this.props.clickResumeUpdate : this.props.clickPauseUpdate;
    let pauseLabel = '';
    if (hasUpdates && isPaused) {
      pauseLabel = `Paused +${updateCount}`;
    } else if (hasUpdates && !isPaused) {
      pauseLabel = `Resuming +${updateCount}`;
    } else if (!hasUpdates && isPaused) {
      pauseLabel = 'Paused';
    }

    return (
      <div className="timeline-control">
        <span className="status-info">
          <EditableTime />
        </span>
        {false && <input type="datetime" onChange={this.handleChange} value={this.state.value} />}
        <a className="button">
          <span className="fa fa-clock-o" />
        </a>
        <a className="button" onClick={pauseAction} title={pauseTitle}>
          {pauseLabel !== '' && <span>{pauseLabel}</span>}
          <span className="fa fa-pause" />
        </a>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    updatePausedAt: state.get('updatePausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    clickPauseUpdate,
    clickResumeUpdate,
  }
)(TimelineControl);
