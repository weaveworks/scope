import React from 'react';

import NodeDetailsHealthItem from './node-details-health-item';

export default class NodeDetailsHealthLinkItem extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.handleClick = this.handleClick.bind(this);
    this.buildHref = this.buildHref.bind(this);
  }

  handleClick(ev, href) {
    ev.preventDefault();
    if (!href) return;

    const {router} = this.props;
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
    const {links, withoutGraph, id, ...props} = this.props;
    const href = this.buildHref(links[id] && links[id].url);

    if (!href) return <NodeDetailsHealthItem {...props} />;

    return (
      <a
        className="node-details-health-link-item"
        href={href}
        onClick={e => this.handleClick(e, href)}>
        {!withoutGraph && <NodeDetailsHealthItem {...props} icon="fa-expand" />}
        {withoutGraph && <NodeDetailsHealthItem label={props.label} icon="fa-expand" />}
      </a>
    );
  }
}
