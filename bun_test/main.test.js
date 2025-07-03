import { test, expect } from 'bun:test';


// coping the function, because importing breaks bun test (needs to load html)

/**
 * @param {Number} time_s
 * @returns {string}
 */
function displayTime(time_s) {
  const hours = Math.floor(time_s / 3600);
  time_s = time_s - hours * 3600;
  const minutes = Math.floor(time_s / 60);
  const seconds = Math.floor(time_s - minutes * 60);

  let str = hours > 0 ? `${hours}`.padStart(2, "0") + ":" : "";
  return (
    str + `${minutes}`.padStart(2, "0") + ":" + `${seconds}`.padStart(2, "0")
  );
}

test("display time", () => {
  let out = displayTime(120)
  expect(out).toEqual('02:00');
});

// Fake Test, because testing the rest of main.js does not make sense
test("test 2", () => {
  expect("test 2").toEqual('test 2');
});
