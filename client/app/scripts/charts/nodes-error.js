import React from 'react';
import classnames from 'classnames';

const NodesError = ({ children, faIconClass, hidden, mainClassName = 'nodes-chart-error'}) => {
  const className = classnames(mainClassName, {
    hide: hidden
  });
  const iconClassName = `fa ${faIconClass}`;

  return (
    <div className={className}>
      <div className="nodes-chart-error-icon-container">
        <div className="nodes-chart-error-icon">
          <span className={iconClassName} />
        </div>
      </div>
      {children}
    </div>
  );
};

export default NodesError;
