import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { isPausedSelector } from '../selectors/timeline';
import { clickPauseUpdate, clickResumeUpdate } from '../actions/app-actions';


class PauseButton extends React.Component {
  render() {
    const { isPaused, hasUpdates, updateCount, updatePausedAt } = this.props;
    const action = isPaused ? this.props.clickResumeUpdate : this.props.clickPauseUpdate;
    const className = classNames('button pause-button', { active: isPaused });

    const title = isPaused ?
      `Paused ${moment(updatePausedAt).fromNow()}` :
      'Pause updates (freezes the nodes in their current layout)';

    let label = '';
    if (hasUpdates && isPaused) {
      label = `Paused +${updateCount}`;
    } else if (hasUpdates && !isPaused) {
      label = `Resuming +${updateCount}`;
    } else if (!hasUpdates && isPaused) {
      label = 'Paused';
    }

    return (
      <a className={className} onClick={action} title={title}>
        {label !== '' && <span className="pause-text">{label}</span>}
        <span className="fa fa-pause" />
      </a>
    );
  }
}

function mapStateToProps(state) {
  return {
    hasUpdates: !state.get('nodesDeltaBuffer').isEmpty(),
    updateCount: state.get('nodesDeltaBuffer').size,
    updatePausedAt: state.get('updatePausedAt'),
    isPaused: isPausedSelector(state),
  };
}

export default connect(
  mapStateToProps,
  {
    clickPauseUpdate,
    clickResumeUpdate,
  }
)(PauseButton);
