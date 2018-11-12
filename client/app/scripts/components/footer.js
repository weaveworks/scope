import React from 'react';
import { connect } from 'react-redux';

import Plugins from './plugins';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import {
  clickDownloadGraph,
  clickForceRelayout,
  toggleHelp,
  toggleTroubleshootingMenu,
  setContrastMode
} from '../actions/app-actions';


class Footer extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleContrastClick = this.handleContrastClick.bind(this);
    this.handleRelayoutClick = this.handleRelayoutClick.bind(this);
  }

  handleContrastClick(ev) {
    ev.preventDefault();
    this.props.setContrastMode(!this.props.contrastMode);
  }

  handleRelayoutClick(ev) {
    ev.preventDefault();
    trackAnalyticsEvent('scope.layout.refresh.click', {
      layout: this.props.topologyViewMode,
    });
    this.props.clickForceRelayout();
  }

  render() {
    const {
      hostname, version, versionUpdate, contrastMode
    } = this.props;

    const otherContrastModeTitle = contrastMode
      ? 'Switch to normal contrast' : 'Switch to high contrast';
    const forceRelayoutTitle = 'Force re-layout (might reduce edge crossings, '
      + 'but may shift nodes around)';
    const versionUpdateTitle = versionUpdate
      ? `New version available: ${versionUpdate.get('version')} Click to download`
      : '';

    return (
      <div className="footer">
        <div className="footer-status">
          {versionUpdate &&
            <a
              className="footer-versionupdate"
              title={versionUpdateTitle}
              href={versionUpdate.get('downloadUrl')}
              target="_blank"
              rel="noopener noreferrer">
              Update available: {versionUpdate.get('version')}
            </a>
          }
          <span className="footer-label">Version</span>
          {version || '...'}
          <span className="footer-label">on</span>
          {hostname}
        </div>

        <div className="footer-plugins">
          <Plugins />
        </div>

        <div className="footer-tools">
          <button
            className="footer-icon"
            onClick={this.handleRelayoutClick}
            title={forceRelayoutTitle}>
            <span className="fa fa-refresh" />
          </button>
          <button onClick={this.handleContrastClick} className="footer-icon" title={otherContrastModeTitle}>
            <span className="fa fa-adjust" />
          </button>
          <button
            onClick={this.props.toggleTroubleshootingMenu}
            className="footer-icon"
            title="Open troubleshooting menu"
            href=""
          >
            <span className="fa fa-bug" />
          </button>
          <button className="footer-icon" onClick={this.props.toggleHelp} title="Show help">
            <span className="fa fa-question" />
          </button>
        </div>

      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    hostname: state.get('hostname'),
    topologyViewMode: state.get('topologyViewMode'),
    version: state.get('version'),
    versionUpdate: state.get('versionUpdate'),
    contrastMode: state.get('contrastMode'),
  };
}

export default connect(
  mapStateToProps,
  {
    clickDownloadGraph,
    clickForceRelayout,
    toggleHelp,
    toggleTroubleshootingMenu,
    setContrastMode
  }
)(Footer);
