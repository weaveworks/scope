/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var NavItemTopology = require('./nav-item-topology.js');
var AdjacencyPreview = require('./adjacency-preview');
var MatrixPreview = require('./matrix-preview');
var ChordPreview = require('./chord-preview');
var NodesPreview = require('./nodes-preview');
var ViewConstants = require('../constants/views');


var Views = React.createClass({

	getInitialState: function() {
		return {
			viewsVisible: false 
		};
	},

	onClick: function(ev) {
		ev.preventDefault();
		AppActions.clickView(ev.currentTarget.rel);
	},

	onMouseOver: function(ev) {
		this.handleMouseDebounced(true);
	},

	onMouseOut: function(ev) {
		this.handleMouseDebounced(false);
	},

	componentWillMount: function() {
		this.handleMouseDebounced = _.debounce(function(isOver) {
			this.setState({
				viewsVisible: isOver 
			});
		}, 200);
	},

	getActiveView: function() {
		var label = ViewConstants[this.props.active].label;

		var views = {
			adjacency: function() {
				return <AdjacencyPreview nodes={this.props.nodes} />;
			},
			chord: function() {
				return <ChordPreview nodes={this.props.nodes} />;
			},
			matrix: function() {
				return <MatrixPreview nodes={this.props.nodes} />;
			},
			nodes: function() {
				return <NodesPreview nodes={this.props.nodes} />;
			}
		};

		var view = views[this.props.active].call(this);

		return (
			<div>
				{view}
				<div className="nav-label">
					{label}
				</div>
			</div>
		);
	},

	render: function() {
		var handleClick = this.onTopologyClick,
			activeView = this.getActiveView(),
			navClassName = "nav navbar-nav",
			nodesClass = this.props.active === 'nodes' ? 'active' : '',
			chordClass = this.props.active === 'chord' ? 'active' : '',
			matrixClass = this.props.active === 'matrix' ? 'active' : '',
			adjacencyClass = this.props.active === 'adjacency' ? 'active' : '';

		return (
			<div className="navbar-views">
				<div className="nav-preview" onMouseOut={this.onMouseOut} onMouseOver={this.onMouseOver}>
					{activeView}
				</div>
				{this.state.viewsVisible && <ul className={navClassName} onMouseOut={this.onMouseOut} onMouseOver={this.onMouseOver}>
					<li className={nodesClass}>
						<a href="#" className="row" rel="nodes" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<NodesPreview nodes={this.props.nodes} />
							</div>
							<div className="col-xs-7 nav-item-label">
								Nodes View
							</div>
						</a>
					</li>
					<li className={adjacencyClass}>
						<a href="#" className="row" rel="adjacency" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<AdjacencyPreview nodes={this.props.nodes} />
							</div>
							<div className="col-xs-7 nav-item-label">
								Adjacency View
							</div>
						</a>
					</li>
					<li className={matrixClass}>
						<a href="#" className="row" rel="matrix" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<MatrixPreview nodes={this.props.nodes} />
							</div>
							<div className="col-xs-7 nav-item-label">
								Matrix View
							</div>
						</a>
					</li>
					<li className={chordClass}>
						<a href="#" className="row" rel="chord" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<ChordPreview nodes={this.props.nodes} />
							</div>
							<div className="col-xs-7 nav-item-label">
								Chord View
							</div>
						</a>
					</li>
				</ul>}
			</div>
		);
	}

});

module.exports = Views;