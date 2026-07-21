import { useState } from 'react'
import Calculator from './components/Calculator'
import ResultCard from './components/ResultCard'
import Comparison from './components/Comparison'
import ThemeToggle from './components/ThemeToggle'

function App() {
  const [savedResults, setSavedResults] = useState([])
  const [resultCounter, setResultCounter] = useState(0)

  const addResult = (data, payload) => {
    const newCounter = resultCounter + 1
    const resultId = `result-${newCounter}`

    setSavedResults(prev => {
      let updated = [...prev]

      // Limite de 2 simulações
      if (updated.length >= 2) {
        updated = updated.slice(1)
      }

      return [...updated, {
        id: resultId,
        data,
        payload,
        timestamp: new Date()
      }]
    })

    setResultCounter(newCounter)
  }

  const removeResult = (resultId) => {
    setSavedResults(prev => prev.filter(r => r.id !== resultId))
  }

  return (
    <>
      <ThemeToggle />
      <div className="container">
        <h1>Calculadora CDI / SELIC</h1>
        <Calculator onSubmit={addResult} />
        <Comparison results={savedResults} />
        <div id="results-container">
          {savedResults.slice().reverse().map(result => (
            <ResultCard
              key={result.id}
              result={result}
              onRemove={removeResult}
            />
          ))}
        </div>
      </div>
    </>
  )
}

export default App
