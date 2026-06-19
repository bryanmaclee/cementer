// Ambient declarations for side-effect CSS imports (Vite bundles them). Lets
// `import "uplot/dist/uPlot.min.css"` typecheck under strict + noUncheckedSideEffectImports.
declare module "*.css";
