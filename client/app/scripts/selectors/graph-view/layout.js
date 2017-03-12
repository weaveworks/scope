import { includes, without, pick } from 'lodash';
import { createSelector } from 'reselect';
import { scaleThreshold } from 'd3-scale';
import { fromJS, Set as makeSet, List as makeList } from 'immutable';

import { NODE_BASE_SIZE } from '../../constants/styles';
import { graphNodesSelector, graphEdgesSelector } from './graph';
import { activeLayoutZoomSelector } from '../zooming';
import {
  canvasCircularExpanseSelector,
  canvasDetailsHorizontalCenterSelector,
  canvasDetailsVerticalCenterSelector,
} from '../canvas';


const circularOffsetAngle = Math.PI / 4;

// make sure circular layouts a bit denser with 3-6 nodes
const radiusDensity = scaleThreshold()
  .domain([3, 6])
  .range([2.5, 3, 2.5]);


const translationToViewportCenterSelector = createSelector(
  [
    canvasDetailsHorizontalCenterSelector,
    canvasDetailsVerticalCenterSelector,
    activeLayoutZoomSelector,
  ],
  (centerX, centerY, zoomState) => {
    const { scaleX, scaleY, translateX, translateY } = zoomState.toJS();
    return {
      x: (-translateX + centerX) / scaleX,
      y: (-translateY + centerY) / scaleY,
    };
  }
);

const selectedNodeIdSelector = createSelector(
  [
    graphNodesSelector,
    state => state.get('selectedNodeId'),
  ],
  (graphNodes, selectedNodeId) => (graphNodes.has(selectedNodeId) ? selectedNodeId : null)
);

// TODO: Combine this with the corresponding nodes decorator.
const focusedNodesIdsSelector = createSelector(
  [
    selectedNodeIdSelector,
    state => state.get('nodes'),
  ],
  (selectedNodeId, nodes) => {
    if (!selectedNodeId || nodes.isEmpty()) {
      return [];
    }

    // The selected node always goes in focus.
    let focusedNodes = makeSet([selectedNodeId]);

    // Add all the nodes the selected node is connected to...
    focusedNodes = focusedNodes.merge(nodes.getIn([selectedNodeId, 'adjacency']) || makeList());

    // ... and also all the nodes that connect to the selected one.
    nodes.forEach((node, nodeId) => {
      const adjacency = node.get('adjacency') || makeList();
      if (adjacency.includes(selectedNodeId)) {
        focusedNodes = focusedNodes.add(nodeId);
      }
    });

    return focusedNodes.toArray();
  }
);

const circularLayoutScalarsSelector = createSelector(
  [
    // TODO: Fix this.
    state => activeLayoutZoomSelector(state).get('scaleX'),
    state => focusedNodesIdsSelector(state).length - 1,
    canvasCircularExpanseSelector,
  ],
  (scale, circularNodesCount, viewportExpanse) => {
    // Here we calculate the zoom factor of the nodes that get selected into focus.
    // The factor is a somewhat arbitrary function (based on what looks good) of the
    // viewport dimensions and the number of nodes in the circular layout. The idea
    // is that the node should never be zoomed more than to cover 1/2 of the viewport
    // (`maxScale`) and then the factor gets decresed asymptotically to the inverse
    // square of the number of circular nodes, with a little constant push to make
    // the layout more stable for a small number of nodes. Finally, the zoom factor is
    // divided by the zoom factor applied to the whole topology layout to cancel it out.
    const maxScale = viewportExpanse / NODE_BASE_SIZE / 2;
    const shrinkFactor = Math.sqrt(circularNodesCount + 10);
    const selectedScale = maxScale / shrinkFactor / scale;

    // Following a similar logic as above, we set the radius of the circular
    // layout based on the viewport dimensions and the number of circular nodes.
    const circularRadius = viewportExpanse / radiusDensity(circularNodesCount) / scale;
    const circularInnerAngle = (2 * Math.PI) / circularNodesCount;

    return { selectedScale, circularRadius, circularInnerAngle };
  }
);

export const selectedScaleSelector = createSelector(
  [
    circularLayoutScalarsSelector,
  ],
  layout => layout.selectedScale
);

// Nodes after the selection circular layout has been applied to dagre engine output.
export const layoutNodesSelector = createSelector(
  [
    selectedNodeIdSelector,
    focusedNodesIdsSelector,
    graphNodesSelector,
    translationToViewportCenterSelector,
    circularLayoutScalarsSelector,
  ],
  (selectedNodeId, focusedNodesIds, graphNodes, translationToCenter, layoutScalars) => {
    const { circularRadius, circularInnerAngle } = layoutScalars;

    // Do nothing if the layout doesn't contain the selected node anymore.
    if (!selectedNodeId) {
      return graphNodes;
    }

    // Fix the selected node in the viewport center.
    let layoutNodes = graphNodes.mergeIn([selectedNodeId], translationToCenter);

    // Put the nodes that are adjacent to the selected one in a circular layout around it.
    const circularNodesIds = without(focusedNodesIds, selectedNodeId);
    layoutNodes = layoutNodes.map((node, nodeId) => {
      const index = circularNodesIds.indexOf(nodeId);
      if (index > -1) {
        const angle = circularOffsetAngle + (index * circularInnerAngle);
        return node.merge({
          x: translationToCenter.x + (circularRadius * Math.sin(angle)),
          y: translationToCenter.y + (circularRadius * Math.cos(angle))
        });
      }
      return node;
    });

    return layoutNodes;
  }
);

// Edges after the selection circular layout has been applied to dagre engine output.
export const layoutEdgesSelector = createSelector(
  [
    graphEdgesSelector,
    layoutNodesSelector,
    focusedNodesIdsSelector,
  ],
  (graphEdges, layoutNodes, focusedNodesIds) => (
    // Update the edges in the circular layout to link the nodes in a straight line.
    graphEdges.map((edge) => {
      const source = edge.get('source');
      const target = edge.get('target');
      if (includes(focusedNodesIds, source) || includes(focusedNodesIds, target)) {
        return edge.set('points', fromJS([
          pick(layoutNodes.get(source).toJS(), ['x', 'y']),
          pick(layoutNodes.get(target).toJS(), ['x', 'y']),
        ]));
      }
      return edge;
    })
  )
);
