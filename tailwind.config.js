import daisyui from "daisyui";


/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,go,templ,js}"],
  theme: {
    extend: {},
  },
  plugins: [
    daisyui
  ],
  daisyui: {
    themes: ["light", "dark", "business"],
    darkTheme: "business"
  },
}

