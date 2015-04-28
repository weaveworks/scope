/** @jsx React.DOM */

var React = require('react');

var _ = require('lodash');

var AdjacencyChart = require('../charts/adjacency-chart');

var chart = AdjacencyChart();

var AdjacencyPreview = React.createClass({

	getGraphState: function() {
		return {
			nodes: this.props.nodes
		}
	},

	componentDidMount: function() {
		chart
			.preview(true)
			.width(62)
			.height(62)
			.create(this.refs.graph.getDOMNode(), this.getGraphState());
	},

	componentDidUpdate: function(prevProps, prevState) {
		chart
			.update(this.refs.graph.getDOMNode(), this.getGraphState());
	},

	render: function() {
		return (
			<div id="adjacency-preview">
	            <div className="graph" ref="graph" />
        	</div>
		);
	}

});

module.exports = AdjacencyPreview;