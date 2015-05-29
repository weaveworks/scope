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

  render: function() {
    const node = this.props.details;

    if (!node) {
      return <div className="node-details" />;
    }

    const style = {
      'background-color': this.getNodeColorDark(node.label_major)
    };

    return (
      <div className="node-details">
        <div className="node-details-header" style={style}>
          <h2 className="node-details-header-label truncate">
            {node.label_major}
          </h2>
          <div className="node-details-header-label-minor truncate">{node.label_minor}</div>
        </div>

        <div className="node-details-content">
          {this.props.details.tables.map(function(table) {
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
