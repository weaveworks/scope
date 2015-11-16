import React from 'react';

import { getNodeColor, getNodeColorDark } from '../utils/color-utils';
import Terminal from './terminal';

export default function EmeddedTerminal({pipe, nodeId, nodes}) {
  const node = nodes.get(nodeId && nodeId.split(';').pop());
  const titleBarColor = node && getNodeColorDark(node.get('rank'), node.get('label_major'));
  const statusBarColor = node && getNodeColor(node.get('rank'), node.get('label_major'));
  const title = node && node.get('label_major');

  // React unmount/remounts when key changes, this is important for cleaning up
  // the term.js and creating a new one for the new pipe.
  return (
    <div className="terminal-embedded">
      <Terminal key={pipe.id} pipe={pipe} titleBarColor={titleBarColor}
      statusBarColor={statusBarColor} title={title} />
    </div>
  );
}
