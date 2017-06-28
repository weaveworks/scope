import React from 'react';
import moment from 'moment';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { isPausedSelector } from '../selectors/time-travel';
import { trackMixpanelEvent } from '../utils/tracking-utils';
import { clickPauseUpdate, clickResumeUpdate } from '../actions/app-actions';


class PauseButton extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleClick = this.handleClick.bind(this);
  }

  handleClick() {
    if (this.props.isPaused) {
      trackMixpanelEvent('scope.time.resume.click', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
        nodesDeltaBufferSize: this.props.updateCount,
      });
      this.props.clickResumeUpdate();
    } else {
      trackMixpanelEvent('scope.time.pause.click', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
      this.props.clickPauseUpdate();
    }
  }

  render() {
    const { isPaused, hasUpdates, updateCount, pausedAt } = this.props;
    const className = classNames('button pause-button', { active: isPaused });

    const title = isPaused ?
      `Paused ${moment(pausedAt).fromNow()}` :
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
      <a className={className} onClick={this.handleClick} title={title}>
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
    pausedAt: state.get('pausedAt'),
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
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
