
// TODO: Consider fetching these from the backend.
export const topologiesWithCapacity = ['hosts'];
export const resourceViewLayers = {
  hosts: ['hosts', 'containers', 'processes'],
  containers: ['containers', 'processes'],
  processes: ['processes'],
};
