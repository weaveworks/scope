import React from 'react';
import { connect } from 'react-redux';
import moment from 'moment';

import Plugins from './plugins';
import { getUpdateBufferSize } from '../utils/update-buffer-utils';
import { contrastModeUrl, isContrastMode } from '../utils/contrast-utils';
import { clickDownloadGraph, clickForceRelayout, clickPauseUpdate,
  clickResumeUpdate, toggleHelp, toggleTroubleshootingMenu } from '../actions/app-actions';
import { basePathSlash } from '../utils/web-api-utils';

class Footer extends React.Component {
  render() {
    const { hostname, updatePausedAt, version, versionUpdate } = this.props;
    const contrastMode = isContrastMode();

    // link url to switch contrast with current UI state
    const otherContrastModeUrl = contrastMode
      ? basePathSlash(window.location.pathname) : contrastModeUrl;
    const otherContrastModeTitle = contrastMode
      ? 'Switch to normal contrast' : 'Switch to high contrast';
    const forceRelayoutTitle = 'Force re-layout (might reduce edge crossings, '
      + 'but may shift nodes around)';

    // pause button
    const isPaused = updatePausedAt !== null;
    const updateCount = getUpdateBufferSize();
    const hasUpdates = updateCount > 0;
    const pausedAgo = moment(updatePausedAt).fromNow();
    const pauseTitle = isPaused
      ? `Paused ${pausedAgo}` : 'Pause updates (freezes the nodes in their current layout)';
    const pauseAction = isPaused ? this.props.clickResumeUpdate : this.props.clickPauseUpdate;
    const pauseClassName = isPaused ? 'footer-icon footer-icon-active' : 'footer-icon';
    let pauseLabel = '';
    if (hasUpdates && isPaused) {
      pauseLabel = `Paused +${updateCount}`;
    } else if (hasUpdates && !isPaused) {
      pauseLabel = `Resuming +${updateCount}`;
    } else if (!hasUpdates && isPaused) {
      pauseLabel = 'Paused';
    }

    const versionUpdateTitle = versionUpdate
      ? `New version available: ${versionUpdate.version}. Click to download`
      : '';

    return (
      <div className="footer">

        <div className="footer-status">
          {versionUpdate && <a
            className="footer-versionupdate"
            title={versionUpdateTitle}
            href={versionUpdate.downloadUrl}
            target="_blank" rel="noopener noreferrer">
            Update available: {versionUpdate.version}
          </a>}
          <span className="footer-label">Version</span>
          {version}
          <span className="footer-label">on</span>
          {hostname}
        </div>

        <div className="footer-plugins">
          <Plugins />
        </div>

        <div className="footer-tools">
          <a className={pauseClassName} onClick={pauseAction} title={pauseTitle}>
            {pauseLabel !== '' && <span className="footer-label">{pauseLabel}</span>}
            <span className="fa fa-pause" />
          </a>
          <a
            className="footer-icon"
            onClick={this.props.clickForceRelayout}
            title={forceRelayoutTitle}>
            <span className="fa fa-refresh" />
          </a>
          <a className="footer-icon" href={otherContrastModeUrl} title={otherContrastModeTitle}>
            <span className="fa fa-adjust" />
          </a>
          <a
            onClick={this.props.toggleTroubleshootingMenu}
            className="footer-icon" title="Open troubleshooting menu"
            href=""
          >
            <span className="fa fa-bug" />
          </a>
          <a className="footer-icon" onClick={this.props.toggleHelp} title="Show help">
            <span className="fa fa-question" />
          </a>
        </div>

      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    hostname: state.get('hostname'),
    updatePausedAt: state.get('updatePausedAt'),
    version: state.get('version'),
    versionUpdate: state.get('versionUpdate')
  };
}

export default connect(
  mapStateToProps,
  {
    clickDownloadGraph,
    clickForceRelayout,
    clickPauseUpdate,
    clickResumeUpdate,
    toggleHelp,
    toggleTroubleshootingMenu
  }
)(Footer);
