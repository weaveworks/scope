import React from 'react';

import { getNodeColor, getNodeColorDark } from '../utils/color-utils';
import Terminal from './terminal';

export default function EmeddedTerminal({pipe, nodeId, details, containerMargin}) {
  const d = details.get(nodeId);
  const titleBarColor = d && getNodeColorDark(d.rank, d.label);
  const statusBarColor = d && getNodeColor(d.rank, d.label);
  const title = d && d.label;

  // React unmount/remounts when key changes, this is important for cleaning up
  // the term.js and creating a new one for the new pipe.
  return (
    <Terminal key={pipe.id} pipe={pipe} titleBarColor={titleBarColor}
      statusBarColor={statusBarColor} containerMargin={containerMargin}
      title={title} />
  );
}
