import React from 'react';
import moment from 'moment';
import { connect } from 'react-redux';

import NodeDetailsHealthItem from './node-details-health-item';
import CloudLink from '../cloud-link';
import { getMetricColor } from '../../utils/metric-utils';
import { darkenColor } from '../../utils/color-utils';
import { trackAnalyticsEvent } from '../../utils/tracking-utils';

/**
 * @param {string} url
 * @param {string} time
 * @returns {string}
 */
export function appendTime(url, time) {
  if (!url || !time) return url;

  // rudimentary check whether we have a cloud link
  const cloudLinkPathEnd = 'notebook/new/';
  const pos = url.indexOf(cloudLinkPathEnd);
  const timeUnix = moment(time).unix();
  if (pos !== -1) {
    let payload;
    const json = decodeURIComponent(url.substr(pos + cloudLinkPathEnd.length));
    try {
      payload = JSON.parse(json);
      payload.time = { queryEnd: timeUnix };
    } catch (e) {
      return url;
    }

    return `${url.substr(0, pos + cloudLinkPathEnd.length)}${encodeURIComponent(JSON.stringify(payload) || '')}`;
  }

  if (url.indexOf('?') !== -1) {
    return `${url}&time=${timeUnix}`;
  }
  return `${url}?time=${timeUnix}`;
}

class NodeDetailsHealthLinkItem extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hovered: false
    };

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseOut = this.onMouseOut.bind(this);
    this.onClick = this.onClick.bind(this);
  }

  onMouseOver() {
    this.setState({hovered: true});
  }

  onMouseOut() {
    this.setState({hovered: false});
  }

  onClick() {
    trackAnalyticsEvent('scope.node.metric.click', { topologyId: this.props.topologyId });
  }

  render() {
    const {
      id, url, monitor, pausedAt, ...props
    } = this.props;
    const metricColor = getMetricColor(id);
    const labelColor = this.state.hovered && !props.valueEmpty && darkenColor(metricColor);

    const timedUrl = monitor === true ? appendTime(url, pausedAt) : '';

    return (
      <CloudLink
        alwaysShow
        className="node-details-health-link-item"
        onMouseOver={this.onMouseOver}
        onMouseOut={this.onMouseOut}
        onClick={this.onClick}
        url={timedUrl}
      >
        <NodeDetailsHealthItem
          {...props}
          hovered={this.state.hovered}
          labelColor={labelColor}
          metricColor={metricColor}
        />
      </CloudLink>
    );
  }
}

function mapStateToProps(state) {
  return {
    monitor: state.get('monitor'),
    pausedAt: state.get('pausedAt'),
  };
}

export default connect(mapStateToProps)(NodeDetailsHealthLinkItem);
