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
        router: this.context.router,
        isCloud: true
      });
    }

    // also show if not in weave cloud?
    if (this.props.alwaysShow) {
      return React.cloneElement(React.Children.only(this.props.children), {isCloud: false});
    }

    return null;
  }
}

CloudFeature.contextTypes = {
  store: PropTypes.object.isRequired,
  router: PropTypes.object,
  serviceStore: PropTypes.object
};

CloudFeature.childContextTypes = {
  store: PropTypes.object,
  router: PropTypes.object
};

export default connect()(CloudFeature);
