import React from 'react';

import { formatMetric } from '../../utils/string-utils';

class NodeDetailsTableNodeMetric extends React.Component {
  render() {
    return (
      <td className="node-details-table-node-metric">
        {formatMetric(this.props.value, this.props)}
      </td>
    );
  }
}

export default NodeDetailsTableNodeMetric;
