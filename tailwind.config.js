import daisyui from "daisyui";

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./include_dir/*.{html,js}", "./components/*.{go,templ}"],
  theme: {
    extend: {},
  },
  plugins: [daisyui],
  daisyui: {
    themes: ["light", "dark", "business"],
    darkTheme: "business",
  },
};
