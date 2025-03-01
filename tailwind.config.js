/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.templ", 
    "./templates/**/*_templ.go",
    "./components/**/*.templ",
    "./views/**/*.templ"
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light", "dark"],
  }
}