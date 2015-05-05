var React = require('react');

var NodesLayouts = React.createClass({

	getInitialState: function() {
		return {
			active: 'layered'
		};
	},

	handleClick: function(ev) {
		ev.preventDefault();
		this.props.onChangeLayout(ev.currentTarget.rel);
	},

	render: function() {
		var cmp = this;
		var layouts = ['circle', 'force', 'layered', 'square'];
		var buttons = layouts.map(function(layout) {
			var classes = "btn btn-default";
			if (layout === cmp.props.activeLayout) {
				classes += " active";
			}

			return (
				<a className={classes} rel={layout} onClick={cmp.handleClick} key={layout}>
					{layout}
				</a>
			);
		});

		return (
			<div id="nodes-layouts" className="btn-group-vertical">
				{buttons}
			</div>
		);
	}

});

module.exports = NodesLayouts;