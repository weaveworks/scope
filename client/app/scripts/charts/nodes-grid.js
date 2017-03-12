/* eslint react/jsx-no-bind: "off", no-multi-comp: "off" */

import React from 'react';
import { connect } from 'react-redux';
import { List as makeList, Map as makeMap } from 'immutable';
import NodeDetailsTable from '../components/node-details/node-details-table';
import { clickNode, sortOrderChanged } from '../actions/app-actions';
import { shownNodesSelector } from '../selectors/node-filters';

import { searchNodeMatchesSelector } from '../selectors/search';
import { canvasMarginsSelector } from '../selectors/viewport';
import { getNodeColor } from '../utils/color-utils';


const IGNORED_COLUMNS = ['docker_container_ports', 'docker_container_id', 'docker_image_id',
  'docker_container_command', 'docker_container_networks'];


function getColumns(nodes) {
  const metricColumns = nodes
    .toList()
    .flatMap((n) => {
      const metrics = (n.get('metrics') || makeList())
        .map(m => makeMap({ id: m.get('id'), label: m.get('label'), dataType: 'number' }));
      return metrics;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  const metadataColumns = nodes
    .toList()
    .flatMap((n) => {
      const metadata = (n.get('metadata') || makeList())
        .map(m => makeMap({ id: m.get('id'), label: m.get('label'), dataType: m.get('dataType') }));
      return metadata;
    })
    .toSet()
    .filter(n => !IGNORED_COLUMNS.includes(n.get('id')))
    .toList()
    .sortBy(m => m.get('label'));

  const relativesColumns = nodes
    .toList()
    .flatMap((n) => {
      const metadata = (n.get('parents') || makeList())
        .map(m => makeMap({ id: m.get('topologyId'), label: m.get('topologyId') }));
      return metadata;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  return relativesColumns.concat(metadataColumns, metricColumns).toJS();
}


function renderIdCell(props) {
  const iconStyle = {
    width: 16,
    flex: 'none',
    color: getNodeColor(props.rank, props.label_major)
  };
  const showSubLabel = Boolean(props.pseudo);

  return (
    <div title={props.label} className="nodes-grid-id-column">
      <div style={iconStyle}><i className="fa fa-square" /></div>
      <div className="truncate">
        {props.label} {showSubLabel &&
          <span className="nodes-grid-label-minor">{props.labelMinor}</span>}
      </div>
    </div>
  );
}


class NodesGrid extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.onClickRow = this.onClickRow.bind(this);
    this.onSortChange = this.onSortChange.bind(this);
  }

  onClickRow(ev, node, el) {
    // TODO: do this better
    if (ev.target.className === 'node-details-table-node-link') {
      return;
    }
    this.props.clickNode(node.id, node.label, el.getBoundingClientRect());
  }

  onSortChange(sortedBy, sortedDesc) {
    this.props.sortOrderChanged(sortedBy, sortedDesc);
  }

  render() {
    const { nodes, height, gridSortedBy, gridSortedDesc, canvasMargins,
      searchNodeMatches, searchQuery } = this.props;
    const cmpStyle = {
      height,
      marginTop: canvasMargins.top,
      paddingLeft: canvasMargins.left,
      paddingRight: canvasMargins.right,
    };
    const tbodyHeight = height - 24 - 18;
    const className = 'scroll-body';
    const tbodyStyle = {
      height: `${tbodyHeight}px`,
    };

    const detailsData = {
      label: this.props.currentTopology && this.props.currentTopology.get('fullName'),
      id: '',
      nodes: nodes
        .toList()
        .filter(n => !(searchQuery && searchNodeMatches.get(n.get('id'), makeMap()).isEmpty()))
        .toJS(),
      columns: getColumns(nodes)
    };

    return (
      <div className="nodes-grid">
        {nodes.size > 0 && <NodeDetailsTable
          style={cmpStyle}
          className={className}
          renderIdCell={renderIdCell}
          tbodyStyle={tbodyStyle}
          topologyId={this.props.currentTopologyId}
          onSortChange={this.onSortChange}
          onClickRow={this.onClickRow}
          sortedBy={gridSortedBy}
          sortedDesc={gridSortedDesc}
          selectedNodeId={this.props.selectedNodeId}
          limit={1000}
          {...detailsData}
          />}
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    nodes: shownNodesSelector(state),
    canvasMargins: canvasMarginsSelector(state),
    gridSortedBy: state.get('gridSortedBy'),
    gridSortedDesc: state.get('gridSortedDesc'),
    currentTopology: state.get('currentTopology'),
    currentTopologyId: state.get('currentTopologyId'),
    searchNodeMatches: searchNodeMatchesSelector(state),
    searchQuery: state.get('searchQuery'),
    selectedNodeId: state.get('selectedNodeId'),
    // TODO: Change this.
    height: state.getIn(['viewport', 'height']) - 190,
  };
}


export default connect(
  mapStateToProps,
  { clickNode, sortOrderChanged }
)(NodesGrid);
