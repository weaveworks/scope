import React from 'react';
import { connect } from 'react-redux';
import find from 'lodash/find';
import map from 'lodash/map';
import isNull from 'lodash/isNull';
import { CircularProgress } from 'weaveworks-ui-components';

import { getImagesForService } from '../../actions/app-actions';

const topologyWhitelist = ['kube-controllers'];

function newImagesAvailable(images, currentId) {
  const current = find(images, i => i.ID === currentId);

  if (current) {
    const timestamp = new Date(current.CreatedAt);
    return !isNull(find(images, i => timestamp < new Date(i.CreatedAt)));
  }

  return false;
}


class NodeDetailsImageStatus extends React.PureComponent {
  constructor(props, context) {
    super(props, context);

    this.getImagesUrl = this.getImagesUrl.bind(this);
  }

  componentDidMount() {
    if (this.shouldRender() && this.props.serviceId) {
      this.props.getImagesForService(this.props.params.orgId, this.props.serviceId);
    }
  }

  getImagesUrl() {
    const { serviceId, params } = this.props;
    return `/flux/${params.orgId}/services/${encodeURIComponent(serviceId)}`;
  }

  shouldRender() {
    const { pseudo, currentTopologyId } = this.props;
    return !pseudo && currentTopologyId && topologyWhitelist.includes(currentTopologyId);
  }

  renderImages() {
    const { errors, containers, isFetching } = this.props;
    const error = !isFetching && errors;

    if (isFetching) {
      return (
        <div className="progress-wrapper"><CircularProgress /></div>
      );
    }

    if (error) {
      return (
        <p>Error: {JSON.stringify(map(errors, 'message'))}</p>
      );
    }

    if (!containers) {
      return 'No service images found';
    }

    return (
      <div className="images">
        {containers.map((container) => {
          const statusText = newImagesAvailable(container.Available, container.Current.ID)
            ? <span className="new-image">New image(s) available</span>
            : 'Image up to date';

          return (
            <div key={container.Name} className="wrapper">
              <div className="node-details-table-node-label">{container.Name}</div>
              <div className="node-details-table-node-value">{statusText}</div>
            </div>
          );
        })}
      </div>
    );
  }

  render() {
    const { containers } = this.props;

    if (!this.shouldRender()) {
      return null;
    }

    return (
      <div className="node-details-content-section image-status">
        <div className="node-details-content-section-header">
          Container Image Status
          {containers &&
            <div>
              <a
                href={this.getImagesUrl()}
                className="node-details-table-node-link">
                  View in Deploy
              </a>
            </div>
          }

        </div>
        {this.renderImages()}
      </div>
    );
  }
}

function mapStateToProps({ scope }, { metadata, name }) {
  const namespace = find(metadata, d => d.id === 'kubernetes_namespace');
  const nodeType = find(metadata, d => d.id === 'kubernetes_node_type');
  const serviceId = (namespace && nodeType) ? `${namespace.value}:${nodeType.value.toLowerCase()}/${name}` : null;
  const { containers, isFetching, errors } = scope.getIn(['serviceImages', serviceId]) || {};

  return {
    isFetching,
    errors,
    currentTopologyId: scope.get('currentTopologyId'),
    containers,
    serviceId
  };
}

export default connect(mapStateToProps, { getImagesForService })(NodeDetailsImageStatus);
