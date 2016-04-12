import React from 'react';
import { addMetric } from '../../actions/app-actions';

import NodeDetailsControlButton from './node-details-control-button';

export default function NodeDetailsControls({metrics, controls, error, nodeId, nodeTopologyId,
                                            pending}) {
  let spinnerClassName = 'fa fa-circle-o-notch fa-spin';
  if (pending) {
    spinnerClassName += ' node-details-controls-spinner';
  } else {
    spinnerClassName += ' node-details-controls-spinner hide';
  }

  const addThisMetric = () => (
    addMetric(nodeId,
              nodeTopologyId,
              metrics[0].id)
  );

  return (
    <div className="node-details-controls">
      {error && <div className="node-details-controls-error" title={error}>
        <span className="node-details-controls-error-icon fa fa-warning" />
        <span className="node-details-controls-error-messages">{error}</span>
      </div>}
      <span className="node-details-controls-buttons">
        {controls && controls.map(control => <NodeDetailsControlButton
          nodeId={nodeId} control={control} pending={pending} key={control.id} />)}
        {metrics && <span
          className="node-control-button"
          title="Show Metrics"
          onClick={addThisMetric}>
            <i className="fa fa-plus"></i>
            <i className="fa fa-line-chart"></i>
        </span>}
      </span>
      {controls && <span title="Applying..." className={spinnerClassName}></span>}
    </div>
  );
}
