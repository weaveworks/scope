/** @jsx React.DOM */

var React = require('react');

var _ = require('lodash');

var ChordChart = require('../charts/chord-chart');
var AppActions = require('../actions/app-actions');

var chart = ChordChart();
var navbarHeight = 160;
var marginTop = 0;
var marginLeft = 0;

var Chord = React.createClass({

	getInitialState: function() {
		return {
			width: window.innerHeight - navbarHeight - marginTop,
			height: window.innerHeight - navbarHeight - marginTop
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
			height: window.innerHeight - navbarHeight - marginTop,
			width: window.innerWidth
		});
	},

	handleResize: function() {
		this.setDimensions();
	},

	render: function() {
		return (
			<div id="chord">
	            <div className="graph" ref="graph" />
        	</div>
		);
	}

});

module.exports = Chord;