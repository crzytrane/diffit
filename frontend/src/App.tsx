//import { useState } from 'react'
import styles from './App.module.css'

function App() {
  //const [count, setCount] = useState(0)

  return (
    <>
      <div className={styles.diffContainer}>
        <UploadableImage imageURL='https://place-hold.it/1280x800/cc0000' />
        <div style={{ justifySelf: 'center' }}>Test</div>
        <UploadableImage imageURL='https://place-hold.it/1280x800/0000cc' />
      </div>
    </>
  )
}

type UploadableImageProps = {
  imageURL: string
}

function UploadableImage({ imageURL }: UploadableImageProps) {
  return (
    <div className={styles.diffItem}>
      <img src={imageURL} className={styles.diffItemImage} alt="Image to diff"></img>
      <form action="http://localhost:4007/" method="post" encType="multipart/form-data" className={styles.diffItemForm}>
        <label htmlFor="file">Choose file to upload:</label>
        <input type="file" name="file" aria-label='image upload' />
        <button type="submit">Upload</button>
      </form>
    </div>
  )
}

export default App
