/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../../internal/view/**/*.templ",
  ],
  theme: {
    extend: {
      fontFamily: {
        display: ["Manrope", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};
