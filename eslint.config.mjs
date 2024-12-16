import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import html from "eslint-plugin-html";

/** @type {import('eslint').Linter.Config[]} */
export default [
  {
    files: ["**/*.{js,mjs,cjs,ts,templ}"],
    plugins: { html },
  },
  { languageOptions: { globals: globals.browser } },
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
];

