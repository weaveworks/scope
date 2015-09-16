const React = require('react');

const Sidebar = React.createClass({

  render: function() {
    return (
      <div className="sidebar">
        {this.props.children}
      </div>
    );
  }

});

module.exports = Sidebar;
