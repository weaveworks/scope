import React from 'react';

export default function NodesError({children, faIconClass, hidden}) {
  let classNames = 'nodes-chart-error';
  if (hidden) {
    classNames += ' hide';
  }
  const iconClassName = `fa ${faIconClass}`;

  return (
    <div className={classNames}>
      <div className="nodes-chart-error-icon">
        <span className={iconClassName} />
      </div>
      {children}
    </div>
  );
}
