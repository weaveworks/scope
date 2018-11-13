import React from 'react';
import { connect } from 'react-redux';

import {
  toggleTroubleshootingMenu,
  resetLocalViewState,
  clickDownloadGraph
} from '../actions/app-actions';
import { getApiPath } from '../utils/web-api-utils';

class DebugMenu extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleClickReset = this.handleClickReset.bind(this);
  }

  handleClickReset(ev) {
    ev.preventDefault();
    this.props.resetLocalViewState();
  }

  render() {
    const reportDownloadUrl = process.env.WEAVE_CLOUD
      ? `${getApiPath()}/api/report`
      : 'api/report';
    return (
      <div className="troubleshooting-menu-wrapper">
        <div className="troubleshooting-menu">
          <div className="troubleshooting-menu-content">
            <h3>Debugging/Troubleshooting</h3>
            <div className="troubleshooting-menu-item">
              <a
                className="footer-icon"
                href={reportDownloadUrl}
                download
                title="Save raw data as JSON"
              >
                <i className="fa fa-code" />
                <span className="description">
                  Save raw data as JSON
                </span>
              </a>
            </div>
            <div className="troubleshooting-menu-item">
              <button
                className="footer-icon"
                onClick={this.props.clickDownloadGraph}
                title="Save canvas as SVG (does not include search highlighting)"
              >
                <i className="fa fa-download" />
                <span className="description">
                  Save canvas as SVG (does not include search highlighting)
                </span>
              </button>
            </div>
            <div className="troubleshooting-menu-item">
              <button
                className="footer-icon"
                title="Reset view state"
                onClick={this.handleClickReset}
              >
                <i className="fa fa-undo" />
                <span className="description">Reset your local view state and reload the page</span>
              </button>
            </div>
            <div className="troubleshooting-menu-item">
              <a
                className="footer-icon"
                title="Report an issue"
                href="https://gitreports.com/issue/weaveworks/scope"
                target="_blank"
                rel="noopener noreferrer"
              >
                <i className="fa fa-bug" />
                <span className="description">Report a bug</span>
              </a>
            </div>
            <div className="help-panel-tools">
              <i
                title="Close menu"
                className="fa fa-times"
                onClick={this.props.toggleTroubleshootingMenu}
              />
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(null, {
  toggleTroubleshootingMenu,
  resetLocalViewState,
  clickDownloadGraph
})(DebugMenu);
