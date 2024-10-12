//import { useState } from 'react'
import './App.css'

function App() {
  //const [count, setCount] = useState(0)

  return (
    <>
      <div className='image-container'>
        <img className='image-item' src="https://place-hold.it/1280x800/0000ff"></img>
        <button>click me!</button>
        <img className='image-item' src="https://place-hold.it/1280x800/ff0000"></img>
      </div>
      <form action="http://localhost:4007/" method="post" encType="multipart/form-data">
        <label htmlFor="file">Choose file to upload:</label>
        <input type="file" id="file" name="file" />
        <button type="submit">Upload</button>
      </form>
    </>
  )
}

export default App
