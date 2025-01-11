import styles from './App.module.css'
import { useCallback, useRef, useState } from "react"
import { UploadButton } from './components/UploadButton'
import { ImageUpload } from './components/ImageUpload'
import { DiffImageCompareViewer } from './components/DiffImageCompareViewer'

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
      {hasDiffImage && <DiffImageCompareViewer diffImageSrc={diffImageSrc} baseImageSrc={baseImageSrc} otherImageSrc={otherImageSrc} />}
      {hasDiffImage && <button type="button" onClick={handleClear} className={styles.clearButton}>Clear</button>}
    </form>
  )
}

export default App
