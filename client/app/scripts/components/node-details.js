const React = require('react');

const NodeDetailsTable = require('./node-details-table');
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
          <h2 className="node-details-header-label truncate">
            n/a
          </h2>
          <div className="node-details-header-label-minor truncate">
            {this.props.nodeId}
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

    const style = {
      'backgroundColor': this.getNodeColorDark(details.label_major)
    };

    return (
      <div className="node-details">
        <div className="node-details-header" style={style}>
          <h2 className="node-details-header-label truncate">
            {details.label_major}
          </h2>
          <div className="node-details-header-label-minor truncate">{details.label_minor}</div>
        </div>

        <div className="node-details-content">
          {details.tables.map(function(table) {
            return <NodeDetailsTable title={table.title} rows={table.rows} isNumeric={table.numeric} />;
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
