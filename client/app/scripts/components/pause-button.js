import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';

import { getUpdateBufferSize } from '../utils/update-buffer-utils';
import { clickPauseUpdate, clickResumeUpdate } from '../actions/app-actions';


class PauseButton extends React.Component {
  render() {
    const isPaused = this.props.updatePausedAt !== null;
    const updateCount = this.props.updateCount;
    const hasUpdates = updateCount > 0;
    const title = isPaused ?
      `Paused ${moment(this.props.updatePausedAt).fromNow()}` :
      'Pause updates (freezes the nodes in their current layout)';
    const action = isPaused ? this.props.clickResumeUpdate : this.props.clickPauseUpdate;
    let label = '';
    if (hasUpdates && isPaused) {
      label = `Paused +${updateCount}`;
    } else if (hasUpdates && !isPaused) {
      label = `Resuming +${updateCount}`;
    } else if (!hasUpdates && isPaused) {
      label = 'Paused';
    }

    return (
      <a className="button pause-button" onClick={action} title={title}>
        {label !== '' && <span className="pause-text">{label}</span>}
        <span className="fa fa-pause" />
      </a>
    );
  }
}

function mapStateToProps(state) {
  return {
    updateCount: getUpdateBufferSize(state),
    updatePausedAt: state.get('updatePausedAt'),
  };
}

export default connect(
  mapStateToProps,
  {
    clickPauseUpdate,
    clickResumeUpdate,
  }
)(PauseButton);
