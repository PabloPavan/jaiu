const path = require("path");

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    path.join(__dirname, "..", "..", "internal", "view", "**", "*.templ"),
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
