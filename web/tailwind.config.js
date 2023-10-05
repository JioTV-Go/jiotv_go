/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
}
