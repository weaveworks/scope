/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var AppStore = require('../stores/app-store');
var NavItemTopology = require('./nav-item-topology.js');

var Topologies = React.createClass({

	getInitialState: function() {
		return {
			topologiesVisible: false 
		};
	},

	onTopologyClick: function(ev) {
		ev.preventDefault();
		AppActions.clickTopology(ev.currentTarget.rel);
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
				topologiesVisible: isOver 
			});
		}, 200);
	},

	getActiveTopology: function() {
		var activeTopology = _.find(this.props.topologies, function(topology) {
			return AppStore.getTopologyForUrl(topology.url) === this.props.active;
		}, this);

		if (activeTopology) {
			var title = ['Topology: ' + activeTopology.name,
				'Nodes: ' + activeTopology.stats.node_count,
				'Connections: ' + activeTopology.stats.node_count].join('\n');

			return (
				<div className="nav-preview" onMouseOver={this.onMouseOver} onMouseOut={this.onMouseOut}>
					<div rel={this.props.active} title={title}
						onClick={this.props.onClick}>
						<div className="nav-topology-frame">
							<span className="nav-topology-nodes">{activeTopology.stats.node_count}</span>
							<span className="nav-topology-divider" />
							<span className="nav-topology-edges">{activeTopology.stats.edge_count}</span>
						</div>
						<div className="nav-label">
							{activeTopology.name}
						</div>
					</div>
				</div>
			);
		}

		return (<div />);
	},

	render: function() {
		var handleClick = this.onTopologyClick,
			onMouseOut = this.onMouseOut,
			onMouseOver = this.onMouseOver,
			activeTopologyId = this.props.active,
			activeTopology = this.getActiveTopology(),
			navClassName = "nav navbar-nav";
			topologies = _.sortBy(this.props.topologies, function(topology) {
				return topology.name;
			});

		return (
			<div className="navbar-topology">
				{activeTopology}
				{this.state.topologiesVisible && <ul className={navClassName} onMouseOut={this.onMouseOut} onMouseOver={this.onMouseOver}>
					{_.map(this.props.topologies, function(topology) {
						var topologyId = AppStore.getTopologyForUrl(topology.url),
							activeClass = topologyId === activeTopologyId ? 'active' : '';

						return (
							<NavItemTopology
								key={topologyId}
								rel={topologyId}
								url={topology.url}
								nodes={topology.stats.node_count}
								edges={topology.stats.edge_count}
								active={activeClass}
								onClick={handleClick}
								name={topology.name} />
						);
					})}
				</ul>}
			</div>
		);
	}

});

module.exports = Topologies;
