import React from 'react';

export default function Sidebar({children, classNames}) {
  const className = `tour-step-anchor sidebar ${classNames}`;
  return (
    <div className={className}>
      {children}
    </div>
  );
}
