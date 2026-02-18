/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "#0b1221",
        card: "#111b33",
        accent: "#25c2a0",
        muted: "#9fb2c7",
        border: "#1f2d4a",
        danger: "#ff6b6b",
        warning: "#f6c177",
        success: "#1fe0ae",
      },
      boxShadow: {
        card: "0 10px 30px rgba(0,0,0,0.25)",
      },
      keyframes: {
        flash: {
          "0%": { backgroundColor: "rgba(37,194,160,0.18)" },
          "100%": { backgroundColor: "transparent" },
        },
      },
      animation: {
        flash: "flash 1.6s ease-out",
      },
    },
  },
  plugins: [],
};
