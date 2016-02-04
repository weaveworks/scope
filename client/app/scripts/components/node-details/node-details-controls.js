import React from 'react';
import { showMetrics } from '../../actions/app-actions';

import NodeDetailsControlButton from './node-details-control-button';

export default class NodeDetailsControls extends React.Component {
  render() {
    let spinnerClassName = 'fa fa-circle-o-notch fa-spin';
    if (this.props.pending) {
      spinnerClassName += ' node-details-controls-spinner';
    } else {
      spinnerClassName += ' node-details-controls-spinner hide';
    }

    return (
      <div className="node-details-controls">
        {this.props.error && <div className="node-details-controls-error" title={this.props.error}>
          <span className="node-details-controls-error-icon fa fa-warning" />
          <span className="node-details-controls-error-messages">{this.props.error}</span>
        </div>}
        <span className="node-details-controls-buttons">
          {this.props.controls && this.props.controls.map(control => {
            return (
              <NodeDetailsControlButton nodeId={this.props.nodeId} control={control}
                pending={this.props.pending} key={control.id} />
            );
          })}
          {this.props.metrics && <span
            className="node-control-button fa fa-line-chart"
            title="Show Metrics"
            onClick={() => showMetrics(this.props.nodeId)} />}
          </span>
        {this.props.controls && <span title="Applying..." className={spinnerClassName}></span>}
      </div>
    );
  }
}
