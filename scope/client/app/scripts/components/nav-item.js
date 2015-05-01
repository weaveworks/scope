/** @jsx React.DOM */

var React = require('react');

var NavItem = React.createClass({

	render: function() {
		return (
			<li className={this.props.active}>
				<a href={this.props.url} rel={this.props.rel} onClick={this.props.onClick}>
					{this.props.children}
				</a>
			</li>
		);
	}

});

module.exports = NavItem;