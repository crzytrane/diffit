import { useCallback, useRef, useState } from 'react'
import { UploadButton } from '../components/UploadButton'
import { ImageUpload } from '../components/ImageUpload'
import { DiffImageCompareViewer } from '../components/DiffImageCompareViewer'
import styles from './ComparePage.module.css'

export function ComparePage() {
  const formRef = useRef<HTMLFormElement>(null)

  const [baseImageSrc, setBaseImageSrc] = useState('')
  const [otherImageSrc, setOtherImageSrc] = useState('')
  const [diffImageSrc, setDiffImageSrc] = useState('')

  const hasBothImageSrc = baseImageSrc !== '' && otherImageSrc !== ''

  const hasDiffImage = diffImageSrc !== ''
  const hasNoImage = !hasDiffImage

  const handleClear = useCallback(() => {
    setBaseImageSrc('')
    setOtherImageSrc('')
    setDiffImageSrc('')
  }, [])

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Compare Images</h1>
      <p className={styles.description}>
        Upload two images to compare them and see the visual differences.
      </p>

      <form ref={formRef} className={styles.form}>
        {hasNoImage && (
          <div className={styles.uploadGrid}>
            <ImageUpload $name="Base" imgSrc={baseImageSrc} setImageSrc={setBaseImageSrc} />
            <ImageUpload $name="Other" imgSrc={otherImageSrc} setImageSrc={setOtherImageSrc} />
          </div>
        )}
        {hasBothImageSrc && hasNoImage && (
          <div className={styles.compareAction}>
            <UploadButton formRef={formRef} setDiffImageSrc={setDiffImageSrc} />
          </div>
        )}
        {hasDiffImage && (
          <>
            <DiffImageCompareViewer
              diffImageSrc={diffImageSrc}
              baseImageSrc={baseImageSrc}
              otherImageSrc={otherImageSrc}
            />
            <div className={styles.clearAction}>
              <button type="button" onClick={handleClear} className={styles.clearButton}>
                Clear & Start Over
              </button>
            </div>
          </>
        )}
      </form>
    </div>
  )
}
