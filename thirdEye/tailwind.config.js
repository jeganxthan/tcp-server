/** @type {import('tailwindcss').Config} */
module.exports = {
  // NOTE: Update this to include the paths to all of your component files.
  content: ["./App.{js,jsx,ts,tsx}", "./src/**/*.{js,jsx,ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        primary: "#3b82f6",
        secondary: "#64748b",
        danger: "#ef4444",
        warning: "#f59e0b",
        success: "#10b981",
        background: "#0f172a",
        card: "#1e293b",
      },
    },
  },
  plugins: [],
};
