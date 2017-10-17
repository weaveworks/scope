import React from 'react';

import CloudLink from '../cloud-link';
import { formatMetric } from '../../utils/string-utils';
import { trackAnalyticsEvent } from '../../utils/tracking-utils';
import { dismissRowClickProps } from './node-details-table-row';

class NodeDetailsTableNodeMetricLink extends React.Component {
  constructor(props) {
    super(props);

    this.onClick = this.onClick.bind(this);
  }

  onClick() {
    trackAnalyticsEvent('scope.node.metric.click', { topologyId: this.props.topologyId });
  }

  render() {
    const {
      url, style, value, valueEmpty
    } = this.props;

    return (
      <td
        className="node-details-table-node-metric"
        style={style}
        {...dismissRowClickProps}
      >
        <CloudLink
          alwaysShow
          url={url}
          className={url && 'node-details-table-node-metric-link'}
          onClick={this.onClick}
        >
          {!valueEmpty && formatMetric(value, this.props)}
        </CloudLink>
      </td>
    );
  }
}

export default NodeDetailsTableNodeMetricLink;
