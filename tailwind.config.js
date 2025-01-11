/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
      "./templates/**/*.templ",  // Updated for templ files
      "./components/**/*.templ", // Optional: if you use a components directory
      "./views/**/*.templ"      // Optional: if you use a views directory
    ],
    theme: {
      extend: {},
    },
    plugins: [require("daisyui")],
    daisyui: {
      themes: ["light", "dark"],
    }
  }