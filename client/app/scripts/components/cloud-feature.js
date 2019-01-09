import React from 'react';
import PropTypes from 'prop-types';
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

    // also show if not in weave cloud?
    if (this.props.alwaysShow) {
      return React.cloneElement(React.Children.only(this.props.children));
    }

    return null;
  }
}

CloudFeature.contextTypes = {
  router: PropTypes.object,
  serviceStore: PropTypes.object,
  store: PropTypes.object.isRequired
};

CloudFeature.childContextTypes = {
  router: PropTypes.object,
  store: PropTypes.object
};

export default connect()(CloudFeature);
