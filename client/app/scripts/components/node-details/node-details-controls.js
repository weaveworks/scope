const React = require('react');

const NodeControlButton = require('../node-control-button');

const NodeDetailsControls = React.createClass({

  render: function() {
    let spinnerClassName = 'fa fa-circle-o-notch fa-spin';
    if (this.props.pending) {
      spinnerClassName += ' node-details-controls-spinner';
    } else {
      spinnerClassName += ' node-details-controls-spinner hide';
    }

    return (
      <div className="node-details-controls">
        <span className="node-details-controls-buttons">
          {this.props.controls && this.props.controls.map(control => {
            return (
              <NodeControlButton control={control} pending={this.props.pending} key={control.id} />
            );
          })}
        </span>
        {this.props.controls && <span title="Applying..." className={spinnerClassName}></span>}
        {this.props.error && <div className="node-details-controls-error" title={this.props.error}>
          <span className="node-details-controls-error-icon fa fa-warning" />
          <span className="node-details-controls-error-messages">{this.props.error}</span>
        </div>}
      </div>
    );
  }

});

module.exports = NodeDetailsControls;
