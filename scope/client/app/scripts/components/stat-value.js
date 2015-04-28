/** @jsx React.DOM */

var React = require('react');

var StatValue = React.createClass({

	render: function() {
		return (
			<div className="stat-value">
				<span className="value">{this.props.value}</span>
				<span className="stat-label">{this.props.label}</span>
			</div>
		);
	}

});

module.exports = StatValue;