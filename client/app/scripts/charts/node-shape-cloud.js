import React from 'react';
import { extent } from 'd3-array';

import { isContrastMode } from '../utils/contrast-utils';

const CLOUD_PATH = 'M 1920,384 Q 1920,225 1807.5,112.5 1695,0 1536,0 H 448 '
  + 'Q 263,0 131.5,131.5 0,263 0,448 0,580 71,689.5 142,799 258,853 '
  + 'q -2,28 -2,43 0,212 150,362 150,150 362,150 158,0 286.5,-88 128.5,-88 '
  + '187.5,-230 70,62 166,62 106,0 181,-75 75,-75 75,-181 0,-75 -41,-138 '
  + '129,-30 213,-134.5 84,-104.5 84,-239.5 z';

function toPoint(stringPair) {
  return stringPair.split(',').map(p => parseFloat(p, 10));
}

function getExtents(svgPath) {
  const points = svgPath.split(' ').filter(s => s.length > 1).map(toPoint);
  return [extent(points, p => p[0]), extent(points, p => p[1])];
}

export default function NodeShapeCloud({highlighted, size, color}) {
  const [[minx, maxx], [miny, maxy]] = getExtents(CLOUD_PATH);
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
      {highlighted &&
          <path className="highlighted" {...pathProps(0.7)} />}
      <path className="border" stroke={color} {...pathProps(0.5)} />
      <path className="shadow" {...pathProps(0.45)} />
      <circle className="node" r={Math.max(2, (size * 0.125))} />
    </g>
  );
}
