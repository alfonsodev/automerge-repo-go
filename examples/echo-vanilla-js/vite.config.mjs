import { resolve } from "path";
import { defineConfig } from "vite";
import wasm from "vite-plugin-wasm";
import topLevelAwait from "vite-plugin-top-level-await";

const outDir = resolve(__dirname, "public/");

export default defineConfig({
  root: "src",
  base: "./",
  plugins: [wasm(), topLevelAwait()],
  build: {
    outDir,
    cssCodeSplit: false, // All CSS in one file
    emptyOutDir: true, // Clears app/static/js before each build
    minify: true,
    sourcemap: true, // Inline sourcemaps for dev, separate for prod

    rollupOptions: {
      input: {
        // Main HTML entry point
        index: resolve(__dirname, "src/index.html"),
        // JavaScript entry point
        main: resolve(__dirname, "src/js/app.js"),
        // If you have other independent JS files you want to bundle, add them here:
        // example: resolve(__dirname, 'app/js/example.js'),
      },
      output: {
        // Ensures the output filenames are predictable (e.g., main.js)
        entryFileNames: "[name].js",
        chunkFileNames: "[name].js", // For any code-split chunks
        assetFileNames: "[name].[ext]", // For any assets like CSS if imported via JS
      },
      onwarn(warning, warn) {
        if (warning.message.includes("htmx.org")) return;
        warn(warning);
      },
    },
  },
  server: {
    open: true, // open the browser automatically
    port: 5173, // default port; feel free to change
  },
});
