/* eslint react/jsx-no-bind: "off", no-multi-comp: "off" */

import React from 'react';
import { Set as makeSet, List as makeList, Map as makeMap } from 'immutable';
import NodesChart from './nodes-chart';
import NodeDetailsTable from '../components/node-details/node-details-table';
import { enterNode, leaveNode } from '../actions/app-actions';


function MiniChart(props) {
  const {width, height} = props;
  return (
    <div style={{height, width}} className="nodes-grid-graph">
      <NodesChart {...props} />
    </div>
  );
}


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
    const {margins, nodes, height, nodeSize} = this.props;
    const rowStyle = { height: nodeSize };
    const tableHeight = nodes.size * rowStyle.height;
    const graphProps = Object.assign({}, this.props, {
      height: tableHeight,
      width: 400,
      noZoom: true,
      nodeSize: nodeSize - 4,
      margins: {top: 0, left: 0, right: 0, bottom: 0},
      nodes: nodes.map(node => node.remove('label').remove('label_minor'))
    });
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
          limit={1000}>
          <MiniChart {...graphProps} />
        </NodeDetailsTable>
      </div>
    );
  }
}
