Rendering is split into many functions that do a piece of the work.
Often a task is done in parts that are merged together.

Many of the parts are memoised - the output is placed in a cache so if
it is needed again it is quickly available.

E.g. to render containers (ContainerRenderer), we do two sub-parts:
 - map from processes to containers
   - each output node is a container node, that has processes as child nodes.
 - map from endpoints to containers
   - the container for each endpoint is found via the container's IP address.
   - each output node is a container node, that has the adjacencies of the endpoints.
Then these two sets of container nodes are merged together.
 - Nodes with the same ID are mapped onto a single output node.
Then we drop deleted containers.

To render container by image name (ContainerImageRenderer),
 - first via ContainerWithImageNameRenderer,
   - joins container (from ContainerRenderer) and image topologies
     - copies across image name, tag, etc., and adds image node as a parent
 - then join these nodes with the relevant container-image node
Finally rename nodes to the image name with no version
and filter empty ones

To render container by DNS name (ContainerHostnameRenderer),
 - first via ContainerWithImageNameRenderer,
   - joins container (from ContainerRenderer) and image topologies
     - copies across image name, tag, etc., and adds image node as a parent
 - then map these nodes to a node with id of the DNS name ("hostname")
 - then take the container topology again (from ContainerRenderer)
   - map these nodes to a node with id of the DNS name ("hostname")
   - remove adjacency, children, etc
Merge these two sets together
and filter out ones with no container children
 
To render pods,
 - map from endpoints to pods
   - the container for each endpoint is found via the pod's IP address.
   - each output node is a container node, that has the adjacencies of the endpoints.
 - map from containers to pods
   - first via ContainerWithImageNameRenderer,
     - joins container and image topologies
     - filtered on running and non-pause containers
   - mapped to pod parent
   - then we find the host for that pod
   - then propagate any single metrics
Then these two sets of pod nodes are merged together.
 - Nodes with the same ID are mapped onto a single output node.

To render hosts,
 - take processes
   - map endpoints to processes
     - map those nodes to hosts via parent, with processes as children
 - take containers, rendered as above
   - map to hosts via parent, with containers as children
 - take containers by image, rendered as above
   - map to hosts via parent, with container images as children
 - take pods, rendered as above
   - map to hosts via parent, with pods as children
 - map from endpoints to hosts
   - the host for each endpoint is found via a 'latest' label
   - each output node is a host node, that has the adjacencies of the endpoints.


Idea: can we distinguish maps that change the node ID from maps that do not?
 - or just change them to Renders

Summaries display a bit about children
 - e.g. the number of containers in a pod

UI fetches the main view via websocket wss://cloud.weave.works/api/app/cold-sky-72/api/topology/hosts/ws?t=5s
Detailed view is polled https://cloud.weave.works/api/app/cold-sky-72/api/topology/hosts/ip-172-20-2-173%3B%3Chost%3E

Detailed view lists all children, grouped.
This is the main reason we add so many children to nodes.

New idea for detailed render:
 - for each topology have a detailed-renderer.

E.g. for host:
 - find all processes, containers, container-images, pods with parent host
 - add adjacencies from regular host rendering


