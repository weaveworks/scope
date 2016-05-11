import React from 'react';

export default function Warning({text}) {
  return (
    <span className="warning warning-icon fa fa-warning" title={text} />
  );
}
