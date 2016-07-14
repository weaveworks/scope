/* eslint react/jsx-no-bind: "off", no-multi-comp: "off" */

import React from 'react';
import { connect } from 'react-redux';
import { List as makeList, Map as makeMap } from 'immutable';
import NodeDetailsTable from '../components/node-details/node-details-table';
import { clickNode, sortOrderChanged, clickPauseUpdate,
  clickResumeUpdate } from '../actions/app-actions';

import { getNodeColor } from '../utils/color-utils';


const IGNORED_COLUMNS = ['docker_container_ports', 'docker_container_id', 'docker_image_id',
  'docker_container_command', 'docker_container_networks'];


function getColumns(nodes) {
  const metricColumns = nodes
    .toList()
    .flatMap(n => {
      const metrics = (n.get('metrics') || makeList())
        .map(m => makeMap({ id: m.get('id'), label: m.get('label') }));
      return metrics;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  const metadataColumns = nodes
    .toList()
    .flatMap(n => {
      const metadata = (n.get('metadata') || makeList())
        .map(m => makeMap({ id: m.get('id'), label: m.get('label') }));
      return metadata;
    })
    .toSet()
    .filter(n => !IGNORED_COLUMNS.includes(n.get('id')))
    .toList()
    .sortBy(m => m.get('label'));

  const relativesColumns = nodes
    .toList()
    .flatMap(n => {
      const metadata = (n.get('parents') || makeList())
        .map(m => makeMap({ id: m.get('topologyId'), label: m.get('topologyId') }));
      return metadata;
    })
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));

  return relativesColumns.concat(metadataColumns.concat(metricColumns)).toJS();
}


function renderIdCell(props, onClick) {
  const style = {
    width: 16,
    flex: 'none',
    color: getNodeColor(props.rank, props.label_major)
  };

  return (
    <div className="nodes-grid-id-column" onClick={onClick}>
      <div className="content">
        <div style={style}><i className="fa fa-square" /></div>
        <div className="truncate">
          {props.label} {props.pseudo &&
            <span className="nodes-grid-label-minor">{props.label_minor}</span>}
        </div>
      </div>
    </div>
  );
}


class NodesGrid extends React.Component {

  constructor(props, context) {
    super(props, context);

    this.renderIdCell = this.renderIdCell.bind(this);
    this.clickRow = this.clickRow.bind(this);

    this.onSortChange = this.onSortChange.bind(this);
    this.onMouseEnterRow = this.onMouseEnterRow.bind(this);
    this.onMouseLeaveRow = this.onMouseLeaveRow.bind(this);
  }

  clickRow(ev, nodeId, nodeLabel) {
    if (ev.target.className === 'node-details-relatives-link') {
      return;
    }
    this.props.clickNode(nodeId, nodeLabel);
  }

  renderIdCell(props) {
    return renderIdCell(props, (ev) => this.clickRow(ev, props.id, props.label));
  }

  onMouseEnterRow() {
    this.props.clickPauseUpdate();
  }

  onMouseLeaveRow() {
    this.props.clickResumeUpdate();
  }

  onSortChange(sortBy, sortedDesc) {
    this.props.sortOrderChanged(sortBy, sortedDesc);
  }

  render() {
    const { margins, nodes, height, gridSortBy, gridSortedDesc,
      searchNodeMatches = makeMap(), searchQuery } = this.props;
    const cmpStyle = {
      height,
      marginTop: margins.top,
      paddingLeft: margins.left,
      paddingRight: margins.right,
    };
    const tbodyHeight = height - 24 - 18;
    const className = 'scroll-body';
    const tbodyStyle = {
      height: `${tbodyHeight}px`,
    };

    const detailsData = {
      label: this.props.topology && this.props.topology.get('fullName'),
      id: '',
      nodes: nodes
        .toList()
        .filter(n => !searchQuery || searchNodeMatches.has(n.get('id')))
        .toJS(),
      columns: getColumns(nodes)
    };

    return (
      <div className="nodes-grid">
        <NodeDetailsTable
          style={cmpStyle}
          className={className}
          renderIdCell={this.renderIdCell}
          tbodyStyle={tbodyStyle}
          topologyId={this.props.topologyId}
          onSortChange={this.onSortChange}
          {...detailsData}
          sortBy={gridSortBy}
          sortedDesc={gridSortedDesc}
          selectedNodeId={this.props.selectedNodeId}
          limit={1000} />
      </div>
    );
  }
}


export default connect(
  () => ({}),
  { clickNode, sortOrderChanged, clickPauseUpdate, clickResumeUpdate }
)(NodesGrid);
