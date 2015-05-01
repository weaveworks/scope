/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');
var AdjacentNode = require('./adjacent-node.js');
var AdjacentNodes = require('./adjacent-nodes.js');

var AdjacencyRow = React.createClass({

	render: function() {
		var nodes = this.props.nodes,
			filterText = this.props.filterText,
			filterAdjacent = this.props.filterAdjacent,
			highlightedAdjacents = this.props.highlightedAdjacents,
			adjacentNodes = _.compact(_.map(this.props.node.adjacency, function(nodeId) {
				return nodes[nodeId];
			}));

		return (
			<tr className={this.props.node.rowClass}>
				<td className="">
					<AdjacentNode node={this.props.node} filterText={filterText} allowFilter={true}
						highlightedAdjacents={highlightedAdjacents} />
				</td>
				<td className="">
					<AdjacentNodes nodes={adjacentNodes} filterText={filterText} filterAdjacent={filterAdjacent}
						highlightedAdjacents={highlightedAdjacents} />
				</td>
			</tr>
		);
	}

});

module.exports = AdjacencyRow;