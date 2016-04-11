import _ from 'lodash';
import React from 'react';

import NodeDetailsControls from './node-details/node-details-controls';
import NodeDetailsHealth from './node-details/node-details-health';
import NodeDetailsInfo from './node-details/node-details-info';
import NodeDetailsLabels from './node-details/node-details-labels';
import NodeDetailsRelatives from './node-details/node-details-relatives';
import NodeDetailsTable from './node-details/node-details-table';
import { clickCloseDetails, clickShowTopologyForNode } from '../actions/app-actions';
import { brightenColor, getNeutralColor, getNodeColorDark } from '../utils/color-utils';
import { resetDocumentTitle, setDocumentTitle } from '../utils/title-utils';

export default class NodeDetails extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClickClose = this.handleClickClose.bind(this);
    this.handleShowTopologyForNode = this.handleShowTopologyForNode.bind(this);
  }

  handleClickClose(ev) {
    ev.preventDefault();
    clickCloseDetails(this.props.nodeId);
  }

  handleShowTopologyForNode(ev) {
    ev.preventDefault();
    clickShowTopologyForNode(this.props.topologyId, this.props.nodeId);
  }

  componentDidMount() {
    this.updateTitle();
  }

  componentWillUnmount() {
    resetDocumentTitle();
  }

  renderTools() {
    const showSwitchTopology = this.props.index > 0;
    const topologyTitle = `View ${this.props.label} in ${this.props.topologyId}`;

    return (
      <div className="node-details-tools-wrapper">
        <div className="node-details-tools">
          {showSwitchTopology && <span title={topologyTitle}
            className="fa fa-exchange" onClick={this.handleShowTopologyForNode} />}
          <span title="Close details" className="fa fa-close" onClick={this.handleClickClose} />
        </div>
      </div>
    );
  }

  renderLoading() {
    const node = this.props.nodes.get(this.props.nodeId);
    const label = node ? node.get('label') : this.props.label;
    const nodeColor = (node ?
                       getNodeColorDark(node.get('rank'), label, node.get('pseudo')) :
                       getNeutralColor());
    const tools = this.renderTools();
    const styles = {
      header: {
        backgroundColor: nodeColor
      }
    };

    return (
      <div className="node-details">
        {tools}
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate">
              {label}
            </h2>
            <div className="node-details-relatives truncate">
              Loading...
            </div>
          </div>
        </div>
        <div className="node-details-content">
          <div className="node-details-content-loading">
            <span className="fa fa-circle-o-notch fa-spin" />
          </div>
        </div>
      </div>
    );
  }

  renderNotAvailable() {
    const tools = this.renderTools();
    return (
      <div className="node-details">
        {tools}
        <div className="node-details-header node-details-header-notavailable">
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label">
              {this.props.label}
            </h2>
            <div className="node-details-relatives truncate">
              n/a
            </div>
          </div>
        </div>
        <div className="node-details-content">
          <p className="node-details-content-info">
            <strong>{this.props.label}</strong> is not visible to Scope when it
             is not communicating.
            Details will become available here when it communicates again.
          </p>
        </div>
      </div>
    );
  }

  renderTable(table) {
    const key = _.snakeCase(table.title);
    return (<NodeDetailsTable title={table.title} key={key} rows={table.rows}
      isNumeric={table.numeric} />);
  }

  render() {
    if (this.props.notFound) {
      return this.renderNotAvailable();
    }

    if (this.props.details) {
      return this.renderDetails();
    }

    return this.renderLoading();
  }

  renderDetails() {
    const details = this.props.details;
    const showControls = details.controls && details.controls.length > 0;
    const nodeColor = getNodeColorDark(details.rank, details.label, details.pseudo);
    const {error, pending} = this.props.nodeControlStatus
      ? this.props.nodeControlStatus.toJS() : {};
    const tools = this.renderTools();
    const styles = {
      controls: {
        backgroundColor: brightenColor(nodeColor)
      },
      header: {
        backgroundColor: nodeColor
      }
    };

    return (
      <div className="node-details">
        {tools}
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate" title={details.label}>
              {details.label}
            </h2>
            <div className="node-details-header-relatives">
              {details.parents && <NodeDetailsRelatives relatives={details.parents} />}
            </div>
          </div>
        </div>

        {showControls && <div className="node-details-controls-wrapper" style={styles.controls}>
          <NodeDetailsControls nodeId={this.props.nodeId}
            controls={details.controls}
            pending={pending}
            error={error} />
        </div>}

        <div className="node-details-content">
          {details.metrics && <div className="node-details-content-section">
            <div className="node-details-content-section-header">Status</div>
            <NodeDetailsHealth metrics={details.metrics} />
          </div>}
          {details.metadata && <div className="node-details-content-section">
            <div className="node-details-content-section-header">Info</div>
            <NodeDetailsInfo rows={details.metadata} />
          </div>}

          {details.connections && details.connections.map(connections => <div
            className="node-details-content-section" key={connections.id}>
              <NodeDetailsTable {...connections} nodes={connections.connections}
                nodeIdKey="nodeId" />
            </div>
          )}

          {details.children && details.children.map(children => <div
            className="node-details-content-section" key={children.topologyId}>
              <NodeDetailsTable {...children} />
            </div>
          )}

          {details.docker_labels && details.docker_labels.length > 0
            && <div className="node-details-content-section">
            <div className="node-details-content-section-header">Docker Labels</div>
            <NodeDetailsLabels rows={details.docker_labels} />
          </div>}
        </div>
      </div>
    );
  }

  componentDidUpdate() {
    this.updateTitle();
  }

  updateTitle() {
    setDocumentTitle(this.props.details && this.props.details.label);
  }
}
