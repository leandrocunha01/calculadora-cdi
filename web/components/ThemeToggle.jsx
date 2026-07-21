import { useState, useEffect } from 'react'

function ThemeToggle() {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    if (isDark) {
      document.body.classList.add('dark-mode')
    } else {
      document.body.classList.remove('dark-mode')
    }
  }, [isDark])

  return (
    <button
      className="theme-toggle"
      onClick={() => setIsDark(!isDark)}
    >
      {isDark ? '☀️' : '🌙'}
    </button>
  )
}

export default ThemeToggle
