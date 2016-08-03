import React from 'react';

export default function Sidebar({children, classNames}) {
  const className = `sidebar ${classNames}`;
  return (
    <div className={className}>
      {children}
    </div>
  );
}
