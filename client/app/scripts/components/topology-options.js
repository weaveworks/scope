const React = require('react');
const _ = require('lodash');
const mui = require('material-ui');
const DropDownMenu = mui.DropDownMenu;

const AppActions = require('../actions/app-actions');

const TopologyOptions = React.createClass({

  componentDidMount: function() {
    this.fixWidth();
  },

  onChange: function(ev, index, item) {
    ev.preventDefault();
    AppActions.changeTopologyOption(item.option, item.payload);
  },

  renderOption: function(items) {
    let selected = 0;
    let key;
    const activeOptions = this.props.activeOptions;
    const menuItems = items.map(function(item, index) {
      if (activeOptions[item.option] && activeOptions[item.option] === item.value) {
        selected = index;
      }
      key = item.option;
      return {
        option: item.option,
        payload: item.value,
        text: item.display
      };
    });

    return (
      <DropDownMenu menuItems={menuItems} onChange={this.onChange} key={key}
        selectedIndex={selected} />
    );
  },

  render: function() {
    const options = _.sortBy(
      _.map(this.props.options, function(items, optionId) {
        _.each(items, function(item) {
          item.option = optionId;
        });
        items.option = optionId;
        return items;
      }),
      'option'
    );

    return (
      <div className="topology-options" ref="container">
        {options.map(function(items) {
          return this.renderOption(items);
        }, this)}
      </div>
    );
  },

  componentDidUpdate: function() {
    this.fixWidth();
  },

  fixWidth: function() {
    const containerNode = this.refs.container.getDOMNode();
    _.each(containerNode.childNodes, function(child) {
      // set drop down width to length of current label
      const label = child.getElementsByClassName('mui-menu-label')[0];
      const width = label.getBoundingClientRect().width + 40;
      child.style.width = width + 'px';
    });
  }
});

module.exports = TopologyOptions;
