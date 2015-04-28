/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AdjacentNode = require('./adjacent-node.js');

var AdjacentNodes = React.createClass({

	render: function() {
		var filterText = this.props.filterText,
			filterAdjacent = this.props.filterAdjacent,
			highlightedAdjacents = this.props.highlightedAdjacents,
			nodes = _.map(this.props.nodes, function(node) {
				return <AdjacentNode key={node.id} node={node} filterText={filterText} allowFilter={filterAdjacent}
					highlightedAdjacents={highlightedAdjacents} />;
			});

		return (
			<div className="adjacent-nodes">
				{nodes}
			</div>
		);
	}

});

module.exports = AdjacentNodes;