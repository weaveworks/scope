
// Cap the number of layers in the resource view to this constant. The reason why we have
// this constant is not just about the style, but also helps us build the selectors.
export const RESOURCE_VIEW_MAX_LAYERS = 3;

// TODO: Consider fetching these from the backend.
export const TOPOLOGIES_WITH_CAPACITY = ['hosts'];
export const RESOURCE_VIEW_LAYERS = {
  hosts: ['hosts', 'containers', 'processes'],
  containers: ['containers', 'processes'],
  processes: ['processes'],
};
