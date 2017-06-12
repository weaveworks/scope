import debug from 'debug';
import { union, size, map, find, reject, each } from 'lodash';

const log = debug('scope:nodes-delta-utils');


// TODO: It would be nice to have a unit test for this function.
export function consolidateNodesDeltas(first, second) {
  let toAdd = union(first.add, second.add);
  let toUpdate = union(first.update, second.update);
  let toRemove = union(first.remove, second.remove);
  log('Consolidating delta buffer',
    'add', size(toAdd),
    'update', size(toUpdate),
    'remove', size(toRemove));

  // check if an added node in first was updated in second -> add second update
  toAdd = map(toAdd, (node) => {
    const updateNode = find(second.update, {id: node.id});
    if (updateNode) {
      toUpdate = reject(toUpdate, {id: node.id});
      return updateNode;
    }
    return node;
  });

  // check if an updated node in first was updated in second -> updated second update
  // no action needed, successive updates are fine

  // check if an added node in first was removed in second -> dont add, dont remove
  each(first.add, (node) => {
    const removedNode = find(second.remove, {id: node.id});
    if (removedNode) {
      toAdd = reject(toAdd, {id: node.id});
      toRemove = reject(toRemove, {id: node.id});
    }
  });

  // check if an updated node in first was removed in second -> remove
  each(first.update, (node) => {
    const removedNode = find(second.remove, {id: node.id});
    if (removedNode) {
      toUpdate = reject(toUpdate, {id: node.id});
    }
  });

  // check if an removed node in first was added in second -> update
  // remove -> add is fine for the store

  log('Consolidated delta buffer',
    'add', size(toAdd),
    'update', size(toUpdate),
    'remove', size(toRemove));

  return {
    add: toAdd.length > 0 ? toAdd : null,
    update: toUpdate.length > 0 ? toUpdate : null,
    remove: toRemove.length > 0 ? toRemove : null
  };
}
