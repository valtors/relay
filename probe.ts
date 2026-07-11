import { renderHook } from "@solidjs/testing-library";
import { createSignal } from "solid-js";

const testHook = () => {
  const [val, setVal] = createSignal(1);
  return { val, setVal };
};

const { result } = renderHook(testHook);
console.log("Type of result:", typeof result);
console.log("Result:", result);
if (typeof result === 'function') {
  console.log("Result() value:", result());
} else if (result && typeof result === 'object' && 'current' in result) {
  console.log("Result.current value:", result.current);
} else {
  console.log("Result is something else");
}
