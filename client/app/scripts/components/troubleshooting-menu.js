import React from 'react';
import { connect } from 'react-redux';

import {
  toggleTroubleshootingMenu,
  resetLocalViewState,
  clickDownloadGraph
} from '../actions/app-actions';

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
    return (
      <div className="troubleshooting-menu-wrapper">
        <div className="troubleshooting-menu">
          <div className="troubleshooting-menu-content">
            <h3>Debugging/Troubleshooting</h3>
            <div className="troubleshooting-menu-item">
              <a
                className="footer-icon"
                href="api/report"
                download
                title="Save raw data as JSON"
              >
                <span className="fa fa-code" />
                <span className="description">
                  Save raw data as JSON
                </span>
              </a>
            </div>
            <div className="troubleshooting-menu-item">
              <a
                href=""
                className="footer-icon"
                onClick={this.props.clickDownloadGraph}
                title="Save canvas as SVG (does not include search highlighting)"
              >
                <span className="fa fa-download" />
                <span className="description">
                  Save canvas as SVG (does not include search highlighting)
                </span>
              </a>
            </div>
            <div className="troubleshooting-menu-item">
              <a
                href=""
                className="footer-icon"
                title="Reset view state"
                onClick={this.handleClickReset}
              >
                <span className="fa fa-undo" />
                <span className="description">Reset your local view state and reload the page</span>
              </a>
            </div>
            <div className="troubleshooting-menu-item">
              <a
                className="footer-icon" title="Report an issue"
                href="https://gitreports.com/issue/weaveworks/scope"
                target="_blank" rel="noopener noreferrer"
              >
                <span className="fa fa-bug" />
                <span className="description">Report a bug</span>
              </a>
            </div>
            <div className="help-panel-tools">
              <span
                title="Close menu"
                className="fa fa-close"
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
