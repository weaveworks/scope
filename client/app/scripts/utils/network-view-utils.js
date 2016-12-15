import { fromJS, List as makeList } from 'immutable';

export function getNetworkNodes(nodes) {
  const networks = {};
  nodes.forEach(node => (node.get('networks') || makeList()).forEach((n) => {
    const networkId = n.get('id');
    networks[networkId] = (networks[networkId] || []).concat([node.get('id')]);
  }));
  return fromJS(networks);
}


export function getAvailableNetworks(nodes) {
  return nodes
    .valueSeq()
    .flatMap(node => node.get('networks') || makeList())
    .toSet()
    .toList()
    .sortBy(m => m.get('label'));
}
