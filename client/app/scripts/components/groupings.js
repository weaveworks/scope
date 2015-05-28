const React = require('react');
const _ = require('lodash');

const AppActions = require('../actions/app-actions');

const GROUPINGS = [{
  id: 'none',
  iconClass: 'fa fa-th',
  needsTopology: false
}, {
  id: 'grouped',
  iconClass: 'fa fa-th-large',
  needsTopology: 'grouped_url'
}];

const Groupings = React.createClass({

  onGroupingClick: function(ev) {
    ev.preventDefault();
    AppActions.clickGrouping(ev.currentTarget.getAttribute('rel'));
  },

  getGroupingsSupportedByTopology: function(topology) {
    return _.filter(GROUPINGS, _.partial(this.isGroupingSupportedByTopology, topology));
  },

  renderGrouping: function(grouping, activeGroupingId) {
    let className = 'groupings-item';
    const isSupportedByTopology = this.isGroupingSupportedByTopology(this.props.currentTopology, grouping);

    if (grouping.id === activeGroupingId) {
      className += ' groupings-item-active';
    } else if (!isSupportedByTopology) {
      className += ' groupings-item-disabled';
    } else {
      className += ' groupings-item-default';
    }

    return (
      <div className={className} key={grouping.id} rel={grouping.id} onClick={isSupportedByTopology && this.onGroupingClick}>
        <span className={grouping.iconClass} />
      </div>
    );
  },

  render: function() {
    const activeGrouping = this.props.active;
    const isGroupingSupported = _.size(this.getGroupingsSupportedByTopology(this.props.currentTopology)) > 1;

    return (
      <div className="groupings">
        {isGroupingSupported && GROUPINGS.map(function(grouping) {
          return this.renderGrouping(grouping, activeGrouping);
        }, this)}
      </div>
    );
  },

  isGroupingSupportedByTopology: function(topology, grouping) {
    return !grouping.needsTopology || topology && topology[grouping.needsTopology];
  }

});

module.exports = Groupings;
