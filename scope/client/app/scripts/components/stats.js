/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var StatValue = require('./stat-value.js');

var Stats = React.createClass({

	render: function() {
		var nodeCount = _.size(this.props.nodes),
			edgeCount = _.reduce(this.props.nodes, function(result, node) {
				return result + _.size(node.adjacency);
			}, 0);

		return (
			<div id="stats">
				<div className="col-xs-6">
					<StatValue value={nodeCount} label="Nodes" />
				</div>
				<div className="col-xs-6">
					<StatValue value={edgeCount} label="Connections" />
				</div>
			</div>
		);
	}

});

module.exports = Stats;