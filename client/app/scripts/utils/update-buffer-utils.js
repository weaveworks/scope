import debug from 'debug';
import Immutable from 'immutable';
import { union, size, map, find, reject, each } from 'lodash';

import { receiveNodesDelta } from '../actions/app-actions';

const log = debug('scope:update-buffer-utils');
const makeList = Immutable.List;
const feedInterval = 1000;
const bufferLength = 100;

let deltaBuffer = makeList();
let updateTimer = null;

function isPaused(getState) {
  return getState().get('updatePausedAt') !== null;
}

export function resetUpdateBuffer() {
  clearTimeout(updateTimer);
  deltaBuffer = deltaBuffer.clear();
}

function maybeUpdate(getState) {
  if (isPaused(getState)) {
    clearTimeout(updateTimer);
    resetUpdateBuffer();
  } else {
    if (deltaBuffer.size > 0) {
      const delta = deltaBuffer.first();
      deltaBuffer = deltaBuffer.shift();
      receiveNodesDelta(delta);
    }
    if (deltaBuffer.size > 0) {
      updateTimer = setTimeout(() => maybeUpdate(getState), feedInterval);
    }
  }
}

// consolidate first buffer entry with second
function consolidateBuffer() {
  const first = deltaBuffer.first();
  deltaBuffer = deltaBuffer.shift();
  const second = deltaBuffer.first();
  let toAdd = union(first.add, second.add);
  let toUpdate = union(first.update, second.update);
  let toRemove = union(first.remove, second.remove);
  log('Consolidating delta buffer', 'add', size(toAdd), 'update',
    size(toUpdate), 'remove', size(toRemove));

  // check if an added node in first was updated in second -> add second update
  toAdd = map(toAdd, node => {
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
  each(first.add, node => {
    const removedNode = find(second.remove, {id: node.id});
    if (removedNode) {
      toAdd = reject(toAdd, {id: node.id});
      toRemove = reject(toRemove, {id: node.id});
    }
  });

  // check if an updated node in first was removed in second -> remove
  each(first.update, node => {
    const removedNode = find(second.remove, {id: node.id});
    if (removedNode) {
      toUpdate = reject(toUpdate, {id: node.id});
    }
  });

  // check if an removed node in first was added in second ->  update
  // remove -> add is fine for the store

  // update buffer
  log('Consolidated delta buffer', 'add', size(toAdd), 'update',
    size(toUpdate), 'remove', size(toRemove));
  deltaBuffer.set(0, {
    add: toAdd.length > 0 ? toAdd : null,
    update: toUpdate.length > 0 ? toUpdate : null,
    remove: toRemove.length > 0 ? toRemove : null
  });
}

export function bufferDeltaUpdate(delta) {
  if (delta.add === null && delta.update === null && delta.remove === null) {
    log('Discarding empty nodes delta');
    return;
  }

  if (deltaBuffer.size >= bufferLength) {
    consolidateBuffer();
  }

  deltaBuffer = deltaBuffer.push(delta);
  log('Buffering node delta, new size', deltaBuffer.size);
}

export function getUpdateBufferSize() {
  return deltaBuffer.size;
}

export function resumeUpdate(getState) {
  maybeUpdate(getState);
}
