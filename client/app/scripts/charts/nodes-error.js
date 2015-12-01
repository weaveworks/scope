import React from 'react';

export default class NodesError extends React.Component {
  render() {
    let classNames = 'nodes-chart-error';
    if (this.props.hidden) {
      classNames += ' hide';
    }
    const iconClassName = 'fa ' + this.props.faIconClass;

    return (
      <div className={classNames}>
        <div className="nodes-chart-error-icon">
          <span className={iconClassName} />
        </div>
        {this.props.children}
      </div>
    );
  }
}
