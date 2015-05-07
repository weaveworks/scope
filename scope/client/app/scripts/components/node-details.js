/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');


var NodeDetails = React.createClass({

	getMetrics: function(metrics) {
		return _.map(metrics, function(value, label) {
			return (
				<div className="node-details-metric">
					<span className="node-details-metric-label">{label}:</span>
					<span className="node-details-metric-value">{value}</span>
				</div>
			);
		});
	},

	render: function() {
		var node = this.props.details;

		if (!node) {
			return <div id="node-details" />;
		}

		var metrics = this.getMetrics(node.aggregate);

		return (
			<div id="node-details">
				<h2>
					{node.label_major} <small>{node.label_minor}</small>
				</h2>
				{metrics.length && <h4 className="node-details-metrics">Metrics</h4>}
				{metrics}
        	</div>
		);
	}

});

module.exports = NodeDetails;