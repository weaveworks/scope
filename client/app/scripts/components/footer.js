import React from 'react';
import moment from 'moment';

import Plugins from './plugins.js';
import { getUpdateBufferSize } from '../utils/update-buffer-utils';
import { contrastModeUrl, isContrastMode } from '../utils/contrast-utils';
import { clickDownloadGraph, clickForceRelayout, clickPauseUpdate,
  clickResumeUpdate, toggleHelp } from '../actions/app-actions';
import { basePathSlash } from '../utils/web-api-utils';

export default function Footer(props) {
  const { hostname, plugins, updatePaused, updatePausedAt, version } = props;
  const contrastMode = isContrastMode();

  // link url to switch contrast with current UI state
  const otherContrastModeUrl = contrastMode
    ? basePathSlash(window.location.pathname) : contrastModeUrl;
  const otherContrastModeTitle = contrastMode
    ? 'Switch to normal contrast' : 'Switch to high contrast';
  const forceRelayoutTitle = 'Force re-layout (might reduce edge crossings, '
    + 'but may shift nodes around)';

  // pause button
  const isPaused = updatePaused;
  const updateCount = getUpdateBufferSize();
  const hasUpdates = updateCount > 0;
  const pausedAgo = moment(updatePausedAt).fromNow();
  const pauseTitle = isPaused
    ? `Paused ${pausedAgo}` : 'Pause updates (freezes the nodes in their current layout)';
  const pauseAction = isPaused ? clickResumeUpdate : clickPauseUpdate;
  const pauseClassName = isPaused ? 'footer-icon footer-icon-active' : 'footer-icon';
  let pauseLabel = '';
  if (hasUpdates && isPaused) {
    pauseLabel = `Paused +${updateCount}`;
  } else if (hasUpdates && !isPaused) {
    pauseLabel = `Resuming +${updateCount}`;
  } else if (!hasUpdates && isPaused) {
    pauseLabel = 'Paused';
  }

  return (
    <div className="footer">

      <div className="footer-status">
        <span className="footer-label">Version</span>
        {version}
        <span className="footer-label">on</span>
        {hostname}
      </div>

      <div className="footer-plugins">
        <Plugins plugins={plugins} />
      </div>

      <div className="footer-tools">
        <a className={pauseClassName} onClick={pauseAction} title={pauseTitle}>
          {pauseLabel !== '' && <span className="footer-label">{pauseLabel}</span>}
          <span className="fa fa-pause" />
        </a>
        <a className="footer-icon" onClick={clickForceRelayout} title={forceRelayoutTitle}>
          <span className="fa fa-refresh" />
        </a>
        <a className="footer-icon" onClick={clickDownloadGraph} title="Save canvas as SVG">
          <span className="fa fa-download" />
        </a>
        <a className="footer-icon" href={otherContrastModeUrl} title={otherContrastModeTitle}>
          <span className="fa fa-adjust" />
        </a>
        <a className="footer-icon" href="https://gitreports.com/issue/weaveworks/scope" target="_blank" title="Report an issue">
          <span className="fa fa-bug" />
        </a>
        <a className="footer-icon" onClick={toggleHelp} title="Show help">
          <span className="fa fa-question" />
        </a>
      </div>

    </div>
  );
}
