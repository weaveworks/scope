/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');
var NavItemTopology = require('./nav-item-topology.js');

var DetailViews = React.createClass({

	onClick: function(ev) {
		ev.preventDefault();
		// AppActions.clickDetailsView(ev.currentTarget.rel);
	},

	getLabel: function(view) {
		if (view.external) {
			return (
				<span className="node-detail-view-label">
					{view.label}
					<span className="glyphicon glyphicon-new-window"></span>
				</span>
			);
		} else {
			return view.label;
		}
	},

	getViews: function() {
		var views = [{
				external: false,
				label: 'Details',
				id: 'explorer'
		}, {
				external: true,
				label: 'Prometheus',
				id: 'prometheus'
		}];

		if (this.props.details && this.props.details.third_party) {
			views = views.concat(this.props.details.third_party);
		}

		return _.map(views, function(view) {
			var className = this.props.active === view.id ? "active" : "";
			var target = view.external ? view.id : '';
			var handleClick = view.external ? null : this.onClick;
			var label = this.getLabel(view);
			var title = view.external ? "Opens in new window" : '';

			return (
				<li className={className}>
					<a href="#" title={title} rel={view.id} target={target} onClick={handleClick}>
						{label}
					</a>
				</li>
			);
		}, this);
	},

	render: function() {
		var views = this.getViews();

		return (
			<ul className="nav nav-tabs">
				{views}
			</ul>
		);
	}

});

module.exports = DetailViews;