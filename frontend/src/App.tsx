import styles from './App.module.css'

function App() {
  return (
    <>
      <form action="http://localhost:4007/" method="post" encType="multipart/form-data" className={styles.container}>
        <UploadableImage $name="Base" imageURL='https://place-hold.it/1280x800/cc0000' />
        <div style={{ justifySelf: 'center' }}>Test</div>
        <UploadableImage $name="Other" imageURL='https://place-hold.it/1280x800/0000cc' />
        <button type="submit" className={styles.containerButton}>Upload</button>
      </form >
    </>
  )
}

type UploadableImageProps = {
  imageURL: string
  $name: string
}

function UploadableImage({ imageURL, $name }: UploadableImageProps) {
  return (
    <div className={styles.image}>
      <img src={imageURL} alt="Image to diff"></img>
      <label htmlFor="file">Choose {$name.toLowerCase()} file to upload</label>
      <input type="file" name={`file-${$name}`} aria-label={`${$name} image upload`} />
    </div>
  )
}

export default App
