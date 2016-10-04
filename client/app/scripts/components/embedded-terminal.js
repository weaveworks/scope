import React from 'react';
import { connect } from 'react-redux';

import { brightenColor, getNodeColorDark } from '../utils/color-utils';
import Terminal from './terminal';

class EmeddedTerminal extends React.Component {
  render() {
    const { pipe, details } = this.props;
    const nodeId = pipe.get('nodeId');
    const node = details.get(nodeId);
    const d = node && node.details;
    const titleBarColor = d && getNodeColorDark(d.rank, d.label, d.pseudo);
    const statusBarColor = d && brightenColor(titleBarColor);
    const title = d && d.label;

    // React unmount/remounts when key changes, this is important for cleaning up
    // the term.js and creating a new one for the new pipe.
    return (
      <div className="terminal-embedded">
        <Terminal key={pipe.get('id')} pipe={pipe} titleBarColor={titleBarColor}
          statusBarColor={statusBarColor}
          title={title} />
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    details: state.get('nodeDetails'),
    pipe: state.get('controlPipes').last()
  };
}

export default connect(
  mapStateToProps
)(EmeddedTerminal);
