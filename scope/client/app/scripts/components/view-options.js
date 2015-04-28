/** @jsx React.DOM */

var React = require('react');
var _ = require('lodash');

var AppActions = require('../actions/app-actions');

var ViewOptions = React.createClass({

	getInitialState: function() {
		return {
			menuVisible: false 
		};
	},

	onClick: function(ev) {
		ev.preventDefault();
		AppActions.clickTopologyMode(ev.currentTarget.rel);
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
				menuVisible: isOver 
			});
		}, 200);
	},

	getActiveViewMode: function() {
		var className = this.props.active === 'class' ? "glyphicon glyphicon-th-large" : "glyphicon glyphicon-th";

		return (
			<div className="nav-preview" onMouseOver={this.onMouseOver} onMouseOut={this.onMouseOut}>
				<div rel={this.props.active} onClick={this.props.onClick}>
					<div className="nav-icon">
						<span className={className}></span>
					</div>
					<div className="nav-label">
						View Options
					</div>
				</div>
			</div>
		);
	},

	render: function() {
		var activeMode = this.props.active,
			activeOptions = this.getActiveViewMode(),
			baseClass = "",
			individualClass = activeMode == 'individual' ? baseClass + ' active' : baseClass,
			classClass = activeMode == 'class' ? baseClass + ' active' : baseClass,
			navClassName = "nav navbar-nav";

		return (
			<div className="navbar-view-options">
				{activeOptions}
				{this.state.menuVisible && <ul className={navClassName} onMouseOut={this.onMouseOut} onMouseOver={this.onMouseOver}>
					<li className={individualClass}>
						<a href="#" className="row" rel="individual" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<span className="glyphicon glyphicon-th"></span>
							</div>
							<div className="col-xs-7 nav-item-label">
								Standard View
							</div>
						</a>
					</li>
					<li className={classClass}>
						<a href="#" className="row" rel="class" onClick={this.onClick}>
							<div className="col-xs-5 nav-item-preview">
								<span className="glyphicon glyphicon-th-large"></span>
							</div>
							<div className="col-xs-7 nav-item-label">
								Group View
							</div>
						</a>
					</li>
				</ul>}
	        </div>
		);
	}

});

module.exports = ViewOptions;