import { spring } from 'react-motion';


export function weakSpring(value) {
  return spring(value, { stiffness: 100, damping: 18, precision: 1 });
}

export function strongSpring(value) {
  return spring(value, { stiffness: 800, damping: 50, precision: 1 });
}
