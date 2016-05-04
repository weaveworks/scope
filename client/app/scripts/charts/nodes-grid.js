/* eslint react/jsx-no-bind: "off", no-multi-comp: "off" */

import React from 'react';
import { Set as makeSet, List as makeList, Map as makeMap } from 'immutable';
import NodeDetailsTable from '../components/node-details/node-details-table';
import { enterNode, leaveNode } from '../actions/app-actions';


const IGNORED_COLUMNS = ['docker_container_ports'];


function getColumns(nodes) {
  const allColumns = nodes.toList().flatMap(n => {
    const metrics = (n.get('metrics') || makeList())
      .map(m => makeMap({ id: m.get('id'), label: m.get('label') }));
    const metadata = (n.get('metadata') || makeList())
      .map(m => makeMap({ id: m.get('id'), label: m.get('label') }));
    return metadata.concat(metrics);
  });
  return makeSet(allColumns).filter(n => !IGNORED_COLUMNS.includes(n.get('id'))).toJS();
}


export default class NodesGrid extends React.Component {

  onMouseOverRow(node) {
    enterNode(node.id);
  }

  onMouseOut() {
    leaveNode();
  }

  render() {
    const { margins, nodes, height } = this.props;
    const cmpStyle = {
      height,
      paddingTop: margins.top,
      paddingBottom: margins.bottom,
      paddingLeft: margins.left,
      paddingRight: margins.right,
    };

    const detailsData = {
      label: 'procs',
      id: '',
      nodes: nodes.toList().toJS(),
      columns: getColumns(nodes)
    };

    return (
      <div className="nodes-grid">
        <NodeDetailsTable
          style={cmpStyle}
          onMouseOut={this.onMouseOut}
          onMouseOverRow={this.onMouseOverRow}
          {...detailsData}
          highlightedNodeIds={this.props.highlightedNodeIds}
          limit={1000} />
      </div>
    );
  }
}
