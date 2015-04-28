/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var NodesChart = require('../charts/nodes-chart');
var NodeDetails = require('./node-details');

var marginBottom = 64;
var marginTop = 64;
var marginLeft = 36;
var marginRight = 36;

var Explorer = React.createClass({

	getInitialState: function() {
		return {
			layout: 'solar',
			width: window.innerWidth - marginLeft - marginRight,
			height: window.innerHeight - marginBottom - marginTop
		};
	},

	componentDidMount: function() {
		window.addEventListener('resize', this.handleResize);
	},

	componentWillUnmount: function() {
		window.removeEventListener('resize', this.handleResize);
	},

	setDimensions: function() {
		this.setState({
			height: window.innerHeight - marginBottom - marginTop,
			width: window.innerWidth - marginLeft - marginRight
		});
	},

	handleResize: function() {
		this.setDimensions();
	},

	getSubTopology: function(topology) {
		var subTopology = {};
		var nodeSet = [];

		_.each(this.props.expandedNodes, function(nodeId) {
			if (topology[nodeId]) {
				subTopology[nodeId] = topology[nodeId];
				nodeSet = _.union(subTopology[nodeId].adjacency, nodeSet);
				_.each(subTopology[nodeId].adjacency, function(adjacentId) {
					var node = _.assign({}, topology[adjacentId]);

					subTopology[adjacentId] = node;
				});
			}
		});

		// weed out edges
		_.each(subTopology, function(node) {
			node.adjacency = _.intersection(node.adjacency, nodeSet);
		});

		return subTopology;
	},

	onNodeClick: function(ev) {
		var nodeId = ev.currentTarget.id;
		AppActions.clickNode(nodeId);
	},

	render: function() {
		var subTopology = this.getSubTopology(this.props.nodes);

		return (
			<div id="explorer">
				<NodeDetails details={this.props.details} />
				<div className="graph">
					<NodesChart
						onNodeClick={this.onNodeClick}
						layout={this.state.layout}
						nodes={subTopology}
						highlightedNodes={this.props.expandedNodes}
						width={this.state.width}
						height={this.state.height}
						context="explorer"
					/>
				</div>
        	</div>
		);
	}

});

module.exports = Explorer;