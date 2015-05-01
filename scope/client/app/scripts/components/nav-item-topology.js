/** @jsx React.DOM */

var React = require('react');

var NavItem = React.createClass({

	render: function() {
		var className = [this.props.active, 'nav-item-topology'].join(' '),
			title = 'Topology: ' + this.props.name
				+ '\nNodes: ' + this.props.nodes
				+ '\nConnections: ' + this.props.edges;

		return (
			<li className={className}>
				<a href={this.props.url} rel={this.props.rel} title={title}
					onClick={this.props.onClick}>
					<div className="col-xs-4 nav-item-topology-frame">
						<span className="nav-item-topology-nodes">{this.props.nodes}</span>
						<span className="nav-item-topology-divider" />
						<span className="nav-item-topology-edges">{this.props.edges}</span>
					</div>
					<div className="col-xs-8 nav-item-label">
						{this.props.name}
					</div>
				</a>
			</li>
		);
	}

});

module.exports = NavItem;