"use client"

import { Moon, Sun } from "lucide-react"
import { useEffect, useState } from "react"

export function ThemeToggle() {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    const storedTheme = window.localStorage.getItem("novel-theme")
    const nextIsDark =
      storedTheme === "dark" ||
      (!storedTheme && window.matchMedia("(prefers-color-scheme: dark)").matches)

    document.documentElement.classList.toggle("dark", nextIsDark)
    setIsDark(nextIsDark)
  }, [])

  function toggleTheme() {
    const nextIsDark = !isDark
    document.documentElement.classList.toggle("dark", nextIsDark)
    window.localStorage.setItem("novel-theme", nextIsDark ? "dark" : "light")
    setIsDark(nextIsDark)
  }

  return (
    <button
      type="button"
      className="icon-button"
      onClick={toggleTheme}
      aria-label="切换主题"
      title="切换主题"
    >
      {isDark ? <Sun size={16} /> : <Moon size={16} />}
    </button>
  )
}
