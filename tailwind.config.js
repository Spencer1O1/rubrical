/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/web/**/*.templ",
    "./internal/web/**/*_templ.go",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "Roboto",
          "Helvetica Neue",
          "Arial",
          "Noto Sans",
          "sans-serif",
        ],
      },
    },
  },
  plugins: [require("@tailwindcss/forms")],
};
