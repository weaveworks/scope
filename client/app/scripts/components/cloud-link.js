import React from 'react';
import { connect } from 'react-redux';
import filterInvalidDOMProps from 'filter-invalid-dom-props';

import CloudFeature from './cloud-feature';

/**
 * CloudLink provides an anchor that allows to set a target
 * that is comprised of Weave Cloud related pieces.
 *
 * We support here relative links with a leading `/` that rewrite
 * the browser url as well as cloud-related placeholders (:instanceId).
 *
 * If no `url` is given, only the children is rendered (no anchor).
 *
 * If you want to render the content even if not on the cloud, set
 * the `alwaysShow` property. A location redirect will be made for
 * clicks instead.
 */
const CloudLink = ({ alwaysShow, ...props }) => (
  <CloudFeature alwaysShow={alwaysShow}>
    <LinkWrapper {...props} />
  </CloudFeature>
);

class LinkWrapper extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleClick = this.handleClick.bind(this);
    this.buildHref = this.buildHref.bind(this);
  }

  handleClick(ev, href) {
    ev.preventDefault();
    if (!href) return;

    const { router, onClick } = this.props;

    if (onClick) {
      onClick();
    }

    if (router && href[0] === '/') {
      router.push(href);
    } else {
      window.location.href = href;
    }
  }

  buildHref(url) {
    const { params } = this.props;
    if (!url || !params || !params.instanceId) return url;
    return url.replace(/:instanceid/gi, encodeURIComponent(params.instanceId));
  }

  render() {
    const { url, children, ...props } = this.props;
    if (!url) {
      return React.isValidElement(children) ? children : (<span>{children}</span>);
    }

    const href = this.buildHref(url);
    return (
      <a {...filterInvalidDOMProps(props)} href={href} onClick={e => this.handleClick(e, href)}>
        {children}
      </a>
    );
  }
}

export default connect()(CloudLink);
