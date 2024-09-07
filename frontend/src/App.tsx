//import { useState } from 'react'
import './App.css'

function App() {
  //const [count, setCount] = useState(0)

  return (
    <>
      <h1>Vite + React</h1>
      <form action="http://localhost:4007/" method="post" encType="multipart/form-data">
        <label htmlFor="file">Choose file to upload:</label>
        <input type="file" id="file" name="file"/>
        <button type="submit">Upload</button>
      </form>
    </>
  )
}

export default App
