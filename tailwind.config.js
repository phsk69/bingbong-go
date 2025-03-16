/** @type {import('tailwindcss').Config} */
export const content = [
  "./templates/**/*.templ",
  "./templates/**/*_templ.go",
  "./components/**/*.templ",
  "./views/**/*.templ"
];
export const theme = {
  extend: {},
};
export const plugins = [require("daisyui")];
export const daisyui = {
  themes: ["light", "dark"],
};