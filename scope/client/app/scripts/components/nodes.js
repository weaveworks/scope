/** @jsx React.DOM */

var React = require('react');

var NodesChart = require('../charts/nodes-chart');
var AppActions = require('../actions/app-actions');
var NodesLayouts = require('./nodes-layouts');

var navbarHeight = 160;
var marginTop = 0;
var marginLeft = 0;

var Nodes = React.createClass({

	getInitialState: function() {
		return {
			layout: 'circle',
			width: window.innerWidth,
			height: window.innerHeight - navbarHeight - marginTop
		};
	},

	onNodeClick: function(ev) {
		AppActions.clickNode(ev.currentTarget.id);
	},

	onChangeLayout: function(layout) {
		this.setState({
			layout: layout
		});
	},

	componentDidMount: function() {
		window.addEventListener('resize', this.handleResize);
	},

	componentWillUnmount: function() {
		window.removeEventListener('resize', this.handleResize);
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
			<div id="nodes">
				<NodesLayouts activeLayout={this.state.layout} onChangeLayout={this.onChangeLayout} />
				<div className="graph">
					<NodesChart
						onNodeClick={this.onNodeClick}
						layout={this.state.layout}
						nodes={this.props.nodes}
						width={this.state.width}
						height={this.state.height}
						context="view"
					/>
				</div>
			</div>
		);
	}

});

module.exports = Nodes;