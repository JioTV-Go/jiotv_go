/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./cmd/jiotv_go/views/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
}
