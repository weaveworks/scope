import { fromJS } from 'immutable';

import { nodeNetworksSelector } from '../selectors/node-networks';

export function getNetworkNodes(state) {
  const networksMap = {};
  nodeNetworksSelector(state).forEach((networks, nodeId) => {
    networks.forEach((network) => {
      const networkId = network.get('id');
      networksMap[networkId] = networksMap[networkId] || [];
      networksMap[networkId].push(nodeId);
    });
  });
  return fromJS(networksMap);
}
