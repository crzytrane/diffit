import appStyles from './App.module.css'
import { ImageContainer } from './components/ImageContainer/ImageContainer'

function App() {

  // return (
  //   <>
  //     <div className={appStyles.grid}>
  //       <ImageContainer />
  //       <ImageContainer />
  //       <ImageControls />
  //     </div>
  //   </>
  // )

  return (
    <>
      <form action="http://localhost:4007/api/files" method="post" encType="multipart/form-data">
        <UploadableImage $name="Base" imageURL='https://place-hold.it/1280x800/cc0000' />
        <div style={{ justifySelf: 'center' }}>Test</div>
        <UploadableImage $name="Other" imageURL='https://place-hold.it/1280x800/0000cc' />
        <button type="submit">Upload</button>
      </form >
    </>
  )
}

type UploadableImageProps = {
  imageURL: string
  $name: string
}

function UploadableImage({ imageURL, $name }: UploadableImageProps) {
  const name = $name.toLowerCase();
  return (
    <div>
      {/* <img src={imageURL} alt="Image to diff"></img> */}
      <label htmlFor="file">Choose {name} file to upload</label>
      <input type="file" name={`file-${name}`} aria-label={`${name} image upload`} />
    </div>
  )
}

export default App
