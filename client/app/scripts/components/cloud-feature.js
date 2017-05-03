import React from 'react';
import { connect } from 'react-redux';

class CloudFeature extends React.Component {
  getChildContext() {
    return {
      store: this.context.serviceStore || this.context.store
    };
  }

  render() {
    if (process.env.WEAVE_CLOUD) {
      return React.cloneElement(React.Children.only(this.props.children), {
        params: this.context.router.params,
        router: this.context.router
      });
    }

    return null;
  }
}

CloudFeature.contextTypes = {
  store: React.PropTypes.object.isRequired,
  router: React.PropTypes.object,
  serviceStore: React.PropTypes.object
};

CloudFeature.childContextTypes = {
  store: React.PropTypes.object,
  router: React.PropTypes.object
};

export default connect()(CloudFeature);
