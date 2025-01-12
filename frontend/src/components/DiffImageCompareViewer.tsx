import { useState } from "react";
import styles from './DiffImageCompareViewer.module.css';

type DiffImageCompareViewerProps = {
    diffImageSrc: string;
    baseImageSrc: string;
    otherImageSrc: string;
}

type Showing = 'diff' | 'base' | 'other' | 'base-diff';

export function DiffImageCompareViewer({ diffImageSrc, baseImageSrc, otherImageSrc }: DiffImageCompareViewerProps) {
    const [showing, setShowing] = useState<Showing>("base-diff");
    return <>
        <button type="button" onClick={() => setShowing("base")} className={styles.baseButton}>Base</button>
        <button type="button" onClick={() => setShowing("base-diff")} className={styles.baseDiffButton}>Overlay</button>
        <button type="button" onClick={() => setShowing("diff")} className={styles.baseDiffButton}>Diff</button>
        <button type="button" onClick={() => setShowing("other")} className={styles.otherButton}>Other</button>
        <div className={styles.grid}>
            {showing.includes('diff') && <img className={styles.diffImage} src={diffImageSrc} alt="Diff" />}
            {showing.includes('base') && <img className={styles.baseImage} src={baseImageSrc} alt="Base image to compare" />}
            {showing.includes('other') && <img className={styles.otherImage} src={otherImageSrc} alt="Other image to compare" />}
        </div></>
}