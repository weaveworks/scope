/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var NodeDetailsTable = require('./node-details-table');
var NodeColorMixin = require('../mixins/node-color-mixin');

var NodeDetails = React.createClass({

    mixins: [
        NodeColorMixin
    ],

	render: function() {
		var node = this.props.details;

		if (!node) {
			return <div className="node-details" />;
		}

		var style = {
			"background-color": this.getNodeColorDark(node.label_major)
		};

		return (
			<div className="node-details">
				<div className="node-details-header" style={style}>
					<h2 className="node-details-header-label">
						{node.label_major}
					</h2>
					<div className="node-details-header-label-minor">{node.label_minor}</div>
				</div>

				<div className="node-details-content">
					{this.props.details.tables.map(function(table) {
						return <NodeDetailsTable title={table.title} rows={table.rows} isNumeric={table.numeric} />;
					})}
				</div>
        	</div>
		);
	}

});

module.exports = NodeDetails;