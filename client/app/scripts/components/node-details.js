import debug from 'debug';
import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { clickCloseDetails, clickShowTopologyForNode } from '../actions/app-actions';
import { brightenColor, getNeutralColor, getNodeColorDark } from '../utils/color-utils';
import { isGenericTable, isPropertyList } from '../utils/node-details-utils';
import { resetDocumentTitle, setDocumentTitle } from '../utils/title-utils';
import { timestampsEqual } from '../utils/time-utils';

import Overlay from './overlay';
import MatchedText from './matched-text';
import NodeDetailsControls from './node-details/node-details-controls';
import NodeDetailsGenericTable from './node-details/node-details-generic-table';
import NodeDetailsPropertyList from './node-details/node-details-property-list';
import NodeDetailsHealth from './node-details/node-details-health';
import NodeDetailsInfo from './node-details/node-details-info';
import NodeDetailsRelatives from './node-details/node-details-relatives';
import NodeDetailsTable from './node-details/node-details-table';
import Warning from './warning';
import CloudFeature from './cloud-feature';
import NodeDetailsImageStatus from './node-details/node-details-image-status';


const log = debug('scope:node-details');

function getTruncationText(count) {
  return 'This section was too long to be handled efficiently and has been truncated'
  + ` (${count} extra entries not included). We are working to remove this limitation.`;
}

