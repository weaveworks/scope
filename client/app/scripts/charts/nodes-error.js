const React = require('react');

const NodesError = React.createClass({

  render: function() {
    let classNames = 'nodes-chart-error';
    if (this.props.hidden) {
      classNames += ' hide';
    }
    let iconClassName = 'fa ' + this.props.faIconClass;

    return (
      <div className={classNames}>
        <div className="nodes-chart-error-icon">
          <span className={iconClassName} />
        </div>
        {this.props.children}
      </div>
    );
  }

});

module.exports = NodesError;
