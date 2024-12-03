import daisyui from "daisyui";
import { addIconSelectors } from "@iconify/tailwind";

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./include_dir/*.{html,js}", "./components/*.{go,templ}"],
  theme: {
    extend: {},
  },
  plugins: [daisyui, addIconSelectors(["tabler"])],
  daisyui: {
    themes: ["light", "dark", "business"],
    darkTheme: "business",
  },
};