class NodeDetails extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClickClose = this.handleClickClose.bind(this);
    this.handleShowTopologyForNode = this.handleShowTopologyForNode.bind(this);
  }

  handleClickClose(ev) {
    ev.preventDefault();
    this.props.clickCloseDetails(this.props.nodeId);
  }

  handleShowTopologyForNode(ev) {
    ev.preventDefault();
    this.props.clickShowTopologyForNode(this.props.topologyId, this.props.nodeId);
  }

  componentDidMount() {
    this.updateTitle();
  }

  componentWillUnmount() {
    resetDocumentTitle();
  }

  static collectMetrics(details) {
    const metrics = details.metrics || [];

    // collect by metric id (id => link)
    const metricLinks = (details.metric_links || [])
      .reduce((agg, link) => Object.assign(agg, {[link.id]: link}), {});

    const availableMetrics = metrics.reduce(
      (agg, m) => Object.assign(agg, {[m.id]: true}),
      {}
    );

    // append links with no metrics as fake metrics
    (details.metric_links || []).forEach((link) => {
      if (availableMetrics[link.id] === undefined) {
        metrics.push({id: link.id, label: link.label});
      }
    });

    return { metrics, metricLinks };
  }

  renderTools() {
    const showSwitchTopology = this.props.nodeId !== this.props.selectedNodeId;
    const topologyTitle = `View ${this.props.label} in ${this.props.topologyId}`;

    return (
      <div className="node-details-tools-wrapper">
        <div className="node-details-tools">
          {showSwitchTopology && <span
            title={topologyTitle}
            className="fa fa-long-arrow-left"
            onClick={this.handleShowTopologyForNode}>
            <span>Show in <span>{this.props.topologyId.replace(/-/g, ' ')}</span></span>
          </span>}
          <span
            title="Close details"
            className="fa fa-close close-details"
            onClick={this.handleClickClose}
          />
        </div>
      </div>
    );
  }

  renderLoading() {
    const node = this.props.nodes.get(this.props.nodeId);
    const label = node ? node.get('label') : this.props.label;
    // NOTE: If we start the fa-spin animation before the node details panel has been
    // mounted, the spinner is displayed blurred the whole time in Chrome (possibly
    // caused by a bug having to do with animating the details panel).
    const spinnerClassName = classNames('fa fa-circle-o-notch', { 'fa-spin': this.props.mounted });
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
            <span className={spinnerClassName} />
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
        <Overlay faded={this.props.transitioning} />
      </div>
    );
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
    const { details, nodeControlStatus, nodeMatches = makeMap(), topologyId } = this.props;
    const showControls = details.controls && details.controls.length > 0;
    const nodeColor = getNodeColorDark(details.rank, details.label, details.pseudo);
    const {error, pending} = nodeControlStatus ? nodeControlStatus.toJS() : {};
    const tools = this.renderTools();
    const styles = {
      controls: {
        backgroundColor: brightenColor(nodeColor)
      },
      header: {
        backgroundColor: nodeColor
      }
    };

    const { metrics, metricLinks } = NodeDetails.collectMetrics(details);

    return (
      <div className="node-details">
        {tools}
        <div className="node-details-header" style={styles.header}>
          <div className="node-details-header-wrapper">
            <h2 className="node-details-header-label truncate" title={details.label}>
              <MatchedText text={details.label} match={nodeMatches.get('label')} />
            </h2>
            <div className="node-details-header-relatives">
              {details.parents && <NodeDetailsRelatives
                matches={nodeMatches.get('parents')}
                relatives={details.parents} />}
            </div>
          </div>
        </div>

        {showControls && <div className="node-details-controls-wrapper" style={styles.controls}>
          <NodeDetailsControls
            nodeId={this.props.nodeId}
            controls={details.controls}
            pending={pending}
            error={error} />
        </div>}

        <div className="node-details-content">
          {metrics.length > 0 && <div className="node-details-content-section">
            <div className="node-details-content-section-header">Status</div>
            <NodeDetailsHealth
              metrics={metrics}
              metricLinks={metricLinks}
              topologyId={topologyId}
              nodeColor={nodeColor}
              />
          </div>}
          {details.metadata && <div className="node-details-content-section">
            <div className="node-details-content-section-header">Info</div>
            <NodeDetailsInfo rows={details.metadata} matches={nodeMatches.get('metadata')} />
          </div>}

          {details.connections && details.connections.filter(cs => cs.connections.length > 0)
            .map(connections => (<div className="node-details-content-section" key={connections.id}>
              <NodeDetailsTable
                {...connections}
                nodes={connections.connections}
                nodeIdKey="nodeId"
              />
            </div>
          ))}

          {details.children && details.children.map(children => (
            <div className="node-details-content-section" key={children.topologyId}>
              <NodeDetailsTable {...children} />
            </div>
          ))}

          {details.tables && details.tables.length > 0 && details.tables.map((table) => {
            if (table.rows.length > 0) {
              return (
                <div className="node-details-content-section" key={table.id}>
                  <div className="node-details-content-section-header">
                    {table.label && table.label.length > 0 && table.label}
                    {table.truncationCount > 0 && <span
                      className="node-details-content-section-header-warning">
                      <Warning text={getTruncationText(table.truncationCount)} />
                    </span>}
                  </div>
                  {this.renderTable(table)}
                </div>
              );
            }
            return null;
          })}

          <CloudFeature>
            <NodeDetailsImageStatus
              name={details.label}
              metadata={details.metadata}
              pseudo={details.pseudo}
            />
          </CloudFeature>
        </div>

        <Overlay faded={this.props.transitioning} />
      </div>
    );
  }

  renderTable(table) {
    const { nodeMatches = makeMap() } = this.props;

    if (isGenericTable(table)) {
      return (
        <NodeDetailsGenericTable
          rows={table.rows} columns={table.columns}
          matches={nodeMatches.get('tables')}
        />
      );
    } else if (isPropertyList(table)) {
      return (
        <NodeDetailsPropertyList
          rows={table.rows} controls={table.controls}
          matches={nodeMatches.get('property-lists')}
        />
      );
    }

    log(`Undefined type '${table.type}' for table ${table.id}`);
    return null;
  }

  componentDidUpdate() {
    this.updateTitle();
  }

  updateTitle() {
    setDocumentTitle(this.props.details && this.props.details.label);
  }
}

function mapStateToProps(state, ownProps) {
  const currentTopologyId = state.get('currentTopologyId');
  return {
    transitioning: !timestampsEqual(state.get('pausedAt'), ownProps.timestamp),
    nodeMatches: state.getIn(['searchNodeMatches', currentTopologyId, ownProps.id]),
    nodes: state.get('nodes'),
    selectedNodeId: state.get('selectedNodeId'),
  };
}

export default connect(
  mapStateToProps,
  { clickCloseDetails, clickShowTopologyForNode }
)(NodeDetails);
