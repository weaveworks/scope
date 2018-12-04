import React from 'react';
import { sample } from 'lodash';

import { findTopologyById } from '../utils/topology-utils';
import NodesError from '../charts/nodes-error';


const LOADING_TEMPLATES = [
  'Loading THINGS',
  'Verifying THINGS',
  'Fetching THINGS',
  'Processing THINGS',
  'Reticulating THINGS',
  'Locating THINGS',
  'Optimizing THINGS',
  'Transporting THINGS',
];


export function getNodeType(topology, topologies) {
  if (!topology || topologies.size === 0) {
    return '';
  }
  let name = topology.get('name');
  if (topology.get('parentId')) {
    const parentTopology = findTopologyById(topologies, topology.get('parentId'));
    name = parentTopology.get('name');
  }
  return name.toLowerCase();
}


function renderTemplate(nodeType, template) {
  return template.replace('THINGS', nodeType);
}


export class Loading extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      template: sample(LOADING_TEMPLATES)
    };
  }

  render() {
    const { itemType, show } = this.props;
    const message = renderTemplate(itemType, this.state.template);
    return (
      <NodesError mainClassName="nodes-chart-loading" faIconClass="far fa-circle" hidden={!show}>
        <div className="heading">{message}</div>
      </NodesError>
    );
  }
}
