import styles from './App.module.css'
import { useCallback, useRef, useState } from "react"
import { UploadButton } from './components/UploadButton'
import { ImageUpload } from './components/ImageUpload'

function App() {
  const formRef = useRef<HTMLFormElement>(null)

  const [baseImageSrc, setBaseImageSrc] = useState("")
  const [otherImageSrc, setOtherImageSrc] = useState("")
  const [diffImageSrc, setDiffImageSrc] = useState("")

  const hasBothImageSrc = baseImageSrc !== "" && otherImageSrc !== ""

  const hasDiffImage = diffImageSrc !== ""
  const hasNoImage = !hasDiffImage

  const handleClear = useCallback(() => {
    setBaseImageSrc("")
    setOtherImageSrc("")
    setDiffImageSrc("")
  }, [])

  return (
    <form ref={formRef} className={styles.grid}>
      {hasNoImage && <ImageUpload $name="Base" imgSrc={baseImageSrc} setImageSrc={setBaseImageSrc} />}
      {hasNoImage && <ImageUpload $name="Other" imgSrc={otherImageSrc} setImageSrc={setOtherImageSrc} />}
      {hasBothImageSrc && hasNoImage && <UploadButton formRef={formRef} setDiffImageSrc={setDiffImageSrc} />}
      {diffImageSrc && <img src={diffImageSrc} alt="Diff"></img>}
      {hasDiffImage && <button type="button" onClick={handleClear}>Clear</button>}
    </form>
  )
}

export default App
