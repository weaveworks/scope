import React from 'react';

import { trackMixpanelEvent } from '../../utils/tracking-utils';
import { getMetricColor } from '../../utils/metric-utils';
import NodeDetailsHealthItem from './node-details-health-item';

export default class NodeDetailsHealthLinkItem extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      hovered: false
    };

    this.handleClick = this.handleClick.bind(this);
    this.buildHref = this.buildHref.bind(this);
    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onMouseOver() {
    this.setState({hovered: true});
  }

  onMouseOut() {
    this.setState({hovered: false});
  }

  handleClick(ev, href) {
    ev.preventDefault();
    if (!href) return;

    const { router, topologyId } = this.props;
    trackMixpanelEvent('scope.node.health.click', { topologyId });

    if (router && href[0] === '/') {
      router.push(href);
    } else {
      location.href = href;
    }
  }

  buildHref(url) {
    if (!url || !this.props.isCloud) return url;
    return url.replace(/:orgid/gi, encodeURIComponent(this.props.params.orgId));
  }

  render() {
    const { links, id, nodeColor, ...props } = this.props;
    const href = this.buildHref(links[id] && links[id].url);
    if (!href) return <NodeDetailsHealthItem {...props} />;

    const hasData = (props.samples && props.samples.length > 0) || props.value !== undefined;
    const labelColor = this.state.hovered && !hasData ? nodeColor : undefined;
    const sparkline = {};
    if (this.state.hovered) {
      sparkline.strokeColor = getMetricColor(id);
      sparkline.strokeWidth = '2px';
    }

    return (
      <a
        className="node-details-health-link-item"
        href={href}
        onClick={e => this.handleClick(e, href)}
        onMouseOver={this.onMouseOver}
        onMouseOut={this.onMouseOut}>
        <NodeDetailsHealthItem
          {...props}
          {...sparkline}
          labelColor={labelColor} />
      </a>
    );
  }
}
