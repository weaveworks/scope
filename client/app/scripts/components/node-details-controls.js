const React = require('react');

const NodeControlButton = require('./node-control-button');

const NodeDetailsControls = React.createClass({

  render: function() {
    return (
      <div className="node-details-controls">
        {this.props.error && <div className="node-details-controls-error" title={this.props.error}>
          <span className="node-details-controls-error-icon fa fa-warning" />
          <span className="node-details-controls-error-messages">{this.props.error}</span>
        </div>}
        {this.props.controls && this.props.controls.map(control => {
          return (
            <NodeControlButton control={control} pending={this.props.pending} />
          );
        })}
      </div>
    );
  }

});

module.exports = NodeDetailsControls;
