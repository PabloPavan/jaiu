/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../templates/**/*.tmpl",
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
