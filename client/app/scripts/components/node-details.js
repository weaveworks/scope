const _ = require('lodash');
const React = require('react');

const NodeDetailsControls = require('./node-details/node-details-controls');
const NodeDetailsTable = require('./node-details/node-details-table');
const NodeColorMixin = require('../mixins/node-color-mixin');
const TitleUtils = require('../utils/title-utils');

const NodeDetails = React.createClass({

  mixins: [
    NodeColorMixin
  ],

  componentDidMount: function() {
    this.updateTitle();
  },

  componentWillUnmount: function() {
    TitleUtils.resetTitle();
  },

  renderLoading: function() {
    return (
      <div className="node-details" />
    );
  },

  renderNotAvailable: function() {
    return (
      <div className="node-details">
        <div className="node-details-header node-details-header-notavailable">
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label">
              n/a
            </h2>
            <div className="node-details-header-label-minor truncate">
              {this.props.nodeId}
            </div>
          </div>
        </div>
        <div className="node-details-content">
          <p className="node-details-content-info">
            This node is not visible to Scope anymore.
            The node will re-appear if it communicates again.
          </p>
        </div>
      </div>
    );
  },

  render: function() {
    const details = this.props.details;
    const nodeExists = this.props.nodes && this.props.nodes.has(this.props.nodeId);

    if (!nodeExists) {
      return this.renderNotAvailable();
    }

    if (!details) {
      return this.renderLoading();
    }

    const nodeColor = this.getNodeColorDark(details.label_major);
    const styles = {
      controls: {
        'backgroundColor': this.brightenColor(nodeColor)
      },
      header: {
        'backgroundColor': nodeColor
      }
    };

    return (
      <div className="node-details">
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate" title={details.label_major}>
              {details.label_major}
            </h2>
            <div className="node-details-header-label-minor truncate" title={details.label_minor}>
              {details.label_minor}
            </div>
          </div>
        </div>

        {details.controls && details.controls.length > 0 && <div className="node-details-controls-wrapper" style={styles.controls}>
          <NodeDetailsControls controls={details.controls}
            pending={this.props.controlPending} error={this.props.controlError} />
        </div>}

        <div className="node-details-content">
          {details.tables.map(function(table) {
            const key = _.snakeCase(table.title);
            return <NodeDetailsTable title={table.title} key={key} rows={table.rows} isNumeric={table.numeric} />;
          })}
        </div>
      </div>
    );
  },

  componentDidUpdate: function() {
    this.updateTitle();
  },

  updateTitle: function() {
    TitleUtils.setTitle(this.props.details && this.props.details.label_major);
  }

});

module.exports = NodeDetails;
