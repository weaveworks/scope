/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var NodeDetailsTable = require('./node-details-table');

var NodeDetails = React.createClass({

	render: function() {
		var node = this.props.details;

		if (!node) {
			return <div className="node-details" />;
		}

		return (
			<div className="node-details">
				<h2 className="node-details-label">
					{node.label_major} <span className="node-details-label-minor">{node.label_minor}</span>
				</h2>

				{this.props.details.tables.map(function(table) {
					return <NodeDetailsTable title={table.title} rows={table.rows} />;
				})}
        	</div>
		);
	}

});

module.exports = NodeDetails;