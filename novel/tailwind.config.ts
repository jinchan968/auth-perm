import type { Config } from "tailwindcss"

const config = {
  darkMode: ["class"],
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        paper: "hsl(var(--paper))",
        ink: "hsl(var(--ink))",
        line: "hsl(var(--line))",
        dim: "hsl(var(--dim))",
        mark: "hsl(var(--mark))",
        wash: "hsl(var(--wash))",
        rust: "hsl(var(--rust))",
        terminal: "hsl(var(--terminal))",
      },
      fontFamily: {
        display: [
          "var(--font-display)",
          "Songti SC",
          "STSong",
          "Noto Serif CJK SC",
          "serif",
        ],
        body: [
          "var(--font-body)",
          "Iowan Old Style",
          "Songti SC",
          "STSong",
          "serif",
        ],
        mono: [
          "var(--font-mono)",
          "SFMono-Regular",
          "ui-monospace",
          "monospace",
        ],
      },
      boxShadow: {
        hairline: "0 1px 0 hsl(var(--line))",
        reader: "0 20px 80px hsl(var(--ink) / 0.08)",
      },
    },
  },
  plugins: [],
} satisfies Config

export default config
