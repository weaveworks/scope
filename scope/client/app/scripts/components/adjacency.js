/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');
var AdjacencyRow = require('./adjacency-row.js');

function sortNodes(nodes) {
	return _.sortBy(nodes, function(node) {
		return [node.label_major, node.label_minor].join('');
	});
}

function filterNodes(nodes, filterText, checkAdjacent) {
	return _.filter(nodes, function(n) {
		return !n.pseudo &&
			(n.label_major.toLowerCase().indexOf(filterText) > -1
			|| n.label_minor && n.label_minor.toLowerCase().indexOf(filterText) > -1
			|| checkAdjacent && filterNodes(_.map(n.adjacency, function(nodeId) {
					return nodes[nodeId];
				}), filterText, false).length > 0);
	});
}

function markRows(nodes) {
	var lastNode;

	return _.each(nodes, function(node) {
		if (lastNode && lastNode.label_major !== node.label_major) {
			node.rowClass = 'adjacency-group-first-row';
		}

		lastNode = node;
	});
}

var Adjacency = React.createClass({

	render: function() {
		var nodes = this.props.nodes,
			filterText = this.props.filterText,
			filterAdjacent = this.props.filterAdjacent,
			highlightedAdjacents = this.props.highlightedAdjacents,
			nodesFiltered = markRows(sortNodes(filterNodes(nodes, this.props.filterText.toLowerCase(), filterAdjacent)));

		return (
			<div className="adjacency">
				<table className="table table-striped adjacency-table">
					<thead>
						<tr>
							<th className="">
								Node
							</th>
							<th className="">
								Connected to
							</th>
						</tr>
					</thead>
					<tbody>
						{nodesFiltered.map(function(node) {
							return (
								<AdjacencyRow key={node.id} node={node} nodes={nodes} filterAdjacent={filterAdjacent} 
									highlightedAdjacents={highlightedAdjacents} filterText={filterText} />
							)
						})}
					</tbody>
				</table>
			</div>
		);
	}

});

module.exports = Adjacency;