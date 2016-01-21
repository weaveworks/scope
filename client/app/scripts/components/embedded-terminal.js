import React from 'react';

import { getNodeColor, getNodeColorDark } from '../utils/color-utils';
import Terminal from './terminal';

export default function EmeddedTerminal({pipe, nodeId, details}) {
  const node = details.get(nodeId);
  const d = node && node.details;
  const titleBarColor = d && getNodeColorDark(d.rank, d.label_major);
  const statusBarColor = d && getNodeColor(d.rank, d.label_major);
  const title = d && d.label_major;

  // React unmount/remounts when key changes, this is important for cleaning up
  // the term.js and creating a new one for the new pipe.
  return (
    <div className="terminal-embedded">
      <Terminal key={pipe.id} pipe={pipe} titleBarColor={titleBarColor}
        statusBarColor={statusBarColor} title={title} />
    </div>
  );
}
