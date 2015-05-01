/** @jsx React.DOM */

var React = require('react');

var _ = require('lodash');

var MatrixChart = require('../charts/matrix-chart');
var AppActions = require('../actions/app-actions');

var chart = MatrixChart();
var navbarHeight = 160;
var marginTop = 40;
var marginBottom = 16;
var marginLeft = 200;
var padding = 16;

var Matrix = React.createClass({

	getInitialState: function() {
		return {
			width: window.innerHeight - marginLeft - 2 * padding,
			height: window.innerHeight - navbarHeight - marginTop - marginBottom - 2 * padding
		};
	},

	getGraphState: function() {
		return {
			nodes: this.props.nodes
		}
	},

	onNodeClick: function(nodeId) {
		AppActions.clickNode(nodeId);
	},

	componentDidMount: function() {
		window.addEventListener('resize', this.handleResize);

		this.setDimensions();
		chart
			.on('node.click', this.onNodeClick)
			.width(this.state.width)
			.height(this.state.height)
			.create(this.refs.graph.getDOMNode(), this.getGraphState());
	},

	componentWillUnmount: function() {
		window.removeEventListener('resize', this.handleResize);
	},

	componentDidUpdate: function(prevProps, prevState) {
		chart
			.width(this.state.width)
			.height(this.state.height)
			.update(this.refs.graph.getDOMNode(), this.getGraphState());
	},

	setDimensions: function() {
		this.setState({
			height: window.innerHeight - navbarHeight - marginTop - marginBottom - 2 * padding,
			width: window.innerWidth - 2 * padding
		});
	},

	handleResize: function() {
		this.setDimensions();
	},

	render: function() {
		return (
			<div id="matrix">
	            <div className="graph" ref="graph" />
        	</div>
		);
	}

});

module.exports = Matrix;