import React from 'react';

import { isContrastMode } from '../utils/contrast-utils';
import { getBorderProps } from './node';


const CLOUD_PATH = 'M32 6.4Q32 3.75 30.125 1.875 28.25 0 25.6 0H7.467Q4.383 0 2.192 2.192 0 4.383 0 7.467 0 9.667 1.183 11.492 2.367 13.317 4.3 14.217q-0.033 0.467-0.033 0.716 0 3.533 2.5 6.034 2.5 2.5 6.033 2.5 2.633 0 4.775-1.467 2.142-1.467 3.125-3.833 1.167 1.033 2.767 1.033 1.767 0 3.016-1.25 1.25-1.25 1.25-3.017 0-1.25-0.683-2.3 2.15-0.5 3.55-2.241 1.4-1.742 1.4-3.992z';

const PATH_EXTENTS = [[0, 32], [0, 23.466666666666665]];


export default function NodeShapeCloud({highlighted, size, color}) {
  const [[minx, maxx], [miny, maxy]] = PATH_EXTENTS;
  const width = (maxx - minx);
  const height = (maxy - miny);
  const cx = width / 2;
  const cy = height / 2;
  const pathSize = (width + height) / 2;
  const baseScale = (size * 2) / pathSize;
  const strokeWidth = isContrastMode() ? 6 / baseScale : 4 / baseScale;

  const pathProps = v => ({
    d: CLOUD_PATH,
    fill: 'none',
    transform: `scale(-${v * baseScale}) translate(-${cx},-${cy})`,
    strokeWidth
  });

  return (
    <g className="shape shape-cloud">
      {highlighted && <path className="highlighted" {...pathProps(0.7)} />}
      <path className="border" stroke={color} {...pathProps(0.5)} {...getBorderProps()} />
      <path className="shadow" {...pathProps(0.45)} />
      <circle className="node" r={Math.max(2, (size * 0.125))} />
    </g>
  );
}
