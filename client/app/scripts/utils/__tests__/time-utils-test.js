import { timer } from '../time-utils';


describe('timer', () => {
  it('records how long a function takes to execute', () => {
    const add100k = (number) => {
      for (let i = 0; i < 100000; i += 1) {
        number += 1;
      }
      return number;
    };

    const timedFn = timer(add100k);
    const result = timedFn(70);
    expect(result).toEqual(100070);
    expect(Number.isInteger(timedFn.time)).toBeTruthy();
  });
});
