/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./cmd/jiotv_go/templates/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
}
