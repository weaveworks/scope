/* eslint react/jsx-no-bind: "off", no-multi-comp: "off" */
import React from 'react';
import styled from 'styled-components';
import { connect } from 'react-redux';
import { List as makeList, Map as makeMap } from 'immutable';
import capitalize from 'lodash/capitalize';

import NodeDetailsTable from '../components/node-details/node-details-table';
import { clickNode, sortOrderChanged } from '../actions/app-actions';
import { shownNodesSelector } from '../selectors/node-filters';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import { findTopologyById } from '../utils/topology-utils';
import { TABLE_VIEW_MODE } from '../constants/naming';

import { windowHeightSelector } from '../selectors/canvas';
import { searchNodeMatchesSelector } from '../selectors/search';
import { getNodeColor } from '../utils/color-utils';


const IGNORED_COLUMNS = ['docker_container_ports', 'docker_container_id', 'docker_image_id',
  'docker_container_command', 'docker_container_networks'];


const Icon = styled.span`
  border-radius: ${props => props.theme.borderRadius.soft};
  background-color: ${props => props.color};
  margin-top: 3px;
  display: block;
  height: 12px;
  width: 12px;
`;

function topologyLabel(topologies, id) {
  const topology = findTopologyById(topologies, id);
  if (!topology) {
    return capitalize(id);
  }
  return topology.get('fullName');
}

function getColumns(nodes, topologies) {
  const metricColumns = nodes
    .toList()
    .flatMap((n) => {
      const metrics = (n.get('metrics') || makeList())
        .filter(m => !m.get('valueEmpty'))
        .map(m => makeMap({ dataType: 'number', id: m.get('id'), label: m.get('label') }));
      return metrics;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  const metadataColumns = nodes
    .toList()
    .flatMap((n) => {
      const metadata = (n.get('metadata') || makeList())
        .map(m => makeMap({ dataType: m.get('dataType'), id: m.get('id'), label: m.get('label') }));
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
        .map(m => makeMap({ id: m.get('topologyId'), label: topologyLabel(topologies, m.get('topologyId')) }));
      return metadata;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  return relativesColumns.concat(metadataColumns, metricColumns).toJS();
}


function renderIdCell({
  rank, label, labelMinor, pseudo
}) {
  const showSubLabel = Boolean(pseudo) && labelMinor;
  const title = showSubLabel ? `${label} (${labelMinor})` : label;

  return (
    <div title={title} className="nodes-grid-id-column">
      <div style={{ flex: 'none', width: 16 }}>
        <Icon color={getNodeColor(rank, label)} />
      </div>
      <div className="truncate">
        {label} {showSubLabel && <span className="nodes-grid-label-minor">{labelMinor}</span>}
      </div>
    </div>
  );
}
class NodesGrid extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onClickRow = this.onClickRow.bind(this);
    this.onSortChange = this.onSortChange.bind(this);
    this.saveTableRef = this.saveTableRef.bind(this);
  }

  onClickRow(ev, node) {
    trackAnalyticsEvent('scope.node.click', {
      layout: TABLE_VIEW_MODE,
      parentTopologyId: this.props.currentTopology.get('parentId'),
      topologyId: this.props.currentTopology.get('id'),
    });
    this.props.clickNode(node.id, node.label, ev.target.getBoundingClientRect());
  }

  onSortChange(sortedBy, sortedDesc) {
    this.props.sortOrderChanged(sortedBy, sortedDesc);
  }

  saveTableRef(ref) {
    this.tableRef = ref;
  }

  render() {
    const {
      nodes, gridSortedBy, gridSortedDesc, searchNodeMatches, searchQuery, windowHeight, topologies
    } = this.props;
    const height =
      this.tableRef ? windowHeight - this.tableRef.getBoundingClientRect().top - 30 : 0;
    const cmpStyle = {
      height,
      paddingLeft: 40,
      paddingRight: 40,
    };
    // TODO: What are 24 and 18? Use a comment or extract into constants.
    const tbodyHeight = height - 24 - 18;
    const className = 'tour-step-anchor scroll-body';
    const tbodyStyle = {
      height: `${tbodyHeight}px`,
    };

    const detailsData = {
      columns: getColumns(nodes, topologies),
      id: '',
      label: this.props.currentTopology && this.props.currentTopology.get('fullName'),
      nodes: nodes
        .toList()
        .filter(n => !(searchQuery && searchNodeMatches.get(n.get('id'), makeMap()).isEmpty()))
        .toJS()
    };

    return (
      <div className="nodes-grid" ref={this.saveTableRef}>
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
    currentTopology: state.get('currentTopology'),
    currentTopologyId: state.get('currentTopologyId'),
    gridSortedBy: state.get('gridSortedBy'),
    gridSortedDesc: state.get('gridSortedDesc'),
    nodes: shownNodesSelector(state),
    searchNodeMatches: searchNodeMatchesSelector(state),
    searchQuery: state.get('searchQuery'),
    selectedNodeId: state.get('selectedNodeId'),
    topologies: state.get('topologies'),
    windowHeight: windowHeightSelector(state),
  };
}


export default connect(
  mapStateToProps,
  { clickNode, sortOrderChanged }
)(NodesGrid);
